package user

import (
	"context"
	"time"
)

type Role struct {
	ID          uint64
	Name        string
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID                  uint64
	RoleID              uint64
	Name                string
	Email               string
	PasswordHash        string
	IsActive            bool
	FailedLoginAttempts uint32
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (uint64, error)
	UpdateFailedLogin(ctx context.Context, id uint64) error
	ResetFailedLogin(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*User, error)
	GetRoleByID(ctx context.Context, roleID uint64) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
}
