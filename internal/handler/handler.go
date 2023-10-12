package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chamrasilva89/reservationWeb/internal/config"
	"github.com/chamrasilva89/reservationWeb/internal/driver"
	"github.com/chamrasilva89/reservationWeb/internal/forms"
	"github.com/chamrasilva89/reservationWeb/internal/helpers"
	"github.com/chamrasilva89/reservationWeb/internal/models"
	"github.com/chamrasilva89/reservationWeb/internal/render"
	"github.com/chamrasilva89/reservationWeb/internal/repository"
	"github.com/chamrasilva89/reservationWeb/internal/repository/dbrepo"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	render.Templates(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Parse form data for start_date and end_date
    sd := r.Form.Get("start_date")
    ed := r.Form.Get("end_date")

    layout := "2006-01-02"
    startDate, err := time.Parse(layout, sd)
    if err != nil {
        helpers.ServerError(w, err)
        return
    }
    endDate, err := time.Parse(layout, ed)
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Parse room_id from the form and handle errors
    roomID, err := strconv.Atoi(r.Form.Get("room_id"))
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Create a reservation object with form data
    reservation := models.Reservation{
        FirstName: r.Form.Get("first_name"),
        LastName:  r.Form.Get("last_name"),
        Email:     r.Form.Get("email"),
        Phone:     r.Form.Get("phone"),
        StartDate: startDate,
        EndDate:   endDate,
        RoomID:    roomID,
    }

    // Create a form object for validation
    form := forms.New(r.PostForm)

    // Check required fields and add validation errors
    form.Required("first_name", "last_name", "email")
    form.MinLength("first_name", 3)
    form.IsEmail("email")

    // If the form is not valid, render the reservation page with validation errors
    if !form.Valid() {
        data := make(map[string]interface{})
        data["reservation"] = reservation

        render.Templates(w, r, "make-reservation.page.tmpl", &models.TemplateData{
            Form: form,
            Data: data,
        })
        return
    }

    // Insert the reservation into the database
    newReservationID, err := m.DB.InsertReservation(reservation)
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Create a room restriction for the reservation
    restriction := models.RoomRestriction{
        StartDate:     startDate,
        EndDate:       endDate,
        RoomID:        roomID,
        ReservationID: newReservationID,
        RestrictionID: 1,
    }

    // Insert the room restriction into the database
    err = m.DB.InsertRoomRestriction(restriction)
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Store the reservation in the session and redirect to the reservation summary page
    m.App.Session.Put(r.Context(), "reservation", reservation)
    http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}


func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		// no availability
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}
	w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"` // message to display on the client side if any error occurs while processing request

}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      false,
		Message: "Available",
	}
	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		panic(err)
	}
	log.Println(string(out))
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("cannor get items from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	m.App.Session.Remove(r.Context(), "reservation")
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Templates(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request){
	render.Templates(w,r,"login.page.tmpl", &models.TemplateData{
		Form:forms.New(nil),
	})
}

// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
    // Renew the session token to keep the user logged in
    _ = m.App.Session.RenewToken(r.Context())

    // Parse the form data from the HTTP request
    err := r.ParseForm()
    if err != nil {
        log.Println(err)
    }

    // Get the user's email and password from the form data
    email := r.Form.Get("email")
    password := r.Form.Get("password")

    // Create a form validation object and define required fields and email validation
    form := forms.New(r.PostForm)
    form.Required("email", "password")
    form.IsEmail("email")

    // If the form is not valid, render the login page again with the validation errors
    if !form.Valid() {
        render.Template(w, r, "login.page.tmpl", &models.TemplateData{
            Form: form,
        })
        return
    }

    // Try to authenticate the user with the provided email and password
    id, _, err := m.DB.Authenticate(email, password)

    // If authentication fails (an error occurs), show an error message and redirect to the login page
    if err != nil {
        m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
        http.Redirect(w, r, "/user/login", http.StatusSeeOther)
        return
    }

    // If authentication is successful, store the user's ID in the session and redirect to the home page
    m.App.Session.Put(r.Context(), "user_id", id)
    m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminAllReservations shows all reservations inu admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// AdminShowReservation shows the reservation in the admin tool
func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
    // Split the request URI to extract the reservation ID
    exploded := strings.Split(r.RequestURI, "/")
    id, err := strconv.Atoi(exploded[4])
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Get the source (src) from the URI and set it in a string map
    src := exploded[3]
    stringMap := make(map[string]string)
    stringMap["src"] = src

    // Get the year and month from the URL query parameters and add them to the string map
    year := r.URL.Query().Get("y")
    month := r.URL.Query().Get("m")
    stringMap["month"] = month
    stringMap["year"] = year

    // Get the reservation from the database using the reservation ID
    res, err := m.DB.GetReservationByID(id)
    if err != nil {
        helpers.ServerError(w, err)
        return
    }

    // Create data for rendering the reservation details
    data := make(map[string]interface{})
    data["reservation"] = res

    // Render the "admin-reservations-show" template with the string map, data, and an empty form
    render.Template(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
        StringMap: stringMap,
        Data:      data,
        Form:      forms.New(nil),
    })
}

