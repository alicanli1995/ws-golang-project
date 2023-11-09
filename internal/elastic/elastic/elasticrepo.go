package elastic

import (
	"github.com/elastic/go-elasticsearch/v7"
	"golang-observer-project/internal/config"
	"golang-observer-project/internal/elastic"
)

var app *config.AppConfig

type elasticRepo struct {
	App           *config.AppConfig
	ElasticClient *elasticsearch.Client
}

func NewElasticRepo(ElasticClient *elasticsearch.Client, a *config.AppConfig) elastic.Operations {
	app = a
	return &elasticRepo{
		App:           a,
		ElasticClient: ElasticClient,
	}
}
