package audit

import (
	"context"
	"time"
)

type AuditLog struct {
	ID          uint64
	UserID      *uint64
	Action      string
	Entity      string
	EntityID    uint64
	DiffSummary []byte
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}

type Repository interface {
	Create(ctx context.Context, log *AuditLog) (uint64, error)
	List(ctx context.Context, limit int32, offset int32) ([]*AuditLog, error)
}
