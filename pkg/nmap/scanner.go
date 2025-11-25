package nmap

import (
	"context"
	"encoding/xml"
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

// Scanner implements the nmap scanner
type Scanner struct {
	path string
}

// NewScanner creates a new nmap scanner
func NewScanner() (*Scanner, error) {
	// Check if nmap is installed
	path, err := exec.LookPath("nmap")
	if err != nil {
		return nil, fmt.Errorf("nmap not found in PATH: %w", err)
	}

	return &Scanner{path: path}, nil
}

// GetName returns the scanner name
func (s *Scanner) GetName() string {
	return "nmap"
}

// ValidateConfig validates the nmap configuration
func (s *Scanner) ValidateConfig(config *scanner.ScanConfig) error {
	if config.Ports != "" {
		// Validate port format
		portRegex := regexp.MustCompile(`^(\d+(-\d+)?)(,\d+(-\d+)?)*$`)
		if !portRegex.MatchString(config.Ports) {
			return fmt.Errorf("invalid port format: %s", config.Ports)
		}
	}

	if config.Timing != "" {
		// Validate timing template (0-5)
		if timing, err := strconv.Atoi(config.Timing); err != nil || timing < 0 || timing > 5 {
			return fmt.Errorf("invalid timing template: %s (must be 0-5)", config.Timing)
		}
	}

	return nil
}

// Scan performs an nmap scan
func (s *Scanner) Scan(ctx context.Context, target string, config *scanner.ScanConfig) (*scanner.ScanResult, error) {
	if err := s.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	startTime := time.Now()

	// Build nmap command
	args := []string{"-oX", "-"} // Output XML to stdout

	// Add port specification
	if config.Ports != "" {
		args = append(args, "-p", config.Ports)
	}

	// Add timing template
	if config.Timing != "" {
		args = append(args, "-T"+config.Timing)
	}

	// Add service detection
	args = append(args, "-sV")

	// Add OS detection
	args = append(args, "-O")

	// Add additional arguments
	if config.Arguments != "" {
		additionalArgs := strings.Fields(config.Arguments)
		args = append(args, additionalArgs...)
	}

	// Add target
	args = append(args, target)

	// Execute nmap command
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

	// Parse XML output
	hosts, parseErr := s.parseNmapXML(output)
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

// NmapRun represents the root XML element
type NmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []NmapHost `xml:"host"`
}

// NmapHost represents a host in the XML output
type NmapHost struct {
	XMLName   xml.Name      `xml:"host"`
	Status    NmapStatus    `xml:"status"`
	Address   []NmapAddress `xml:"address"`
	Hostnames NmapHostnames `xml:"hostnames"`
	Ports     NmapPorts     `xml:"ports"`
	OS        NmapOS        `xml:"os"`
}

// NmapStatus represents host status
type NmapStatus struct {
	State string `xml:"state,attr"`
}

// NmapAddress represents an address
type NmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

// NmapHostnames contains hostnames
type NmapHostnames struct {
	Hostnames []NmapHostname `xml:"hostname"`
}

// NmapHostname represents a hostname
type NmapHostname struct {
	Name string `xml:"name,attr"`
}

// NmapPorts contains port information
type NmapPorts struct {
	Ports []NmapPort `xml:"port"`
}

// NmapPort represents a port
type NmapPort struct {
	Protocol string      `xml:"protocol,attr"`
	PortID   int         `xml:"portid,attr"`
	State    NmapState   `xml:"state"`
	Service  NmapService `xml:"service"`
}

// NmapState represents port state
type NmapState struct {
	State string `xml:"state,attr"`
}

// NmapService represents service information
type NmapService struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
	Info    string `xml:"extrainfo,attr"`
}

// NmapOS represents OS information
type NmapOS struct {
	OSMatches []NmapOSMatch `xml:"osmatch"`
}

// NmapOSMatch represents an OS match
type NmapOSMatch struct {
	Name     string `xml:"name,attr"`
	Accuracy int    `xml:"accuracy,attr"`
}

// parseNmapXML parses nmap XML output
func (s *Scanner) parseNmapXML(xmlData []byte) ([]*models.Host, error) {
	var nmapRun NmapRun
	if err := xml.Unmarshal(xmlData, &nmapRun); err != nil {
		return nil, fmt.Errorf("failed to parse nmap XML: %w", err)
	}

	var hosts []*models.Host

	for _, nmapHost := range nmapRun.Hosts {
		host := &models.Host{
			ID:        uuid.New(),
			Status:    nmapHost.Status.State,
			CreatedAt: time.Now(),
		}

		// Get IP address
		for _, addr := range nmapHost.Address {
			if addr.AddrType == "ipv4" {
				host.IPAddress = addr.Addr
				break
			}
		}

		// Get hostname
		if len(nmapHost.Hostnames.Hostnames) > 0 {
			host.Hostname = nmapHost.Hostnames.Hostnames[0].Name
		}

		// Get OS information
		if len(nmapHost.OS.OSMatches) > 0 {
			osMatch := nmapHost.OS.OSMatches[0]
			host.OS = osMatch.Name
			host.OSConfidence = osMatch.Accuracy
		}

		hosts = append(hosts, host)
	}

	return hosts, nil
}
