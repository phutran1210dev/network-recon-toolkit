package scanner

import (
	"context"
	"github.com/netrecon/toolkit/internal/models"
)

// Scanner defines the interface for network scanners
type Scanner interface {
	// Scan performs a network scan on the given target
	Scan(ctx context.Context, target string, config *ScanConfig) (*ScanResult, error)

	// GetName returns the scanner name
	GetName() string

	// ValidateConfig validates the scanner configuration
	ValidateConfig(config *ScanConfig) error
}

// ScanConfig holds configuration for a scan
type ScanConfig struct {
	Ports     string            `json:"ports"`     // Port range (e.g., "1-1000", "80,443,8080")
	Timing    string            `json:"timing"`    // Timing template (0-5 for nmap)
	Arguments string            `json:"arguments"` // Additional scanner arguments
	Output    string            `json:"output"`    // Output format
	Timeout   int               `json:"timeout"`   // Timeout in seconds
	Threads   int               `json:"threads"`   // Number of threads
	Options   map[string]string `json:"options"`   // Scanner-specific options
}

// ScanResult holds the results of a network scan
type ScanResult struct {
	Target    string         `json:"target"`
	Scanner   string         `json:"scanner"`
	Status    string         `json:"status"`
	StartTime string         `json:"start_time"`
	EndTime   string         `json:"end_time"`
	Duration  string         `json:"duration"`
	Hosts     []*models.Host `json:"hosts"`
	RawOutput string         `json:"raw_output"`
	Error     string         `json:"error,omitempty"`
}

// ScannerManager manages multiple scanners
type ScannerManager struct {
	scanners map[string]Scanner
}

// NewScannerManager creates a new scanner manager
func NewScannerManager() *ScannerManager {
	return &ScannerManager{
		scanners: make(map[string]Scanner),
	}
}

// RegisterScanner registers a scanner with the manager
func (sm *ScannerManager) RegisterScanner(scanner Scanner) {
	sm.scanners[scanner.GetName()] = scanner
}

// GetScanner returns a scanner by name
func (sm *ScannerManager) GetScanner(name string) (Scanner, bool) {
	scanner, exists := sm.scanners[name]
	return scanner, exists
}

// ListScanners returns all available scanner names
func (sm *ScannerManager) ListScanners() []string {
	var names []string
	for name := range sm.scanners {
		names = append(names, name)
	}
	return names
}

// ScanOptions provides additional options for scanning
type ScanOptions struct {
	SaveToDatabase bool   `json:"save_to_database"`
	OutputFile     string `json:"output_file"`
	OutputFormat   string `json:"output_format"`
	Verbose        bool   `json:"verbose"`
}
