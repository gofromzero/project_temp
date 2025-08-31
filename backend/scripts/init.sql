-- Multi-Tenant Admin Database Initialization
CREATE DATABASE IF NOT EXISTS multi_tenant_admin;
USE multi_tenant_admin;

-- Basic health check table for initial setup
CREATE TABLE IF NOT EXISTS system_health (
    id INT PRIMARY KEY AUTO_INCREMENT,
    component VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial health check data
INSERT INTO system_health (component, status) VALUES 
('database', 'healthy'),
('redis', 'healthy')
ON DUPLICATE KEY UPDATE 
    status = VALUES(status), 
    checked_at = CURRENT_TIMESTAMP;