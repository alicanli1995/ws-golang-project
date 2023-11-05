package handlers

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"golang-vigilate-project/internal/helpers"
	"golang-vigilate-project/internal/models"
	"log"
	"net/http"
	"sort"
)

type ByHost []models.ScheduleResponse

func (a ByHost) Len() int           { return len(a) }
func (a ByHost) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByHost) Less(i, j int) bool { return a[i].Host < a[j].Host }

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var response models.ListEntriesResponse
	var list []models.ScheduleResponse

	for k, v := range repo.App.MonitorMap {
		var item models.ScheduleResponse
		item.ID = k
		hs, err := repo.DB.GetHostServiceByID(k)
		if err != nil {
			printTemplateError(w, err)
			log.Printf("error getting host service for id %d: %s\n", k, err)
			return
		}
		item.ScheduleText = fmt.Sprintf("@every %d%s", hs.SchedulerNumber, hs.SchedulerUnit)
		item.LastRunFromHS = hs.LastCheck
		item.HostServiceID = hs.ID
		item.Host = hs.HostName
		item.Service = hs.Service.ServiceName

		item.EntryID = v
		entry := &cron.Entry{
			ID:       v,
			Schedule: app.Scheduler.Entry(v).Schedule,
			Next:     app.Scheduler.Entry(v).Next,
			Prev:     app.Scheduler.Entry(v).Prev,
		}
		item.Entry = *entry

		list = append(list, item)
	}

	sort.Sort(ByHost(list))

	response.Entries = list
	response.OK = true
	response.Message = "Schedule entries"

	helpers.RenderJSON(w, r, response)
}
