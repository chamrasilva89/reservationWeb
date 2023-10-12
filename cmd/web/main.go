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

	fmt.Println(fmt.Sprintf("Application started on : %s", portNumber))
	//_ = http.ListenAndServe(portNumber, nil)
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	//
	srv := &http.Server{
		Addr:    portNumber, // configure the bind address
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	//what i am going to put in session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	//connect to database
	log.Println("connecting to db")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=reservation user=postgres password=ChamVish@123")
	if err != nil {
		log.Fatal("Cannot connect to database")
		return nil, err
	}
	log.Println("Database Connection Established")
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("Cannot create template cache")
		return nil, err
	}
	app.TemplateCache = tc
	app.UseCache = false
	repo := handler.NewRepo(&app, db)
	handler.NewHandlers(repo)

	render.NewRenderer(&app)
	helpers.NewHelpers(&app)
	//http.HandleFunc("/", handler.Repo.Home)
	//http.HandleFunc("/about", handler.Repo.About)
	return db, nil
}
