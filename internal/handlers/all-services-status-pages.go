package handlers

import (
	"golang-observer-project/internal/helpers"
	"net/http"
)

// AllHealthyServices lists all healthy services
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("healthy")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// return services object to the JSON response
	helpers.RenderJSON(w, services)
}

// AllWarningServices lists all warning services
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {

	services, err := repo.DB.GetServicesByStatus("warning")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// return services object to the JSON response
	helpers.RenderJSON(w, services)
}

// AllProblemServices lists all problem services
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("problem")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// return services object to the JSON response
	helpers.RenderJSON(w, services)
}

// AllPendingServices lists all pending services
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	// get all host services with status pending
	services, err := repo.DB.GetServicesByStatus("pending")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// return services object to the JSON response
	helpers.RenderJSON(w, services)
}
