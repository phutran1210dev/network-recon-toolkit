-- Migration: 001_create_initial_tables.up.sql
-- Create initial database schema for network reconnaissance toolkit

-- Create scan_targets table
CREATE TABLE IF NOT EXISTS scan_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('ip', 'range', 'domain')),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create scan_results table
CREATE TABLE IF NOT EXISTS scan_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id UUID NOT NULL REFERENCES scan_targets(id) ON DELETE CASCADE,
    scan_type VARCHAR(50) NOT NULL CHECK (scan_type IN ('nmap', 'masscan')),
    status VARCHAR(50) NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    raw_output TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create hosts table
CREATE TABLE IF NOT EXISTS hosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL REFERENCES scan_results(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    hostname VARCHAR(255),
    status VARCHAR(50) NOT NULL CHECK (status IN ('up', 'down', 'filtered')),
    os VARCHAR(255),
    os_confidence INTEGER DEFAULT 0 CHECK (os_confidence >= 0 AND os_confidence <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create ports table
CREATE TABLE IF NOT EXISTS ports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    number INTEGER NOT NULL CHECK (number > 0 AND number <= 65535),
    protocol VARCHAR(10) NOT NULL CHECK (protocol IN ('tcp', 'udp')),
    state VARCHAR(20) NOT NULL CHECK (state IN ('open', 'closed', 'filtered')),
    service VARCHAR(100),
    version VARCHAR(255),
    product VARCHAR(255),
    extra_info TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create vulnerabilities table
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    port_id UUID NOT NULL REFERENCES ports(id) ON DELETE CASCADE,
    cve VARCHAR(50),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT NOT NULL,
    solution TEXT,
    reference_links TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create scan_configurations table
CREATE TABLE IF NOT EXISTS scan_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    scanner VARCHAR(50) NOT NULL CHECK (scanner IN ('nmap', 'masscan')),
    ports VARCHAR(255),
    arguments TEXT,
    timing VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_scan_targets_type ON scan_targets(type);
CREATE INDEX idx_scan_results_target_id ON scan_results(target_id);
CREATE INDEX idx_scan_results_status ON scan_results(status);
CREATE INDEX idx_hosts_scan_id ON hosts(scan_id);
CREATE INDEX idx_hosts_ip_address ON hosts(ip_address);
CREATE INDEX idx_ports_host_id ON ports(host_id);
CREATE INDEX idx_ports_number_protocol ON ports(number, protocol);
CREATE INDEX idx_vulnerabilities_port_id ON vulnerabilities(port_id);
CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for scan_targets updated_at
CREATE TRIGGER update_scan_targets_updated_at 
    BEFORE UPDATE ON scan_targets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();