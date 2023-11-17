package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang-observer-project/internal/certificateutils"
	"golang-observer-project/internal/channeldata"
	"golang-observer-project/internal/helpers"
	"golang-observer-project/internal/models"
	"golang-observer-project/internal/sms"
	"html/template"
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
	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		log.Println(err)
		return
	}
	locationNow := time.Now().In(location)
	hs.LastCheck = locationNow

	err = repo.DB.UpdateHostService(hs)
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

func (repo *DBRepo) broadcastMessageJsonObject(ch, eventName string, data map[string]interface{}) error {
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

	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		log.Println(err)
		return
	}
	hs.LastCheck = time.Now().In(location)

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
	helpers.RenderJSON(w, resp)

}

func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostServices) (string, string) {
	var msg, newStatus string

	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = repo.testHTTP(h.URL, h, hs)
		break
	case HTTPS:
		msg, newStatus = repo.testHTTPS(h.URL, h, hs)
		break
	case SSLCertificate:
		msg, newStatus = repo.testSSLCert(h.URL, h, hs)
		break
	}

	if newStatus != hs.Status {
		repo.pushStatusChangeEvent(h, hs, newStatus)
		// add to the event log
		repo.addEvents(h, hs, newStatus, msg)

		if repo.App.PreferenceMap["notify_via_email"] == "1" {
			if hs.Status != "pending" {
				mm := channeldata.MailData{
					ToName:    repo.App.PreferenceMap["notify_name"],
					ToAddress: repo.App.PreferenceMap["notify_email"],
				}
				if newStatus == "healthy" {
					mm.Subject = fmt.Sprintf("HEALTHY : service %s on host %s", hs.Service.ServiceName, h.HostName)
					mm.Content = template.HTML(fmt.Sprintf("Service %s on host %s is now <strong>HEALTHY</strong>",
						hs.Service.ServiceName, h.HostName))
				} else if newStatus == "problem" {
					mm.Subject = fmt.Sprintf("PROBLEM : service %s on host %s", hs.Service.ServiceName, h.HostName)
					mm.Content = template.HTML(fmt.Sprintf("Service %s on host %s is now <strong>PROBLEM</strong>",
						hs.Service.ServiceName, h.HostName))
				} else if newStatus == "warning" {
					mm.Subject = fmt.Sprintf("WARNING : service %s on host %s", hs.Service.ServiceName, h.HostName)
					mm.Content = template.HTML(fmt.Sprintf("Service %s on host %s is now <strong>WARNING</strong>",
						hs.Service.ServiceName, h.HostName))
				}

				helpers.SendEmail(mm)
			}
		}

		if repo.App.PreferenceMap["notify_via_sms"] == "1" {
			if hs.Status != "pending" {
				to := repo.App.PreferenceMap["sms_notify_number"]
				msg := ""

				if newStatus == "healthy" {
					msg = fmt.Sprintf("Service %s on host %s is now HEALTHY",
						hs.Service.ServiceName, h.HostName)
				} else if newStatus == "problem" {
					msg = fmt.Sprintf("Service %s on host %s is now PROBLEM",
						hs.Service.ServiceName, h.HostName)
				} else if newStatus == "warning" {
					msg = fmt.Sprintf("Service %s on host %s is now WARNING",
						hs.Service.ServiceName, h.HostName)
				}

				err := sms.SendTextTwilio(to, msg, repo.App)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}

	repo.pushScheduleChangeEvent(hs, newStatus)

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
	data["old_status"] = hs.Status
	data["message"] = fmt.Sprintf("%s is %s", hs.Service.ServiceName, newStatus)
	data["last_check"] = time.Now().Format("2006-01-02 15:04:05")

	_ = repo.broadcastMessage("public-channel", "host-service-status-changed", data)
}

func (repo *DBRepo) pushScheduleChangeEvent(hs models.HostServices, newStatus string) {

	yearOne := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)
	data := make(map[string]string)
	data["host_service_id"] = strconv.Itoa(hs.ID)
	data["service_id"] = strconv.Itoa(hs.ServiceID)
	data["host_id"] = strconv.Itoa(hs.HostID)
	if app.Scheduler.Entry(repo.App.MonitorMap[hs.ID]).Next.After(yearOne) {
		data["next_run"] = app.Scheduler.Entry(repo.App.MonitorMap[hs.ID]).Next.Format("2006-01-02 15:04:05")
	} else {
		data["next_run"] = "Pending..."
	}
	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		log.Println(err)
		return
	}
	locationNow := time.Now().In(location)

	data["last_run"] = locationNow.Format("2006-01-02 15:04:05")
	data["host"] = hs.HostName
	data["service"] = hs.Service.ServiceName
	data["schedule"] = fmt.Sprintf("@every %d%s", hs.SchedulerNumber, hs.SchedulerUnit)
	data["status"] = newStatus
	data["icon"] = hs.Service.Icon

	_ = repo.broadcastMessage("public-channel", "schedule-changed-event", data)

}

func (repo *DBRepo) testHTTP(url string, h models.Host, hs models.HostServices) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "https://", "http://", -1)

	resp, err := http.Get(url)
	if err != nil {
		cpTime, _ := repo.addElastic(url, h, hs)
		data := make(map[string]interface{})
		data["service_info"] = cpTime
		_ = repo.broadcastMessageJsonObject("public-channel", "host-service-check-response", data)
		return err.Error(), "problem"
	}

	cpTime, _ := repo.addElastic(url, h, hs)
	data := make(map[string]interface{})
	data["service_info"] = cpTime

	_ = repo.broadcastMessageJsonObject("public-channel", "host-service-check-response", data)

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

func (repo *DBRepo) addElastic(url string, h models.Host, hs models.HostServices) (*models.ComputeTimes, error) {
	computeTimes := helpers.ComputeTime(url)

	computeTimes.ID = uuid.New().String()
	computeTimes.Host = h
	computeTimes.HostServices = hs
	computeTimes.CreatedAt = time.Now()
	computeTimes.UpdatedAt = time.Now()

	err := repo.ElasticClient.AddDocument("performances", computeTimes.ID, *computeTimes)
	if err != nil {
		return nil, err
	}
	return computeTimes, err
}

// testHTTPS tests an url with https
func (repo *DBRepo) testHTTPS(url string, h models.Host, hs models.HostServices) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "http://", "https://", -1)

	resp, err := http.Get(url)
	if err != nil {
		_, _ = repo.addElastic(url, h, hs)
		return err.Error(), "problem"
	}

	cpTime, _ := repo.addElastic(url, h, hs)
	data := make(map[string]interface{})
	data["service_info"] = cpTime

	_ = repo.broadcastMessageJsonObject("public-channel", "host-service-check-response", data)

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

func (repo *DBRepo) testSSLCert(url string, h models.Host, hs models.HostServices) (string, string) {
	if strings.HasPrefix(url, "https://") {
		url = strings.Replace(url, "https://", "", -1)
	}

	if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "", -1)
	}

	// scanning ssl cert for expiry date
	var certDetailsChannel chan certificateutils.CertificateDetails
	var errorsChannel chan error
	certDetailsChannel = make(chan certificateutils.CertificateDetails, 1)
	errorsChannel = make(chan error, 1)

	var messages, newStatus string

	scanHost(url, certDetailsChannel, errorsChannel)

	for i, certDetailsInQueue := 0, len(certDetailsChannel); i < certDetailsInQueue; i++ {
		certDetails := <-certDetailsChannel
		certificateutils.CheckExpirationStatus(&certDetails, 30)

		if certDetails.ExpiringSoon {

			if certDetails.DaysUntilExpiration < 7 {
				messages = certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
				newStatus = "problem"
			} else {
				messages = certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
				newStatus = "warning"
			}
		} else {
			messages = certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
			newStatus = "healthy"
		}
	}

	return messages, newStatus

}

// scanHost gets cert details from an internet host
func scanHost(hostname string, certDetailsChannel chan certificateutils.CertificateDetails, errorsChannel chan error) {

	res, err := certificateutils.GetCertificateDetails(hostname, 10)
	if err != nil {
		errorsChannel <- err
	} else {
		certDetailsChannel <- res
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
		data["next_run"] = repo.App.Scheduler.Entry(scheduleID).Next.Format("2006-01-02 15:04:05")
		data["last_run"] = time.Now().Format("2006-01-02 15:04:05")
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

func (repo *DBRepo) GetDocumentsInLastXMinutes(w http.ResponseWriter, r *http.Request) {
	indexName := chi.URLParam(r, "indexName")
	minutes, err := strconv.Atoi(chi.URLParam(r, "minutes"))
	if err != nil {
		log.Println(err)
		return
	}
	hostID, err := strconv.Atoi(chi.URLParam(r, "hostID"))
	if err != nil {
		log.Println(err)
		return
	}

	serviceID, err := strconv.Atoi(chi.URLParam(r, "serviceID"))
	if err != nil {
		log.Println(err)
		return
	}

	computeTimes, err := repo.ElasticClient.GetDocumentsByIDAndInLastXMinutes(indexName, minutes, hostID, serviceID)
	if err != nil {
		log.Println(err)
		return
	}

	helpers.RenderJSON(w, computeTimes)
}
