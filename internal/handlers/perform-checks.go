package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi"
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

	// update the status and last check time
	hs.Status = newStatus
	hs.LastCheck = time.Now()

	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Printf("error updating host service: %s\n", err)
		okay = false
	}

	// broadcast the status change to pusher

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

	return newStatus, msg
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
