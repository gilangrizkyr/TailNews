package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/domain/article"
	"news-portal/apps/api/internal/domain/page"
	"news-portal/apps/api/internal/pkg/auth"
	"news-portal/apps/api/internal/platform"
	"news-portal/apps/api/internal/repository/mysql"
	transporthttp "news-portal/apps/api/internal/transport/http"
	"news-portal/apps/api/internal/transport/http/dto"
)

func setupPublicTestServer(t *testing.T) (*httptest.Server, *http.Client, *platform.RedisCache, *platform.RedisViewCounter, *mysql.Repository) {
	cfg, err := config.Load()
	require.NoError(t, err)

	tempDir := t.TempDir()
	cfg.MediaStorageDir = tempDir

	ctx := context.Background()
	db, err := platform.NewMySQLPool(ctx, cfg)
	require.NoError(t, err)

	rdb, err := platform.NewRedisClient(ctx, cfg)
	require.NoError(t, err)

	storage, err := platform.NewLocalStorage(tempDir)
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := mysql.NewRepository(db)
	cache := platform.NewRedisCache(rdb, logger)
	viewCounter := platform.NewRedisViewCounter(rdb, repo.Article, logger)

	handler := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config:      cfg,
		Logger:      logger,
		DB:          db,
		Redis:       rdb,
		Storage:     storage,
		Cache:       cache,
		ViewCounter: viewCounter,
	})

	server := httptest.NewServer(handler)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	t.Cleanup(func() {
		server.Close()
		_ = db.Close()
		_ = rdb.Close()
	})

	return server, client, cache, viewCounter, repo
}

func TestPublicAPI_Homepage(t *testing.T) {
	server, client, cache, _, _ := setupPublicTestServer(t)
	ctx := context.Background()

	// Invalidate home cache first to verify fresh query
	_ = cache.Delete(ctx, "news:v1:feed:home")

	resp, err := client.Get(server.URL + "/api/v1/homepage")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.True(t, env.Success)

	var home dto.HomepageResponse
	require.NoError(t, json.Unmarshal(env.Data, &home))

	assert.NotEmpty(t, home.Headlines, "Homepage should contain headlines")
	assert.NotEmpty(t, home.Latest, "Homepage should contain latest articles")
	assert.NotEmpty(t, home.Categories, "Homepage should contain categories")
	assert.NotEmpty(t, home.NextCursor, "Homepage should have a next_cursor")

	// Verify L2 cache hit on second request
	var cachedHome dto.HomepageResponse
	hit, err := cache.Get(ctx, "news:v1:feed:home", &cachedHome)
	assert.NoError(t, err)
	assert.True(t, hit, "Homepage should be stored in Redis L2 cache")
	assert.Equal(t, len(home.Headlines), len(cachedHome.Headlines))
}

func TestPublicAPI_Articles_CursorPagination(t *testing.T) {
	server, client, _, _, _ := setupPublicTestServer(t)

	// Page 1: Limit 5
	resp1, err := client.Get(server.URL + "/api/v1/articles?limit=5")
	require.NoError(t, err)
	defer resp1.Body.Close()

	require.Equal(t, http.StatusOK, resp1.StatusCode)
	var env1 Envelope
	require.NoError(t, json.NewDecoder(resp1.Body).Decode(&env1))

	var page1 dto.CursorArticlesResponse
	require.NoError(t, json.Unmarshal(env1.Data, &page1))

	assert.Len(t, page1.Articles, 5)
	assert.True(t, page1.HasMore)
	assert.NotEmpty(t, page1.NextCursor)

	// Page 2: Using NextCursor
	resp2, err := client.Get(fmt.Sprintf("%s/api/v1/articles?cursor=%s&limit=5", server.URL, page1.NextCursor))
	require.NoError(t, err)
	defer resp2.Body.Close()

	require.Equal(t, http.StatusOK, resp2.StatusCode)
	var env2 Envelope
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&env2))

	var page2 dto.CursorArticlesResponse
	require.NoError(t, json.Unmarshal(env2.Data, &page2))

	assert.Len(t, page2.Articles, 5)

	// Verify zero overlap between page 1 and page 2
	page1IDs := make(map[uint64]bool)
	for _, a := range page1.Articles {
		page1IDs[a.ID] = true
	}

	for _, a := range page2.Articles {
		assert.False(t, page1IDs[a.ID], "Article ID %d appears in both page 1 and page 2; keyset cursor violated", a.ID)
	}
}

func TestPublicAPI_Article_Detail_And_PreviewToken(t *testing.T) {
	server, client, _, _, repo := setupPublicTestServer(t)
	ctx := context.Background()

	// 1. Fetch existing published article from seeded database
	publishedArticles, err := repo.Article.ListHeadlines(ctx, 1)
	require.NoError(t, err)
	require.NotEmpty(t, publishedArticles)

	pubSlug := publishedArticles[0].Slug
	resp, err := client.Get(fmt.Sprintf("%s/api/v1/articles/%s", server.URL, pubSlug))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	var art dto.ArticleResponse
	require.NoError(t, json.Unmarshal(env.Data, &art))
	assert.Equal(t, pubSlug, art.Slug)
	assert.NotEmpty(t, art.Title)
	assert.NotEmpty(t, art.CategoryName)

	// 2. Non-existent slug -> 404
	badResp, err := client.Get(server.URL + "/api/v1/articles/this-slug-definitely-does-not-exist-999")
	require.NoError(t, err)
	defer badResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, badResp.StatusCode)

	// 3. Draft article cannot be read publicly without preview token
	draftSlug := fmt.Sprintf("draft-article-%d", time.Now().UnixNano())
	draftID, err := repo.Article.Create(ctx, &article.Article{
		Title:      "Artikel Draf Rahasia",
		Slug:       draftSlug,
		Excerpt:    "Draf eksklusif",
		Body:       "<p>Isi draf rahasia.</p>",
		CategoryID: 1,
		AuthorID:   1,
		CreatedBy:  1,
		Status:     article.StatusDraft,
	})
	require.NoError(t, err)

	unauthDraftResp, err := client.Get(fmt.Sprintf("%s/api/v1/articles/%s", server.URL, draftSlug))
	require.NoError(t, err)
	defer unauthDraftResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, unauthDraftResp.StatusCode, "Draft article without token should return 404 to reader")

	// 4. Draft article with valid signed preview token -> 200 OK
	cfg, _ := config.Load()
	token, err := transporthttpPreviewToken(cfg.JWTSecret, draftID)
	require.NoError(t, err)

	previewResp, err := client.Get(fmt.Sprintf("%s/api/v1/articles/%s?token=%s", server.URL, draftSlug, token))
	require.NoError(t, err)
	defer previewResp.Body.Close()

	assert.Equal(t, http.StatusOK, previewResp.StatusCode)
	var prevEnv Envelope
	require.NoError(t, json.NewDecoder(previewResp.Body).Decode(&prevEnv))
	var prevArt dto.ArticleResponse
	require.NoError(t, json.Unmarshal(prevEnv.Data, &prevArt))
	assert.Equal(t, draftSlug, prevArt.Slug)
	assert.Equal(t, "draft", prevArt.Status)
}

func TestPublicAPI_Taxonomies_And_Author_Feeds(t *testing.T) {
	server, client, _, _, _ := setupPublicTestServer(t)

	// 1. Categories
	catResp, err := client.Get(server.URL + "/api/v1/categories")
	require.NoError(t, err)
	defer catResp.Body.Close()
	assert.Equal(t, http.StatusOK, catResp.StatusCode)

	var catEnv Envelope
	require.NoError(t, json.NewDecoder(catResp.Body).Decode(&catEnv))
	var categories []dto.CategoryResponse
	require.NoError(t, json.Unmarshal(catEnv.Data, &categories))
	require.NotEmpty(t, categories)

	firstCat := categories[0]

	// 2. Articles by Category
	catArtResp, err := client.Get(fmt.Sprintf("%s/api/v1/categories/%s/articles?limit=5", server.URL, firstCat.Slug))
	require.NoError(t, err)
	defer catArtResp.Body.Close()
	assert.Equal(t, http.StatusOK, catArtResp.StatusCode)

	var catArtEnv Envelope
	require.NoError(t, json.NewDecoder(catArtResp.Body).Decode(&catArtEnv))
	var catArts dto.CursorArticlesResponse
	require.NoError(t, json.Unmarshal(catArtEnv.Data, &catArts))
	assert.NotEmpty(t, catArts.Articles)

	// 3. Author Profile & Feed
	authorResp, err := client.Get(server.URL + "/api/v1/authors/rina-wijaya")
	require.NoError(t, err)
	defer authorResp.Body.Close()
	assert.Equal(t, http.StatusOK, authorResp.StatusCode)

	var authEnv Envelope
	require.NoError(t, json.NewDecoder(authorResp.Body).Decode(&authEnv))
	var authorProfile dto.AuthorProfileResponse
	require.NoError(t, json.Unmarshal(authEnv.Data, &authorProfile))
	assert.Equal(t, "rina-wijaya", authorProfile.Author.Slug)
	assert.NotEmpty(t, authorProfile.Articles)
}

func TestPublicAPI_Search_Trending_And_Pages(t *testing.T) {
	server, client, _, _, repo := setupPublicTestServer(t)
	ctx := context.Background()

	// 1. Full-text Search
	searchResp, err := client.Get(server.URL + "/api/v1/search?q=pemerintah&limit=5")
	require.NoError(t, err)
	defer searchResp.Body.Close()
	assert.Equal(t, http.StatusOK, searchResp.StatusCode)

	var searchEnv Envelope
	require.NoError(t, json.NewDecoder(searchResp.Body).Decode(&searchEnv))
	var searchResults []dto.ArticleResponse
	require.NoError(t, json.Unmarshal(searchEnv.Data, &searchResults))
	assert.NotEmpty(t, searchResults)

	// 2. Trending Articles
	trendingResp, err := client.Get(server.URL + "/api/v1/trending?window=today&limit=5")
	require.NoError(t, err)
	defer trendingResp.Body.Close()
	assert.Equal(t, http.StatusOK, trendingResp.StatusCode)

	var trendEnv Envelope
	require.NoError(t, json.NewDecoder(trendingResp.Body).Decode(&trendEnv))
	var trendingResults []dto.ArticleResponse
	require.NoError(t, json.Unmarshal(trendEnv.Data, &trendingResults))
	assert.NotEmpty(t, trendingResults)

	// 3. Static Editorial Page
	pageSlug := fmt.Sprintf("tentang-kami-%d", time.Now().UnixNano())
	_, err = repo.Page.Create(ctx, &page.Page{
		Title:       "Tentang Portal Berita",
		Slug:        pageSlug,
		Content:     "<p>Portal berita terpercaya dan independen.</p>",
		IsPublished: true,
	})
	require.NoError(t, err)

	pageResp, err := client.Get(fmt.Sprintf("%s/api/v1/pages/%s", server.URL, pageSlug))
	require.NoError(t, err)
	defer pageResp.Body.Close()
	assert.Equal(t, http.StatusOK, pageResp.StatusCode)

	var pageEnv Envelope
	require.NoError(t, json.NewDecoder(pageResp.Body).Decode(&pageEnv))
	var pageData dto.PageResponse
	require.NoError(t, json.Unmarshal(pageEnv.Data, &pageData))
	assert.Equal(t, pageSlug, pageData.Slug)
	assert.Contains(t, pageData.Content, "Portal berita terpercaya")
}

func TestPublicAPI_ViewCounter_Deduplication_And_BatchFlush(t *testing.T) {
	server, _, _, viewCounter, repo := setupPublicTestServer(t)
	ctx := context.Background()

	// Get an article
	headlines, err := repo.Article.ListHeadlines(ctx, 1)
	require.NoError(t, err)
	require.NotEmpty(t, headlines)
	artID := headlines[0].ID
	initialViews := headlines[0].ViewCount

	// View 1: IP A, UA A -> should record
	req1, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/articles/%d/view", server.URL, artID), nil)
	req1.RemoteAddr = "192.168.1.100:1234"
	req1.Header.Set("User-Agent", "Mozilla/5.0 ClientA")
	resp1, err := http.DefaultClient.Do(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()

	var env1 Envelope
	require.NoError(t, json.NewDecoder(resp1.Body).Decode(&env1))
	var res1 dto.RecordViewResponse
	require.NoError(t, json.Unmarshal(env1.Data, &res1))
	assert.True(t, res1.Recorded, "First view should be recorded")

	// View 2: Same IP A, same UA A immediately -> deduplicated!
	req2, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/articles/%d/view", server.URL, artID), nil)
	req2.RemoteAddr = "192.168.1.100:1234"
	req2.Header.Set("User-Agent", "Mozilla/5.0 ClientA")
	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	var env2 Envelope
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&env2))
	var res2 dto.RecordViewResponse
	require.NoError(t, json.Unmarshal(env2.Data, &res2))
	assert.False(t, res2.Recorded, "Second view from same IP and UA should be deduplicated")

	// View 3: Different IP B, UA B -> should record
	req3, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/articles/%d/view", server.URL, artID), nil)
	req3.RemoteAddr = "192.168.1.200:5678"
	req3.Header.Set("User-Agent", "Mozilla/5.0 ClientB")
	resp3, err := http.DefaultClient.Do(req3)
	require.NoError(t, err)
	defer resp3.Body.Close()

	var env3 Envelope
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&env3))
	var res3 dto.RecordViewResponse
	require.NoError(t, json.Unmarshal(env3.Data, &res3))
	assert.True(t, res3.Recorded, "View from different IP/UA should be recorded")

	// Flush batch to MySQL
	flushed, err := viewCounter.Flush(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, flushed, 1)

	// Verify database record has updated views (+2 total recorded)
	updatedArt, err := repo.Article.GetByID(ctx, artID)
	require.NoError(t, err)
	assert.Equal(t, initialViews+2, updatedArt.ViewCount)
}

func TestPublicAPI_Singleflight_Concurrency(t *testing.T) {
	server, client, cache, _, repo := setupPublicTestServer(t)
	ctx := context.Background()

	headlines, err := repo.Article.ListHeadlines(ctx, 1)
	require.NoError(t, err)
	slug := headlines[0].Slug

	// Invalidate cache key to force concurrent cache miss
	_ = cache.Delete(ctx, fmt.Sprintf("news:v1:article:slug:%s", slug))

	concurrency := 20
	var wg sync.WaitGroup
	wg.Add(concurrency)

	errorsChan := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			resp, err := client.Get(fmt.Sprintf("%s/api/v1/articles/%s", server.URL, slug))
			if err != nil {
				errorsChan <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errorsChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}
		}()
	}

	wg.Wait()
	close(errorsChan)

	for err := range errorsChan {
		require.NoError(t, err)
	}
}

// Helper: generate signed preview token for test
func transporthttpPreviewToken(secret string, articleID uint64) (string, error) {
	return auth.GeneratePreviewToken(secret, articleID, 2*time.Hour)
}
