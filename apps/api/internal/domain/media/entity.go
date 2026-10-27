package media

import (
	"context"
	"time"
)

type Media struct {
	ID           uint64
	Filename     string
	FilePath     string
	MimeType     string
	SizeBytes    uint64
	Width        uint32
	Height       uint32
	AltText      string
	Caption      string
	Photographer string
	Variants     map[string]string // e.g., "thumbnail": "/media/thumb_xyz.webp", "medium": ...
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id uint64) (*Media, error)
	Create(ctx context.Context, m *Media) (uint64, error)
	List(ctx context.Context, limit int32, offset int32) ([]*Media, error)
	Delete(ctx context.Context, id uint64) error
}
