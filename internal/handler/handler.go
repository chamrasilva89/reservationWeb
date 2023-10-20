package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chamrasilva89/reservationWeb/internal/config"
	"github.com/chamrasilva89/reservationWeb/internal/driver"
	"github.com/chamrasilva89/reservationWeb/internal/forms"
	"github.com/chamrasilva89/reservationWeb/internal/helpers"
	"github.com/chamrasilva89/reservationWeb/internal/models"
	"github.com/chamrasilva89/reservationWeb/internal/render"
	"github.com/chamrasilva89/reservationWeb/internal/repository"
	"github.com/chamrasilva89/reservationWeb/internal/repository/dbrepo"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
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

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Templates(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
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
		render.Templates(w, r, "login.page.tmpl", &models.TemplateData{
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
	render.Templates(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
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
	render.Templates(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{
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

	render.Templates(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{
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
	render.Templates(w, r, "admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

// AdminPostShowReservation posts a reservation
func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")

	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(), "flash", "Changes saved")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

// AdminProcessReservation marks a reservation as processed
func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	// Extract the 'id' and 'src' parameters from the URL
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	// Call the UpdateProcessedForReservation method to mark the reservation as processed
	err := m.DB.UpdateProcessedForReservation(id, 1)
	if err != nil {
		log.Println(err)
	}

	// Get the 'y' and 'm' query parameters from the URL
	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	// Set a flash message using the session to indicate that the reservation is marked as processed
	m.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")

	if year == "" {
		// If 'year' is empty, redirect to the reservations page for the specified 'src'
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		// If 'year' is not empty, redirect to the reservations calendar page with 'year' and 'month' query parameters
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AddCustomer(w http.ResponseWriter, r *http.Request) {
	var emptyCustomer models.Customer
	data := make(map[string]interface{})
	data["customer"] = emptyCustomer
	render.Templates(w, r, "add-customer.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostCustomer(w http.ResponseWriter, r *http.Request) {
	// Parse the form to handle form fields and file uploads
	err := r.ParseMultipartForm(10 << 20) // 10MB maximum file size
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	// Retrieve the uploaded files
	files := r.MultipartForm.File["photos"]
	fmt.Println("Customer intert started ", files)
	// Create a reservation object with form data
	customer := models.Customer{
		CustomerCode:     r.Form.Get("customerCode"),
		CustomerName:     r.Form.Get("customerName"),
		ContactNo:        r.Form.Get("contactNo"),
		ContactPerson:    r.Form.Get("contactPerson"),
		MobileNo:         r.Form.Get("mobileNo"),
		BusinessName:     r.Form.Get("businessName"),
		Email:            r.Form.Get("email"),
		LocationDetails:  r.Form.Get("locationDetails"),
		NatureOfBusiness: r.Form.Get("natureOfBusiness"),
		MarketedBy:       r.Form.Get("marketedBy"),
		MarketerName:     r.Form.Get("marketerName"),
		MarketerEmail:    r.Form.Get("marketerEmail"),
		Status:           r.Form.Get("status"),
	}
	fmt.Println("Customer Code:", r.Form.Get("customerCode"))
	fmt.Println("Customer Name:", r.Form.Get("customerName"))
	fmt.Println("Customer data", customer.CustomerCode, customer.CustomerName, customer.ContactNo)
	// Create a form object for validation
	form := forms.New(r.PostForm)

	// Check required fields and add validation errors
	form.Required("customerCode", "customerName", "contactPerson", "contactNo", "mobileNo", "email")
	form.IsEmail("email")

	// If the form is not valid, render the reservation page with validation errors
	if !form.Valid() {
		data := make(map[string]interface{})
		data["customer"] = customer

		render.Templates(w, r, "add-customer.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert the reservation into the database
	newCustomerID, err := m.DB.InsertCustomer(customer)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	// If files were uploaded, save them in a directory with the customer code
	if len(files) > 0 {
		customerCode := r.Form.Get("customerCode")
		//customerDir := fmt.Sprintf("/path/to/upload_directory/%s", customerCode)
		//customerDir := "E:\\GO\\FF\\" + customerCode
		customerDir := fmt.Sprintf("%s\\%s", config.FileUploadPath, customerCode)
		if _, err := os.Stat(customerDir); os.IsNotExist(err) {
			// The directory doesn't exist, so create it
			if err := os.MkdirAll(customerDir, os.ModePerm); err != nil {
				helpers.ServerError(w, err)
				return
			}
		}

		for _, file := range files {
			//uniqueFilename := generateUniqueFilename()

			// Extract the file extension from the original filename
			originalFilename := file.Filename
			//fileExtension := filepath.Ext(originalFilename)

			// Append the file extension to the unique filename
			uniqueFilenameWithExtension := originalFilename

			filePath := filepath.Join(customerDir, uniqueFilenameWithExtension)

			destinationFile, err := os.Create(filePath)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
			defer destinationFile.Close()

			sourceFile, err := file.Open()
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
			defer sourceFile.Close()

			_, err = io.Copy(destinationFile, sourceFile)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
			fmt.Println("File name with Extension", uniqueFilenameWithExtension)
			// Save the file path in the database
			_, err = m.DB.InsertFile(customerCode, filePath, newCustomerID, uniqueFilenameWithExtension)
			if err != nil {
				helpers.ServerError(w, err)
				return
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "New Customer Added Successfully..! ID :"+strconv.Itoa(newCustomerID))
	// Store the reservation in the session and redirect to the reservation summary page
	m.App.Session.Put(r.Context(), "customer", customer)
	http.Redirect(w, r, "/customer/all", http.StatusSeeOther)
}

func generateUniqueFilename() string {
	// Generate a UUID to ensure a unique filename
	id := uuid.New()
	return id.String()
}

func (m *Repository) AllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := m.DB.AllCustomers()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["customer"] = customers

	render.Templates(w, r, "all-customers.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) ShowCustomerDetails(w http.ResponseWriter, r *http.Request) {
	// Split the request URI to extract the customer ID
	fmt.Println("Inside customer show 1")
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("Inside customer show ", id)

	// Get the customer information from the database using the customer ID
	res, err := m.DB.GetCustomerByID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the "no rows in result set" error by rendering an empty page
			render.Templates(w, r, "empty-page.tmpl", &models.TemplateData{})
			return
		}
		helpers.ServerError(w, err)
		return
	}

	// Get the attachments for the customer from the database
	attachments, err := m.DB.GetAttachmentsByCustomerID(id)

	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the "no rows in result set" error by rendering an empty page
			render.Templates(w, r, "empty-page.tmpl", &models.TemplateData{})
			return
		}
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("attachments", attachments)

	// Create data for rendering the customer details
	data := make(map[string]interface{})
	data["customer"] = res

	// Create a slice to store valid attachment file paths
	validAttachments := []string{}

	// Check if each attachment file exists
	for _, attachment := range attachments.Attachments {
		if _, err := os.Stat(attachment.FilePath); err == nil {
			// The file exists, add it to the validAttachments slice
			validAttachments = append(validAttachments, attachment.FilePath)
		}
	}

	// Add valid attachments to the data
	data["attachments"] = validAttachments

	// Render the "customer-details.page.tmpl" template with the data
	render.Templates(w, r, "customer-details.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) ShowCustomerTradeLicense(w http.ResponseWriter, r *http.Request) {
	// Split the request URI to extract the customer ID
	fmt.Println("Inside Trade License")
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})

	// Get the trade information from the database using the customer ID
	res, err := m.DB.GetTradeLicenseInforByID(id)
	if err != nil {
		res = models.TradeLicense{} // Modify this to match your data structure
	}
	res.CustomerId = id
	// Create data for rendering the customer details
	data["tradelicense"] = res

	validAttachments := []string{}

	if _, err := os.Stat(res.FilePath); err == nil {
		// The file exists, add it to the validAttachments slice
		validAttachments = append(validAttachments, res.FilePath)
	}
	// Add valid attachments to the data
	data["tradeattachments"] = validAttachments

	// Render the "customer-trade-license.page.tmpl" template with the data
	render.Templates(w, r, "customer-trade-license.page.tmpl", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) ShowCustomerPartners(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside ShowCustomerPartners")

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("Customer ID:", id)

	data := make(map[string]interface{})

	// Get the trade information from the database using the customer ID
	fmt.Println("Getting trade information from the database...")
	res, err := m.DB.GetTradeShareInforByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("Trade information retrieved successfully.")

	fmt.Println("Printing trade information:")
	for _, shareholder := range res {
		fmt.Printf("Trade License ID: %d\n", shareholder.TradeLicenseID)
		fmt.Printf("Customer ID: %d\n", shareholder.CustomerId)
		fmt.Printf("Shareholder ID: %d\n", shareholder.ShareHolderID)
		fmt.Printf("Customer Code: %s\n", shareholder.ShEmirateID)
		fmt.Printf("Shareholder Name: %s\n", shareholder.ShareHolderName)
	}

	// Create data for rendering the customer details
	data["partners"] = res
	data["id"] = id
	validAttachments := []string{}

	// Iterate through partner records
	for _, partner := range res {
		if partner.ShIDFilepath != "" {
			if _, err := os.Stat(partner.ShIDFilepath); err == nil {
				// The file exists, add it to the validAttachments slice
				validAttachments = append(validAttachments, partner.ShIDFilepath)
			}
		}

		if partner.ShPassFilepath != "" {
			if _, err := os.Stat(partner.ShPassFilepath); err == nil {
				// The file exists, add it to the validAttachments slice
				validAttachments = append(validAttachments, partner.ShPassFilepath)
			}
		}
	}

	fmt.Println("Valid Attachments:")
	for _, attachment := range validAttachments {
		fmt.Println(attachment)
	}

	// Render the "customer-partner-details.tmpl" template with the data
	render.Templates(w, r, "customer-partner-details.page.tmpl", &models.TemplateData{
		Data:       data,
		Form:       forms.New(nil),
		CustomerID: id,
	})
}

func (m *Repository) ShowCustomerMemorandum(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inside ShowCustomerMemorandum")

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("Customer ID:", id)

	data := make(map[string]interface{})

	// Get the trade information from the database using the customer ID
	fmt.Println("Getting trade information from the database...")
	res, err := m.DB.GetMemorandumInforByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("Trade information retrieved successfully.")

	fmt.Println("Printing trade information:")
	for _, memorandum := range res {
		fmt.Printf("Trade License ID: %d\n", memorandum.TradeLicenseID)
		fmt.Printf("Customer ID: %d\n", memorandum.CustomerId)
		fmt.Printf("Representative: %s\n", memorandum.RepresentativeName)
		fmt.Printf("EMI ID: %s\n", memorandum.RepEmID)
		fmt.Printf("Passsport: %s\n", memorandum.RepPassport)
	}

	// Create data for rendering the customer details
	data["memorandum"] = res
	data["id"] = id

	validAttachments := []string{}

	// Iterate through partner records
	for _, partner := range res {
		if partner.RepIDFilepath != "" {
			if _, err := os.Stat(partner.RepIDFilepath); err == nil {
				// The file exists, add it to the validAttachments slice
				validAttachments = append(validAttachments, partner.RepIDFilepath)
			}
		}

		if partner.RepPassFilepath != "" {
			if _, err := os.Stat(partner.RepPassFilepath); err == nil {
				// The file exists, add it to the validAttachments slice
				validAttachments = append(validAttachments, partner.RepPassFilepath)
			}
		}
	}

	fmt.Println("Valid Attachments:")
	for _, attachment := range validAttachments {
		fmt.Println(attachment)
	}

	// Render the "customer-partner-details.tmpl" template with the data
	render.Templates(w, r, "customer-memorandum-details.page.tmpl", &models.TemplateData{
		Data:       data,
		Form:       forms.New(nil),
		CustomerID: id,
	})
}

func (m *Repository) PostTradeLicense(w http.ResponseWriter, r *http.Request) {
	// Parse the form to handle form fields and file uploads
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	fmt.Println("Trade Post Started")
	err = r.ParseMultipartForm(10 << 20) // 10MB maximum file size
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Retrieve the uploaded files
	files := r.MultipartForm.File["photos"]
	var filePath string
	var fileName string
	fmt.Println("Trade Post Started 2", files)
	// Create a customer ID as an integer
	customerIDStr := r.Form.Get("customerId")
	customerIDint, err := strconv.Atoi(customerIDStr)
	if err != nil {
		customerIDint = 0
	}
	fmt.Println("Trade Post Started 3", customerIDStr)
	// Parse date fields
	parseDate := func(fieldName string) (time.Time, error) {
		dateStr := r.Form.Get(fieldName)
		if dateStr == "" {
			return time.Time{}, nil
		}
		return time.Parse("2006-01-02", dateStr)
	}

	establishmentDate, _ := parseDate("establishmentDate")
	registrationDate, _ := parseDate("registrationDate")
	licenseExpiryDate, _ := parseDate("licenseExpiryDate")
	fmt.Println("Trade Post Started 4", establishmentDate, registrationDate, licenseExpiryDate)
	// If files were uploaded, save them in a directory with the customer code
	if len(files) == 1 {
		customerCode, err := m.DB.GetCustomerCodeByID(id)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		tradeLicenseDir := r.Form.Get("tradelicenseid")
		//customerDir := fmt.Sprintf("E:\\GO\\FF\\%s\\%s", customerCode, tradeLicenseDir)
		customerDir := fmt.Sprintf("%s%s\\%s", config.FileUploadPath, customerCode, tradeLicenseDir)
		if _, err := os.Stat(customerDir); os.IsNotExist(err) {
			if err := os.MkdirAll(customerDir, os.ModePerm); err != nil {
				helpers.ServerError(w, err)
				return
			}
		}

		file := files[0] // Assuming the file slice contains only one file

		// Use the original filename as it is
		filePath = filepath.Join(customerDir, file.Filename)
		fileName = file.Filename
		destinationFile, err := os.Create(filePath)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		defer destinationFile.Close()

		sourceFile, err := file.Open()
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		defer sourceFile.Close()

		_, err = io.Copy(destinationFile, sourceFile)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		fmt.Println("File name with Extension", file.Filename)
	}

	// Create a trade license object with form data
	tradeLicense := models.TradeLicense{
		CustomerId:       customerIDint,
		Emirate:          r.Form.Get("emirate"),
		TradeLicenseNo:   r.Form.Get("tradelicenseid"),
		MohreNo:          r.Form.Get("mohreno"),
		EstablishDate:    establishmentDate,
		RegistrationDate: registrationDate,
		LicenseExpiry:    licenseExpiryDate,
		TradeName:        r.Form.Get("tradeName"),
		LegalStatus:      r.Form.Get("legalState"),
		FilePath:         filePath,
		FileName:         fileName,
	}

	// Create a form object for validation
	form := forms.New(r.PostForm)

	// Check required fields and add validation errors
	form.Required("tradelicenseid", "licenseExpiryDate", "mohreno")

	// If the form is not valid, render the trade license page with validation errors
	if !form.Valid() {
		data := make(map[string]interface{})
		data["tradelicense"] = tradeLicense

		render.Templates(w, r, "customer-trade-license.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert the trade license into the database
	newTradeLicenseID, err := m.DB.InsertTradeLicense(tradeLicense)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Set flash message and redirect
	m.App.Session.Put(r.Context(), "flash", "New Trade License Added Successfully. ID: "+strconv.Itoa(newTradeLicenseID))
	http.Redirect(w, r, "/customer/all", http.StatusSeeOther)
}

func (m *Repository) AddPartner(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	var emptyCustomer models.TradeLicenseHolder
	data := make(map[string]interface{})
	data["partner"] = emptyCustomer
	render.Templates(w, r, "customer-add-partner.page.tmpl", &models.TemplateData{
		Form:       forms.New(nil),
		Data:       data,
		CustomerID: id,
	})
}

func (m *Repository) AddMemorandum(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	customerCode, err := m.DB.GetCustomerCodeByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var emptyCustomer models.Memorandum
	data := make(map[string]interface{})
	data["memorandum"] = emptyCustomer
	data["CustomerCode"] = customerCode
	render.Templates(w, r, "customer-add-memorandum.page.tmpl", &models.TemplateData{
		Form:       forms.New(nil),
		Data:       data,
		CustomerID: id,
	})
}

func (m *Repository) PostPartner(w http.ResponseWriter, r *http.Request) {
	// Extract the customer ID from the request URI
	exploded := strings.Split(r.RequestURI, "/")
	if len(exploded) < 3 {
		// Handle the case where there are not enough segments in the URI.
		http.Error(w, "Invalid request URI", http.StatusBadRequest)
		fmt.Println("Invalid request URI")
		return
	}

	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Println("Error converting ID:", err)
		return
	}
	fmt.Println("PostPartner", id)

	// Parse the form to handle form fields and file uploads
	err = r.ParseMultipartForm(10 << 20) // 10MB maximum file size
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Println("Error parsing form:", err)
		return
	}

	// Retrieve the uploaded files
	idfiles := r.MultipartForm.File["shIDFilepath"]
	passfiles := r.MultipartForm.File["ShPassFilepath"]

	var idfilePath string
	var passfilePath string
	var customerCode string
	// Parse date fields
	parseDate := func(fieldName string) (time.Time, error) {
		dateStr := r.Form.Get(fieldName)
		if dateStr == "" {
			return time.Time{}, nil
		}
		return time.Parse("2006-01-02", dateStr)
	}

	idExpireDate, _ := parseDate("shEmIDExp")
	passExpireDate, _ := parseDate("shPassportExp")

	// Function to handle file uploads
	handleFileUpload := func(files []*multipart.FileHeader, filePath *string) {
		if len(files) == 1 {
			customerCode, err = m.DB.GetCustomerCodeByID(id)
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error getting customer code:", err)
				return
			}

			partnerDir := "partners"
			//customerDir := fmt.Sprintf("E:\\GO\\FF\\%s\\%s", customerCode, partnerDir)
			customerDir := fmt.Sprintf("%s%s\\%s", config.FileUploadPath, customerCode, partnerDir)
			if _, err := os.Stat(customerDir); os.IsNotExist(err) {
				if err := os.MkdirAll(customerDir, os.ModePerm); err != nil {
					helpers.ServerError(w, err)
					fmt.Println("Error creating customer directory:", err)
					return
				}
			}

			file := files[0] // Assuming the file slice contains only one file

			// Use the original filename as it is
			*filePath = filepath.Join(customerDir, file.Filename)
			destinationFile, err := os.Create(*filePath)
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error creating destination file:", err)
				return
			}
			defer destinationFile.Close()

			sourceFile, err := file.Open()
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error opening source file:", err)
				return
			}
			defer sourceFile.Close()

			_, err = io.Copy(destinationFile, sourceFile)
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error copying files:", err)
				return
			}

			fmt.Printf("Added file: %s\n", file.Filename)
		}
	}

	// Handle file uploads for ID and Passport
	handleFileUpload(idfiles, &idfilePath)
	handleFileUpload(passfiles, &passfilePath)

	// Create a reservation object with form data
	partner := models.TradeLicenseHolder{
		CustomerId:      id,
		CustomerCode:    customerCode,
		CustomerName:    r.Form.Get("customerName"),
		ShareHolderRole: r.Form.Get("shareHolderRole"),
		ShNationality:   r.Form.Get("shareHolderNationality"),
		ShareHolderName: r.Form.Get("shareHolderName"),
		ShEmirateID:     r.Form.Get("shEmirateID"),
		ShEmIDExp:       idExpireDate,
		ShIDFilepath:    idfilePath,
		ShPassport:      r.Form.Get("shPassport"),
		ShPassportExp:   passExpireDate,
		ShPassFilepath:  passfilePath,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	fmt.Println("Partner emi File:", idfilePath)
	fmt.Println("partner pass file:", passfilePath)

	// Create a form object for validation
	form := forms.New(r.PostForm)

	// Check required fields and add validation errors
	form.Required("shareHolderName", "shEmirateID", "shPassport")

	// If the form is not valid, render the reservation page with validation errors
	if !form.Valid() {
		data := make(map[string]interface{})
		data["partners"] = partner

		render.Templates(w, r, "customer-add-partner.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert the reservation into the database
	newCustomerID, err := m.DB.InsertPartner(partner)
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Println("Error inserting partner:", err)
		return
	}
	url := fmt.Sprintf("/customer/partners/%d", id)
	m.App.Session.Put(r.Context(), "flash", "New Partner Added Successfully..! ID :"+strconv.Itoa(newCustomerID))

	// Store the reservation in the session and redirect to the reservation summary page
	m.App.Session.Put(r.Context(), "partner", partner)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (m *Repository) PostRepresentative(w http.ResponseWriter, r *http.Request) {
	// Extract the customer ID from the request URI
	exploded := strings.Split(r.RequestURI, "/")
	if len(exploded) < 3 {
		// Handle the case where there are not enough segments in the URI.
		http.Error(w, "Invalid request URI", http.StatusBadRequest)
		fmt.Println("Invalid request URI")
		return
	}

	id, err := strconv.Atoi(exploded[3])
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Println("Error converting ID:", err)
		return
	}
	fmt.Println("PostRepresentative", id)

	// Parse the form to handle form fields and file uploads
	err = r.ParseMultipartForm(10 << 20) // 10MB maximum file size
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Println("Error parsing form:", err)
		return
	}

	// Retrieve the uploaded files
	idfiles := r.MultipartForm.File["repIDFilepath"]
	passfiles := r.MultipartForm.File["repPassFilepath"]

	var idfilePath string
	var passfilePath string
	var customerCode string
	// Parse date fields
	parseDate := func(fieldName string) (time.Time, error) {
		dateStr := r.Form.Get(fieldName)
		if dateStr == "" {
			return time.Time{}, nil
		}
		return time.Parse("2006-01-02", dateStr)
	}

	idExpireDate, _ := parseDate("repEmIDExp")
	passExpireDate, _ := parseDate("repPassportExp")

	// Function to handle file uploads
	handleFileUpload := func(files []*multipart.FileHeader, filePath *string) {
		if len(files) == 1 {
			customerCode, err = m.DB.GetCustomerCodeByID(id)
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error getting customer code:", err)
				return
			}

			partnerDir := "memorandum"
			//customerDir := fmt.Sprintf("E:\\GO\\FF\\%s\\%s", customerCode, partnerDir)
			customerDir := fmt.Sprintf("%s%s\\%s", config.FileUploadPath, customerCode, partnerDir)
			if _, err := os.Stat(customerDir); os.IsNotExist(err) {
				if err := os.MkdirAll(customerDir, os.ModePerm); err != nil {
					helpers.ServerError(w, err)
					fmt.Println("Error creating customer directory:", err)
					return
				}
			}

			file := files[0] // Assuming the file slice contains only one file

			// Use the original filename as it is
			*filePath = filepath.Join(customerDir, file.Filename)
			destinationFile, err := os.Create(*filePath)
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error creating destination file:", err)
				return
			}
			defer destinationFile.Close()

			sourceFile, err := file.Open()
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error opening source file:", err)
				return
			}
			defer sourceFile.Close()

			_, err = io.Copy(destinationFile, sourceFile)
			if err != nil {
				helpers.ServerError(w, err)
				fmt.Println("Error copying files:", err)
				return
			}

			fmt.Printf("Added file: %s\n", file.Filename)
		}
	}

	// Handle file uploads for ID and Passport
	handleFileUpload(idfiles, &idfilePath)
	handleFileUpload(passfiles, &passfilePath)

	// Create a reservation object with form data
	partner := models.Memorandum{
		CustomerId:         id,
		CustomerCode:       customerCode,
		RepresentativeName: r.Form.Get("representativeName"),
		RepNoOfShares:      r.Form.Get("repNoOfShares"),
		RepEmID:            r.Form.Get("repEmID"),
		RepEmIDExp:         idExpireDate,
		RepIDFilepath:      idfilePath,
		RepPassport:        r.Form.Get("repPassport"),
		RepPassportExp:     passExpireDate,
		RepPassFilepath:    passfilePath,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	fmt.Println("Rep emi File:", idfilePath)
	fmt.Println("Rep pass file:", passfilePath)

	// Create a form object for validation
	form := forms.New(r.PostForm)

	// Check required fields and add validation errors
	form.Required("representativeName", "repNoOfShares")

	// If the form is not valid, render the reservation page with validation errors
	if !form.Valid() {
		data := make(map[string]interface{})
		data["memorandum"] = partner

		render.Templates(w, r, "customer-add-memorandum.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Insert the reservation into the database
	newCustomerID, err := m.DB.InsertMemorandum(partner)
	if err != nil {
		helpers.ServerError(w, err)
		fmt.Println("Error inserting partner:", err)
		return
	}
	url := fmt.Sprintf("/customer/memorandum/%d", id)
	m.App.Session.Put(r.Context(), "flash", "New Representative Added Successfully..! ID :"+strconv.Itoa(newCustomerID))

	// Store the reservation in the session and redirect to the reservation summary page
	m.App.Session.Put(r.Context(), "memorandum", partner)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
