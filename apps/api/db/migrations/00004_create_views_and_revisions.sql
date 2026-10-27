-- +goose Up
-- Create article_views_daily table for trending analytics
CREATE TABLE IF NOT EXISTS article_views_daily (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    article_id BIGINT UNSIGNED NOT NULL,
    view_date DATE NOT NULL,
    views INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT uk_article_views_date UNIQUE (article_id, view_date),
    CONSTRAINT fk_article_views_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_article_views_trending ON article_views_daily (view_date, views DESC);

-- Create article_revisions table for editorial history
CREATE TABLE IF NOT EXISTS article_revisions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    article_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(280) NOT NULL,
    excerpt VARCHAR(500) NOT NULL,
    body MEDIUMTEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_revisions_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,
    CONSTRAINT fk_revisions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_revisions_article ON article_revisions (article_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS article_revisions;
DROP TABLE IF EXISTS article_views_daily;
