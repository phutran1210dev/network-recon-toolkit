package models

import (
	"github.com/google/uuid"
	"time"
)

// ScanTarget represents a target for network scanning
type ScanTarget struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Target      string    `json:"target" db:"target"`
	Type        string    `json:"type" db:"type"` // ip, range, domain
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ScanResult represents the result of a network scan
type ScanResult struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TargetID  uuid.UUID  `json:"target_id" db:"target_id"`
	ScanType  string     `json:"scan_type" db:"scan_type"` // nmap, masscan
	Status    string     `json:"status" db:"status"`       // running, completed, failed
	StartTime time.Time  `json:"start_time" db:"start_time"`
	EndTime   *time.Time `json:"end_time" db:"end_time"`
	RawOutput string     `json:"raw_output" db:"raw_output"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// Host represents a discovered host
type Host struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ScanID       uuid.UUID `json:"scan_id" db:"scan_id"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	Hostname     string    `json:"hostname" db:"hostname"`
	Status       string    `json:"status" db:"status"` // up, down, filtered
	OS           string    `json:"os" db:"os"`
	OSConfidence int       `json:"os_confidence" db:"os_confidence"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Port represents an open port on a host
type Port struct {
	ID        uuid.UUID `json:"id" db:"id"`
	HostID    uuid.UUID `json:"host_id" db:"host_id"`
	Number    int       `json:"number" db:"number"`
	Protocol  string    `json:"protocol" db:"protocol"` // tcp, udp
	State     string    `json:"state" db:"state"`       // open, closed, filtered
	Service   string    `json:"service" db:"service"`
	Version   string    `json:"version" db:"version"`
	Product   string    `json:"product" db:"product"`
	ExtraInfo string    `json:"extra_info" db:"extra_info"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Vulnerability represents a detected vulnerability
type Vulnerability struct {
	ID             uuid.UUID `json:"id" db:"id"`
	PortID         uuid.UUID `json:"port_id" db:"port_id"`
	CVE            string    `json:"cve" db:"cve"`
	Severity       string    `json:"severity" db:"severity"` // low, medium, high, critical
	Description    string    `json:"description" db:"description"`
	Solution       string    `json:"solution" db:"solution"`
	ReferenceLinks string    `json:"reference_links" db:"reference_links"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ScanConfiguration represents scan parameters
type ScanConfiguration struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Scanner   string    `json:"scanner" db:"scanner"` // nmap, masscan
	Ports     string    `json:"ports" db:"ports"`
	Arguments string    `json:"arguments" db:"arguments"`
	Timing    string    `json:"timing" db:"timing"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
