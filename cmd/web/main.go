package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/chamrasilva89/reservationWeb/internal/config"
	"github.com/chamrasilva89/reservationWeb/internal/driver"
	"github.com/chamrasilva89/reservationWeb/internal/handler"
	"github.com/chamrasilva89/reservationWeb/internal/helpers"
	"github.com/chamrasilva89/reservationWeb/internal/models"
	"github.com/chamrasilva89/reservationWeb/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
    // Print a message to indicate that the application has started on the specified port
    fmt.Println(fmt.Sprintf("Application started on : %s", portNumber))

    // Initialize the database and handle potential errors
    db, err := run()
    if err != nil {
        log.Fatal(err)
    }
    defer db.SQL.Close()

    // Create an HTTP server with the specified configuration and start listening for incoming requests
    srv := &http.Server{
        Addr:    portNumber, // Configure the bind address
        Handler: routes(&app),
    }
    err = srv.ListenAndServe()
    log.Fatal(err)
}

func run() (*driver.DB, error) {
    // Register types that will be stored in session data
	/*models.Reservation, models.User, models.Room, and models.Restriction are registered with the gob package. This is used for serializing and deserializing data when working with sessions.*/
    gob.Register(models.Reservation{})
    gob.Register(models.User{})
    gob.Register(models.Room{})
    gob.Register(models.Restriction{})
    app.InProduction = false

    // Initialize loggers for information and error logs
    infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
    app.InfoLog = infoLog

    errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
    app.ErrorLog = errorLog

    // Create a new session manager with specific settings
    session = scs.New()
    session.Lifetime = 24 * time.Hour
    session.Cookie.Persist = true
    session.Cookie.SameSite = http.SameSiteLaxMode
    session.Cookie.Secure = app.InProduction

    app.Session = session

    // Connect to the database
    log.Println("Connecting to the database")
    db, err := driver.ConnectSQL("host=localhost port=5432 dbname=reservation user=postgres password=ChamVish@123")
    if err != nil {
        log.Fatal("Cannot connect to the database")
        return nil, err
    }
    log.Println("Database Connection Established")

    // Create the template cache for rendering
    tc, err := render.CreateTemplateCache()
    if err != nil {
        log.Fatal("Cannot create template cache")
        return nil, err
    }
    app.TemplateCache = tc
    app.UseCache = false

    // Initialize repository and handlers with the application configuration and database connection
    repo := handler.NewRepo(&app, db)
    handler.NewHandlers(repo)

    // Initialize the renderer and helpers
    render.NewRenderer(&app)
    helpers.NewHelpers(&app)

    return db, nil
}
