-- Migration: 001_create_initial_tables.down.sql
-- Drop all tables in reverse order

DROP TRIGGER IF EXISTS update_scan_targets_updated_at ON scan_targets;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_vulnerabilities_severity;
DROP INDEX IF EXISTS idx_vulnerabilities_port_id;
DROP INDEX IF EXISTS idx_ports_number_protocol;
DROP INDEX IF EXISTS idx_ports_host_id;
DROP INDEX IF EXISTS idx_hosts_ip_address;
DROP INDEX IF EXISTS idx_hosts_scan_id;
DROP INDEX IF EXISTS idx_scan_results_status;
DROP INDEX IF EXISTS idx_scan_results_target_id;
DROP INDEX IF EXISTS idx_scan_targets_type;

DROP TABLE IF EXISTS scan_configurations;
DROP TABLE IF EXISTS vulnerabilities;
DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS hosts;
DROP TABLE IF EXISTS scan_results;
DROP TABLE IF EXISTS scan_targets;