package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang-observer-project/internal/config"
	"golang-observer-project/internal/driver"
	"golang-observer-project/internal/elastic"
	"golang-observer-project/internal/helpers"
	"golang-observer-project/internal/models"
	"golang-observer-project/internal/repository"
	"golang-observer-project/internal/repository/dbrepo"
	"golang-observer-project/token"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

// Repo is the repository
var Repo *DBRepo
var app *config.AppConfig

// DBRepo is the db repo
type DBRepo struct {
	App           *config.AppConfig
	DB            repository.DatabaseRepo
	TokenMaker    token.Maker
	ElasticClient elastic.Operations
}

// NewHandlers creates the handlers
func NewHandlers(repo *DBRepo, a *config.AppConfig, tokenMaker token.Maker, elasticClient elastic.Operations) {
	Repo = repo
	app = a
	Repo.TokenMaker = tokenMaker
	Repo.ElasticClient = elasticClient
}

// NewPostgresqlHandlers creates db repo for postgres
func NewPostgresqlHandlers(db *driver.DB, a *config.AppConfig, tokenMaker token.Maker, elasticClient elastic.Operations) *DBRepo {
	return &DBRepo{
		App:           a,
		DB:            dbrepo.NewPostgresRepo(db.SQL, a),
		TokenMaker:    tokenMaker,
		ElasticClient: elasticClient,
	}
}

// AdminDashboard displays the dashboard
func (repo *DBRepo) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	pending, healthy, warning, problem, err := repo.DB.GetAllServicesStatusCounts()
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	var response models.DashResponse

	response.OK = true
	response.Message = "Dashboard data retrieved"
	response.Healthy = healthy
	response.Warning = warning
	response.Problem = problem
	response.Pending = pending

	// get all hosts
	hosts, err := repo.DB.AllHosts()
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	response.Hosts = hosts
	// return services object to the JSON response
	helpers.RenderJSON(w, r, response)
}

// Events display the events page
func (repo *DBRepo) Events(w http.ResponseWriter, r *http.Request) {
	events, err := repo.DB.AllEvents()
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	// return services object to the JSON response
	helpers.RenderJSON(w, r, events)
}

type settingsUpdateRequest struct {
	UpdatePreferences []settingUpdateObj `json:"UpdatePreferences"`
}

type settingUpdateObj struct {
	Name       string `json:"Name"`
	Preference string `json:"Preference"`
}

// PostSettings saves site settings
func (repo *DBRepo) PostSettings(w http.ResponseWriter, r *http.Request) {
	prefMap := make(map[string]string)
	var req settingsUpdateRequest
	err := helpers.ReadJSONBody(r, &req)
	if err != nil {
		log.Println(err)
	}

	for _, v := range req.UpdatePreferences {
		prefMap[v.Name] = v.Preference
	}

	//if r.Form.Get("sms_enabled") == "0" {
	//	prefMap["notify_via_sms"] = "0"
	//}

	err = repo.DB.InsertOrUpdateSitePreferences(prefMap)
	if err != nil {
		log.Println(err)
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	// update app config
	for k, v := range prefMap {
		app.PreferenceMap[k] = v
	}

	var jsonResp jsonResp
	jsonResp.OK = true
	jsonResp.Message = "Settings updated"

	helpers.RenderJSON(w, r, jsonResp)

}

// AllHosts displays list of all hosts
func (repo *DBRepo) AllHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := repo.DB.AllHosts()
	if err != nil {
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	var response models.HostsJsonResponse
	response.OK = true
	response.Message = "Hosts retrieved"
	response.Hosts = hosts

	helpers.RenderJSON(w, r, response)
}

// Host shows the host add/edit form
func (repo *DBRepo) Host(w http.ResponseWriter, r *http.Request) {
	// get the host id from the url
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var response models.HostJsonResponse
	response.OK = true
	response.Message = "Host retrieved"

	var h models.Host
	if id > 0 {
		host, err := repo.DB.FindHostByID(id)
		if err != nil {
			log.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
		h = host
	}

	response.Host = h

	// return services object to the JSON response
	helpers.RenderJSON(w, r, response)

}

// PostHost adds a host
func (repo *DBRepo) PostHost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var host models.Host
	var hostID int

	var req models.HostPostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
	}

	host.HostName = req.HostName
	host.CanonicalName = req.CanonicalName
	host.URL = req.URL
	host.IP = req.IP
	host.IPV6 = req.IPV6
	host.Location = req.Location
	host.OS = req.OS
	host.Active = req.Active

	if id > 0 {
		log.Println("updating host")
		host.ID = id
		err := repo.DB.UpdateHost(host)
		if err != nil {
			log.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
		hostID = id
	} else {
		log.Println("inserting host")
		ID, err := repo.DB.InsertHost(host)
		if err != nil {
			log.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
		hostID = ID
	}

	repo.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, "/admin/host/"+strconv.Itoa(hostID), http.StatusSeeOther)

}

// ToggleHostService toggles host service on/off
func (repo *DBRepo) ToggleHostService(w http.ResponseWriter, r *http.Request) {
	var req models.ToggleServiceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
	}

	var response models.ServiceJSON
	response.OK = true
	err = repo.DB.UpdateHostServiceStatus(req.HostID, req.ServiceID, req.Active)
	if err != nil {
		log.Println(err)
		response.OK = false
	}

	hs, _ := repo.DB.GetHostServiceByHostIDServiceID(req.HostID, req.ServiceID)
	h, _ := repo.DB.FindHostByID(req.HostID)

	if req.Active == 1 {
		repo.pushScheduleChangeEvent(hs, "pending")
		repo.pushStatusChangeEvent(h, hs, "pending")
		repo.addToMonitorMap(hs)
	} else {
		repo.removeFromMonitorMap(hs)
	}

	// return services object to the JSON response
	helpers.RenderJSON(w, r, response)
}

// AllUsers lists all admin users
func (repo *DBRepo) AllUsers(w http.ResponseWriter, r *http.Request) {
	u, err := repo.DB.AllUsers()
	if err != nil {
		ClientError(w, r, http.StatusBadRequest)
		return
	}

	helpers.RenderJSON(w, r, u)
}

// OneUser displays the add/edit user page
func (repo *DBRepo) OneUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
	}

	var user models.User

	if id > 0 {

		u, err := repo.DB.GetUserById(id)
		if err != nil {
			ClientError(w, r, http.StatusBadRequest)
			return
		}
		user = u

	} else {
		var u models.User
		user = u
	}

	helpers.RenderJSON(w, r, user)
}

// PostOneUser adds/edits a user
func (repo *DBRepo) PostOneUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Println(err)
	}

	var req models.UserRequest
	var u models.User

	// read JSON from request body
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
	}

	if id > 0 {
		u, _ = repo.DB.GetUserById(id)
		u.FirstName = req.FirstName
		u.LastName = req.LastName
		u.Email = req.Email
		u.UserActive = req.UserActive

		err := repo.DB.UpdateUser(u)
		if err != nil {
			log.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
		if len(req.Password) > 0 {
			// changing password
			err := repo.DB.UpdatePassword(id, fmt.Sprintf("%s", req.Password))
			if err != nil {
				log.Println(err)
				ClientError(w, r, http.StatusBadRequest)
				return
			}
		}
	} else {
		u.AccessLevel = 3
		u.FirstName = req.FirstName
		u.LastName = req.LastName
		u.Email = req.Email
		u.UserActive = req.UserActive
		u.Password = []byte(req.Password)

		_, err := repo.DB.InsertUser(u)
		if err != nil {
			log.Println(err)
			ClientError(w, r, http.StatusBadRequest)
			return
		}
	}

	var jsonResp jsonResp
	jsonResp.OK = true
	if id > 0 {
		jsonResp.Message = "User updated"
	} else {
		jsonResp.Message = "User added"
	}

	helpers.RenderJSON(w, r, jsonResp)
}

// DeleteUser soft deletes a user
func (repo *DBRepo) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_ = repo.DB.DeleteUser(id)
	var jsonResp jsonResp
	jsonResp.OK = true
	jsonResp.Message = "User deleted"

	helpers.RenderJSON(w, r, jsonResp)
}

// ClientError will display error page for client error i.e. bad request
func ClientError(w http.ResponseWriter, r *http.Request, status int) {
	switch status {
	case http.StatusNotFound:
		show404(w, r)
	case http.StatusInternalServerError:
		show500(w, r)
	default:
		http.Error(w, http.StatusText(status), status)
	}
}

// ServerError will display error page for internal server error
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = log.Output(2, trace)
	show500(w, r)
}

func show404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	http.ServeFile(w, r, "./ui/static/404.html")
}

func show500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	http.ServeFile(w, r, "./ui/static/500.html")
}

func printTemplateError(w http.ResponseWriter, err error) {
	_, _ = fmt.Fprint(w, fmt.Sprintf(`<small><span class='text-danger'>Error executing template: %s</span></small>`, err))
}

func (repo *DBRepo) Preferences(w http.ResponseWriter, r *http.Request) {
	var refs models.PreferencesResponse
	refs.OK = true
	refs.Message = "Preferences retrieved"

	allRefs, err := repo.DB.AllPreferences()
	if err != nil {
		log.Println(err)
		refs.OK = false
		refs.Message = err.Error()
	}

	refs.Preferences = allRefs

	helpers.RenderJSON(w, r, refs)

}

func (repo *DBRepo) SetSystemPref(w http.ResponseWriter, r *http.Request) {
	var jsonResp jsonResp
	jsonResp.OK = true
	jsonResp.Message = "Preference updated"

	var req models.SystemPrefRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println(err)
	}

	err = repo.DB.UpdateSystemPref(req.PrefName, req.PrefValue)
	if err != nil {
		jsonResp.OK = false
		jsonResp.Message = err.Error()
	}

	repo.App.PreferenceMap["monitoring_live"] = req.PrefValue

	helpers.RenderJSON(w, r, jsonResp)
}

// ToggleMonitoring toggles monitoring on/off
func (repo *DBRepo) ToggleMonitoring(w http.ResponseWriter, r *http.Request) {
	var reactJsResponse models.ToggleMonitoringRequest
	err := json.NewDecoder(r.Body).Decode(&reactJsResponse)
	if err != nil {
		log.Println(err)
	}

	if reactJsResponse.Enabled == true {
		log.Println("Starting monitoring...")
		repo.App.PreferenceMap["monitoring_live"] = "1"
		repo.StartMonitoring()
		repo.App.Scheduler.Start()
	} else {
		log.Println("Stopping monitoring...")
		repo.App.PreferenceMap["monitoring_live"] = "0"
		for _, v := range repo.App.MonitorMap {
			repo.App.Scheduler.Remove(v)
		}

		for k := range repo.App.MonitorMap {
			delete(repo.App.MonitorMap, k)
		}

		// delete all entries from scheduler, be sure to stop the scheduler
		for _, v := range repo.App.Scheduler.Entries() {
			repo.App.Scheduler.Remove(v.ID)
		}

		repo.App.Scheduler.Stop()

		data := make(map[string]string)
		data["message"] = "Monitoring stopped"

		err := app.WsClient.Trigger("public-channel",
			"app-stopping", data)
		if err != nil {
			log.Println(err)
		}

	}

	var jsonResp jsonResp
	jsonResp.OK = true
	jsonResp.Message = "Monitoring updated"

	out, _ := json.MarshalIndent(jsonResp, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}
