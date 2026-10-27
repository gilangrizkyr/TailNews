package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/domain/user"
	"news-portal/apps/api/internal/pkg/auth"
	domainerrors "news-portal/apps/api/internal/pkg/errors"
	"news-portal/apps/api/internal/transport/http/dto"
)

var (
	ErrAccountLocked     = errors.New("account is temporarily locked due to multiple failed login attempts")
	ErrAccountDeactivated = errors.New("account has been deactivated by administrator")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthUsecase struct {
	userRepo user.Repository
	cfg      *config.Config
}

func NewAuthUsecase(userRepo user.Repository, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, email string, password string) (*dto.AuthResponse, string, error) {
	usr, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !usr.IsActive {
		return nil, "", ErrAccountDeactivated
	}

	// Check account lock
	if usr.LockedUntil != nil && usr.LockedUntil.After(time.Now().UTC()) {
		return nil, "", ErrAccountLocked
	}

	// Verify password
	valid, err := auth.VerifyPassword(password, usr.PasswordHash)
	if err != nil || !valid {
		_ = u.userRepo.UpdateFailedLogin(ctx, usr.ID)
		return nil, "", ErrInvalidCredentials
	}

	// Reset failed attempts upon successful login
	if usr.FailedLoginAttempts > 0 {
		_ = u.userRepo.ResetFailedLogin(ctx, usr.ID)
	}

	role, err := u.userRepo.GetRoleByID(ctx, usr.RoleID)
	roleName := "writer"
	if err == nil && role != nil {
		roleName = role.Name
	}

	// Generate access token (15 mins) & refresh token (7 days)
	accessTTL := time.Duration(u.cfg.JWTAccessTTLMinutes) * time.Minute
	refreshTTL := time.Duration(u.cfg.JWTRefreshTTLDays) * 24 * time.Hour

	accessToken, err := auth.GenerateAccessToken(u.cfg.JWTSecret, usr, roleName, accessTTL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(u.cfg.JWTSecret, usr.ID, refreshTTL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(accessTTL.Seconds()),
		User: dto.UserResponse{
			ID:        usr.ID,
			RoleID:    usr.RoleID,
			RoleName:  roleName,
			Name:      usr.Name,
			Email:     usr.Email,
			IsActive:  usr.IsActive,
			CreatedAt: usr.CreatedAt,
		},
	}, refreshToken, nil
}

func (u *AuthUsecase) Refresh(ctx context.Context, refreshTokenStr string) (*dto.AuthResponse, error) {
	userID, err := auth.ValidateRefreshToken(u.cfg.JWTSecret, refreshTokenStr)
	if err != nil {
		return nil, domainerrors.ErrUnauthorized
	}

	usr, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || !usr.IsActive {
		return nil, domainerrors.ErrUnauthorized
	}

	role, err := u.userRepo.GetRoleByID(ctx, usr.RoleID)
	roleName := "writer"
	if err == nil && role != nil {
		roleName = role.Name
	}

	accessTTL := time.Duration(u.cfg.JWTAccessTTLMinutes) * time.Minute
	accessToken, err := auth.GenerateAccessToken(u.cfg.JWTSecret, usr, roleName, accessTTL)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(accessTTL.Seconds()),
		User: dto.UserResponse{
			ID:        usr.ID,
			RoleID:    usr.RoleID,
			RoleName:  roleName,
			Name:      usr.Name,
			Email:     usr.Email,
			IsActive:  usr.IsActive,
			CreatedAt: usr.CreatedAt,
		},
	}, nil
}

func (u *AuthUsecase) GetProfile(ctx context.Context, userID uint64) (*dto.UserResponse, error) {
	usr, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	role, _ := u.userRepo.GetRoleByID(ctx, usr.RoleID)
	roleName := "writer"
	if role != nil {
		roleName = role.Name
	}
	return &dto.UserResponse{
		ID:        usr.ID,
		RoleID:    usr.RoleID,
		RoleName:  roleName,
		Name:      usr.Name,
		Email:     usr.Email,
		IsActive:  usr.IsActive,
		CreatedAt: usr.CreatedAt,
	}, nil
}
