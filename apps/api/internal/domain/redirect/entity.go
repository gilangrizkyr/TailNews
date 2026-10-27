package redirect

import (
	"context"
	"time"
)

type Redirect struct {
	ID         uint64
	FromPath   string
	ToPath     string
	HTTPStatus int16
	HitCount   uint64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Repository interface {
	GetByFromPath(ctx context.Context, fromPath string) (*Redirect, error)
	CreateOrUpdate(ctx context.Context, fromPath string, toPath string, status int16) error
	IncrementHit(ctx context.Context, id uint64) error
	List(ctx context.Context, limit int32, offset int32) ([]*Redirect, error)
}
