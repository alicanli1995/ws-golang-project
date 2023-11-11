package main

import (
	"flag"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/pusher/pusher-http-go"
	"github.com/robfig/cron/v3"
	"golang-observer-project/internal/channeldata"
	"golang-observer-project/internal/config"
	"golang-observer-project/internal/driver"
	"golang-observer-project/internal/elastic/elastic"
	"golang-observer-project/internal/handlers"
	"golang-observer-project/internal/helpers"
	"golang-observer-project/internal/token"
	"log"
	"os"
	"time"
)

func setupApp() (*string, error) {
	// read flags
	insecurePort := flag.String("port", ":4000", "port to listen on")
	identifier := flag.String("identifier", "observer", "unique identifier")
	domain := flag.String("domain", "localhost", "domain name (e.g. example.com)")
	inProduction := flag.Bool("production", false, "application is in production")
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbPort := flag.String("dbport", "5432", "database port")
	dbUser := flag.String("dbuser", "postgres", "database user")
	dbPass := flag.String("dbpass", "postgres", "database password")
	databaseName := flag.String("db", "postgres", "database name")
	dbSsl := flag.String("dbssl", "disable", "database ssl setting")
	pusherHost := flag.String("pusherHost", "localhost", "pusher host")
	pusherPort := flag.String("pusherPort", "4001", "pusher port")
	pusherApp := flag.String("pusherApp", "1", "pusher app id")
	pusherKey := flag.String("pusherKey", "abc123", "pusher key")
	pusherSecret := flag.String("pusherSecret", "123abc", "pusher secret")
	pusherSecure := flag.Bool("pusherSecure", false, "pusher server uses SSL (true or false)")
	jwtSecret := flag.String("jwtSecret", "jwtSecretManagerTry1234512345123", "secret key for signing JWTs")

	flag.Parse()

	if *dbUser == "" || *dbHost == "" || *dbPort == "" || *databaseName == "" || *identifier == "" {
		fmt.Println("Missing required flags.")
		os.Exit(1)
	}

	log.Println("Connecting to database....")
	dsnString := ""

	// when developing locally, we often don't have a db password
	if *dbPass == "" {
		dsnString = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			*dbHost,
			*dbPort,
			*dbUser,
			*databaseName,
			*dbSsl)
	} else {
		dsnString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			*dbHost,
			*dbPort,
			*dbUser,
			*dbPass,
			*databaseName,
			*dbSsl)
	}

	db, err := driver.ConnectPostgres(dsnString)
	if err != nil {
		log.Fatal("Cannot connect to database!", err)
	}

	// start mail channel
	log.Println("Initializing mail channel and worker pool....")
	mailQueue := make(chan channeldata.MailJob, maxWorkerPoolSize)

	// Start the email dispatcher
	log.Println("Starting email dispatcher....")
	dispatcher := NewDispatcher(mailQueue, maxJobMaxWorkers)
	dispatcher.run()

	// define application configuration
	a := config.AppConfig{
		DB:           db,
		InProduction: *inProduction,
		Domain:       *domain,
		PusherSecret: *pusherSecret,
		MailQueue:    mailQueue,
		Version:      observerVersion,
		Identifier:   *identifier,
	}

	app = a

	tokenMaker, err := token.NewJWTMaker(*jwtSecret)
	if err != nil {
		log.Fatal("cannot create token maker")
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}

	client, err := elasticsearch.NewClient(cfg)

	elasticClient := elastic.NewElasticRepo(client, &app)

	repo = handlers.NewPostgresqlHandlers(db, &app, tokenMaker, elasticClient)
	handlers.NewHandlers(repo, &app, tokenMaker, elasticClient)

	log.Println("Getting preferences...")
	preferenceMap = make(map[string]string)
	preferences, err := repo.DB.AllPreferences()
	if err != nil {
		log.Fatal("Cannot read preferences:", err)
	}

	for _, pref := range preferences {
		preferenceMap[pref.Name] = pref.Preference
	}

	preferenceMap["pusher-host"] = *pusherHost
	preferenceMap["pusher-port"] = *pusherPort
	preferenceMap["pusher-key"] = *pusherKey
	preferenceMap["identifier"] = *identifier
	preferenceMap["version"] = observerVersion

	app.PreferenceMap = preferenceMap

	// create a pusher client
	wsClient = pusher.Client{
		AppID:  *pusherApp,
		Secret: *pusherSecret,
		Key:    *pusherKey,
		Secure: *pusherSecure,
		Host:   fmt.Sprintf("%s:%s", *pusherHost, *pusherPort),
	}

	log.Println("Host", fmt.Sprintf("%s:%s", *pusherHost, *pusherPort))
	log.Println("Secure", *pusherSecure)

	app.WsClient = wsClient

	monitorMap := make(map[int]cron.EntryID)

	app.MonitorMap = monitorMap

	// set up a time zone with Istanbul
	localZone, _ := time.LoadLocation("Europe/Istanbul")
	scheduler := cron.New(cron.WithLocation(localZone), cron.WithChain(
		cron.DelayIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))

	app.Scheduler = scheduler

	go handlers.Repo.StartMonitoring()

	if app.PreferenceMap["monitoring_live"] == "1" {
		app.Scheduler.Start()
	}

	helpers.NewHelpers(&app)

	return insecurePort, err
}

// createDirIfNotExist creates a directory if it does not exist
func createDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
