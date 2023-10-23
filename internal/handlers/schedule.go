package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"golang-vigilate-project/internal/helpers"
	"golang-vigilate-project/internal/models"
	"log"
	"net/http"
	"sort"
)

type ByHost []models.Schedule

func (a ByHost) Len() int           { return len(a) }
func (a ByHost) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByHost) Less(i, j int) bool { return a[i].Host < a[j].Host }

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var items []models.Schedule

	for k, v := range repo.App.MonitorMap {
		var item models.Schedule
		item.ID = k
		item.EntryID = v
		item.Entry = app.Scheduler.Entry(v)
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

		items = append(items, item)
	}

	// sort by host name
	sort.Sort(ByHost(items))

	data := make(jet.VarMap)
	data.Set("items", items)

	err := helpers.RenderPage(w, r, "schedule", data, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
