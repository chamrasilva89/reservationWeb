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

    // Serve static files from the "static" directory
    fileServer := http.FileServer(http.Dir("./static"))
    mux.Handle("/static/*", http.StripPrefix("/static", fileServer))


	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashboard)

		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Get("/process-reservation/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservation/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)

		mux.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)
	})

    return mux
}
