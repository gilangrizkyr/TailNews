package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/domain/article"
	"news-portal/apps/api/internal/domain/author"
	"news-portal/apps/api/internal/domain/category"
	"news-portal/apps/api/internal/domain/media"
	"news-portal/apps/api/internal/domain/page"
	"news-portal/apps/api/internal/domain/tag"
	"news-portal/apps/api/internal/pkg/auth"
	"news-portal/apps/api/internal/pkg/cursor"
	domainerrors "news-portal/apps/api/internal/pkg/errors"
	"news-portal/apps/api/internal/platform"
	"news-portal/apps/api/internal/transport/http/dto"
)

type PublicUsecase struct {
	articleRepo  article.Repository
	categoryRepo category.Repository
	tagRepo      tag.Repository
	authorRepo   author.Repository
	pageRepo     page.Repository
	mediaRepo    media.Repository
	cache        platform.CachePlatform
	viewCounter  platform.ViewCounter
	cfg          *config.Config
}

func NewPublicUsecase(
	articleRepo article.Repository,
	categoryRepo category.Repository,
	tagRepo tag.Repository,
	authorRepo author.Repository,
	pageRepo page.Repository,
	mediaRepo media.Repository,
	cache platform.CachePlatform,
	viewCounter platform.ViewCounter,
	cfg *config.Config,
) *PublicUsecase {
	return &PublicUsecase{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		authorRepo:   authorRepo,
		pageRepo:     pageRepo,
		mediaRepo:    mediaRepo,
		cache:        cache,
		viewCounter:  viewCounter,
		cfg:          cfg,
	}
}

// GetHomepage returns aggregated data for the public homepage (cached for 30s).
func (u *PublicUsecase) GetHomepage(ctx context.Context) (*dto.HomepageResponse, error) {
	cacheKey := "news:v1:feed:home"
	var resp dto.HomepageResponse

	err := u.cache.Remember(ctx, cacheKey, 30*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		headlinesSummary, err := u.articleRepo.ListHeadlines(fetchCtx, 5)
		if err != nil {
			return nil, err
		}

		breakingSummary, err := u.articleRepo.ListBreaking(fetchCtx, 5)
		if err != nil {
			return nil, err
		}

		latestSummary, err := u.articleRepo.ListCursor(fetchCtx, nil, 10)
		if err != nil {
			return nil, err
		}

		// Trending articles past 24h
		trendingSummary, err := u.articleRepo.GetTrending(fetchCtx, time.Now().AddDate(0, 0, -1), 5)
		if err != nil || len(trendingSummary) == 0 {
			// Fallback to headlines if trending has no recorded views yet
			trendingSummary = headlinesSummary
		}

		categoriesList, err := u.categoryRepo.List(fetchCtx)
		if err != nil {
			return nil, err
		}

		// Map to DTOs
		headlines := u.mapSummariesToDTOs(fetchCtx, headlinesSummary)
		breaking := u.mapSummariesToDTOs(fetchCtx, breakingSummary)
		latest := u.mapSummariesToDTOs(fetchCtx, latestSummary)
		trending := u.mapSummariesToDTOs(fetchCtx, trendingSummary)

		cats := make([]dto.CategoryResponse, len(categoriesList))
		for i, c := range categoriesList {
			cats[i] = dto.CategoryResponse{
				ID:          c.ID,
				Name:        c.Name,
				Slug:        c.Slug,
				Description: c.Description,
				ParentID:    c.ParentID,
				SortOrder:   c.SortOrder,
				IsNav:       c.IsNav,
				CreatedAt:   c.CreatedAt,
				UpdatedAt:   c.UpdatedAt,
			}
		}

		nextCursor := ""
		if len(latestSummary) > 0 {
			last := latestSummary[len(latestSummary)-1]
			if last.PublishedAt != nil {
				nextCursor = cursor.Encode(*last.PublishedAt, last.ID)
			}
		}

		return dto.HomepageResponse{
			Headlines:  headlines,
			Breaking:   breaking,
			Latest:     latest,
			NextCursor: nextCursor,
			Trending:   trending,
			Categories: cats,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListArticles returns keyset cursor paginated published articles (cached for 60s).
func (u *PublicUsecase) ListArticles(ctx context.Context, cursorStr string, limit int32) (*dto.CursorArticlesResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 12
	}

	cur, err := cursor.Decode(cursorStr)
	if err != nil {
		return nil, domainerrors.ErrBadRequest
	}

	cacheKey := fmt.Sprintf("news:v1:feed:articles:%s:%d", cursorStr, limit)
	var resp dto.CursorArticlesResponse

	err = u.cache.Remember(ctx, cacheKey, 60*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		// Fetch limit + 1 to detect if next page exists
		summaries, err := u.articleRepo.ListCursor(fetchCtx, cur, limit+1)
		if err != nil {
			return nil, err
		}

		hasMore := false
		if int32(len(summaries)) > limit {
			hasMore = true
			summaries = summaries[:limit]
		}

		nextCursor := ""
		if len(summaries) > 0 && hasMore {
			last := summaries[len(summaries)-1]
			if last.PublishedAt != nil {
				nextCursor = cursor.Encode(*last.PublishedAt, last.ID)
			}
		}

		return dto.CursorArticlesResponse{
			Articles:   u.mapSummariesToDTOs(fetchCtx, summaries),
			NextCursor: nextCursor,
			HasMore:    hasMore,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListArticlesByCategory returns keyset cursor paginated articles by category slug (cached for 60s).
func (u *PublicUsecase) ListArticlesByCategory(ctx context.Context, catSlug string, cursorStr string, limit int32) (*dto.CursorArticlesResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 12
	}

	cat, err := u.categoryRepo.GetBySlug(ctx, catSlug)
	if err != nil {
		return nil, err
	}

	cur, err := cursor.Decode(cursorStr)
	if err != nil {
		return nil, domainerrors.ErrBadRequest
	}

	cacheKey := fmt.Sprintf("news:v1:feed:category:%s:%s:%d", catSlug, cursorStr, limit)
	var resp dto.CursorArticlesResponse

	err = u.cache.Remember(ctx, cacheKey, 60*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		summaries, err := u.articleRepo.ListByCategoryCursor(fetchCtx, cat.ID, cur, limit+1)
		if err != nil {
			return nil, err
		}

		hasMore := false
		if int32(len(summaries)) > limit {
			hasMore = true
			summaries = summaries[:limit]
		}

		nextCursor := ""
		if len(summaries) > 0 && hasMore {
			last := summaries[len(summaries)-1]
			if last.PublishedAt != nil {
				nextCursor = cursor.Encode(*last.PublishedAt, last.ID)
			}
		}

		return dto.CursorArticlesResponse{
			Articles:   u.mapSummariesToDTOs(fetchCtx, summaries),
			NextCursor: nextCursor,
			HasMore:    hasMore,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListArticlesByTag returns keyset cursor paginated articles by tag slug (cached for 60s).
func (u *PublicUsecase) ListArticlesByTag(ctx context.Context, tagSlug string, cursorStr string, limit int32) (*dto.CursorArticlesResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 12
	}

	tagItem, err := u.tagRepo.GetBySlug(ctx, tagSlug)
	if err != nil {
		return nil, err
	}

	cur, err := cursor.Decode(cursorStr)
	if err != nil {
		return nil, domainerrors.ErrBadRequest
	}

	cacheKey := fmt.Sprintf("news:v1:feed:tag:%s:%s:%d", tagSlug, cursorStr, limit)
	var resp dto.CursorArticlesResponse

	err = u.cache.Remember(ctx, cacheKey, 60*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		summaries, err := u.articleRepo.ListByTagCursor(fetchCtx, tagItem.ID, cur, limit+1)
		if err != nil {
			return nil, err
		}

		hasMore := false
		if int32(len(summaries)) > limit {
			hasMore = true
			summaries = summaries[:limit]
		}

		nextCursor := ""
		if len(summaries) > 0 && hasMore {
			last := summaries[len(summaries)-1]
			if last.PublishedAt != nil {
				nextCursor = cursor.Encode(*last.PublishedAt, last.ID)
			}
		}

		return dto.CursorArticlesResponse{
			Articles:   u.mapSummariesToDTOs(fetchCtx, summaries),
			NextCursor: nextCursor,
			HasMore:    hasMore,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuthorProfileAndArticles returns author info and keyset cursor paginated articles (cached for 60s).
func (u *PublicUsecase) GetAuthorProfileAndArticles(ctx context.Context, authorSlug string, cursorStr string, limit int32) (*dto.AuthorProfileResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 12
	}

	authItem, err := u.authorRepo.GetBySlug(ctx, authorSlug)
	if err != nil {
		return nil, err
	}

	cur, err := cursor.Decode(cursorStr)
	if err != nil {
		return nil, domainerrors.ErrBadRequest
	}

	cacheKey := fmt.Sprintf("news:v1:feed:author:%s:%s:%d", authorSlug, cursorStr, limit)
	var resp dto.AuthorProfileResponse

	err = u.cache.Remember(ctx, cacheKey, 60*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		summaries, err := u.articleRepo.ListByAuthorCursor(fetchCtx, authItem.ID, cur, limit+1)
		if err != nil {
			return nil, err
		}

		hasMore := false
		if int32(len(summaries)) > limit {
			hasMore = true
			summaries = summaries[:limit]
		}

		nextCursor := ""
		if len(summaries) > 0 && hasMore {
			last := summaries[len(summaries)-1]
			if last.PublishedAt != nil {
				nextCursor = cursor.Encode(*last.PublishedAt, last.ID)
			}
		}

		return dto.AuthorProfileResponse{
			Author: dto.AuthorResponse{
				ID:            authItem.ID,
				UserID:        authItem.UserID,
				Name:          authItem.Name,
				Slug:          authItem.Slug,
				Bio:           authItem.Bio,
				AvatarMediaID: authItem.AvatarMediaID,
				SocialLinks:   authItem.SocialLinks,
				CreatedAt:     authItem.CreatedAt,
				UpdatedAt:     authItem.UpdatedAt,
			},
			Articles:   u.mapSummariesToDTOs(fetchCtx, summaries),
			NextCursor: nextCursor,
			HasMore:    hasMore,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetArticleBySlug retrieves article detail by slug with optional preview token for draft review.
func (u *PublicUsecase) GetArticleBySlug(ctx context.Context, artSlug string, previewToken string) (*dto.ArticleResponse, error) {
	// If preview token is supplied, bypass cache and allow draft preview
	if previewToken != "" {
		artID, err := auth.ValidatePreviewToken(u.cfg.JWTSecret, previewToken)
		if err != nil {
			return nil, domainerrors.ErrUnauthorized
		}

		art, err := u.articleRepo.GetByID(ctx, artID)
		if err != nil {
			return nil, err
		}
		return u.enrichArticleDetail(ctx, art)
	}

	// Normal published article read path with L2 Cache and singleflight anti-stampede
	cacheKey := fmt.Sprintf("news:v1:article:slug:%s", artSlug)
	var resp dto.ArticleResponse

	err := u.cache.Remember(ctx, cacheKey, 120*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		art, err := u.articleRepo.GetBySlug(fetchCtx, artSlug)
		if err != nil {
			return nil, err
		}

		// Ensure only published articles are served publicly
		if art.Status != article.StatusPublished {
			return nil, domainerrors.ErrNotFound
		}

		detail, err := u.enrichArticleDetail(fetchCtx, art)
		if err != nil {
			return nil, err
		}
		return *detail, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchArticles performs boolean full-text search against article titles and excerpts (cached for 120s).
func (u *PublicUsecase) SearchArticles(ctx context.Context, query string, limit int32) ([]dto.ArticleResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []dto.ArticleResponse{}, nil
	}

	if limit <= 0 || limit > 50 {
		limit = 15
	}

	cacheKey := fmt.Sprintf("news:v1:search:%s:%d", query, limit)
	var resp []dto.ArticleResponse

	err := u.cache.Remember(ctx, cacheKey, 120*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		summaries, err := u.articleRepo.SearchFulltext(fetchCtx, query, limit)
		if err != nil {
			return nil, err
		}
		return u.mapSummariesToDTOs(fetchCtx, summaries), nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetTrending returns top articles ranked by views within a period ("today" or "week").
func (u *PublicUsecase) GetTrending(ctx context.Context, window string, limit int32) ([]dto.ArticleResponse, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	since := time.Now().AddDate(0, 0, -1)
	if window == "week" {
		since = time.Now().AddDate(0, 0, -7)
	}

	cacheKey := fmt.Sprintf("news:v1:trending:%s:%d", window, limit)
	var resp []dto.ArticleResponse

	err := u.cache.Remember(ctx, cacheKey, 120*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		summaries, err := u.articleRepo.GetTrending(fetchCtx, since, limit)
		if err != nil {
			return nil, err
		}

		// If period views is empty, fallback to headlines
		if len(summaries) == 0 {
			summaries, err = u.articleRepo.ListHeadlines(fetchCtx, limit)
			if err != nil {
				return nil, err
			}
		}

		return u.mapSummariesToDTOs(fetchCtx, summaries), nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetPageBySlug returns static editorial page content by slug (cached for 600s).
func (u *PublicUsecase) GetPageBySlug(ctx context.Context, slug string) (*dto.PageResponse, error) {
	cacheKey := fmt.Sprintf("news:v1:page:%s", slug)
	var resp dto.PageResponse

	err := u.cache.Remember(ctx, cacheKey, 600*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		p, err := u.pageRepo.GetBySlug(fetchCtx, slug)
		if err != nil {
			return nil, err
		}
		if !p.IsPublished {
			return nil, domainerrors.ErrNotFound
		}

		return dto.PageResponse{
			ID:          p.ID,
			Title:       p.Title,
			Slug:        p.Slug,
			Content:     p.Content,
			IsPublished: p.IsPublished,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListCategories returns all navigation categories (cached for 300s).
func (u *PublicUsecase) ListCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	cacheKey := "news:v1:categories"
	var resp []dto.CategoryResponse

	err := u.cache.Remember(ctx, cacheKey, 300*time.Second, &resp, func(fetchCtx context.Context) (interface{}, error) {
		categoriesList, err := u.categoryRepo.List(fetchCtx)
		if err != nil {
			return nil, err
		}

		cats := make([]dto.CategoryResponse, len(categoriesList))
		for i, c := range categoriesList {
			cats[i] = dto.CategoryResponse{
				ID:          c.ID,
				Name:        c.Name,
				Slug:        c.Slug,
				Description: c.Description,
				ParentID:    c.ParentID,
				SortOrder:   c.SortOrder,
				IsNav:       c.IsNav,
				CreatedAt:   c.CreatedAt,
				UpdatedAt:   c.UpdatedAt,
			}
		}
		return cats, nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil
}

// RecordView logs an article view asynchronously with hourly IP+UA deduplication.
func (u *PublicUsecase) RecordView(ctx context.Context, articleID uint64, ip string, ua string) (*dto.RecordViewResponse, error) {
	recorded, err := u.viewCounter.RecordView(ctx, articleID, ip, ua)
	if err != nil {
		return nil, err
	}
	return &dto.RecordViewResponse{
		Recorded:  recorded,
		ArticleID: articleID,
	}, nil
}

// Helper: mapSummariesToDTOs enriches slice of article summaries with category, author, and cover URL.
func (u *PublicUsecase) mapSummariesToDTOs(ctx context.Context, summaries []*article.ArticleSummary) []dto.ArticleResponse {
	res := make([]dto.ArticleResponse, len(summaries))
	for i, s := range summaries {
		catName, catSlug := s.CategoryName, s.CategorySlug
		if catName == "" {
			cat, _ := u.categoryRepo.GetByID(ctx, s.CategoryID)
			if cat != nil {
				catName, catSlug = cat.Name, cat.Slug
			}
		}

		authName, authSlug := s.AuthorName, s.AuthorSlug
		if authName == "" {
			a, _ := u.authorRepo.GetByID(ctx, s.AuthorID)
			if a != nil {
				authName, authSlug = a.Name, a.Slug
			}
		}

		coverURL := s.CoverURL
		if coverURL == "" && s.CoverMediaID != nil {
			m, _ := u.mediaRepo.GetByID(ctx, *s.CoverMediaID)
			if m != nil {
				coverURL = m.FilePath
			}
		}

		res[i] = dto.ArticleResponse{
			ID:           s.ID,
			Title:        s.Title,
			Slug:         s.Slug,
			Excerpt:      s.Excerpt,
			CoverMediaID: s.CoverMediaID,
			CoverURL:     coverURL,
			CategoryID:   s.CategoryID,
			CategoryName: catName,
			CategorySlug: catSlug,
			AuthorID:     s.AuthorID,
			AuthorName:   authName,
			AuthorSlug:   authSlug,
			IsHeadline:   s.IsHeadline,
			IsBreaking:   s.IsBreaking,
			ViewCount:    s.ViewCount,
			PublishedAt:  s.PublishedAt,
		}
	}
	return res
}

// Helper: enrichArticleDetail enriches single full article entity.
func (u *PublicUsecase) enrichArticleDetail(ctx context.Context, art *article.Article) (*dto.ArticleResponse, error) {
	cat, _ := u.categoryRepo.GetByID(ctx, art.CategoryID)
	catName, catSlug := "", ""
	if cat != nil {
		catName, catSlug = cat.Name, cat.Slug
	}

	authItem, _ := u.authorRepo.GetByID(ctx, art.AuthorID)
	authName, authSlug := "", ""
	if authItem != nil {
		authName, authSlug = authItem.Name, authItem.Slug
	}

	coverURL := ""
	if art.CoverMediaID != nil {
		m, _ := u.mediaRepo.GetByID(ctx, *art.CoverMediaID)
		if m != nil {
			coverURL = m.FilePath
		}
	}

	tags, _ := u.articleRepo.GetTags(ctx, art.ID)
	tagDTOs := make([]dto.TagResponse, len(tags))
	for i, t := range tags {
		tagDTOs[i] = dto.TagResponse{
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
		Tags:           tagDTOs,
	}, nil
}
