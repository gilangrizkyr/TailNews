package article

import (
	"context"
	"time"

	"news-portal/apps/api/internal/domain/tag"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusReview    Status = "review"
	StatusScheduled Status = "scheduled"
	StatusPublished Status = "published"
)

type Article struct {
	ID             uint64
	Title          string
	Slug           string
	Excerpt        string
	Body           string
	CoverMediaID   *uint64
	CategoryID     uint64
	AuthorID       uint64
	CreatedBy      uint64
	Status         Status
	IsHeadline     bool
	IsBreaking     bool
	ViewCount      uint64
	PublishedAt    *time.Time
	SEOTitle       string
	SEODescription string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time

	// Enriched fields for presentation
	CategoryName string
	CategorySlug string
	AuthorName   string
	AuthorSlug   string
	CoverURL     string
	Tags         []*tag.Tag
}

type ArticleSummary struct {
	ID           uint64
	Title        string
	Slug         string
	Excerpt      string
	CoverMediaID *uint64
	CategoryID   uint64
	AuthorID     uint64
	IsHeadline   bool
	IsBreaking   bool
	ViewCount    uint64
	PublishedAt  *time.Time
	CategoryName string
	CategorySlug string
	AuthorName   string
	AuthorSlug   string
	CoverURL     string
}

type Cursor struct {
	PublishedAt time.Time
	ID          uint64
}

type ArticleRevision struct {
	ID        uint64
	ArticleID uint64
	UserID    uint64
	Title     string
	Slug      string
	Excerpt   string
	Body      string
	CreatedAt time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id uint64) (*Article, error)
	GetBySlug(ctx context.Context, slug string) (*Article, error)
	Create(ctx context.Context, a *Article) (uint64, error)
	Update(ctx context.Context, a *Article) error
	SoftDelete(ctx context.Context, id uint64) error
	PermanentDelete(ctx context.Context, id uint64) error

	// Keyset Cursor Pagination
	ListCursor(ctx context.Context, cursor *Cursor, limit int32) ([]*ArticleSummary, error)
	ListByCategoryCursor(ctx context.Context, categoryID uint64, cursor *Cursor, limit int32) ([]*ArticleSummary, error)
	ListByAuthorCursor(ctx context.Context, authorID uint64, cursor *Cursor, limit int32) ([]*ArticleSummary, error)
	ListByTagCursor(ctx context.Context, tagID uint64, cursor *Cursor, limit int32) ([]*ArticleSummary, error)

	// Feeds
	ListHeadlines(ctx context.Context, limit int32) ([]*ArticleSummary, error)
	ListBreaking(ctx context.Context, limit int32) ([]*ArticleSummary, error)
	SearchFulltext(ctx context.Context, query string, limit int32) ([]*ArticleSummary, error)

	// Views and Scheduled Jobs
	IncrementViews(ctx context.Context, id uint64, count int64) error
	GetScheduledToPublish(ctx context.Context, now time.Time) ([]*ArticleSummary, error)
	PublishScheduled(ctx context.Context, id uint64) error

	// Tags & Revisions
	SetTags(ctx context.Context, articleID uint64, tagIDs []uint64) error
	GetTags(ctx context.Context, articleID uint64) ([]*tag.Tag, error)
	CreateRevision(ctx context.Context, rev *ArticleRevision) (uint64, error)
	ListRevisions(ctx context.Context, articleID uint64) ([]*ArticleRevision, error)

	// Admin
	ListAdmin(ctx context.Context, limit int32, offset int32) ([]*Article, error)
}
