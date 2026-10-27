package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/platform"
	transporthttp "news-portal/apps/api/internal/transport/http"
	"news-portal/apps/api/internal/transport/http/dto"
)

type Envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Meta map[string]interface{} `json:"meta"`
}

func setupAdminTestServer(t *testing.T) (*httptest.Server, *http.Client, string) {
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

	handler := transporthttp.NewRouter(transporthttp.RouterConfig{
		Config:  cfg,
		Logger:  logger,
		DB:      db,
		Redis:   rdb,
		Storage: storage,
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

	return server, client, tempDir
}

func loginUser(t *testing.T, serverURL string, client *http.Client, email, password string) (string, dto.UserResponse) {
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(loginPayload)

	resp, err := client.Post(serverURL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	require.True(t, env.Success)

	var loginData dto.AuthResponse
	require.NoError(t, json.Unmarshal(env.Data, &loginData))
	require.NotEmpty(t, loginData.AccessToken)

	return loginData.AccessToken, loginData.User
}

func TestAdminAPI_Auth_Login_Success_And_Failure(t *testing.T) {
	server, client, _ := setupAdminTestServer(t)

	// 1. Valid Admin Login
	adminToken, adminUser := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")
	assert.NotEmpty(t, adminToken)
	assert.Equal(t, "Super Admin", adminUser.Name)
	assert.Equal(t, "admin", adminUser.RoleName)

	// 2. Invalid Password
	badPayload, _ := json.Marshal(map[string]string{
		"email":    "admin@newsportal.id",
		"password": "WrongPassword999!",
	})
	badResp, err := client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(badPayload))
	require.NoError(t, err)
	defer badResp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, badResp.StatusCode)

	// 3. Non-existent User
	noUserPayload, _ := json.Marshal(map[string]string{
		"email":    "nonexistent@newsportal.id",
		"password": "AnyPassword123!",
	})
	noUserResp, err := client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(noUserPayload))
	require.NoError(t, err)
	defer noUserResp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, noUserResp.StatusCode)
}

func TestAdminAPI_Auth_Me_Refresh_And_Logout(t *testing.T) {
	server, client, _ := setupAdminTestServer(t)

	adminToken, _ := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")

	// 1. GET /auth/me with Bearer Token
	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/auth/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.True(t, env.Success)

	var me dto.UserResponse
	require.NoError(t, json.Unmarshal(env.Data, &me))
	assert.Equal(t, "admin@newsportal.id", me.Email)

	// 2. GET /auth/me without Bearer Token -> Unauthorized
	unauthReq, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/auth/me", nil)
	unauthResp, err := client.Do(unauthReq)
	require.NoError(t, err)
	defer unauthResp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, unauthResp.StatusCode)

	// 3. POST /auth/refresh using CookieJar stored refresh_token
	refreshResp, err := client.Post(server.URL+"/api/v1/auth/refresh", "application/json", nil)
	require.NoError(t, err)
	defer refreshResp.Body.Close()

	assert.Equal(t, http.StatusOK, refreshResp.StatusCode)
	var refreshEnv Envelope
	require.NoError(t, json.NewDecoder(refreshResp.Body).Decode(&refreshEnv))
	assert.True(t, refreshEnv.Success)

	var refreshData dto.AuthResponse
	require.NoError(t, json.Unmarshal(refreshEnv.Data, &refreshData))
	assert.NotEmpty(t, refreshData.AccessToken)

	// 4. POST /auth/logout -> clears cookie
	logoutResp, err := client.Post(server.URL+"/api/v1/auth/logout", "application/json", nil)
	require.NoError(t, err)
	defer logoutResp.Body.Close()
	assert.Equal(t, http.StatusOK, logoutResp.StatusCode)
}

func TestAdminAPI_RBAC_ArticleDeletion(t *testing.T) {
	server, client, _ := setupAdminTestServer(t)

	// 1. Login as Writer (Role: writer)
	writerToken, writerUser := loginUser(t, server.URL, client, "rina.writer@newsportal.id", "AdminSecret123!")
	assert.Equal(t, "writer", writerUser.RoleName)

	// 2. Writer creates an article
	createReq := dto.ArticleUpsertRequest{
		Title:      fmt.Sprintf("Writer Article %d", time.Now().UnixNano()),
		Excerpt:    "A test excerpt written by author",
		Body:       "<p>Valid body content</p>",
		CategoryID: 1,
		AuthorID:   1,
		Status:     "draft",
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/articles", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+writerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	var createdArt dto.ArticleResponse
	require.NoError(t, json.Unmarshal(env.Data, &createdArt))
	artID := createdArt.ID

	// 3. Writer attempts to DELETE article -> Forbidden (403)
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/admin/articles/%d", server.URL, artID), nil)
	delReq.Header.Set("Authorization", "Bearer "+writerToken)
	delResp, err := client.Do(delReq)
	require.NoError(t, err)
	defer delResp.Body.Close()

	assert.Equal(t, http.StatusForbidden, delResp.StatusCode)

	// 4. Admin deletes article -> OK (200)
	adminToken, _ := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")
	adminDelReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/admin/articles/%d", server.URL, artID), nil)
	adminDelReq.Header.Set("Authorization", "Bearer "+adminToken)
	adminDelResp, err := client.Do(adminDelReq)
	require.NoError(t, err)
	defer adminDelResp.Body.Close()

	assert.Equal(t, http.StatusOK, adminDelResp.StatusCode)
}

func TestAdminAPI_Article_FullLifecycle_And_301Redirect(t *testing.T) {
	server, client, _ := setupAdminTestServer(t)

	adminToken, _ := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")

	// 1. Create Article with unsanitized HTML (verify bluemonday sanitization)
	origSlug := fmt.Sprintf("slug-v1-%d", time.Now().UnixNano())
	createReq := dto.ArticleUpsertRequest{
		Title:      "Investigasi Teknologi Terbaru",
		Slug:       origSlug,
		Excerpt:    "Ringkasan artikel investigasi mendalam",
		Body:       "<p>Berita asli.</p><script>alert('xss attack');</script><iframe src=\"https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ\"></iframe>",
		CategoryID: 1,
		AuthorID:   1,
		Status:     "published",
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/articles", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	var createdArt dto.ArticleResponse
	require.NoError(t, json.Unmarshal(env.Data, &createdArt))

	assert.Equal(t, origSlug, createdArt.Slug)
	// Verify script was stripped by bluemonday, but youtube iframe remains
	assert.NotContains(t, createdArt.Body, "<script>")
	assert.Contains(t, createdArt.Body, "youtube-nocookie.com")

	artID := createdArt.ID

	// 2. Update Article Slug (Should trigger automatic 301 redirect entry)
	newSlug := fmt.Sprintf("slug-v2-%d", time.Now().UnixNano())
	updateReq := dto.ArticleUpsertRequest{
		Title:      "Investigasi Teknologi Terbaru (Revisi)",
		Slug:       newSlug,
		Excerpt:    createdArt.Excerpt,
		Body:       "<p>Konten hasil revisi mendalam.</p>",
		CategoryID: 1,
		AuthorID:   1,
		Status:     "published",
	}
	updBody, _ := json.Marshal(updateReq)
	updReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/admin/articles/%d", server.URL, artID), bytes.NewReader(updBody))
	updReq.Header.Set("Authorization", "Bearer "+adminToken)
	updReq.Header.Set("Content-Type", "application/json")

	updResp, err := client.Do(updReq)
	require.NoError(t, err)
	defer updResp.Body.Close()

	require.Equal(t, http.StatusOK, updResp.StatusCode)

	// 3. Verify Redirects API returns the automatic 301 redirect
	redirReq, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/admin/redirects", nil)
	redirReq.Header.Set("Authorization", "Bearer "+adminToken)
	redirResp, err := client.Do(redirReq)
	require.NoError(t, err)
	defer redirResp.Body.Close()

	require.Equal(t, http.StatusOK, redirResp.StatusCode)
	var redirEnv Envelope
	require.NoError(t, json.NewDecoder(redirResp.Body).Decode(&redirEnv))
	var redirs []dto.RedirectResponse
	require.NoError(t, json.Unmarshal(redirEnv.Data, &redirs))

	foundRedirect := false
	expectedSource := "/" + origSlug
	expectedTarget := "/" + newSlug
	for _, r := range redirs {
		if r.FromPath == expectedSource && r.ToPath == expectedTarget {
			foundRedirect = true
			assert.Equal(t, int16(301), r.HTTPStatus)
			break
		}
	}
	assert.True(t, foundRedirect, "Expected 301 redirect was not created automatically")

	// 4. Test Article Revisions
	revReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/admin/articles/%d/revisions", server.URL, artID), nil)
	revReq.Header.Set("Authorization", "Bearer "+adminToken)
	revResp, err := client.Do(revReq)
	require.NoError(t, err)
	defer revResp.Body.Close()

	assert.Equal(t, http.StatusOK, revResp.StatusCode)
	var revEnv Envelope
	require.NoError(t, json.NewDecoder(revResp.Body).Decode(&revEnv))
	var revisions []dto.ArticleRevisionResponse
	require.NoError(t, json.Unmarshal(revEnv.Data, &revisions))
	require.GreaterOrEqual(t, len(revisions), 2)

	firstRev := revisions[0]

	// 5. Restore first revision
	restReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/admin/articles/%d/restore/%d", server.URL, artID, firstRev.ID), nil)
	restReq.Header.Set("Authorization", "Bearer "+adminToken)
	restResp, err := client.Do(restReq)
	require.NoError(t, err)
	defer restResp.Body.Close()

	assert.Equal(t, http.StatusOK, restResp.StatusCode)
	var restEnv Envelope
	require.NoError(t, json.NewDecoder(restResp.Body).Decode(&restEnv))
	var restoredArt dto.ArticleResponse
	require.NoError(t, json.Unmarshal(restEnv.Data, &restoredArt))
	assert.Equal(t, firstRev.Title, restoredArt.Title)

	// 6. Generate Preview Token
	prevReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/admin/articles/%d/preview-token", server.URL, artID), nil)
	prevReq.Header.Set("Authorization", "Bearer "+adminToken)
	prevResp, err := client.Do(prevReq)
	require.NoError(t, err)
	defer prevResp.Body.Close()

	assert.Equal(t, http.StatusOK, prevResp.StatusCode)
	var prevEnv Envelope
	require.NoError(t, json.NewDecoder(prevResp.Body).Decode(&prevEnv))
	var prevData map[string]string
	require.NoError(t, json.Unmarshal(prevEnv.Data, &prevData))
	assert.NotEmpty(t, prevData["token"])
	assert.Contains(t, prevData["url"], "/preview?token=")

	// 7. List Articles (Pagination)
	listReq, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/admin/articles?page=1&limit=5", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listResp, err := client.Do(listReq)
	require.NoError(t, err)
	defer listResp.Body.Close()

	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	var listEnv Envelope
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&listEnv))
	var listArts []dto.ArticleResponse
	require.NoError(t, json.Unmarshal(listEnv.Data, &listArts))
	assert.NotEmpty(t, listArts)
}

func TestAdminAPI_Taxonomies_CRUD(t *testing.T) {
	server, client, _ := setupAdminTestServer(t)

	adminToken, _ := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")

	// 1. Category CRUD
	catSlug := fmt.Sprintf("cat-%d", time.Now().UnixNano())
	catPayload, _ := json.Marshal(map[string]interface{}{
		"name":        "Kategori Uji",
		"slug":        catSlug,
		"description": "Deskripsi kategori",
		"sort_order":  10,
		"is_nav":      true,
	})
	catReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/categories", bytes.NewReader(catPayload))
	catReq.Header.Set("Authorization", "Bearer "+adminToken)
	catReq.Header.Set("Content-Type", "application/json")
	catResp, err := client.Do(catReq)
	require.NoError(t, err)
	defer catResp.Body.Close()
	require.Equal(t, http.StatusCreated, catResp.StatusCode)

	var catEnv Envelope
	require.NoError(t, json.NewDecoder(catResp.Body).Decode(&catEnv))
	var cat dto.CategoryResponse
	require.NoError(t, json.Unmarshal(catEnv.Data, &cat))
	assert.Equal(t, catSlug, cat.Slug)

	// 2. Tag CRUD
	tagSlug := fmt.Sprintf("tag-%d", time.Now().UnixNano())
	tagPayload, _ := json.Marshal(map[string]string{
		"name": "Tag Uji",
		"slug": tagSlug,
	})
	tagReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/tags", bytes.NewReader(tagPayload))
	tagReq.Header.Set("Authorization", "Bearer "+adminToken)
	tagReq.Header.Set("Content-Type", "application/json")
	tagResp, err := client.Do(tagReq)
	require.NoError(t, err)
	defer tagResp.Body.Close()
	require.Equal(t, http.StatusCreated, tagResp.StatusCode)

	var tagEnv Envelope
	require.NoError(t, json.NewDecoder(tagResp.Body).Decode(&tagEnv))
	var tag dto.TagResponse
	require.NoError(t, json.Unmarshal(tagEnv.Data, &tag))
	assert.Equal(t, tagSlug, tag.Slug)

	// 3. Author CRUD
	authSlug := fmt.Sprintf("author-%d", time.Now().UnixNano())
	authPayload, _ := json.Marshal(map[string]interface{}{
		"name": "Jurnalis Investigasi",
		"slug": authSlug,
		"bio":  "Penulis spesialis data dan investigasi",
	})
	authReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/authors", bytes.NewReader(authPayload))
	authReq.Header.Set("Authorization", "Bearer "+adminToken)
	authReq.Header.Set("Content-Type", "application/json")
	authResp, err := client.Do(authReq)
	require.NoError(t, err)
	defer authResp.Body.Close()
	require.Equal(t, http.StatusCreated, authResp.StatusCode)

	var authEnv Envelope
	require.NoError(t, json.NewDecoder(authResp.Body).Decode(&authEnv))
	var author dto.AuthorResponse
	require.NoError(t, json.Unmarshal(authEnv.Data, &author))
	assert.Equal(t, authSlug, author.Slug)
}

func TestAdminAPI_Media_Upload_And_Variants(t *testing.T) {
	server, client, tempDir := setupAdminTestServer(t)

	adminToken, _ := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")

	// Create in-memory valid 1250x600 PNG image
	img := image.NewRGBA(image.Rect(0, 0, 1250, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 1250; x++ {
			img.Set(x, y, color.RGBA{R: 20, G: 80, B: 200, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	require.NoError(t, png.Encode(&imgBuf, img))

	// Prepare multipart form upload
	var reqBody bytes.Buffer
	writer := multipart.NewWriter(&reqBody)
	part, err := writer.CreateFormFile("file", "test_banner.png")
	require.NoError(t, err)
	_, err = io.Copy(part, &imgBuf)
	require.NoError(t, err)

	_ = writer.WriteField("caption", "Banner pengujian berita")
	_ = writer.WriteField("alt_text", "Banner berita")
	require.NoError(t, writer.Close())

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/admin/media/upload", &reqBody)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	var med dto.MediaResponse
	require.NoError(t, json.Unmarshal(env.Data, &med))

	assert.Greater(t, med.ID, uint64(0))
	assert.Equal(t, uint32(1250), med.Width)
	assert.Equal(t, uint32(600), med.Height)
	assert.NotEmpty(t, med.Variants["thumbnail"])
	assert.NotEmpty(t, med.Variants["medium"])
	assert.NotEmpty(t, med.Variants["large"])

	// Verify disk presence of variants
	relThumb := med.Variants["thumbnail"]
	relMedium := med.Variants["medium"]
	relLarge := med.Variants["large"]

	// The media files start with /media/ prefix in URL, map to tempDir
	thumbDisk := filepath.Join(tempDir, filepath.Base(relThumb))
	medDisk := filepath.Join(tempDir, filepath.Base(relMedium))
	largeDisk := filepath.Join(tempDir, filepath.Base(relLarge))

	_, errThumb := os.Stat(thumbDisk)
	_, errMed := os.Stat(medDisk)
	_, errLarge := os.Stat(largeDisk)

	assert.NoError(t, errThumb, "Thumbnail variant file should exist on disk")
	assert.NoError(t, errMed, "Medium variant file should exist on disk")
	assert.NoError(t, errLarge, "Large variant file should exist on disk")
}

func TestAdminAPI_Dashboard_Stats(t *testing.T) {
	server, client, _ := setupAdminTestServer(t)

	adminToken, _ := loginUser(t, server.URL, client, "admin@newsportal.id", "AdminSecret123!")

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/admin/dashboard/stats", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var env Envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.True(t, env.Success)

	var stats dto.DashboardStatsResponse
	require.NoError(t, json.Unmarshal(env.Data, &stats))

	assert.Greater(t, stats.TotalArticles, int64(0))
	assert.Greater(t, stats.PublishedArticles, int64(0))
	assert.Greater(t, stats.TotalCategories, int64(0))
}
