package usecase

import (
	"context"

	"news-portal/apps/api/internal/domain/redirect"
	"news-portal/apps/api/internal/transport/http/dto"
)

type RedirectAdminUsecase struct {
	redirectRepo redirect.Repository
}

func NewRedirectAdminUsecase(redirectRepo redirect.Repository) *RedirectAdminUsecase {
	return &RedirectAdminUsecase{redirectRepo: redirectRepo}
}

func (u *RedirectAdminUsecase) List(ctx context.Context, limit int32, offset int32) ([]dto.RedirectResponse, error) {
	rows, err := u.redirectRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	res := make([]dto.RedirectResponse, len(rows))
	for i, r := range rows {
		res[i] = dto.RedirectResponse{
			ID:         r.ID,
			FromPath:   r.FromPath,
			ToPath:     r.ToPath,
			HTTPStatus: r.HTTPStatus,
			HitCount:   r.HitCount,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}
	}
	return res, nil
}

func (u *RedirectAdminUsecase) CreateOrUpdate(ctx context.Context, req *dto.RedirectRequest) error {
	return u.redirectRepo.CreateOrUpdate(ctx, req.FromPath, req.ToPath, req.HTTPStatus)
}
