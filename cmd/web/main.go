package main

import (
	"database/sql"
	"encoding/gob"
	"github.com/pusher/pusher-http-go"
	"golang-observer-project/internal/config"
	"golang-observer-project/internal/handlers"
	"golang-observer-project/internal/models"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

var app config.AppConfig
var repo *handlers.DBRepo
var preferenceMap map[string]string
var wsClient pusher.Client

const observerVersion = "1.0.0"
const maxWorkerPoolSize = 5
const maxJobMaxWorkers = 5

func init() {
	gob.Register(models.User{})
	_ = os.Setenv("TZ", "UTC")
}

// the main is the application entry point
func main() {
	// set up application
	insecurePort, err := setupApp()
	if err != nil {
		log.Fatal(err)
	}

	// close channels & db when the application ends
	defer close(app.MailQueue)
	defer func(SQL *sql.DB) {
		_ = SQL.Close()
	}(app.DB.SQL)

	// print info
	log.Printf("******************************************")
	log.Printf("** %sObserver%s v%s built in %s", "\033[31m", "\033[0m", observerVersion, runtime.Version())
	log.Printf("**----------------------------------------")
	log.Printf("** Running with %d Processors", runtime.NumCPU())
	log.Printf("** Running on %s", runtime.GOOS)
	log.Printf("******************************************")

	// create http server
	srv := &http.Server{
		Addr:              *insecurePort,
		Handler:           routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	log.Printf("Starting HTTP server on port %s....", *insecurePort)

	// start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
