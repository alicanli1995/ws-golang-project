package dbrepo

import (
	"context"
	"golang-vigilate-project/internal/models"
	"log"
	"time"
)

// InsertHost inserts a host into the database
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO hosts (
		    host_name, canonical_name, url, ip, ipv6, location, os, created_at, updated_at) VALUES 
		    ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var newID int
	err := m.DB.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return newID, nil
}
