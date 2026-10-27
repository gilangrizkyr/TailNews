-- +goose Up
-- Create redirects table
CREATE TABLE IF NOT EXISTS redirects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    from_path VARCHAR(500) NOT NULL UNIQUE,
    to_path VARCHAR(500) NOT NULL,
    http_status SMALLINT NOT NULL DEFAULT 301,
    hit_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_redirects_from ON redirects (from_path);

-- Create pages table for static dynamic pages (About, Privacy, etc.)
CREATE TABLE IF NOT EXISTS pages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL UNIQUE,
    content MEDIUMTEXT NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create settings table
CREATE TABLE IF NOT EXISTS settings (
    key_name VARCHAR(100) PRIMARY KEY,
    value JSON NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert initial essential settings
INSERT INTO settings (key_name, value) VALUES
('site_info', '{"name":"Portal Berita Modern","tagline":"Aktual, Cepat, dan Terpercaya","description":"Portal berita modern berkecepatan tinggi berskala nasional."}'),
('navigation', '{"primary":[{"name":"Home","url":"/"},{"name":"Politik","url":"/category/politik"},{"name":"Ekonomi","url":"/category/ekonomi"},{"name":"Teknologi","url":"/category/teknologi"},{"name":"Olahraga","url":"/category/olahraga"},{"name":"Hiburan","url":"/category/hiburan"}]}');

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NULL,
    action VARCHAR(50) NOT NULL,
    entity VARCHAR(50) NOT NULL,
    entity_id BIGINT UNSIGNED NOT NULL,
    diff_summary JSON NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_audit_logs_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_audit_logs_entity ON audit_logs (entity, entity_id);
CREATE INDEX idx_audit_logs_user ON audit_logs (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS redirects;
