package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/domain/article"
	"news-portal/apps/api/internal/domain/author"
	"news-portal/apps/api/internal/domain/category"
	"news-portal/apps/api/internal/domain/page"
	"news-portal/apps/api/internal/domain/tag"
	"news-portal/apps/api/internal/platform"
	"news-portal/apps/api/internal/repository/mysql"
)

func setupTestRepo(t *testing.T) (*mysql.Repository, context.Context) {
	cfg, err := config.Load()
	require.NoError(t, err)

	ctx := context.Background()
	db, err := platform.NewMySQLPool(ctx, cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return mysql.NewRepository(db), ctx
}

func TestCategoryRepository_CRUD(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	slug := fmt.Sprintf("test-category-%d", time.Now().UnixNano())
	cat := &category.Category{
		Name:        "Test Category",
		Slug:        slug,
		Description: "Kategori pengujian sistem",
		SortOrder:   99,
		IsNav:       true,
	}

	// Create
	id, err := repo.Category.Create(ctx, cat)
	require.NoError(t, err)
	assert.Greater(t, id, uint64(0))

	// GetByID
	fetched, err := repo.Category.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, cat.Name, fetched.Name)
	assert.Equal(t, slug, fetched.Slug)

	// GetBySlug
	fetchedSlug, err := repo.Category.GetBySlug(ctx, slug)
	require.NoError(t, err)
	assert.Equal(t, id, fetchedSlug.ID)

	// Update
	fetched.Name = "Updated Test Category"
	err = repo.Category.Update(ctx, fetched)
	require.NoError(t, err)

	updated, err := repo.Category.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Test Category", updated.Name)

	// Delete
	err = repo.Category.Delete(ctx, id)
	require.NoError(t, err)

	_, err = repo.Category.GetByID(ctx, id)
	assert.Error(t, err)
}

func TestTagRepository_CRUD(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	slug := fmt.Sprintf("test-tag-%d", time.Now().UnixNano())
	tObj := &tag.Tag{
		Name: "Test Tag",
		Slug: slug,
	}

	id, err := repo.Tag.Create(ctx, tObj)
	require.NoError(t, err)
	assert.Greater(t, id, uint64(0))

	fetched, err := repo.Tag.GetBySlug(ctx, slug)
	require.NoError(t, err)
	assert.Equal(t, tObj.Name, fetched.Name)

	// GetOrCreate
	existing, err := repo.Tag.GetOrCreate(ctx, "Test Tag", slug)
	require.NoError(t, err)
	assert.Equal(t, id, existing.ID)

	// Clean up
	err = repo.Tag.Delete(ctx, id)
	require.NoError(t, err)
}

func TestAuthorRepository_CRUD(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	slug := fmt.Sprintf("test-author-%d", time.Now().UnixNano())
	a := &author.Author{
		Name: "Test Author",
		Slug: slug,
		Bio:  "Penulis pengujian integrasi",
	}

	id, err := repo.Author.Create(ctx, a)
	require.NoError(t, err)
	assert.Greater(t, id, uint64(0))

	fetched, err := repo.Author.GetBySlug(ctx, slug)
	require.NoError(t, err)
	assert.Equal(t, a.Name, fetched.Name)

	fetched.Bio = "Bio yang diperbarui"
	err = repo.Author.Update(ctx, fetched)
	require.NoError(t, err)

	updated, err := repo.Author.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Bio yang diperbarui", updated.Bio)
}

func TestArticleRepository_KeysetCursor(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	// 1. Fetch First Page (limit 5)
	firstPage, err := repo.Article.ListCursor(ctx, nil, 5)
	require.NoError(t, err)
	require.Len(t, firstPage, 5)

	// Ensure sorted descending
	for i := 0; i < len(firstPage)-1; i++ {
		require.NotNil(t, firstPage[i].PublishedAt)
		require.NotNil(t, firstPage[i+1].PublishedAt)
		assert.True(t, firstPage[i].PublishedAt.After(*firstPage[i+1].PublishedAt) ||
			(firstPage[i].PublishedAt.Equal(*firstPage[i+1].PublishedAt) && firstPage[i].ID > firstPage[i+1].ID))
	}

	// 2. Fetch Second Page using cursor from last item
	lastItem := firstPage[len(firstPage)-1]
	cursor := &article.Cursor{
		PublishedAt: *lastItem.PublishedAt,
		ID:          lastItem.ID,
	}

	secondPage, err := repo.Article.ListCursor(ctx, cursor, 5)
	require.NoError(t, err)
	require.Len(t, secondPage, 5)

	// Second page items must be strictly older than lastItem
	assert.True(t, secondPage[0].PublishedAt.Before(*lastItem.PublishedAt) ||
		(secondPage[0].PublishedAt.Equal(*lastItem.PublishedAt) && secondPage[0].ID < lastItem.ID))
}

func TestArticleRepository_Feeds(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	headlines, err := repo.Article.ListHeadlines(ctx, 3)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(headlines), 3)

	breaking, err := repo.Article.ListBreaking(ctx, 3)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(breaking), 3)
}

func TestArticleRepository_FulltextSearch(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	// Search for 'Politik' or 'Ekonomi'
	results, err := repo.Article.SearchFulltext(ctx, "+Kebijakan +Sektor", 5)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestArticleRepository_RevisionsAndTags(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	// Grab any existing article
	articles, err := repo.Article.ListCursor(ctx, nil, 1)
	require.NoError(t, err)
	require.NotEmpty(t, articles)

	targetID := articles[0].ID

	// Create revision
	rev := &article.ArticleRevision{
		ArticleID: targetID,
		UserID:    1,
		Title:     "Revisi Judul Artikel",
		Slug:      articles[0].Slug,
		Excerpt:   "Revisi excerpt",
		Body:      "<p>Konten revisi.</p>",
	}
	revID, err := repo.Article.CreateRevision(ctx, rev)
	require.NoError(t, err)
	assert.Greater(t, revID, uint64(0))

	revisions, err := repo.Article.ListRevisions(ctx, targetID)
	require.NoError(t, err)
	assert.NotEmpty(t, revisions)
	assert.Equal(t, "Revisi Judul Artikel", revisions[0].Title)

	// Tags
	err = repo.Article.SetTags(ctx, targetID, []uint64{1, 2})
	require.NoError(t, err)

	tags, err := repo.Article.GetTags(ctx, targetID)
	require.NoError(t, err)
	assert.Len(t, tags, 2)
}

func TestRedirectRepository_CRUD(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	fromPath := fmt.Sprintf("/old-path-%d", time.Now().UnixNano())
	toPath := "/new-path"

	err := repo.Redirect.CreateOrUpdate(ctx, fromPath, toPath, 301)
	require.NoError(t, err)

	r, err := repo.Redirect.GetByFromPath(ctx, fromPath)
	require.NoError(t, err)
	assert.Equal(t, toPath, r.ToPath)
	assert.Equal(t, int16(301), r.HTTPStatus)

	err = repo.Redirect.IncrementHit(ctx, r.ID)
	require.NoError(t, err)

	updated, err := repo.Redirect.GetByFromPath(ctx, fromPath)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), updated.HitCount)
}

func TestPageRepository_CRUD(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	slug := fmt.Sprintf("test-page-%d", time.Now().UnixNano())
	p := &page.Page{
		Title:       "Halaman Uji",
		Slug:        slug,
		Content:     "<p>Isi halaman uji integrasi.</p>",
		IsPublished: true,
	}

	id, err := repo.Page.Create(ctx, p)
	require.NoError(t, err)
	assert.Greater(t, id, uint64(0))

	fetched, err := repo.Page.GetBySlug(ctx, slug)
	require.NoError(t, err)
	assert.Equal(t, p.Title, fetched.Title)

	err = repo.Page.Delete(ctx, id)
	require.NoError(t, err)
}

func TestSettingRepository_CRUD(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	key := fmt.Sprintf("test_setting_%d", time.Now().UnixNano())
	val := []byte(`{"enabled":true,"limit":50}`)

	err := repo.Setting.Upsert(ctx, key, val)
	require.NoError(t, err)

	s, err := repo.Setting.Get(ctx, key)
	require.NoError(t, err)
	assert.JSONEq(t, string(val), string(s.Value))
}
