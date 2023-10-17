package main

import (
	"net/http"

	"github.com/chamrasilva89/reservationWeb/internal/config"
	"github.com/chamrasilva89/reservationWeb/internal/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	// Create a new Chi router instance
	mux := chi.NewRouter()

	// Use Chi middleware for recovering from panics
	mux.Use(middleware.Recoverer)

	// Use the NoSurf middleware (not shown in this code snippet, but it's assumed to be part of your application)
	mux.Use(NoSurf)

	// Use the SessionLoad middleware (not shown in this code snippet, but it's assumed to be part of your application)
	mux.Use(SessionLoad)

	// Define routes and associate them with their corresponding handlers
	mux.Get("/", handler.Repo.Home)
	mux.Get("/about", handler.Repo.About)
	mux.Get("/generals", handler.Repo.Generals)
	mux.Get("/majors", handler.Repo.Majors)

	mux.Get("/make-reservation", handler.Repo.Reservation)
	mux.Post("/make-reservation", handler.Repo.PostReservation)
	mux.Get("/search-availability", handler.Repo.Availability)
	mux.Get("/reservation-summary", handler.Repo.ReservationSummary)

	mux.Get("/contact", handler.Repo.Contact)

	mux.Post("/search-availability", handler.Repo.PostAvailability)
	mux.Post("/search-availability-g", handler.Repo.AvailabilityJSON)

	mux.Get("/user/login", handler.Repo.ShowLogin)
	mux.Post("/user/login", handler.Repo.PostShowLogin)
	mux.Get("/user/logout", handler.Repo.Logout)

	mux.Get("/add-customer", handler.Repo.AddCustomer)
	mux.Post("/add-customer", handler.Repo.PostCustomer)
	mux.Get("/customer-all", handler.Repo.AllCustomers)
	mux.Get("/customer-details/{id}", handler.Repo.ShowCustomerDetails)
	mux.Get("/customer-trade-license/{id}", handler.Repo.ShowCustomerTradeLicense)
	mux.Post("/customer-trade-license/{id}", handler.Repo.PostTradeLicense)
	mux.Get("/customer-partners/{id}", handler.Repo.ShowCustomerPartners)
	mux.Get("/customer-memorandum/{id}", handler.Repo.ShowCustomerMemorandum)
	mux.Get("/customer-add-partner/{id}", handler.Repo.AddPartner)
	mux.Get("/customer-add-memorandum/{id}", handler.Repo.AddPartner)
	mux.Post("/add-partner", handler.Repo.PostPartner)
	// Serve static files from the "static" directory
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Get("/dashboard", handler.Repo.AdminDashboard)

		mux.Get("/reservations-new", handler.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handler.Repo.AdminAllReservations)
		/*mux.Get("/reservations-calendar", handler.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handler.Repo.AdminPostReservationsCalendar)
		mux.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)*/
		mux.Get("/process-reservation/{src}/{id}/do", handler.Repo.AdminProcessReservation)
		mux.Get("/reservations/{src}/{id}/show", handler.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handler.Repo.AdminPostShowReservation)
	})

	return mux
}
