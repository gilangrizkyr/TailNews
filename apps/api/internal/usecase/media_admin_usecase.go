package usecase

import (
	"context"
	"io"

	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/domain/media"
	"news-portal/apps/api/internal/platform"
	"news-portal/apps/api/internal/transport/http/dto"
)

type MediaAdminUsecase struct {
	mediaRepo media.Repository
	storage   platform.StoragePlatform
	cfg       *config.Config
}

func NewMediaAdminUsecase(
	mediaRepo media.Repository,
	storage platform.StoragePlatform,
	cfg *config.Config,
) *MediaAdminUsecase {
	return &MediaAdminUsecase{
		mediaRepo: mediaRepo,
		storage:   storage,
		cfg:       cfg,
	}
}

func (u *MediaAdminUsecase) Upload(ctx context.Context, r io.Reader, filename string, altText string, caption string, photographer string) (*dto.MediaResponse, error) {
	maxBytes := u.cfg.MaxUploadSizeMB * 1024 * 1024
	processed, err := u.storage.SaveImage(r, filename, maxBytes)
	if err != nil {
		return nil, err
	}

	m := &media.Media{
		Filename:     processed.StoredFilename,
		FilePath:     processed.FilePath,
		MimeType:     processed.MimeType,
		SizeBytes:    processed.SizeBytes,
		Width:        processed.Width,
		Height:       processed.Height,
		AltText:      altText,
		Caption:      caption,
		Photographer: photographer,
		Variants:     processed.Variants,
	}

	id, err := u.mediaRepo.Create(ctx, m)
	if err != nil {
		return nil, err
	}
	m.ID = id

	return &dto.MediaResponse{
		ID:           m.ID,
		Filename:     m.Filename,
		FilePath:     m.FilePath,
		MimeType:     m.MimeType,
		SizeBytes:    m.SizeBytes,
		Width:        m.Width,
		Height:       m.Height,
		AltText:      m.AltText,
		Caption:      m.Caption,
		Photographer: m.Photographer,
		Variants:     m.Variants,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func (u *MediaAdminUsecase) List(ctx context.Context, limit int32, offset int32) ([]dto.MediaResponse, error) {
	rows, err := u.mediaRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	res := make([]dto.MediaResponse, len(rows))
	for i, m := range rows {
		res[i] = dto.MediaResponse{
			ID:           m.ID,
			Filename:     m.Filename,
			FilePath:     m.FilePath,
			MimeType:     m.MimeType,
			SizeBytes:    m.SizeBytes,
			Width:        m.Width,
			Height:       m.Height,
			AltText:      m.AltText,
			Caption:      m.Caption,
			Photographer: m.Photographer,
			Variants:     m.Variants,
			CreatedAt:    m.CreatedAt,
			UpdatedAt:    m.UpdatedAt,
		}
	}
	return res, nil
}

func (u *MediaAdminUsecase) Delete(ctx context.Context, id uint64) error {
	m, err := u.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	_ = u.storage.DeleteImage(m.FilePath, m.Variants)
	return u.mediaRepo.Delete(ctx, id)
}
