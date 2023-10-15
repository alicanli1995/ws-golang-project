package handlers

import (
	"github.com/CloudyKit/jet/v6"
	"golang-vigilate-project/internal/helpers"
	"net/http"
)

// AllHealthyServices lists all healthy services
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("healthy")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// create a map of host id and host name
	hostMap := make(jet.VarMap)
	hostMap.Set("services", services)
	err = helpers.RenderPage(w, r, "healthy", hostMap, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllWarningServices lists all warning services
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {

	services, err := repo.DB.GetServicesByStatus("warning")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// create a map of host id and host name
	hostMap := make(jet.VarMap)
	hostMap.Set("services", services)

	err = helpers.RenderPage(w, r, "warning", hostMap, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllProblemServices lists all problem services
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("problem")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// create a map of host id and host name
	hostMap := make(jet.VarMap)
	hostMap.Set("services", services)
	err = helpers.RenderPage(w, r, "problems", hostMap, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllPendingServices lists all pending services
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	// get all host services with status pending
	services, err := repo.DB.GetServicesByStatus("pending")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	// create a map of host id and host name
	hostMap := make(jet.VarMap)
	hostMap.Set("services", services)

	err = helpers.RenderPage(w, r, "pending", hostMap, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
