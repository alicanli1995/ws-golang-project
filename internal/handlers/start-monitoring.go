package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type job struct {
	HostServiceID int
}

func (j job) Run() {
	Repo.ScheduledCheck(j.HostServiceID)
}

func (repo *DBRepo) StartMonitoring() {
	if app.PreferenceMap["monitoring_live"] == "1" {
		data := make(map[string]string)
		data["message"] = "Monitoring started"

		err := app.WsClient.Trigger("public-channel",
			"app-starting", data)
		if err != nil {
			log.Println(err)
		}

		servicesToMonitor, err := repo.DB.GetServicesToMonitor()
		if err != nil {
			log.Println(err)
		}

		for _, service := range servicesToMonitor {
			var sch string
			if service.SchedulerUnit == "d" {
				sch = fmt.Sprintf("@every %d%s", service.SchedulerNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %d%s", service.SchedulerNumber, service.SchedulerUnit)
			}
			var j job
			j.HostServiceID = service.ID
			jobID, err := app.Scheduler.AddJob(sch, j)
			if err != nil {
				log.Println(err)
			}

			app.MonitorMap[service.ID] = jobID

			payload := make(map[string]string)
			payload["message"] = fmt.Sprintf("Monitoring %s", service.Service.ServiceName)
			payload["host_service_id"] = strconv.Itoa(service.ID)

			yearOne := time.Date(0001, 11, 17, 20, 34, 58, 0, time.UTC)
			if app.Scheduler.Entry(app.MonitorMap[service.ID]).Next.After(yearOne) {
				payload["next_run"] = app.Scheduler.Entry(app.MonitorMap[service.ID]).Next.Format("2006-01-02 15:04:05")
			} else {
				payload["next_run"] = "Pending..."
			}

			payload["host"] = service.HostName
			payload["service"] = service.Service.ServiceName
			if service.LastCheck.After(yearOne) {
				payload["last_run"] = service.LastCheck.Format("2006-01-02 15:04:05")
			} else {
				payload["last_run"] = "Pending..."
			}

			payload["schedule"] = fmt.Sprintf("@every %d%s", service.SchedulerNumber, service.SchedulerUnit)

			err = app.WsClient.Trigger("public-channel", "next-run-event", payload)
			if err != nil {
				log.Println(err)
			}

			err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload)
			if err != nil {
				log.Println(err)
			}

		}

	}
}
