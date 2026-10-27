package author

import (
	"context"
	"time"
)

type Author struct {
	ID            uint64
	UserID        *uint64
	Name          string
	Slug          string
	Bio           string
	AvatarMediaID *uint64
	SocialLinks   map[string]string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id uint64) (*Author, error)
	GetBySlug(ctx context.Context, slug string) (*Author, error)
	List(ctx context.Context) ([]*Author, error)
	Create(ctx context.Context, a *Author) (uint64, error)
	Update(ctx context.Context, a *Author) error
}
