-- +goose Up
-- Create articles table
CREATE TABLE IF NOT EXISTS articles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(280) NOT NULL UNIQUE,
    excerpt VARCHAR(500) NOT NULL,
    body MEDIUMTEXT NOT NULL,
    cover_media_id BIGINT UNSIGNED NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    author_id BIGINT UNSIGNED NOT NULL,
    created_by BIGINT UNSIGNED NOT NULL,
    status ENUM('draft', 'review', 'scheduled', 'published') NOT NULL DEFAULT 'draft',
    is_headline BOOLEAN NOT NULL DEFAULT FALSE,
    is_breaking BOOLEAN NOT NULL DEFAULT FALSE,
    view_count BIGINT UNSIGNED NOT NULL DEFAULT 0,
    published_at DATETIME NULL,
    seo_title VARCHAR(255) NULL,
    seo_description VARCHAR(300) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    CONSTRAINT fk_articles_cover FOREIGN KEY (cover_media_id) REFERENCES media (id) ON DELETE SET NULL,
    CONSTRAINT fk_articles_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE RESTRICT,
    CONSTRAINT fk_articles_author FOREIGN KEY (author_id) REFERENCES authors (id) ON DELETE RESTRICT,
    CONSTRAINT fk_articles_creator FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Keyset cursor pagination & query optimization indexes
CREATE INDEX idx_articles_cursor ON articles (status, deleted_at, published_at DESC, id DESC);
CREATE INDEX idx_articles_category_cursor ON articles (category_id, status, deleted_at, published_at DESC, id DESC);
CREATE INDEX idx_articles_author_cursor ON articles (author_id, status, deleted_at, published_at DESC, id DESC);
CREATE INDEX idx_articles_headline ON articles (is_headline, status, deleted_at, published_at DESC, id DESC);
CREATE INDEX idx_articles_breaking ON articles (is_breaking, status, deleted_at, published_at DESC, id DESC);
CREATE INDEX idx_articles_scheduled ON articles (status, published_at);
CREATE FULLTEXT INDEX idx_articles_search ON articles (title, excerpt);

-- Create article_tag pivot table
CREATE TABLE IF NOT EXISTS article_tag (
    article_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (article_id, tag_id),
    CONSTRAINT fk_article_tag_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,
    CONSTRAINT fk_article_tag_tag FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_article_tag_tag ON article_tag (tag_id);

-- +goose Down
DROP TABLE IF EXISTS article_tag;
DROP TABLE IF EXISTS articles;
