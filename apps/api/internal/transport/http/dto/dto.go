package dto

import "time"

// ==========================================
// Auth DTOs
// ==========================================
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresIn   int64        `json:"expires_in"` // in seconds
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID        uint64    `json:"id"`
	RoleID    uint64    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// ==========================================
// Category DTOs
// ==========================================
type CategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Slug        string  `json:"slug" validate:"omitempty,max=120"`
	Description string  `json:"description" validate:"max=500"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int32   `json:"sort_order"`
	IsNav       bool    `json:"is_nav"`
}

type CategoryResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	ParentID    *uint64   `json:"parent_id"`
	SortOrder   int32     `json:"sort_order"`
	IsNav       bool      `json:"is_nav"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ==========================================
// Tag DTOs
// ==========================================
type TagRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"omitempty,max=120"`
}

type TagResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==========================================
// Author DTOs
// ==========================================
type AuthorRequest struct {
	UserID        *uint64           `json:"user_id"`
	Name          string            `json:"name" validate:"required,min=2,max=100"`
	Slug          string            `json:"slug" validate:"omitempty,max=120"`
	Bio           string            `json:"bio" validate:"max=1000"`
	AvatarMediaID *uint64           `json:"avatar_media_id"`
	SocialLinks   map[string]string `json:"social_links"`
}

type AuthorResponse struct {
	ID            uint64            `json:"id"`
	UserID        *uint64           `json:"user_id"`
	Name          string            `json:"name"`
	Slug          string            `json:"slug"`
	Bio           string            `json:"bio"`
	AvatarMediaID *uint64           `json:"avatar_media_id"`
	SocialLinks   map[string]string `json:"social_links"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// ==========================================
// Media DTOs
// ==========================================
type MediaResponse struct {
	ID           uint64            `json:"id"`
	Filename     string            `json:"filename"`
	FilePath     string            `json:"file_path"`
	MimeType     string            `json:"mime_type"`
	SizeBytes    uint64            `json:"size_bytes"`
	Width        uint32            `json:"width"`
	Height       uint32            `json:"height"`
	AltText      string            `json:"alt_text"`
	Caption      string            `json:"caption"`
	Photographer string            `json:"photographer"`
	Variants     map[string]string `json:"variants"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ==========================================
// Article DTOs
// ==========================================
type ArticleUpsertRequest struct {
	Title          string     `json:"title" validate:"required,min=5,max=255"`
	Slug           string     `json:"slug" validate:"omitempty,max=280"`
	Excerpt        string     `json:"excerpt" validate:"required,min=10,max=500"`
	Body           string     `json:"body" validate:"required"`
	CoverMediaID   *uint64    `json:"cover_media_id"`
	CategoryID     uint64     `json:"category_id" validate:"required"`
	AuthorID       uint64     `json:"author_id" validate:"required"`
	Status         string     `json:"status" validate:"required,oneof=draft review scheduled published"`
	IsHeadline     bool       `json:"is_headline"`
	IsBreaking     bool       `json:"is_breaking"`
	PublishedAt    *time.Time `json:"published_at"`
	SEOTitle       string     `json:"seo_title" validate:"max=255"`
	SEODescription string     `json:"seo_description" validate:"max=300"`
	TagIDs         []uint64   `json:"tag_ids"`
}

type ArticleResponse struct {
	ID             uint64            `json:"id"`
	Title          string            `json:"title"`
	Slug           string            `json:"slug"`
	Excerpt        string            `json:"excerpt"`
	Body           string            `json:"body"`
	CoverMediaID   *uint64           `json:"cover_media_id"`
	CoverURL       string            `json:"cover_url,omitempty"`
	CategoryID     uint64            `json:"category_id"`
	CategoryName   string            `json:"category_name,omitempty"`
	CategorySlug   string            `json:"category_slug,omitempty"`
	AuthorID       uint64            `json:"author_id"`
	AuthorName     string            `json:"author_name,omitempty"`
	AuthorSlug     string            `json:"author_slug,omitempty"`
	CreatedBy      uint64            `json:"created_by"`
	Status         string            `json:"status"`
	IsHeadline     bool              `json:"is_headline"`
	IsBreaking     bool              `json:"is_breaking"`
	ViewCount      uint64            `json:"view_count"`
	PublishedAt    *time.Time        `json:"published_at"`
	SEOTitle       string            `json:"seo_title"`
	SEODescription string            `json:"seo_description"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Tags           []TagResponse     `json:"tags,omitempty"`
	RevisionsCount int               `json:"revisions_count,omitempty"`
}

type ArticleRevisionResponse struct {
	ID        uint64    `json:"id"`
	ArticleID uint64    `json:"article_id"`
	UserID    uint64    `json:"user_id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Excerpt   string    `json:"excerpt"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// ==========================================
// Redirect DTOs
// ==========================================
type RedirectRequest struct {
	FromPath   string `json:"from_path" validate:"required,startswith=/"`
	ToPath     string `json:"to_path" validate:"required,startswith=/"`
	HTTPStatus int16  `json:"http_status" validate:"required,oneof=301 302"`
}

type RedirectResponse struct {
	ID         uint64    `json:"id"`
	FromPath   string    `json:"from_path"`
	ToPath     string    `json:"to_path"`
	HTTPStatus int16     `json:"http_status"`
	HitCount   uint64    `json:"hit_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ==========================================
// Dashboard Stats DTO
// ==========================================
type DashboardStatsResponse struct {
	TotalArticles     int64 `json:"total_articles"`
	PublishedArticles int64 `json:"published_articles"`
	DraftArticles     int64 `json:"draft_articles"`
	ScheduledArticles int64 `json:"scheduled_articles"`
	TotalViews        int64 `json:"total_views"`
	TotalCategories   int64 `json:"total_categories"`
	TotalAuthors      int64 `json:"total_authors"`
	TotalMedia        int64 `json:"total_media"`
}
