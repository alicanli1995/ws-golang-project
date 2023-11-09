package elastic

import "golang-observer-project/internal/models"

type Operations interface {
	AddDocument(indexName string, documentID string, times models.ComputeTimes) error
	GetDocumentsByIDAndInLastXMinutes(indexName string, minutes int, hostID int, serviceID int) ([]models.ComputeTimes, error)
}
