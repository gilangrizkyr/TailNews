package setting

import (
	"context"
	"time"
)

type Setting struct {
	KeyName   string
	Value     []byte
	UpdatedAt time.Time
}

type Repository interface {
	Get(ctx context.Context, keyName string) (*Setting, error)
	Upsert(ctx context.Context, keyName string, value []byte) error
	List(ctx context.Context) ([]*Setting, error)
}
