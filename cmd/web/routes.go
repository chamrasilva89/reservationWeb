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

	// Create a route group for routes starting with "/customer"
	mux.Route("/customer", func(customerMux chi.Router) {
		// Apply the Auth middleware to all routes in this group
		customerMux.Use(Auth)

		customerMux.Get("/add", handler.Repo.AddCustomer)
		customerMux.Post("/add", handler.Repo.PostCustomer)
		customerMux.Get("/all", handler.Repo.AllCustomers)
		customerMux.Get("/details/{id}", handler.Repo.ShowCustomerDetails)
		customerMux.Get("/trade-license/{id}", handler.Repo.ShowCustomerTradeLicense)
		customerMux.Post("/trade-license/{id}", handler.Repo.PostTradeLicense)
		customerMux.Get("/partners/{id}", handler.Repo.ShowCustomerPartners)
		customerMux.Get("/memorandum/{id}", handler.Repo.ShowCustomerMemorandum)
		customerMux.Get("/add-partner/{id}", handler.Repo.AddPartner)
		customerMux.Post("/add-partner/{id}", handler.Repo.PostPartner)
		customerMux.Get("/add-memorandum/{id}", handler.Repo.AddMemorandum)
		customerMux.Post("/add-memorandum/{id}", handler.Repo.PostRepresentative)
	})

	// Serve static files from the "static" directory
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(adminMux chi.Router) {
		adminMux.Use(Auth)
		adminMux.Get("/dashboard", handler.Repo.AdminDashboard)

		adminMux.Get("/reservations-new", handler.Repo.AdminNewReservations)
		adminMux.Get("/reservations-all", handler.Repo.AdminAllReservations)
		// Add more admin routes here

		// Example:
		// adminMux.Get("/process-reservation/{src}/{id}/do", handler.Repo.AdminProcessReservation)
		// adminMux.Get("/reservations/{src}/{id}/show", handler.Repo.AdminShowReservation)
		// adminMux.Post("/reservations/{src}/{id}", handler.Repo.AdminPostShowReservation)
	})

	return mux
}
