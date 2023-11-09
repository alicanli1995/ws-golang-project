package models

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/robfig/cron/v3"
	"time"
)

var (
	// ErrNoRecord no record found in database error
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrInvalidCredentials invalid username/password error
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail duplicate email error
	ErrDuplicateEmail = errors.New("models: duplicate email")
	// ErrInactiveAccount inactive account error
	ErrInactiveAccount = errors.New("models: Inactive Account")
)

// User model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	UserActive  int
	AccessLevel int
	Email       string
	Password    []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
	Preferences map[string]string
}

// User model
type UserRequest struct {
	ID          int
	FirstName   string
	LastName    string
	UserActive  int
	AccessLevel int
	Email       string
	Password    string
}

// Preference model
type Preference struct {
	ID         int
	Name       string
	Preference string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Host model
type Host struct {
	ID            int
	HostName      string
	CanonicalName string
	URL           string
	IP            string
	IPV6          string
	Location      string
	OS            string
	Active        int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	HostServices  []HostServices
}

// Services model
type Services struct {
	ID          int
	ServiceName string
	Active      int
	Icon        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HostServices model
type HostServices struct {
	ID              int
	HostID          int
	ServiceID       int
	Active          int
	SchedulerNumber int
	SchedulerUnit   string
	Status          string
	LastCheck       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Service         Services
	HostName        string
	LastMessage     string
}

// Schedule model
type Schedule struct {
	ID            int
	EntryID       cron.EntryID
	Entry         cron.Entry
	Host          string
	Service       string
	LastRunFromHS time.Time
	HostServiceID int
	ScheduleText  string
}

type Event struct {
	ID            int
	EventType     string
	HostServiceID int
	HostID        int
	ServiceName   string
	HostName      string
	Message       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type HostJsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Host    Host   `json:"host"`
}

type HostsJsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Hosts   []Host `json:"hosts"`
}

type ServiceJSON struct {
	OK bool `json:"ok"`
}

type SystemPrefRequest struct {
	PrefName  string `json:"pref_name"`
	PrefValue string `json:"pref_value"`
}

type PreferencesResponse struct {
	OK          bool         `json:"ok"`
	Message     string       `json:"message"`
	Preferences []Preference `json:"preferences"`
}

type DashResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Healthy int    `json:"healthy"`
	Warning int    `json:"warning"`
	Problem int    `json:"problem"`
	Pending int    `json:"pending"`
	Hosts   []Host `json:"hosts"`
}

type ToggleMonitoringRequest struct {
	Enabled bool `json:"enabled"`
}

type HostPostRequest struct {
	HostID        int    `json:"ID"`
	HostName      string `json:"HostName"`
	CanonicalName string `json:"CanonicalName"`
	URL           string `json:"URL"`
	IP            string `json:"IP"`
	IPV6          string `json:"IPV6"`
	Location      string `json:"Location"`
	OS            string `json:"OS"`
	Active        int    `json:"Active"`
}

type ToggleServiceRequest struct {
	HostID    int `json:"host_id"`
	ServiceID int `json:"service_id"`
	Active    int `json:"active"`
}

type ScheduleResponse struct {
	ID            int
	Host          string
	Service       string
	LastRunFromHS time.Time
	HostServiceID int
	ScheduleText  string

	EntryID cron.EntryID
	Entry   cron.Entry
}

type ListEntriesResponse struct {
	OK      bool               `json:"ok"`
	Message string             `json:"message"`
	Entries []ScheduleResponse `json:"entries"`
}

type LoginResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	User    User   `json:"user"`
	Token   string `json:"token"`
}

type CreateUserParams struct {
	Username       string
	HashedPassword []byte
	FullName       string
	Email          string
}

type CreateSessionsParams struct {
	Email        string    `json:"email"`
	ID           uuid.UUID `json:"id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Session struct {
	ID           uuid.UUID          `json:"id"`
	Email        string             `json:"email"`
	RefreshToken string             `json:"refresh_token"`
	UserAgent    string             `json:"user_agent"`
	ClientIp     string             `json:"client_ip"`
	IsBlocked    bool               `json:"is_blocked"`
	ExpiresAt    time.Time          `json:"expires_at"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
}

type ComputeTimes struct {
	ID           string         `json:"id"`
	DNSDone      time.Duration  `json:"DNSDone"`
	ConnectTime  time.Duration  `json:"ConnectTime"`
	TLSHandshake time.Duration  `json:"TLSHandshake"`
	FirstByte    time.Duration  `json:"FirstByte"`
	TotalTime    time.Duration  `json:"TotalTime"`
	Host         Host           `json:"Host"`
	HostServices []HostServices `json:"HostServices"`
	CreatedAt    time.Time      `json:"CreatedAt"`
	UpdatedAt    time.Time      `json:"UpdatedAt"`
}
