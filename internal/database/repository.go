package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/netrecon/toolkit/internal/models"
)

// Repository provides database operations
type Repository struct {
	db *DB
}

// NewRepository creates a new repository instance
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// ScanTarget operations
func (r *Repository) CreateScanTarget(target *models.ScanTarget) error {
	target.ID = uuid.New()
	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()

	query := `
		INSERT INTO scan_targets (id, target, type, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query, target.ID, target.Target, target.Type, target.Description, target.CreatedAt, target.UpdatedAt)
	return err
}

func (r *Repository) GetScanTarget(id uuid.UUID) (*models.ScanTarget, error) {
	target := &models.ScanTarget{}
	query := `
		SELECT id, target, type, description, created_at, updated_at
		FROM scan_targets WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&target.ID, &target.Target, &target.Type, &target.Description,
		&target.CreatedAt, &target.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return target, nil
}

func (r *Repository) ListScanTargets() ([]*models.ScanTarget, error) {
	query := `
		SELECT id, target, type, description, created_at, updated_at
		FROM scan_targets ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []*models.ScanTarget
	for rows.Next() {
		target := &models.ScanTarget{}
		err := rows.Scan(&target.ID, &target.Target, &target.Type, &target.Description,
			&target.CreatedAt, &target.UpdatedAt)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

// ScanResult operations
func (r *Repository) CreateScanResult(result *models.ScanResult) error {
	result.ID = uuid.New()
	result.CreatedAt = time.Now()

	query := `
		INSERT INTO scan_results (id, target_id, scan_type, status, start_time, end_time, raw_output, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query, result.ID, result.TargetID, result.ScanType, result.Status,
		result.StartTime, result.EndTime, result.RawOutput, result.CreatedAt)
	return err
}

func (r *Repository) UpdateScanResult(result *models.ScanResult) error {
	query := `
		UPDATE scan_results 
		SET status = $2, end_time = $3, raw_output = $4
		WHERE id = $1`

	_, err := r.db.Exec(query, result.ID, result.Status, result.EndTime, result.RawOutput)
	return err
}

func (r *Repository) GetScanResult(id uuid.UUID) (*models.ScanResult, error) {
	result := &models.ScanResult{}
	query := `
		SELECT id, target_id, scan_type, status, start_time, end_time, raw_output, created_at
		FROM scan_results WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&result.ID, &result.TargetID, &result.ScanType, &result.Status,
		&result.StartTime, &result.EndTime, &result.RawOutput, &result.CreatedAt)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) ListScanResults(targetID uuid.UUID) ([]*models.ScanResult, error) {
	query := `
		SELECT id, target_id, scan_type, status, start_time, end_time, raw_output, created_at
		FROM scan_results WHERE target_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.ScanResult
	for rows.Next() {
		result := &models.ScanResult{}
		err := rows.Scan(&result.ID, &result.TargetID, &result.ScanType, &result.Status,
			&result.StartTime, &result.EndTime, &result.RawOutput, &result.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Host operations
func (r *Repository) CreateHost(host *models.Host) error {
	host.ID = uuid.New()
	host.CreatedAt = time.Now()

	query := `
		INSERT INTO hosts (id, scan_id, ip_address, hostname, status, os, os_confidence, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query, host.ID, host.ScanID, host.IPAddress, host.Hostname,
		host.Status, host.OS, host.OSConfidence, host.CreatedAt)
	return err
}

func (r *Repository) GetHostsByScanID(scanID uuid.UUID) ([]*models.Host, error) {
	query := `
		SELECT id, scan_id, ip_address, hostname, status, os, os_confidence, created_at
		FROM hosts WHERE scan_id = $1 ORDER BY ip_address`

	rows, err := r.db.Query(query, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*models.Host
	for rows.Next() {
		host := &models.Host{}
		err := rows.Scan(&host.ID, &host.ScanID, &host.IPAddress, &host.Hostname,
			&host.Status, &host.OS, &host.OSConfidence, &host.CreatedAt)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

// Port operations
func (r *Repository) CreatePort(port *models.Port) error {
	port.ID = uuid.New()
	port.CreatedAt = time.Now()

	query := `
		INSERT INTO ports (id, host_id, number, protocol, state, service, version, product, extra_info, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(query, port.ID, port.HostID, port.Number, port.Protocol,
		port.State, port.Service, port.Version, port.Product, port.ExtraInfo, port.CreatedAt)
	return err
}

func (r *Repository) GetPortsByHostID(hostID uuid.UUID) ([]*models.Port, error) {
	query := `
		SELECT id, host_id, number, protocol, state, service, version, product, extra_info, created_at
		FROM ports WHERE host_id = $1 ORDER BY number`

	rows, err := r.db.Query(query, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []*models.Port
	for rows.Next() {
		port := &models.Port{}
		err := rows.Scan(&port.ID, &port.HostID, &port.Number, &port.Protocol,
			&port.State, &port.Service, &port.Version, &port.Product, &port.ExtraInfo, &port.CreatedAt)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	return ports, nil
}
