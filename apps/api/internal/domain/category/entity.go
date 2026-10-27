package category

import (
	"context"
	"time"
)

type Category struct {
	ID          uint64
	Name        string
	Slug        string
	Description string
	ParentID    *uint64
	SortOrder   int32
	IsNav       bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id uint64) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	List(ctx context.Context) ([]*Category, error)
	ListNav(ctx context.Context) ([]*Category, error)
	Create(ctx context.Context, cat *Category) (uint64, error)
	Update(ctx context.Context, cat *Category) error
	Delete(ctx context.Context, id uint64) error
}
