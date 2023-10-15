package dbrepo

import (
	"context"
	"database/sql"
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

	stmt := `
		INSERT INTO host_services (host_id, service_id,active, scheduler_number, scheduler_unit,
		                           status, created_at, updated_at) VALUES
		                           ($1, 1, 0, 3, 'm', 'pending' , $2, $3)`

	_, err = m.DB.ExecContext(ctx, stmt, newID, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return newID, err
	}

	return newID, nil
}

// FindHostByID finds a host by id
func (m *postgresDBRepo) FindHostByID(id int) (models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, host_name, canonical_name, url, ip, ipv6, location, os, active,created_at, updated_at
		FROM hosts WHERE id = $1`

	var host models.Host
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&host.ID,
		&host.HostName,
		&host.CanonicalName,
		&host.URL,
		&host.IP,
		&host.IPV6,
		&host.Location,
		&host.OS,
		&host.Active,
		&host.CreatedAt,
		&host.UpdatedAt,
	)
	if err != nil {
		log.Println(err)
		return host, err
	}

	// get all services for host
	query = `
		SELECT hs.id, hs.host_id, hs.service_id, hs.active, hs.scheduler_number, hs.scheduler_unit,
		       hs.last_check, hs.status, hs.created_at, hs.updated_at,
		       s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
		FROM host_services hs
		LEFT JOIN services s ON (s.id = hs.service_id)
		WHERE host_id = $1`

	rows, err := m.DB.QueryContext(ctx, query, host.ID)
	if err != nil {
		log.Println(err)
		return host, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var services []models.HostServices

	for rows.Next() {
		var s models.HostServices
		err = rows.Scan(
			&s.ID,
			&s.HostID,
			&s.ServiceID,
			&s.Active,
			&s.SchedulerNumber,
			&s.SchedulerUnit,
			&s.LastCheck,
			&s.Status,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.Service.ID,
			&s.Service.ServiceName,
			&s.Service.Active,
			&s.Service.Icon,
			&s.Service.CreatedAt,
			&s.Service.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			return host, err
		}
		services = append(services, s)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return host, err
	}

	host.HostServices = services

	return host, nil
}

// UpdateHost updates a host in the database
func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	log.Println(h)
	query := `
		UPDATE hosts SET host_name = $1, canonical_name = $2, url = $3, ip = $4, 
		                 ipv6 = $5, location = $6, os = $7, active = $8, updated_at = $9
		WHERE id = $10`

	_, err := m.DB.ExecContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		h.ID,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// AllHosts returns a slice of all hosts
func (m *postgresDBRepo) AllHosts() ([]models.Host, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at
		FROM hosts ORDER BY host_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host

	for rows.Next() {
		s := &models.Host{}
		err = rows.Scan(
			&s.ID,
			&s.HostName,
			&s.CanonicalName,
			&s.URL,
			&s.IP,
			&s.IPV6,
			&s.Location,
			&s.OS,
			&s.Active,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, *s)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return hosts, nil
}

// UpdateHostServiceStatus updates the status of a host service
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE host_services SET active = $1, updated_at = $2 WHERE host_id = $3 AND service_id = $4`

	_, err := m.DB.ExecContext(ctx, stmt, active, time.Now(), hostID, serviceID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
