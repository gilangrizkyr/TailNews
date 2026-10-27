package tag

import (
	"context"
	"time"
)

type Tag struct {
	ID        uint64
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id uint64) (*Tag, error)
	GetBySlug(ctx context.Context, slug string) (*Tag, error)
	List(ctx context.Context) ([]*Tag, error)
	Create(ctx context.Context, t *Tag) (uint64, error)
	Delete(ctx context.Context, id uint64) error
	GetOrCreate(ctx context.Context, name string, slug string) (*Tag, error)
}
