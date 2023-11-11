package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aquasecurity/esquery"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"golang-observer-project/internal/models"
	"io"
	"log"
	"strings"
	"time"
)

// AddDocument adds a document to an index
func (elastic *elasticRepo) AddDocument(indexName string, documentID string, ct models.ComputeTimes) error {
	docJSON, err := json.Marshal(ct)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: documentID,
		Body:       strings.NewReader(string(docJSON)),
	}

	res, err := req.Do(context.Background(), elastic.ElasticClient)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	if res.IsError() {
		return fmt.Errorf("error indexing document ID=%s", documentID)
	}

	log.Println("Indexed document ID=", documentID)

	return nil
}

// GetDocumentsByIDAndInLastXMinutes returns all documents in the last X minutes
func (elastic *elasticRepo) GetDocumentsByIDAndInLastXMinutes(indexName string, minutes int, hostID int, serviceID int) ([]models.ComputeTimes, error) {

	res, err := esquery.Search().
		Query(esquery.Bool().
			Must(esquery.Term("Host.ID", hostID)).
			Must(esquery.Term("HostServices.ID", serviceID)).
			Filter(esquery.Range("CreatedAt").
				Gte("now-"+fmt.Sprintf("%dm", minutes)).Lte("now"))).
		Size(10000).
		Sort("CreatedAt", "desc").
		Run(
			elastic.ElasticClient,
			elastic.ElasticClient.Search.WithIndex(indexName),
			elastic.ElasticClient.Search.WithContext(context.TODO()),
		)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("error getting response: %s", res.String())
	}

	var computeTimes []models.ComputeTimes

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, err
	} else {
		log.Println("Found a total of", int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)), "documents")
	}

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		computeTime := models.ComputeTimes{}
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		computeTime.ID = source["id"].(string)
		computeTime.ConnectTime = time.Duration(source["ConnectTime"].(float64))
		computeTime.DNSDone = time.Duration(source["DNSDone"].(float64))
		computeTime.TLSHandshake = time.Duration(source["TLSHandshake"].(float64))
		computeTime.TotalTime = time.Duration(source["TotalTime"].(float64))
		computeTime.Host.HostName = source["Host"].(map[string]interface{})["HostName"].(string)
		computeTime.Host.ID = int(source["Host"].(map[string]interface{})["ID"].(float64))
		computeTime.CreatedAt, _ = time.Parse(time.RFC3339, source["CreatedAt"].(string))
		computeTime.ResponseStatus = int(source["ResponseStatus"].(float64))
		computeTime.HostServices.ID = int(source["HostServices"].(map[string]interface{})["ID"].(float64))

		computeTimes = append(computeTimes, computeTime)
	}

	return computeTimes, nil
}
