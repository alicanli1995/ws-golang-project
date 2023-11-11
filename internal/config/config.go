package config

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/pusher/pusher-http-go"
	"github.com/robfig/cron/v3"
	"golang-observer-project/internal/channeldata"
	"golang-observer-project/internal/driver"
	"html/template"
)

// AppConfig holds application configuration
type AppConfig struct {
	DB            *driver.DB
	InProduction  bool
	Domain        string
	MonitorMap    map[int]cron.EntryID
	PreferenceMap map[string]string
	Scheduler     *cron.Cron
	WsClient      pusher.Client
	PusherSecret  string
	TemplateCache map[string]*template.Template
	MailQueue     chan channeldata.MailJob
	Version       string
	Identifier    string
	ElasticConfig *elasticsearch.Client
}
