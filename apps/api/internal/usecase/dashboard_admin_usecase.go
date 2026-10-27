package usecase

import (
	"context"
	"database/sql"

	"news-portal/apps/api/internal/transport/http/dto"
)

type DashboardAdminUsecase struct {
	db *sql.DB
}

func NewDashboardAdminUsecase(db *sql.DB) *DashboardAdminUsecase {
	return &DashboardAdminUsecase{db: db}
}

func (u *DashboardAdminUsecase) GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	stats := &dto.DashboardStatsResponse{}

	// Query article stats
	row := u.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'published' AND deleted_at IS NULL THEN 1 END) as published,
			COUNT(CASE WHEN status = 'draft' AND deleted_at IS NULL THEN 1 END) as draft,
			COUNT(CASE WHEN status = 'scheduled' AND deleted_at IS NULL THEN 1 END) as scheduled,
			COALESCE(SUM(view_count), 0) as total_views
		FROM articles
	`)
	if err := row.Scan(&stats.TotalArticles, &stats.PublishedArticles, &stats.DraftArticles, &stats.ScheduledArticles, &stats.TotalViews); err != nil {
		return nil, err
	}

	// Categories count
	_ = u.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories`).Scan(&stats.TotalCategories)

	// Authors count
	_ = u.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM authors`).Scan(&stats.TotalAuthors)

	// Media count
	_ = u.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM media`).Scan(&stats.TotalMedia)

	return stats, nil
}
