package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"news-portal/apps/api/internal/domain/article"
	"news-portal/apps/api/internal/domain/audit"
	"news-portal/apps/api/internal/domain/author"
	"news-portal/apps/api/internal/domain/category"
	"news-portal/apps/api/internal/domain/media"
	"news-portal/apps/api/internal/domain/page"
	"news-portal/apps/api/internal/domain/redirect"
	"news-portal/apps/api/internal/domain/setting"
	"news-portal/apps/api/internal/domain/tag"
	"news-portal/apps/api/internal/domain/user"
	domainerrors "news-portal/apps/api/internal/pkg/errors"
)

// Repository aggregates all MySQL domain repositories.
type Repository struct {
	db      *sql.DB
	queries *Queries

	User     user.Repository
	Category category.Repository
	Tag      tag.Repository
	Author   author.Repository
	Media    media.Repository
	Article  article.Repository
	Redirect redirect.Repository
	Page     page.Repository
	Setting  setting.Repository
	Audit    audit.Repository
}

// NewRepository constructs and returns all domain repository implementations.
func NewRepository(db *sql.DB) *Repository {
	q := New(db)
	repo := &Repository{
		db:      db,
		queries: q,
	}

	repo.User = &userRepo{db: db, q: q}
	repo.Category = &categoryRepo{db: db, q: q}
	repo.Tag = &tagRepo{db: db, q: q}
	repo.Author = &authorRepo{db: db, q: q}
	repo.Media = &mediaRepo{db: db, q: q}
	repo.Article = &articleRepo{db: db, q: q}
	repo.Redirect = &redirectRepo{db: db, q: q}
	repo.Page = &pageRepo{db: db, q: q}
	repo.Setting = &settingRepo{db: db, q: q}
	repo.Audit = &auditRepo{db: db, q: q}

	return repo
}

// ==========================================
// User Repository
// ==========================================
type userRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *userRepo) GetByID(ctx context.Context, id uint64) (*user.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	var lockedUntil *time.Time
	if row.LockedUntil.Valid {
		lockedUntil = &row.LockedUntil.Time
	}
	return &user.User{
		ID:                  row.ID,
		RoleID:              row.RoleID,
		Name:                row.Name,
		Email:               row.Email,
		PasswordHash:        row.PasswordHash,
		IsActive:            row.IsActive,
		FailedLoginAttempts: row.FailedLoginAttempts,
		LockedUntil:         lockedUntil,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	var lockedUntil *time.Time
	if row.LockedUntil.Valid {
		lockedUntil = &row.LockedUntil.Time
	}
	return &user.User{
		ID:                  row.ID,
		RoleID:              row.RoleID,
		Name:                row.Name,
		Email:               row.Email,
		PasswordHash:        row.PasswordHash,
		IsActive:            row.IsActive,
		FailedLoginAttempts: row.FailedLoginAttempts,
		LockedUntil:         lockedUntil,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

func (r *userRepo) Create(ctx context.Context, u *user.User) (uint64, error) {
	res, err := r.q.CreateUser(ctx, CreateUserParams{
		RoleID:       u.RoleID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsActive:     u.IsActive,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *userRepo) UpdateFailedLogin(ctx context.Context, id uint64) error {
	return r.q.UpdateUserFailedLogin(ctx, id)
}

func (r *userRepo) ResetFailedLogin(ctx context.Context, id uint64) error {
	return r.q.ResetUserFailedLogin(ctx, id)
}

func (r *userRepo) List(ctx context.Context) ([]*user.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*user.User, len(rows))
	for i, row := range rows {
		res[i] = &user.User{
			ID:        row.ID,
			RoleID:    row.RoleID,
			Name:      row.Name,
			Email:     row.Email,
			IsActive:  row.IsActive,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return res, nil
}

func (r *userRepo) GetRoleByID(ctx context.Context, roleID uint64) (*user.Role, error) {
	row, err := r.q.GetRoleByID(ctx, roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	var perms []string
	_ = json.Unmarshal(row.Permissions, &perms)
	return &user.Role{
		ID:          row.ID,
		Name:        row.Name,
		Permissions: perms,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *userRepo) ListRoles(ctx context.Context) ([]*user.Role, error) {
	rows, err := r.q.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*user.Role, len(rows))
	for i, row := range rows {
		var perms []string
		_ = json.Unmarshal(row.Permissions, &perms)
		res[i] = &user.Role{
			ID:          row.ID,
			Name:        row.Name,
			Permissions: perms,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
	}
	return res, nil
}

// ==========================================
// Category Repository
// ==========================================
type categoryRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *categoryRepo) GetByID(ctx context.Context, id uint64) (*category.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapCategoryRow(&row), nil
}

func (r *categoryRepo) GetBySlug(ctx context.Context, slug string) (*category.Category, error) {
	row, err := r.q.GetCategoryBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapCategoryRow(&row), nil
}

func (r *categoryRepo) List(ctx context.Context) ([]*category.Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*category.Category, len(rows))
	for i, row := range rows {
		res[i] = mapCategoryRow(&row)
	}
	return res, nil
}

func (r *categoryRepo) ListNav(ctx context.Context) ([]*category.Category, error) {
	rows, err := r.q.ListNavCategories(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*category.Category, len(rows))
	for i, row := range rows {
		res[i] = mapCategoryRow(&row)
	}
	return res, nil
}

func (r *categoryRepo) Create(ctx context.Context, c *category.Category) (uint64, error) {
	var parentID sql.NullInt64
	if c.ParentID != nil {
		parentID = sql.NullInt64{Int64: int64(*c.ParentID), Valid: true}
	}
	var desc sql.NullString
	if c.Description != "" {
		desc = sql.NullString{String: c.Description, Valid: true}
	}

	res, err := r.q.CreateCategory(ctx, CreateCategoryParams{
		Name:        c.Name,
		Slug:        c.Slug,
		Description: desc,
		ParentID:    parentID,
		SortOrder:   c.SortOrder,
		IsNav:       c.IsNav,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *categoryRepo) Update(ctx context.Context, c *category.Category) error {
	var parentID sql.NullInt64
	if c.ParentID != nil {
		parentID = sql.NullInt64{Int64: int64(*c.ParentID), Valid: true}
	}
	var desc sql.NullString
	if c.Description != "" {
		desc = sql.NullString{String: c.Description, Valid: true}
	}

	return r.q.UpdateCategory(ctx, UpdateCategoryParams{
		Name:        c.Name,
		Slug:        c.Slug,
		Description: desc,
		ParentID:    parentID,
		SortOrder:   c.SortOrder,
		IsNav:       c.IsNav,
		ID:          c.ID,
	})
}

func (r *categoryRepo) Delete(ctx context.Context, id uint64) error {
	return r.q.DeleteCategory(ctx, id)
}

func mapCategoryRow(row *Categories) *category.Category {
	var parentID *uint64
	if row.ParentID.Valid {
		val := uint64(row.ParentID.Int64)
		parentID = &val
	}
	return &category.Category{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description.String,
		ParentID:    parentID,
		SortOrder:   row.SortOrder,
		IsNav:       row.IsNav,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// ==========================================
// Tag Repository
// ==========================================
type tagRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *tagRepo) GetByID(ctx context.Context, id uint64) (*tag.Tag, error) {
	row, err := r.q.GetTagByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &tag.Tag{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *tagRepo) GetBySlug(ctx context.Context, slug string) (*tag.Tag, error) {
	row, err := r.q.GetTagBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &tag.Tag{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *tagRepo) List(ctx context.Context) ([]*tag.Tag, error) {
	rows, err := r.q.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*tag.Tag, len(rows))
	for i, row := range rows {
		res[i] = &tag.Tag{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return res, nil
}

func (r *tagRepo) Create(ctx context.Context, t *tag.Tag) (uint64, error) {
	res, err := r.q.CreateTag(ctx, CreateTagParams{
		Name: t.Name,
		Slug: t.Slug,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *tagRepo) Delete(ctx context.Context, id uint64) error {
	return r.q.DeleteTag(ctx, id)
}

func (r *tagRepo) GetOrCreate(ctx context.Context, name string, slug string) (*tag.Tag, error) {
	existing, err := r.GetBySlug(ctx, slug)
	if err == nil {
		return existing, nil
	}
	id, err := r.Create(ctx, &tag.Tag{Name: name, Slug: slug})
	if err != nil {
		// In case of race condition, retry fetch
		return r.GetBySlug(ctx, slug)
	}
	return &tag.Tag{ID: id, Name: name, Slug: slug, CreatedAt: time.Now()}, nil
}

// ==========================================
// Author Repository
// ==========================================
type authorRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *authorRepo) GetByID(ctx context.Context, id uint64) (*author.Author, error) {
	row, err := r.q.GetAuthorByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapAuthorRow(&row), nil
}

func (r *authorRepo) GetBySlug(ctx context.Context, slug string) (*author.Author, error) {
	row, err := r.q.GetAuthorBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapAuthorRow(&row), nil
}

func (r *authorRepo) List(ctx context.Context) ([]*author.Author, error) {
	rows, err := r.q.ListAuthors(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*author.Author, len(rows))
	for i, row := range rows {
		res[i] = mapAuthorRow(&row)
	}
	return res, nil
}

func (r *authorRepo) Create(ctx context.Context, a *author.Author) (uint64, error) {
	var userID sql.NullInt64
	if a.UserID != nil {
		userID = sql.NullInt64{Int64: int64(*a.UserID), Valid: true}
	}
	var avatarID sql.NullInt64
	if a.AvatarMediaID != nil {
		avatarID = sql.NullInt64{Int64: int64(*a.AvatarMediaID), Valid: true}
	}
	var bio sql.NullString
	if a.Bio != "" {
		bio = sql.NullString{String: a.Bio, Valid: true}
	}
	socialRaw, _ := json.Marshal(a.SocialLinks)

	res, err := r.q.CreateAuthor(ctx, CreateAuthorParams{
		UserID:        userID,
		Name:          a.Name,
		Slug:          a.Slug,
		Bio:           bio,
		AvatarMediaID: avatarID,
		SocialLinks:   socialRaw,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *authorRepo) Update(ctx context.Context, a *author.Author) error {
	var avatarID sql.NullInt64
	if a.AvatarMediaID != nil {
		avatarID = sql.NullInt64{Int64: int64(*a.AvatarMediaID), Valid: true}
	}
	var bio sql.NullString
	if a.Bio != "" {
		bio = sql.NullString{String: a.Bio, Valid: true}
	}
	socialRaw, _ := json.Marshal(a.SocialLinks)

	return r.q.UpdateAuthor(ctx, UpdateAuthorParams{
		Name:          a.Name,
		Slug:          a.Slug,
		Bio:           bio,
		AvatarMediaID: avatarID,
		SocialLinks:   socialRaw,
		ID:            a.ID,
	})
}

func mapAuthorRow(row *Authors) *author.Author {
	var userID *uint64
	if row.UserID.Valid {
		val := uint64(row.UserID.Int64)
		userID = &val
	}
	var avatarID *uint64
	if row.AvatarMediaID.Valid {
		val := uint64(row.AvatarMediaID.Int64)
		avatarID = &val
	}
	var social map[string]string
	if len(row.SocialLinks) > 0 {
		_ = json.Unmarshal(row.SocialLinks, &social)
	}
	return &author.Author{
		ID:            row.ID,
		UserID:        userID,
		Name:          row.Name,
		Slug:          row.Slug,
		Bio:           row.Bio.String,
		AvatarMediaID: avatarID,
		SocialLinks:   social,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

// ==========================================
// Media Repository
// ==========================================
type mediaRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *mediaRepo) GetByID(ctx context.Context, id uint64) (*media.Media, error) {
	row, err := r.q.GetMediaByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapMediaRow(&row), nil
}

func (r *mediaRepo) Create(ctx context.Context, m *media.Media) (uint64, error) {
	var caption sql.NullString
	if m.Caption != "" {
		caption = sql.NullString{String: m.Caption, Valid: true}
	}
	var photographer sql.NullString
	if m.Photographer != "" {
		photographer = sql.NullString{String: m.Photographer, Valid: true}
	}
	variantsRaw, _ := json.Marshal(m.Variants)

	res, err := r.q.CreateMedia(ctx, CreateMediaParams{
		Filename:     m.Filename,
		FilePath:     m.FilePath,
		MimeType:     m.MimeType,
		SizeBytes:    m.SizeBytes,
		Width:        m.Width,
		Height:       m.Height,
		AltText:      m.AltText,
		Caption:      caption,
		Photographer: photographer,
		Variants:     variantsRaw,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *mediaRepo) List(ctx context.Context, limit int32, offset int32) ([]*media.Media, error) {
	rows, err := r.q.ListMedia(ctx, ListMediaParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*media.Media, len(rows))
	for i, row := range rows {
		res[i] = mapMediaRow(&row)
	}
	return res, nil
}

func (r *mediaRepo) Delete(ctx context.Context, id uint64) error {
	return r.q.DeleteMedia(ctx, id)
}

func mapMediaRow(row *Media) *media.Media {
	var variants map[string]string
	if len(row.Variants) > 0 {
		_ = json.Unmarshal(row.Variants, &variants)
	}
	return &media.Media{
		ID:           row.ID,
		Filename:     row.Filename,
		FilePath:     row.FilePath,
		MimeType:     row.MimeType,
		SizeBytes:    row.SizeBytes,
		Width:        row.Width,
		Height:       row.Height,
		AltText:      row.AltText,
		Caption:      row.Caption.String,
		Photographer: row.Photographer.String,
		Variants:     variants,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// ==========================================
// Article Repository
// ==========================================
type articleRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *articleRepo) GetByID(ctx context.Context, id uint64) (*article.Article, error) {
	row, err := r.q.GetArticleByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapArticleRow(&row), nil
}

func (r *articleRepo) GetBySlug(ctx context.Context, slug string) (*article.Article, error) {
	row, err := r.q.GetArticleBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return mapArticleRow(&row), nil
}

func (r *articleRepo) Create(ctx context.Context, a *article.Article) (uint64, error) {
	var coverID sql.NullInt64
	if a.CoverMediaID != nil {
		coverID = sql.NullInt64{Int64: int64(*a.CoverMediaID), Valid: true}
	}
	var publishedAt sql.NullTime
	if a.PublishedAt != nil {
		publishedAt = sql.NullTime{Time: *a.PublishedAt, Valid: true}
	}
	var seoTitle sql.NullString
	if a.SEOTitle != "" {
		seoTitle = sql.NullString{String: a.SEOTitle, Valid: true}
	}
	var seoDesc sql.NullString
	if a.SEODescription != "" {
		seoDesc = sql.NullString{String: a.SEODescription, Valid: true}
	}

	res, err := r.q.CreateArticle(ctx, CreateArticleParams{
		Title:          a.Title,
		Slug:           a.Slug,
		Excerpt:        a.Excerpt,
		Body:           a.Body,
		CoverMediaID:   coverID,
		CategoryID:     a.CategoryID,
		AuthorID:       a.AuthorID,
		CreatedBy:      a.CreatedBy,
		Status:         ArticlesStatus(a.Status),
		IsHeadline:     a.IsHeadline,
		IsBreaking:     a.IsBreaking,
		ViewCount:      a.ViewCount,
		PublishedAt:    publishedAt,
		SeoTitle:       seoTitle,
		SeoDescription: seoDesc,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *articleRepo) Update(ctx context.Context, a *article.Article) error {
	var coverID sql.NullInt64
	if a.CoverMediaID != nil {
		coverID = sql.NullInt64{Int64: int64(*a.CoverMediaID), Valid: true}
	}
	var publishedAt sql.NullTime
	if a.PublishedAt != nil {
		publishedAt = sql.NullTime{Time: *a.PublishedAt, Valid: true}
	}
	var seoTitle sql.NullString
	if a.SEOTitle != "" {
		seoTitle = sql.NullString{String: a.SEOTitle, Valid: true}
	}
	var seoDesc sql.NullString
	if a.SEODescription != "" {
		seoDesc = sql.NullString{String: a.SEODescription, Valid: true}
	}

	return r.q.UpdateArticle(ctx, UpdateArticleParams{
		Title:          a.Title,
		Slug:           a.Slug,
		Excerpt:        a.Excerpt,
		Body:           a.Body,
		CoverMediaID:   coverID,
		CategoryID:     a.CategoryID,
		AuthorID:       a.AuthorID,
		Status:         ArticlesStatus(a.Status),
		IsHeadline:     a.IsHeadline,
		IsBreaking:     a.IsBreaking,
		PublishedAt:    publishedAt,
		SeoTitle:       seoTitle,
		SeoDescription: seoDesc,
		ID:             a.ID,
	})
}

func (r *articleRepo) SoftDelete(ctx context.Context, id uint64) error {
	return r.q.SoftDeleteArticle(ctx, id)
}

func (r *articleRepo) PermanentDelete(ctx context.Context, id uint64) error {
	return r.q.PermanentDeleteArticle(ctx, id)
}

func (r *articleRepo) ListCursor(ctx context.Context, cursor *article.Cursor, limit int32) ([]*article.ArticleSummary, error) {
	if cursor == nil {
		rows, err := r.q.ListArticlesFirstPage(ctx, limit)
		if err != nil {
			return nil, err
		}
		return mapArticleSummaryRows(rows), nil
	}

	rows, err := r.q.ListArticlesCursor(ctx, ListArticlesCursorParams{
		PublishedAt:   sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		PublishedAt_2: sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		ID:            cursor.ID,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}
	return mapArticleCursorRows(rows), nil
}

func (r *articleRepo) ListByCategoryCursor(ctx context.Context, categoryID uint64, cursor *article.Cursor, limit int32) ([]*article.ArticleSummary, error) {
	if cursor == nil {
		rows, err := r.q.ListArticlesByCategoryFirstPage(ctx, ListArticlesByCategoryFirstPageParams{
			CategoryID: categoryID,
			Limit:      limit,
		})
		if err != nil {
			return nil, err
		}
		return mapCategorySummaryRows(rows), nil
	}

	rows, err := r.q.ListArticlesByCategoryCursor(ctx, ListArticlesByCategoryCursorParams{
		CategoryID:    categoryID,
		PublishedAt:   sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		PublishedAt_2: sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		ID:            cursor.ID,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}
	return mapCategoryCursorRows(rows), nil
}

func (r *articleRepo) ListByAuthorCursor(ctx context.Context, authorID uint64, cursor *article.Cursor, limit int32) ([]*article.ArticleSummary, error) {
	if cursor == nil {
		rows, err := r.q.ListArticlesByAuthorFirstPage(ctx, ListArticlesByAuthorFirstPageParams{
			AuthorID: authorID,
			Limit:    limit,
		})
		if err != nil {
			return nil, err
		}
		return mapAuthorSummaryRows(rows), nil
	}

	rows, err := r.q.ListArticlesByAuthorCursor(ctx, ListArticlesByAuthorCursorParams{
		AuthorID:      authorID,
		PublishedAt:   sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		PublishedAt_2: sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		ID:            cursor.ID,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}
	return mapAuthorCursorRows(rows), nil
}

func (r *articleRepo) ListByTagCursor(ctx context.Context, tagID uint64, cursor *article.Cursor, limit int32) ([]*article.ArticleSummary, error) {
	if cursor == nil {
		rows, err := r.q.ListArticlesByTagFirstPage(ctx, ListArticlesByTagFirstPageParams{
			TagID: tagID,
			Limit: limit,
		})
		if err != nil {
			return nil, err
		}
		return mapTagSummaryRows(rows), nil
	}

	rows, err := r.q.ListArticlesByTagCursor(ctx, ListArticlesByTagCursorParams{
		TagID:         tagID,
		PublishedAt:   sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		PublishedAt_2: sql.NullTime{Time: cursor.PublishedAt, Valid: true},
		ID:            cursor.ID,
		Limit:         limit,
	})
	if err != nil {
		return nil, err
	}
	return mapTagCursorRows(rows), nil
}

func (r *articleRepo) ListHeadlines(ctx context.Context, limit int32) ([]*article.ArticleSummary, error) {
	rows, err := r.q.ListHeadlines(ctx, limit)
	if err != nil {
		return nil, err
	}
	res := make([]*article.ArticleSummary, len(rows))
	for i, row := range rows {
		var coverID *uint64
		if row.CoverMediaID.Valid {
			val := uint64(row.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if row.PublishedAt.Valid {
			pubAt = &row.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           row.ID,
			Title:        row.Title,
			Slug:         row.Slug,
			Excerpt:      row.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   row.CategoryID,
			AuthorID:     row.AuthorID,
			IsHeadline:   row.IsHeadline,
			IsBreaking:   row.IsBreaking,
			ViewCount:    row.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res, nil
}

func (r *articleRepo) ListBreaking(ctx context.Context, limit int32) ([]*article.ArticleSummary, error) {
	rows, err := r.q.ListBreakingNews(ctx, limit)
	if err != nil {
		return nil, err
	}
	res := make([]*article.ArticleSummary, len(rows))
	for i, row := range rows {
		var coverID *uint64
		if row.CoverMediaID.Valid {
			val := uint64(row.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if row.PublishedAt.Valid {
			pubAt = &row.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           row.ID,
			Title:        row.Title,
			Slug:         row.Slug,
			Excerpt:      row.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   row.CategoryID,
			AuthorID:     row.AuthorID,
			IsHeadline:   row.IsHeadline,
			IsBreaking:   row.IsBreaking,
			ViewCount:    row.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res, nil
}

func (r *articleRepo) SearchFulltext(ctx context.Context, query string, limit int32) ([]*article.ArticleSummary, error) {
	rows, err := r.q.SearchArticlesFulltext(ctx, SearchArticlesFulltextParams{
		Query: query,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*article.ArticleSummary, len(rows))
	for i, row := range rows {
		var coverID *uint64
		if row.CoverMediaID.Valid {
			val := uint64(row.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if row.PublishedAt.Valid {
			pubAt = &row.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           row.ID,
			Title:        row.Title,
			Slug:         row.Slug,
			Excerpt:      row.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   row.CategoryID,
			AuthorID:     row.AuthorID,
			IsHeadline:   row.IsHeadline,
			IsBreaking:   row.IsBreaking,
			ViewCount:    row.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res, nil
}

func (r *articleRepo) IncrementViews(ctx context.Context, id uint64, count int64) error {
	return r.q.IncrementViewCountBatch(ctx, IncrementViewCountBatchParams{
		ViewCount: uint64(count),
		ID:        id,
	})
}

func (r *articleRepo) RecordDailyViews(ctx context.Context, articleID uint64, viewDate time.Time, count int64) error {
	return r.q.UpsertDailyView(ctx, UpsertDailyViewParams{
		ArticleID: articleID,
		ViewDate:  viewDate,
		Views:     uint32(count),
	})
}

func (r *articleRepo) GetTrending(ctx context.Context, since time.Time, limit int32) ([]*article.ArticleSummary, error) {
	rows, err := r.q.GetTopArticlesByViewsPeriod(ctx, GetTopArticlesByViewsPeriodParams{
		ViewDate: since,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*article.ArticleSummary, len(rows))
	for i, row := range rows {
		var coverID *uint64
		if row.CoverMediaID.Valid {
			cid := uint64(row.CoverMediaID.Int64)
			coverID = &cid
		}
		var pubAt *time.Time
		if row.PublishedAt.Valid {
			pubAt = &row.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           row.ID,
			Title:        row.Title,
			Slug:         row.Slug,
			Excerpt:      row.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   row.CategoryID,
			AuthorID:     row.AuthorID,
			IsHeadline:   row.IsHeadline,
			IsBreaking:   row.IsBreaking,
			ViewCount:    row.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res, nil
}


func (r *articleRepo) GetScheduledToPublish(ctx context.Context, now time.Time) ([]*article.ArticleSummary, error) {
	rows, err := r.q.GetScheduledArticlesToPublish(ctx, sql.NullTime{Time: now, Valid: true})
	if err != nil {
		return nil, err
	}
	res := make([]*article.ArticleSummary, len(rows))
	for i, row := range rows {
		var pubAt *time.Time
		if row.PublishedAt.Valid {
			pubAt = &row.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:          row.ID,
			Title:       row.Title,
			Slug:        row.Slug,
			CategoryID:  row.CategoryID,
			AuthorID:    row.AuthorID,
			PublishedAt: pubAt,
		}
	}
	return res, nil
}

func (r *articleRepo) PublishScheduled(ctx context.Context, id uint64) error {
	return r.q.PublishScheduledArticle(ctx, id)
}

func (r *articleRepo) SetTags(ctx context.Context, articleID uint64, tagIDs []uint64) error {
	if err := r.q.ClearArticleTags(ctx, articleID); err != nil {
		return err
	}
	for _, tID := range tagIDs {
		if err := r.q.AddArticleTag(ctx, AddArticleTagParams{ArticleID: articleID, TagID: tID}); err != nil {
			return err
		}
	}
	return nil
}

func (r *articleRepo) GetTags(ctx context.Context, articleID uint64) ([]*tag.Tag, error) {
	rows, err := r.q.GetTagsByArticleID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	res := make([]*tag.Tag, len(rows))
	for i, row := range rows {
		res[i] = &tag.Tag{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return res, nil
}

func (r *articleRepo) CreateRevision(ctx context.Context, rev *article.ArticleRevision) (uint64, error) {
	res, err := r.q.CreateRevision(ctx, CreateRevisionParams{
		ArticleID: rev.ArticleID,
		UserID:    rev.UserID,
		Title:     rev.Title,
		Slug:      rev.Slug,
		Excerpt:   rev.Excerpt,
		Body:      rev.Body,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *articleRepo) ListRevisions(ctx context.Context, articleID uint64) ([]*article.ArticleRevision, error) {
	rows, err := r.q.ListRevisionsByArticleID(ctx, articleID)
	if err != nil {
		return nil, err
	}
	res := make([]*article.ArticleRevision, len(rows))
	for i, row := range rows {
		res[i] = &article.ArticleRevision{
			ID:        row.ID,
			ArticleID: row.ArticleID,
			UserID:    row.UserID,
			Title:     row.Title,
			Slug:      row.Slug,
			Excerpt:   row.Excerpt,
			Body:      row.Body,
			CreatedAt: row.CreatedAt,
		}
	}
	return res, nil
}

func (r *articleRepo) ListAdmin(ctx context.Context, limit int32, offset int32) ([]*article.Article, error) {
	rows, err := r.q.ListArticlesAdmin(ctx, ListArticlesAdminParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*article.Article, len(rows))
	for i, row := range rows {
		var coverID *uint64
		if row.CoverMediaID.Valid {
			cid := uint64(row.CoverMediaID.Int64)
			coverID = &cid
		}
		var pubAt *time.Time
		if row.PublishedAt.Valid {
			pubAt = &row.PublishedAt.Time
		}
		items[i] = &article.Article{
			ID:           row.ID,
			Title:        row.Title,
			Slug:         row.Slug,
			Excerpt:      row.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   row.CategoryID,
			AuthorID:     row.AuthorID,
			CreatedBy:    row.CreatedBy,
			Status:       article.Status(row.Status),
			IsHeadline:   row.IsHeadline,
			IsBreaking:   row.IsBreaking,
			ViewCount:    row.ViewCount,
			PublishedAt:  pubAt,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
	}
	return items, nil
}

func mapArticleRow(row *Articles) *article.Article {
	var coverID *uint64
	if row.CoverMediaID.Valid {
		val := uint64(row.CoverMediaID.Int64)
		coverID = &val
	}
	var pubAt *time.Time
	if row.PublishedAt.Valid {
		pubAt = &row.PublishedAt.Time
	}
	var delAt *time.Time
	if row.DeletedAt.Valid {
		delAt = &row.DeletedAt.Time
	}
	return &article.Article{
		ID:             row.ID,
		Title:          row.Title,
		Slug:           row.Slug,
		Excerpt:        row.Excerpt,
		Body:           row.Body,
		CoverMediaID:   coverID,
		CategoryID:     row.CategoryID,
		AuthorID:       row.AuthorID,
		CreatedBy:      row.CreatedBy,
		Status:         article.Status(row.Status),
		IsHeadline:     row.IsHeadline,
		IsBreaking:     row.IsBreaking,
		ViewCount:      row.ViewCount,
		PublishedAt:    pubAt,
		SEOTitle:       row.SeoTitle.String,
		SEODescription: row.SeoDescription.String,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		DeletedAt:      delAt,
	}
}

func mapArticleSummaryRows(rows []ListArticlesFirstPageRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapArticleCursorRows(rows []ListArticlesCursorRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapCategorySummaryRows(rows []ListArticlesByCategoryFirstPageRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapCategoryCursorRows(rows []ListArticlesByCategoryCursorRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapAuthorSummaryRows(rows []ListArticlesByAuthorFirstPageRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapAuthorCursorRows(rows []ListArticlesByAuthorCursorRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapTagSummaryRows(rows []ListArticlesByTagFirstPageRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

func mapTagCursorRows(rows []ListArticlesByTagCursorRow) []*article.ArticleSummary {
	res := make([]*article.ArticleSummary, len(rows))
	for i, r := range rows {
		var coverID *uint64
		if r.CoverMediaID.Valid {
			val := uint64(r.CoverMediaID.Int64)
			coverID = &val
		}
		var pubAt *time.Time
		if r.PublishedAt.Valid {
			pubAt = &r.PublishedAt.Time
		}
		res[i] = &article.ArticleSummary{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Excerpt:      r.Excerpt,
			CoverMediaID: coverID,
			CategoryID:   r.CategoryID,
			AuthorID:     r.AuthorID,
			IsHeadline:   r.IsHeadline,
			IsBreaking:   r.IsBreaking,
			ViewCount:    r.ViewCount,
			PublishedAt:  pubAt,
		}
	}
	return res
}

// ==========================================
// Redirect Repository
// ==========================================
type redirectRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *redirectRepo) GetByFromPath(ctx context.Context, fromPath string) (*redirect.Redirect, error) {
	row, err := r.q.GetRedirectByFromPath(ctx, fromPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &redirect.Redirect{
		ID:         row.ID,
		FromPath:   row.FromPath,
		ToPath:     row.ToPath,
		HTTPStatus: row.HttpStatus,
		HitCount:   row.HitCount,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *redirectRepo) CreateOrUpdate(ctx context.Context, fromPath string, toPath string, status int16) error {
	_, err := r.q.CreateRedirect(ctx, CreateRedirectParams{
		FromPath:   fromPath,
		ToPath:     toPath,
		HttpStatus: status,
	})
	return err
}

func (r *redirectRepo) IncrementHit(ctx context.Context, id uint64) error {
	return r.q.IncrementRedirectHit(ctx, id)
}

func (r *redirectRepo) List(ctx context.Context, limit int32, offset int32) ([]*redirect.Redirect, error) {
	rows, err := r.q.ListRedirects(ctx, ListRedirectsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*redirect.Redirect, len(rows))
	for i, row := range rows {
		res[i] = &redirect.Redirect{
			ID:         row.ID,
			FromPath:   row.FromPath,
			ToPath:     row.ToPath,
			HTTPStatus: row.HttpStatus,
			HitCount:   row.HitCount,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		}
	}
	return res, nil
}

// ==========================================
// Page Repository
// ==========================================
type pageRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *pageRepo) GetBySlug(ctx context.Context, slug string) (*page.Page, error) {
	row, err := r.q.GetPageBySlug(ctx, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &page.Page{
		ID:          row.ID,
		Title:       row.Title,
		Slug:        row.Slug,
		Content:     row.Content,
		IsPublished: row.IsPublished,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *pageRepo) List(ctx context.Context) ([]*page.Page, error) {
	rows, err := r.q.ListPages(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*page.Page, len(rows))
	for i, row := range rows {
		res[i] = &page.Page{
			ID:          row.ID,
			Title:       row.Title,
			Slug:        row.Slug,
			IsPublished: row.IsPublished,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
	}
	return res, nil
}

func (r *pageRepo) Create(ctx context.Context, p *page.Page) (uint64, error) {
	res, err := r.q.CreatePage(ctx, CreatePageParams{
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		IsPublished: p.IsPublished,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *pageRepo) Update(ctx context.Context, p *page.Page) error {
	return r.q.UpdatePage(ctx, UpdatePageParams{
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		IsPublished: p.IsPublished,
		ID:          p.ID,
	})
}

func (r *pageRepo) Delete(ctx context.Context, id uint64) error {
	return r.q.DeletePage(ctx, id)
}

// ==========================================
// Setting Repository
// ==========================================
type settingRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *settingRepo) Get(ctx context.Context, keyName string) (*setting.Setting, error) {
	row, err := r.q.GetSetting(ctx, keyName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerrors.ErrNotFound
		}
		return nil, err
	}
	return &setting.Setting{
		KeyName:   row.KeyName,
		Value:     row.Value,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *settingRepo) Upsert(ctx context.Context, keyName string, val []byte) error {
	return r.q.UpsertSetting(ctx, UpsertSettingParams{
		KeyName: keyName,
		Value:   val,
	})
}

func (r *settingRepo) List(ctx context.Context) ([]*setting.Setting, error) {
	rows, err := r.q.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*setting.Setting, len(rows))
	for i, row := range rows {
		res[i] = &setting.Setting{
			KeyName:   row.KeyName,
			Value:     row.Value,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return res, nil
}

// ==========================================
// Audit Repository
// ==========================================
type auditRepo struct {
	db *sql.DB
	q  *Queries
}

func (r *auditRepo) Create(ctx context.Context, l *audit.AuditLog) (uint64, error) {
	var userID sql.NullInt64
	if l.UserID != nil {
		userID = sql.NullInt64{Int64: int64(*l.UserID), Valid: true}
	}
	res, err := r.q.CreateAuditLog(ctx, CreateAuditLogParams{
		UserID:      userID,
		Action:      l.Action,
		Entity:      l.Entity,
		EntityID:    l.EntityID,
		DiffSummary: l.DiffSummary,
		IpAddress:   l.IPAddress,
		UserAgent:   l.UserAgent,
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (r *auditRepo) List(ctx context.Context, limit int32, offset int32) ([]*audit.AuditLog, error) {
	rows, err := r.q.ListAuditLogs(ctx, ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*audit.AuditLog, len(rows))
	for i, row := range rows {
		var uID *uint64
		if row.UserID.Valid {
			val := uint64(row.UserID.Int64)
			uID = &val
		}
		res[i] = &audit.AuditLog{
			ID:          row.ID,
			UserID:      uID,
			Action:      row.Action,
			Entity:      row.Entity,
			EntityID:    row.EntityID,
			DiffSummary: row.DiffSummary,
			IPAddress:   row.IpAddress,
			UserAgent:   row.UserAgent,
			CreatedAt:   row.CreatedAt,
		}
	}
	return res, nil
}

// Compile-time checks to verify interface implementations
var (
	_ user.Repository     = (*userRepo)(nil)
	_ category.Repository = (*categoryRepo)(nil)
	_ tag.Repository      = (*tagRepo)(nil)
	_ author.Repository   = (*authorRepo)(nil)
	_ media.Repository    = (*mediaRepo)(nil)
	_ article.Repository  = (*articleRepo)(nil)
	_ redirect.Repository = (*redirectRepo)(nil)
	_ page.Repository     = (*pageRepo)(nil)
	_ setting.Repository  = (*settingRepo)(nil)
	_ audit.Repository    = (*auditRepo)(nil)
)
