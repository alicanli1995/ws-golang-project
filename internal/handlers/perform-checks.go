package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang-vigilate-project/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HTTP           = 1
	HTTPS          = 2
	SSLCertificate = 3
)

type jsonResp struct {
	OK            bool      `json:"ok"`
	Message       string    `json:"message"`
	ServiceID     int       `json:"service_id"`
	HostServiceID int       `json:"host_service_id"`
	HostID        int       `json:"host_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	LastCheck     time.Time `json:"last_check"`
}

// ScheduledCheck is used to check a host service on a schedule
func (repo *DBRepo) ScheduledCheck(hsID int) {

	hs, err := repo.DB.GetHostServiceByID(hsID)
	if err != nil {
		log.Println(err)
		return
	}

	h, err := repo.DB.FindHostByID(hs.HostID)
	if err != nil {
		log.Println(err)
		return
	}

	newStatus, msg := repo.testServiceForHost(h, hs)

	if newStatus != hs.Status {
		repo.updateHostServiceStatusCount(h, hs, newStatus, msg)
	}

}

func (repo *DBRepo) updateHostServiceStatusCount(h models.Host, hs models.HostServices, newStatus, msg string) {

	// update the status and last check time
	hs.Status = newStatus
	hs.LastMessage = msg
	hs.LastCheck = time.Now()

	err := repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Println(err)
		return
	}

	pending, healthy, warning, problem, err := repo.DB.GetAllServicesStatusCounts()
	if err != nil {
		log.Println(err)
		return
	}

	data := make(map[string]string)
	data["pending_count"] = strconv.Itoa(pending)
	data["healthy_count"] = strconv.Itoa(healthy)
	data["warning_count"] = strconv.Itoa(warning)
	data["problem_count"] = strconv.Itoa(problem)

	_ = repo.broadcastMessage("public-channel", "host-service-count-changed", data)
}

func (repo *DBRepo) broadcastMessage(ch, eventName string, data map[string]string) error {
	err := repo.App.WsClient.Trigger(ch, eventName, data)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (repo *DBRepo) PerformCheck(w http.ResponseWriter, r *http.Request) {
	hostServiceID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	oldStatus := chi.URLParam(r, "oldStatus")
	okay := true

	hs, err := repo.DB.GetHostServiceByID(hostServiceID)
	if err != nil {
		log.Printf("error getting host service by id: %s\n", err)
		okay = false
	}

	h, err := repo.DB.FindHostByID(hs.HostID)
	if err != nil {
		log.Printf("error getting host by id: %s\n", err)
		okay = false
	}

	newStatus, msg := repo.testServiceForHost(h, hs)
	repo.addEvents(h, hs, newStatus, msg)
	if newStatus != hs.Status {
		repo.pushStatusChangeEvent(h, hs, newStatus)
	}

	// update the status and last check time
	hs.Status = newStatus
	hs.LastMessage = msg
	hs.LastCheck = time.Now()

	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Printf("error updating host service: %s\n", err)
		okay = false
	}

	var resp jsonResp

	if okay {
		resp = jsonResp{
			OK:            true,
			Message:       msg,
			ServiceID:     hs.ServiceID,
			HostServiceID: hs.ID,
			HostID:        hs.HostID,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			LastCheck:     time.Now(),
		}
	} else {
		resp = jsonResp{
			OK:      false,
			Message: msg,
		}
	}

	// send the response
	out, _ := json.MarshalIndent(resp, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}

func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostServices) (string, string) {
	var msg, newStatus string

	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = repo.testHTTP(h.URL)
		break
	}

	if newStatus != hs.Status {
		repo.pushStatusChangeEvent(h, hs, newStatus)
		// add to the event log
		repo.addEvents(h, hs, newStatus, msg)
	}

	repo.pushScheduleChangeEvent(hs, newStatus)

	// TODO - Send email or sms if appropriate

	return newStatus, msg
}

func (repo *DBRepo) addEvents(h models.Host, hs models.HostServices, newStatus string, msg string) {
	err := repo.DB.InsertEvent(models.Event{
		EventType:     newStatus,
		HostServiceID: hs.ID,
		HostID:        hs.HostID,
		ServiceName:   hs.Service.ServiceName,
		HostName:      h.HostName,
		Message:       msg,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		log.Println(err)
	}
}

func (repo *DBRepo) pushStatusChangeEvent(h models.Host, hs models.HostServices, newStatus string) {
	data := make(map[string]string)
	data["host_id"] = strconv.Itoa(hs.HostID)
	data["host_service_id"] = strconv.Itoa(hs.ID)
	data["service_name"] = hs.Service.ServiceName
	data["host_name"] = h.HostName
	data["icon"] = hs.Service.Icon
	data["status"] = newStatus
	data["message"] = fmt.Sprintf("%s is %s", hs.Service.ServiceName, newStatus)
	data["last_check"] = time.Now().Format("2006-01-02 15:04:05 PM")

	_ = repo.broadcastMessage("public-channel", "host-service-status-changed", data)
}

func (repo *DBRepo) pushScheduleChangeEvent(hs models.HostServices, newStatus string) {

	yearOne := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)
	data := make(map[string]string)
	data["host_service_id"] = strconv.Itoa(hs.ID)
	data["service_id"] = strconv.Itoa(hs.ServiceID)
	data["host_id"] = strconv.Itoa(hs.HostID)
	if app.Scheduler.Entry(repo.App.MonitorMap[hs.ID]).Next.After(yearOne) {
		data["next_run"] = app.Scheduler.Entry(repo.App.MonitorMap[hs.ID]).Next.Format("2006-01-02 15:04:05 PM")
	} else {
		data["next_run"] = "Pending..."
	}
	data["last_run"] = time.Now().Format("2006-01-02 15:04:05 PM")
	data["host"] = hs.HostName
	data["service"] = hs.Service.ServiceName
	data["schedule"] = fmt.Sprintf("@every %d%s", hs.SchedulerNumber, hs.SchedulerUnit)
	data["status"] = newStatus
	data["icon"] = hs.Service.Icon

	_ = repo.broadcastMessage("public-channel", "schedule-changed-event", data)

}

func (repo *DBRepo) testHTTP(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "https://", "http://", -1)

	resp, err := http.Get(url)
	if err != nil {
		return err.Error(), "problem"
	}

	defer func(resp *http.Response) {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp)

	if resp.StatusCode != http.StatusOK {
		return resp.Status, "problem"
	} else {
		return resp.Status, "healthy"
	}
}

func (repo *DBRepo) addToMonitorMap(hs models.HostServices) {
	if repo.App.PreferenceMap["monitoring_live"] == "1" {
		var j job
		j.HostServiceID = hs.ID
		scheduleID, err := repo.App.Scheduler.AddJob(fmt.Sprintf("@every %d%s", hs.SchedulerNumber, hs.SchedulerUnit), j)
		if err != nil {
			log.Println(err)
			return
		}
		repo.App.MonitorMap[hs.ID] = scheduleID
		data := make(map[string]string)
		data["host_service_id"] = strconv.Itoa(hs.ID)
		data["service_id"] = strconv.Itoa(hs.ServiceID)
		data["host_id"] = strconv.Itoa(hs.HostID)
		data["next_run"] = repo.App.Scheduler.Entry(scheduleID).Next.Format("2006-01-02 15:04:05 PM")
		data["last_run"] = time.Now().Format("2006-01-02 15:04:05 PM")
		data["host"] = hs.HostName
		data["service"] = hs.Service.ServiceName
		data["schedule"] = fmt.Sprintf("@every %d%s", hs.SchedulerNumber, hs.SchedulerUnit)
		data["status"] = hs.Status
		data["icon"] = hs.Service.Icon
		data["message"] = fmt.Sprintf("%s is %s", hs.Service.ServiceName, hs.Status)

		_ = repo.broadcastMessage("public-channel", "schedule-changed-event", data)
	}
}

func (repo *DBRepo) removeFromMonitorMap(hs models.HostServices) {
	if repo.App.PreferenceMap["monitoring_live"] == "1" {
		if scheduleID, ok := repo.App.MonitorMap[hs.ID]; ok {
			repo.App.Scheduler.Remove(scheduleID)
			delete(repo.App.MonitorMap, hs.ID)

			data := make(map[string]string)
			data["host_service_id"] = strconv.Itoa(hs.ID)
			err := repo.broadcastMessage("public-channel", "schedule-item-removed-event", data)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
