package repository

import "golang-vigilate-project/internal/models"

// DatabaseRepo is the database repository
type DatabaseRepo interface {
	// preferences
	AllPreferences() ([]models.Preference, error)
	UpdateSystemPref(name, value string) error
	SetSystemPref(name, value string) error
	InsertOrUpdateSitePreferences(pm map[string]string) error

	// users and authentication
	GetUserById(id int) (models.User, error)
	InsertUser(u models.User) (int, error)
	UpdateUser(u models.User) error
	DeleteUser(id int) error
	UpdatePassword(id int, newPassword string) error
	Authenticate(email, testPassword string) (int, string, error)
	AllUsers() ([]*models.User, error)
	InsertRememberMeToken(id int, token string) error
	DeleteToken(token string) error
	CheckForToken(id int, token string) bool

	// hosts
	InsertHost(h models.Host) (int, error)
	FindHostByID(id int) (models.Host, error)
	UpdateHost(h models.Host) error
	AllHosts() ([]models.Host, error)
	UpdateHostService(hs models.HostServices) error
	GetAllServicesStatusCounts() (int, int, int, int, error)
	GetServicesByStatus(status string) ([]models.HostServices, error)
	GetHostServiceByID(id int) (models.HostServices, error)
	UpdateHostServiceStatus(hostID, serviceID, active int) error
	GetServicesToMonitor() ([]models.HostServices, error)
	GetHostServiceByHostIDServiceID(hostID, serviceID int) (models.HostServices, error)
	AllEvents() ([]models.Event, error)
	InsertEvent(e models.Event) error

	//sessions
	CreateSession(params models.CreateSessionsParams) (models.Session, error)
}
