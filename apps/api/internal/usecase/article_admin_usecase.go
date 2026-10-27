package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/domain/article"
	"news-portal/apps/api/internal/domain/audit"
	"news-portal/apps/api/internal/domain/author"
	"news-portal/apps/api/internal/domain/category"
	"news-portal/apps/api/internal/domain/media"
	"news-portal/apps/api/internal/domain/redirect"
	"news-portal/apps/api/internal/pkg/auth"
	domainerrors "news-portal/apps/api/internal/pkg/errors"
	"news-portal/apps/api/internal/pkg/sanitize"
	"news-portal/apps/api/internal/pkg/slug"
	"news-portal/apps/api/internal/transport/http/dto"
)

type ArticleAdminUsecase struct {
	articleRepo  article.Repository
	categoryRepo category.Repository
	authorRepo   author.Repository
	mediaRepo    media.Repository
	redirectRepo redirect.Repository
	auditRepo    audit.Repository
	cfg          *config.Config
}

func NewArticleAdminUsecase(
	articleRepo article.Repository,
	categoryRepo category.Repository,
	authorRepo author.Repository,
	mediaRepo media.Repository,
	redirectRepo redirect.Repository,
	auditRepo audit.Repository,
	cfg *config.Config,
) *ArticleAdminUsecase {
	return &ArticleAdminUsecase{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		authorRepo:   authorRepo,
		mediaRepo:    mediaRepo,
		redirectRepo: redirectRepo,
		auditRepo:    auditRepo,
		cfg:          cfg,
	}
}

func (u *ArticleAdminUsecase) Create(ctx context.Context, req *dto.ArticleUpsertRequest, userID uint64, ip string, ua string) (*dto.ArticleResponse, error) {
	// Auto generate slug if empty
	artSlug := req.Slug
	if artSlug == "" {
		artSlug = slug.Generate(req.Title)
	}

	// Sanitize body HTML
	cleanBody := sanitize.HTML(req.Body)

	art := &article.Article{
		Title:          req.Title,
		Slug:           artSlug,
		Excerpt:        req.Excerpt,
		Body:           cleanBody,
		CoverMediaID:   req.CoverMediaID,
		CategoryID:     req.CategoryID,
		AuthorID:       req.AuthorID,
		CreatedBy:      userID,
		Status:         article.Status(req.Status),
		IsHeadline:     req.IsHeadline,
		IsBreaking:     req.IsBreaking,
		PublishedAt:    req.PublishedAt,
		SEOTitle:       req.SEOTitle,
		SEODescription: req.SEODescription,
	}

	id, err := u.articleRepo.Create(ctx, art)
	if err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}
	art.ID = id

	// Associate tags
	if len(req.TagIDs) > 0 {
		_ = u.articleRepo.SetTags(ctx, id, req.TagIDs)
	}

	// Create initial revision snapshot
	_, _ = u.articleRepo.CreateRevision(ctx, &article.ArticleRevision{
		ArticleID: id,
		UserID:    userID,
		Title:     art.Title,
		Slug:      art.Slug,
		Excerpt:   art.Excerpt,
		Body:      art.Body,
	})

	// Audit Log
	diff, _ := json.Marshal(map[string]interface{}{"title": art.Title, "status": art.Status})
	_, _ = u.auditRepo.Create(ctx, &audit.AuditLog{
		UserID:      &userID,
		Action:      "CREATE",
		Entity:      "article",
		EntityID:    id,
		DiffSummary: diff,
		IPAddress:   ip,
		UserAgent:   ua,
	})

	return u.GetByID(ctx, id)
}

func (u *ArticleAdminUsecase) Update(ctx context.Context, id uint64, req *dto.ArticleUpsertRequest, userID uint64, ip string, ua string) (*dto.ArticleResponse, error) {
	existing, err := u.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newSlug := req.Slug
	if newSlug == "" {
		newSlug = slug.Generate(req.Title)
	}

	// Check if slug changed; if so, create 301 redirect
	if existing.Slug != newSlug {
		oldPath := fmt.Sprintf("/%s", existing.Slug)
		newPath := fmt.Sprintf("/%s", newSlug)
		_ = u.redirectRepo.CreateOrUpdate(ctx, oldPath, newPath, 301)
	}

	cleanBody := sanitize.HTML(req.Body)

	existing.Title = req.Title
	existing.Slug = newSlug
	existing.Excerpt = req.Excerpt
	existing.Body = cleanBody
	existing.CoverMediaID = req.CoverMediaID
	existing.CategoryID = req.CategoryID
	existing.AuthorID = req.AuthorID
	existing.Status = article.Status(req.Status)
	existing.IsHeadline = req.IsHeadline
	existing.IsBreaking = req.IsBreaking
	existing.PublishedAt = req.PublishedAt
	existing.SEOTitle = req.SEOTitle
	existing.SEODescription = req.SEODescription

	if err := u.articleRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	if req.TagIDs != nil {
		_ = u.articleRepo.SetTags(ctx, id, req.TagIDs)
	}

	// Create revision snapshot on update
	_, _ = u.articleRepo.CreateRevision(ctx, &article.ArticleRevision{
		ArticleID: id,
		UserID:    userID,
		Title:     existing.Title,
		Slug:      existing.Slug,
		Excerpt:   existing.Excerpt,
		Body:      existing.Body,
	})

	// Audit Log
	diff, _ := json.Marshal(map[string]interface{}{"title": existing.Title, "status": existing.Status})
	_, _ = u.auditRepo.Create(ctx, &audit.AuditLog{
		UserID:      &userID,
		Action:      "UPDATE",
		Entity:      "article",
		EntityID:    id,
		DiffSummary: diff,
		IPAddress:   ip,
		UserAgent:   ua,
	})

	return u.GetByID(ctx, id)
}

func (u *ArticleAdminUsecase) SoftDelete(ctx context.Context, id uint64, userID uint64, ip string, ua string) error {
	existing, err := u.articleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.articleRepo.SoftDelete(ctx, id); err != nil {
		return err
	}

	diff, _ := json.Marshal(map[string]interface{}{"deleted_slug": existing.Slug})
	_, _ = u.auditRepo.Create(ctx, &audit.AuditLog{
		UserID:      &userID,
		Action:      "DELETE",
		Entity:      "article",
		EntityID:    id,
		DiffSummary: diff,
		IPAddress:   ip,
		UserAgent:   ua,
	})

	return nil
}

func (u *ArticleAdminUsecase) RestoreRevision(ctx context.Context, articleID uint64, revisionID uint64, userID uint64, ip string, ua string) (*dto.ArticleResponse, error) {
	art, err := u.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	revisions, err := u.articleRepo.ListRevisions(ctx, articleID)
	if err != nil {
		return nil, err
	}

	var targetRev *article.ArticleRevision
	for _, rev := range revisions {
		if rev.ID == revisionID {
			targetRev = rev
			break
		}
	}
	if targetRev == nil {
		return nil, domainerrors.ErrNotFound
	}

	art.Title = targetRev.Title
	art.Slug = targetRev.Slug
	art.Excerpt = targetRev.Excerpt
	art.Body = targetRev.Body

	if err := u.articleRepo.Update(ctx, art); err != nil {
		return nil, err
	}

	// Create new revision indicating restore
	_, _ = u.articleRepo.CreateRevision(ctx, &article.ArticleRevision{
		ArticleID: articleID,
		UserID:    userID,
		Title:     art.Title,
		Slug:      art.Slug,
		Excerpt:   art.Excerpt,
		Body:      art.Body,
	})

	diff, _ := json.Marshal(map[string]interface{}{"restored_from_revision": revisionID})
	_, _ = u.auditRepo.Create(ctx, &audit.AuditLog{
		UserID:      &userID,
		Action:      "RESTORE",
		Entity:      "article",
		EntityID:    articleID,
		DiffSummary: diff,
		IPAddress:   ip,
		UserAgent:   ua,
	})

	return u.GetByID(ctx, articleID)
}

func (u *ArticleAdminUsecase) GeneratePreviewToken(articleID uint64) (string, error) {
	return auth.GeneratePreviewToken(u.cfg.JWTSecret, articleID, 2*time.Hour)
}

func (u *ArticleAdminUsecase) GetByID(ctx context.Context, id uint64) (*dto.ArticleResponse, error) {
	art, err := u.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tags, _ := u.articleRepo.GetTags(ctx, id)
	revisions, _ := u.articleRepo.ListRevisions(ctx, id)

	cat, _ := u.categoryRepo.GetByID(ctx, art.CategoryID)
	catName, catSlug := "", ""
	if cat != nil {
		catName, catSlug = cat.Name, cat.Slug
	}

	auth, _ := u.authorRepo.GetByID(ctx, art.AuthorID)
	authName, authSlug := "", ""
	if auth != nil {
		authName, authSlug = auth.Name, auth.Slug
	}

	coverURL := ""
	if art.CoverMediaID != nil {
		m, _ := u.mediaRepo.GetByID(ctx, *art.CoverMediaID)
		if m != nil {
			coverURL = m.FilePath
		}
	}

	tagDtos := make([]dto.TagResponse, len(tags))
	for i, t := range tags {
		tagDtos[i] = dto.TagResponse{
			ID:   t.ID,
			Name: t.Name,
			Slug: t.Slug,
		}
	}

	return &dto.ArticleResponse{
		ID:             art.ID,
		Title:          art.Title,
		Slug:           art.Slug,
		Excerpt:        art.Excerpt,
		Body:           art.Body,
		CoverMediaID:   art.CoverMediaID,
		CoverURL:       coverURL,
		CategoryID:     art.CategoryID,
		CategoryName:   catName,
		CategorySlug:   catSlug,
		AuthorID:       art.AuthorID,
		AuthorName:     authName,
		AuthorSlug:     authSlug,
		CreatedBy:      art.CreatedBy,
		Status:         string(art.Status),
		IsHeadline:     art.IsHeadline,
		IsBreaking:     art.IsBreaking,
		ViewCount:      art.ViewCount,
		PublishedAt:    art.PublishedAt,
		SEOTitle:       art.SEOTitle,
		SEODescription: art.SEODescription,
		CreatedAt:      art.CreatedAt,
		UpdatedAt:      art.UpdatedAt,
		Tags:           tagDtos,
		RevisionsCount: len(revisions),
	}, nil
}

func (u *ArticleAdminUsecase) ListRevisions(ctx context.Context, articleID uint64) ([]dto.ArticleRevisionResponse, error) {
	revs, err := u.articleRepo.ListRevisions(ctx, articleID)
	if err != nil {
		return nil, err
	}
	res := make([]dto.ArticleRevisionResponse, len(revs))
	for i, r := range revs {
		res[i] = dto.ArticleRevisionResponse{
			ID:        r.ID,
			ArticleID: r.ArticleID,
			UserID:    r.UserID,
			Title:     r.Title,
			Slug:      r.Slug,
			Excerpt:   r.Excerpt,
			Body:      r.Body,
			CreatedAt: r.CreatedAt,
		}
	}
	return res, nil
}

func (u *ArticleAdminUsecase) List(ctx context.Context, page int32, limit int32) ([]*dto.ArticleResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	articles, err := u.articleRepo.ListAdmin(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.ArticleResponse, len(articles))
	for i, art := range articles {
		cat, _ := u.categoryRepo.GetByID(ctx, art.CategoryID)
		catName, catSlug := "", ""
		if cat != nil {
			catName, catSlug = cat.Name, cat.Slug
		}

		auth, _ := u.authorRepo.GetByID(ctx, art.AuthorID)
		authName, authSlug := "", ""
		if auth != nil {
			authName, authSlug = auth.Name, auth.Slug
		}

		coverURL := ""
		if art.CoverMediaID != nil {
			m, _ := u.mediaRepo.GetByID(ctx, *art.CoverMediaID)
			if m != nil {
				coverURL = m.FilePath
			}
		}

		res[i] = &dto.ArticleResponse{
			ID:             art.ID,
			Title:          art.Title,
			Slug:           art.Slug,
			Excerpt:        art.Excerpt,
			CoverMediaID:   art.CoverMediaID,
			CoverURL:       coverURL,
			CategoryID:     art.CategoryID,
			CategoryName:   catName,
			CategorySlug:   catSlug,
			AuthorID:       art.AuthorID,
			AuthorName:     authName,
			AuthorSlug:     authSlug,
			CreatedBy:      art.CreatedBy,
			Status:         string(art.Status),
			IsHeadline:     art.IsHeadline,
			IsBreaking:     art.IsBreaking,
			ViewCount:      art.ViewCount,
			PublishedAt:    art.PublishedAt,
			SEOTitle:       art.SEOTitle,
			SEODescription: art.SEODescription,
			CreatedAt:      art.CreatedAt,
			UpdatedAt:      art.UpdatedAt,
		}
	}
	return res, nil
}

