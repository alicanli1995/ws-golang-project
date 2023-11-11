package dbrepo

import (
	"context"
	"database/sql"
	"golang-observer-project/internal/models"
	"log"
	"time"
)

// InsertHost inserts a host into the database
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO hosts (
		    host_name, canonical_name, url, ip, ipv6, location, os, active, created_at, updated_at) VALUES 
		    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	var newID int
	err := m.DB.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	query = `
	 SELECT id from services
	     `
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var serviceID int
		err = rows.Scan(&serviceID)
		if err != nil {
			return 0, err
		}

		stmt := `
		INSERT INTO host_services (host_id, service_id, active, scheduler_number, scheduler_unit,
		                           status, created_at, updated_at) VALUES
		                           ($1, $2, 0, 3, 'm', 'pending' , $3, $4)`

		_, err = m.DB.ExecContext(ctx, stmt, newID, serviceID, time.Now(), time.Now())
		if err != nil {
			log.Println(err)
			return newID, err
		}
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
		WHERE host_id = $1
		ORDER BY s.service_name`

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
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

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

	// get all services for host

	query = `
		SELECT hs.id, hs.host_id, hs.service_id, hs.active, hs.scheduler_number, hs.scheduler_unit,
		       hs.last_check, hs.status, hs.created_at, hs.updated_at,
		       s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
		FROM host_services hs
		LEFT JOIN services s ON (s.id = hs.service_id)
		ORDER BY s.service_name`

	rows, err = m.DB.QueryContext(ctx, query)
	if err != nil {
		log.Println(err)
		return nil, err
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
			return nil, err
		}
		services = append(services, s)

		for i, host := range hosts {
			if host.ID == s.HostID {
				hosts[i].HostServices = append(hosts[i].HostServices, s)
			}
		}

	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return hosts, nil
}

// UpdateHostService updates the status of a host service
func (m *postgresDBRepo) UpdateHostService(hs models.HostServices) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE host_services SET host_id = $1, service_id = $2, active = $3, scheduler_number = $4,
		                         scheduler_unit = $5, last_check = $6, status = $7, updated_at = $8,
		                         last_message = $9
		WHERE id = $10`

	_, err := m.DB.ExecContext(ctx, stmt, hs.HostID, hs.ServiceID, hs.Active, hs.SchedulerNumber,
		hs.SchedulerUnit, hs.LastCheck, hs.Status, time.Now(), hs.LastMessage, hs.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// UpdateHostServiceStatus updates the active status of a host service
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update host_services set active = $1 where host_id = $2 and service_id = $3`

	_, err := m.DB.ExecContext(ctx, stmt, active, hostID, serviceID)
	if err != nil {
		return err
	}
	return nil
}

// GetAllServicesStatusCounts returns the number of services with each status
func (m *postgresDBRepo) GetAllServicesStatusCounts() (int, int, int, int, error) {
	query := `
		SELECT (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'pending') as pending,
		       (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'healthy') as healthy,
		       (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'warning') as warning,
		       (SELECT COUNT(id) FROM host_services WHERE active = 1 AND status = 'problem') as problem`

	var healthy, warning, problem, pending int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query).Scan(&pending, &healthy, &warning, &problem)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return pending, healthy, warning, problem, nil
}

// GetServicesByStatus returns a slice of host services with a given status
func (m *postgresDBRepo) GetServicesByStatus(status string) ([]models.HostServices, error) {
	query := `
		select hs.id,
			   hs.host_id,
			   hs.service_id,
			   hs.active,
			   hs.scheduler_number,
			   hs.scheduler_unit,
			   hs.last_check,
			   hs.status,
			   hs.created_at,
			   hs.updated_at,
			   h.host_name,
			   s.service_name,
			   hs.last_message
		from host_services hs
		left join hosts h on hs.host_id = h.id
		left join services s on hs.service_id = s.id
		where status = $1 and hs.active = 1
		order by h.host_name, s.service_name`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, status)

	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var hosts []models.HostServices

	for rows.Next() {
		var hs models.HostServices
		err = rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.SchedulerNumber,
			&hs.SchedulerUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.HostName,
			&hs.Service.ServiceName,
			&hs.LastMessage,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, hs)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

// GetHostServiceByID returns a host service by id
func (m *postgresDBRepo) GetHostServiceByID(id int) (models.HostServices, error) {
	query := `
			select hs.id,
				hs.host_id,
				hs.service_id,
				hs.active,
				hs.scheduler_number,
				hs.scheduler_unit,
				hs.last_check,
				hs.status,
				hs.created_at,
				hs.updated_at,
				s.service_name,
				s.active,
				s.icon,
				s.created_at,
				s.updated_at,
				h.host_name,
				hs.last_message
		from host_services hs
		left join hosts h on hs.host_id = h.id
		left join services s on hs.service_id = s.id
		where hs.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var hs models.HostServices

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&hs.ID,
		&hs.HostID,
		&hs.ServiceID,
		&hs.Active,
		&hs.SchedulerNumber,
		&hs.SchedulerUnit,
		&hs.LastCheck,
		&hs.Status,
		&hs.CreatedAt,
		&hs.UpdatedAt,
		&hs.Service.ServiceName,
		&hs.Service.Active,
		&hs.Service.Icon,
		&hs.Service.CreatedAt,
		&hs.Service.UpdatedAt,
		&hs.HostName,
		&hs.LastMessage,
	)
	if err != nil {
		return hs, err
	}

	return hs, nil
}

// GetServicesToMonitor returns a slice of host services to monitor
func (m *postgresDBRepo) GetServicesToMonitor() ([]models.HostServices, error) {
	query := `
		select hs.id,
			   hs.host_id,
			   hs.service_id,
			   hs.active,
			   hs.scheduler_number,
			   hs.scheduler_unit,
			   hs.last_check,
			   hs.status,
			   hs.created_at,
			   hs.updated_at,
			   s.id,
			   s.service_name,
			   s.active,	
			   s.icon,
			   s.created_at,	
			   s.updated_at,
			   h.host_name,
			   hs.last_message
		from host_services hs
		left join hosts h on hs.host_id = h.id
		left join services s on hs.service_id = s.id 
		where hs.active = 1 and h.active = 1
		order by h.host_name, s.service_name`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var hosts []models.HostServices

	for rows.Next() {
		var hs models.HostServices
		err = rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.SchedulerNumber,
			&hs.SchedulerUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.Service.ID,
			&hs.Service.ServiceName,
			&hs.Service.Active,
			&hs.Service.Icon,
			&hs.Service.CreatedAt,
			&hs.Service.UpdatedAt,
			&hs.HostName,
			&hs.LastMessage,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, hs)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}

// GetHostServiceByHostIDServiceID returns a host service by host id and service id
func (m *postgresDBRepo) GetHostServiceByHostIDServiceID(hostID, serviceID int) (models.HostServices, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select hs.id,
			   hs.host_id,	
			   hs.service_id,	
			   hs.active,
			   hs.scheduler_number,
			   hs.scheduler_unit,
			   hs.last_check,
			   hs.status,
			   hs.created_at,
			   hs.updated_at,	
			   s.id,	
			   s.service_name,
			   s.active,
			   s.icon,	
			   s.created_at,	
			   s.updated_at,
			   h.host_name,
			   hs.last_message
		from host_services hs
		left join hosts h on hs.host_id = h.id
		left join services s on hs.service_id = s.id
		where hs.host_id = $1 and hs.service_id = $2`

	var hs models.HostServices

	err := m.DB.QueryRowContext(ctx, query, hostID, serviceID).Scan(
		&hs.ID,
		&hs.HostID,
		&hs.ServiceID,
		&hs.Active,
		&hs.SchedulerNumber,
		&hs.SchedulerUnit,
		&hs.LastCheck,
		&hs.Status,
		&hs.CreatedAt,
		&hs.UpdatedAt,
		&hs.Service.ID,
		&hs.Service.ServiceName,
		&hs.Service.Active,
		&hs.Service.Icon,
		&hs.Service.CreatedAt,
		&hs.Service.UpdatedAt,
		&hs.HostName,
		&hs.LastMessage,
	)
	if err != nil {
		return hs, err
	}

	return hs, nil
}

// InsertEvent inserts an event into the database
func (m *postgresDBRepo) InsertEvent(event models.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		insert into events (host_service_id, event_type
		,host_id, service_name, host_name, message, created_at, updated_at) VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := m.DB.ExecContext(ctx, query,
		event.HostServiceID,
		event.EventType,
		event.HostID,
		event.ServiceName,
		event.HostName,
		event.Message,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		log.Println(err)
	}

	return nil
}

// AllEvents returns a slice of all events
func (m *postgresDBRepo) AllEvents() ([]models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select id, host_service_id, event_type, host_id, service_name, host_name, message, created_at, updated_at
		from events order by created_at desc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var events []models.Event

	for rows.Next() {
		var e models.Event
		err = rows.Scan(
			&e.ID,
			&e.HostServiceID,
			&e.EventType,
			&e.HostID,
			&e.ServiceName,
			&e.HostName,
			&e.Message,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
