package main

import (
	"github.com/go-chi/chi/v5"
	"golang-observer-project/internal/handlers"
	"net/http"
)

func routes() http.Handler {

	mux := chi.NewRouter()

	// default middleware
	mux.Use(CorsMiddleware())

	// login
	mux.Post("/login", handlers.Repo.Login)

	mux.Route("/pusher", func(mux chi.Router) {
		mux.Use(authMiddleware(handlers.Repo.TokenMaker))

		mux.Post("/auth", handlers.Repo.PusherAuth)
	})

	mux.Get("/user/logout", handlers.Repo.Logout)

	// admin routes
	mux.Route("/admin", func(mux chi.Router) {
		// all admin routes are protected
		mux.Use(authMiddleware(handlers.Repo.TokenMaker))

		// private message
		mux.Get("/private-message", handlers.Repo.SendPrivateMessage)

		// overview
		mux.Get("/overview", handlers.Repo.AdminDashboard)

		// events
		mux.Get("/events", handlers.Repo.Events)

		// settings
		mux.Post("/settings", handlers.Repo.PostSettings)

		// service status pages (all hosts)
		mux.Get("/all-healthy", handlers.Repo.AllHealthyServices)
		mux.Get("/all-warning", handlers.Repo.AllWarningServices)
		mux.Get("/all-problem", handlers.Repo.AllProblemServices)
		mux.Get("/all-pending", handlers.Repo.AllPendingServices)

		// users
		mux.Get("/users", handlers.Repo.AllUsers)
		mux.Get("/user/{id}", handlers.Repo.OneUser)
		mux.Post("/user/{id}", handlers.Repo.PostOneUser)
		mux.Delete("/user/delete/{id}", handlers.Repo.DeleteUser)

		// schedule
		mux.Get("/schedule", handlers.Repo.ListEntries)

		//preferences
		mux.Get("/preferences", handlers.Repo.Preferences)
		mux.Post("/preferences/set-system-pref", handlers.Repo.SetSystemPref)
		mux.Post("/preferences/toggle-monitoring", handlers.Repo.ToggleMonitoring)

		// hosts
		mux.Get("/host/all", handlers.Repo.AllHosts)
		mux.Get("/host/{id}", handlers.Repo.Host)
		mux.Post("/host/{id}", handlers.Repo.PostHost)
		mux.Post("/host/toggle-service", handlers.Repo.ToggleHostService)
		mux.Get("/perform-check/{id}/{oldStatus}", handlers.Repo.PerformCheck)

		// elastic
		mux.Get("/get-documents-in-last-x-minutes/{indexName}/{hostID}/{serviceID}/{minutes}", handlers.Repo.GetDocumentsInLastXMinutes)
	})

	return mux
}
