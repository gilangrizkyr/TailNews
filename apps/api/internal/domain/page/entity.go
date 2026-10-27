package page

import (
	"context"
	"time"
)

type Page struct {
	ID          uint64
	Title       string
	Slug        string
	Content     string
	IsPublished bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	GetBySlug(ctx context.Context, slug string) (*Page, error)
	List(ctx context.Context) ([]*Page, error)
	Create(ctx context.Context, p *Page) (uint64, error)
	Update(ctx context.Context, p *Page) error
	Delete(ctx context.Context, id uint64) error
}
