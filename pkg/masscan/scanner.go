package masscan

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/netrecon/toolkit/internal/models"
	"github.com/netrecon/toolkit/internal/scanner"
)

// Scanner implements the masscan scanner
type Scanner struct {
	path string
}

// NewScanner creates a new masscan scanner
func NewScanner() (*Scanner, error) {
	// Check if masscan is installed
	path, err := exec.LookPath("masscan")
	if err != nil {
		return nil, fmt.Errorf("masscan not found in PATH: %w", err)
	}

	return &Scanner{path: path}, nil
}

// GetName returns the scanner name
func (s *Scanner) GetName() string {
	return "masscan"
}

// ValidateConfig validates the masscan configuration
func (s *Scanner) ValidateConfig(config *scanner.ScanConfig) error {
	if config.Ports == "" {
		return fmt.Errorf("ports must be specified for masscan")
	}

	// Validate port format
	portRegex := regexp.MustCompile(`^(\d+(-\d+)?)(,\d+(-\d+)?)*$`)
	if !portRegex.MatchString(config.Ports) {
		return fmt.Errorf("invalid port format: %s", config.Ports)
	}

	if config.Threads > 0 && config.Threads > 100000 {
		return fmt.Errorf("thread count too high: %d (max 100000)", config.Threads)
	}

	return nil
}

// Scan performs a masscan scan
func (s *Scanner) Scan(ctx context.Context, target string, config *scanner.ScanConfig) (*scanner.ScanResult, error) {
	if err := s.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	startTime := time.Now()

	// Build masscan command
	args := []string{}

	// Add target
	args = append(args, target)

	// Add ports
	args = append(args, "-p", config.Ports)

	// Add rate (threads)
	if config.Threads > 0 {
		args = append(args, "--rate", strconv.Itoa(config.Threads))
	} else {
		args = append(args, "--rate", "1000") // Default rate
	}

	// Output in JSON format
	args = append(args, "--output-format", "json")

	// Additional arguments
	if config.Arguments != "" {
		additionalArgs := strings.Fields(config.Arguments)
		args = append(args, additionalArgs...)
	}

	// Execute masscan command
	cmd := exec.CommandContext(ctx, s.path, args...)
	output, err := cmd.Output()
	if err != nil {
		endTime := time.Now()
		return &scanner.ScanResult{
			Target:    target,
			Scanner:   s.GetName(),
			Status:    "failed",
			StartTime: startTime.Format(time.RFC3339),
			EndTime:   endTime.Format(time.RFC3339),
			Duration:  endTime.Sub(startTime).String(),
			RawOutput: string(output),
			Error:     err.Error(),
		}, err
	}

	endTime := time.Now()

	// Parse JSON output
	hosts, parseErr := s.parseMasscanJSON(output)
	if parseErr != nil {
		return &scanner.ScanResult{
			Target:    target,
			Scanner:   s.GetName(),
			Status:    "completed_with_errors",
			StartTime: startTime.Format(time.RFC3339),
			EndTime:   endTime.Format(time.RFC3339),
			Duration:  endTime.Sub(startTime).String(),
			RawOutput: string(output),
			Error:     parseErr.Error(),
		}, nil
	}

	return &scanner.ScanResult{
		Target:    target,
		Scanner:   s.GetName(),
		Status:    "completed",
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   endTime.Format(time.RFC3339),
		Duration:  endTime.Sub(startTime).String(),
		Hosts:     hosts,
		RawOutput: string(output),
	}, nil
}

// MasscanResult represents a masscan JSON result
type MasscanResult struct {
	IP        string `json:"ip"`
	Timestamp string `json:"timestamp"`
	Ports     []struct {
		Port   int    `json:"port"`
		Proto  string `json:"proto"`
		Status string `json:"status"`
		Reason string `json:"reason"`
		TTL    int    `json:"ttl"`
	} `json:"ports"`
}

// parseMasscanJSON parses masscan JSON output
func (s *Scanner) parseMasscanJSON(jsonData []byte) ([]*models.Host, error) {
	// Masscan outputs one JSON object per line
	lines := strings.Split(strings.TrimSpace(string(jsonData)), "\n")

	hostMap := make(map[string]*models.Host)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result MasscanResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			// Skip malformed lines
			continue
		}

		// Get or create host
		host, exists := hostMap[result.IP]
		if !exists {
			host = &models.Host{
				ID:        uuid.New(),
				IPAddress: result.IP,
				Status:    "up",
				CreatedAt: time.Now(),
			}
			hostMap[result.IP] = host
		}

		// Add ports to host
		for _, portInfo := range result.Ports {
			_ = &models.Port{
				ID:        uuid.New(),
				HostID:    host.ID,
				Number:    portInfo.Port,
				Protocol:  portInfo.Proto,
				State:     portInfo.Status,
				CreatedAt: time.Now(),
			}

			// Note: We can't directly add ports to host model here
			// This would need to be handled by the calling code
		}
	}

	// Convert map to slice
	var hosts []*models.Host
	for _, host := range hostMap {
		hosts = append(hosts, host)
	}

	return hosts, nil
}

// GetPortsFromJSON extracts port information from masscan JSON output
func (s *Scanner) GetPortsFromJSON(jsonData []byte, hostID uuid.UUID) ([]*models.Port, error) {
	lines := strings.Split(strings.TrimSpace(string(jsonData)), "\n")

	var ports []*models.Port

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result MasscanResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			continue
		}

		for _, portInfo := range result.Ports {
			port := &models.Port{
				ID:        uuid.New(),
				HostID:    hostID,
				Number:    portInfo.Port,
				Protocol:  portInfo.Proto,
				State:     portInfo.Status,
				CreatedAt: time.Now(),
			}
			ports = append(ports, port)
		}
	}

	return ports, nil
}
