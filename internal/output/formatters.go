package output

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"

	"github.com/netrecon/toolkit/internal/models"
	"github.com/netrecon/toolkit/internal/scanner"
)

// Formatter defines the interface for output formatters
type Formatter interface {
	Format(result *scanner.ScanResult) ([]byte, error)
	GetMimeType() string
	GetFileExtension() string
}

// JSONFormatter formats output as JSON
type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *scanner.ScanResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}

func (f *JSONFormatter) GetMimeType() string {
	return "application/json"
}

func (f *JSONFormatter) GetFileExtension() string {
	return "json"
}

// XMLFormatter formats output as XML
type XMLFormatter struct{}

func (f *XMLFormatter) Format(result *scanner.ScanResult) ([]byte, error) {
	type XMLScanResult struct {
		XMLName xml.Name `xml:"scanResult"`
		*scanner.ScanResult
	}

	xmlResult := &XMLScanResult{ScanResult: result}
	return xml.MarshalIndent(xmlResult, "", "  ")
}

func (f *XMLFormatter) GetMimeType() string {
	return "application/xml"
}

func (f *XMLFormatter) GetFileExtension() string {
	return "xml"
}

// CSVFormatter formats output as CSV
type CSVFormatter struct{}

func (f *CSVFormatter) Format(result *scanner.ScanResult) ([]byte, error) {
	var output []byte

	// Create CSV writer to a buffer would be better, but for simplicity:
	records := [][]string{
		{"Target", "Scanner", "Status", "Start Time", "End Time", "Duration", "Host Count"},
		{result.Target, result.Scanner, result.Status, result.StartTime, result.EndTime, result.Duration, fmt.Sprintf("%d", len(result.Hosts))},
	}

	// Add host information
	if len(result.Hosts) > 0 {
		records = append(records, []string{}) // Empty line
		records = append(records, []string{"IP Address", "Hostname", "Status", "OS", "OS Confidence"})

		for _, host := range result.Hosts {
			records = append(records, []string{
				host.IPAddress,
				host.Hostname,
				host.Status,
				host.OS,
				fmt.Sprintf("%d", host.OSConfidence),
			})
		}
	}

	// Convert to CSV format (simplified)
	csvContent := ""
	for _, record := range records {
		for i, field := range record {
			if i > 0 {
				csvContent += ","
			}
			csvContent += fmt.Sprintf("\"%s\"", field)
		}
		csvContent += "\n"
	}

	return []byte(csvContent), nil
}

func (f *CSVFormatter) GetMimeType() string {
	return "text/csv"
}

func (f *CSVFormatter) GetFileExtension() string {
	return "csv"
}

// HTMLFormatter formats output as HTML report
type HTMLFormatter struct{}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Network Reconnaissance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .section { margin-bottom: 30px; }
        .host { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .port { background-color: #f9f9f9; margin: 5px 0; padding: 10px; border-left: 4px solid #007cba; }
        .status-up { color: green; font-weight: bold; }
        .status-down { color: red; font-weight: bold; }
        .status-filtered { color: orange; font-weight: bold; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .error { color: red; background-color: #ffe6e6; padding: 10px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Network Reconnaissance Report</h1>
        <p><strong>Target:</strong> {{.Target}}</p>
        <p><strong>Scanner:</strong> {{.Scanner}}</p>
        <p><strong>Status:</strong> <span class="status-{{.Status}}">{{.Status}}</span></p>
        <p><strong>Start Time:</strong> {{.StartTime}}</p>
        <p><strong>End Time:</strong> {{.EndTime}}</p>
        <p><strong>Duration:</strong> {{.Duration}}</p>
        <p><strong>Hosts Found:</strong> {{len .Hosts}}</p>
    </div>

    {{if .Error}}
    <div class="error">
        <h3>Errors</h3>
        <pre>{{.Error}}</pre>
    </div>
    {{end}}

    {{if .Hosts}}
    <div class="section">
        <h2>Discovered Hosts</h2>
        {{range .Hosts}}
        <div class="host">
            <h3>Host: {{.IPAddress}} {{if .Hostname}}({{.Hostname}}){{end}}</h3>
            <p><strong>Status:</strong> <span class="status-{{.Status}}">{{.Status}}</span></p>
            {{if .OS}}<p><strong>OS:</strong> {{.OS}} ({{.OSConfidence}}% confidence)</p>{{end}}
        </div>
        {{end}}
    </div>
    {{end}}

    <div class="section">
        <h2>Raw Output</h2>
        <pre style="background-color: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto;">{{.RawOutput}}</pre>
    </div>

    <div class="section">
        <p><em>Report generated on {{.Timestamp}}</em></p>
    </div>
</body>
</html>
`

func (f *HTMLFormatter) Format(result *scanner.ScanResult) ([]byte, error) {
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Add timestamp to result
	data := struct {
		*scanner.ScanResult
		Timestamp string
	}{
		ScanResult: result,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	var output []byte
	buf := &bytesBuffer{data: &output}

	if err := tmpl.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return output, nil
}

func (f *HTMLFormatter) GetMimeType() string {
	return "text/html"
}

func (f *HTMLFormatter) GetFileExtension() string {
	return "html"
}

// bytesBuffer implements io.Writer for []byte
type bytesBuffer struct {
	data *[]byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	*b.data = append(*b.data, p...)
	return len(p), nil
}

// FormatterManager manages output formatters
type FormatterManager struct {
	formatters map[string]Formatter
}

// NewFormatterManager creates a new formatter manager
func NewFormatterManager() *FormatterManager {
	fm := &FormatterManager{
		formatters: make(map[string]Formatter),
	}

	// Register default formatters
	fm.RegisterFormatter("json", &JSONFormatter{})
	fm.RegisterFormatter("xml", &XMLFormatter{})
	fm.RegisterFormatter("csv", &CSVFormatter{})
	fm.RegisterFormatter("html", &HTMLFormatter{})

	return fm
}

// RegisterFormatter registers a new formatter
func (fm *FormatterManager) RegisterFormatter(name string, formatter Formatter) {
	fm.formatters[name] = formatter
}

// GetFormatter returns a formatter by name
func (fm *FormatterManager) GetFormatter(name string) (Formatter, bool) {
	formatter, exists := fm.formatters[name]
	return formatter, exists
}

// ListFormatters returns all available formatter names
func (fm *FormatterManager) ListFormatters() []string {
	var names []string
	for name := range fm.formatters {
		names = append(names, name)
	}
	return names
}

// FormatAndSave formats scan results and saves to file
func (fm *FormatterManager) FormatAndSave(result *scanner.ScanResult, format string, filename string) error {
	formatter, exists := fm.GetFormatter(format)
	if !exists {
		return fmt.Errorf("formatter '%s' not available. Available formatters: %v", format, fm.ListFormatters())
	}

	data, err := formatter.Format(result)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
