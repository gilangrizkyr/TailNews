package http

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"news-portal/apps/api/internal/config"
	"news-portal/apps/api/internal/platform"
	"news-portal/apps/api/internal/repository/mysql"
	"news-portal/apps/api/internal/transport/http/handler"
	"news-portal/apps/api/internal/transport/http/middleware"
	"news-portal/apps/api/internal/usecase"
)

// RouterConfig defines the dependencies required to construct the HTTP router.
type RouterConfig struct {
	Config  *config.Config
	Logger  *slog.Logger
	DB      *sql.DB
	Redis   *redis.Client
	Storage platform.StoragePlatform
}

// NewRouter builds and returns the fully wired Chi HTTP handler.
func NewRouter(rc RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Core Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.StructuredLogger(rc.Logger))
	r.Use(middleware.Recoverer(rc.Logger))
	r.Use(middleware.CORS(rc.Config.CORSAllowedOrigins))
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.CleanPath)

	// Liveness & Readiness Probes
	healthHandler := handler.NewHealthHandler(rc.DB, rc.Redis)
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// Static Media Delivery (for local development)
	if rc.Config.MediaStorageDir != "" {
		fs := http.StripPrefix("/media/", http.FileServer(http.Dir(rc.Config.MediaStorageDir)))
		r.Get("/media/*", func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		})
	}

	// Initialize Validation, Storage & Repositories
	validate := validator.New()
	storage := rc.Storage
	if storage == nil {
		storage, _ = platform.NewLocalStorage(rc.Config.MediaStorageDir)
	}
	repo := mysql.NewRepository(rc.DB)

	// Initialize Usecases
	authUsecase := usecase.NewAuthUsecase(repo.User, rc.Config)
	articleAdminUsecase := usecase.NewArticleAdminUsecase(
		repo.Article,
		repo.Category,
		repo.Author,
		repo.Media,
		repo.Redirect,
		repo.Audit,
		rc.Config,
	)
	taxonomyAdminUsecase := usecase.NewTaxonomyAdminUsecase(repo.Category, repo.Tag, repo.Author)
	mediaAdminUsecase := usecase.NewMediaAdminUsecase(repo.Media, storage, rc.Config)
	redirectAdminUsecase := usecase.NewRedirectAdminUsecase(repo.Redirect)
	dashboardAdminUsecase := usecase.NewDashboardAdminUsecase(rc.DB)

	// Initialize Handlers
	authHandler := handler.NewAuthHandler(authUsecase, validate)
	articleAdminHandler := handler.NewAdminArticleHandler(articleAdminUsecase, validate)
	taxonomyAdminHandler := handler.NewAdminTaxonomyHandler(taxonomyAdminUsecase, validate)
	mediaAdminHandler := handler.NewAdminMediaHandler(mediaAdminUsecase)
	redirectAdminHandler := handler.NewAdminRedirectHandler(redirectAdminUsecase, validate)
	dashboardAdminHandler := handler.NewAdminDashboardHandler(dashboardAdminUsecase)

	// API Namespace
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"message":"pong"}}`))
		})

		// Public Auth Endpoints
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/login", authHandler.Login)
			auth.Post("/refresh", authHandler.Refresh)
			auth.Post("/logout", authHandler.Logout)

			// Authenticated Profile
			auth.Group(func(protected chi.Router) {
				protected.Use(middleware.AuthMiddleware(rc.Config.JWTSecret))
				protected.Get("/me", authHandler.Me)
			})
		})

		// Protected Admin Endpoints
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.AuthMiddleware(rc.Config.JWTSecret))

			// Dashboard Stats
			admin.Get("/dashboard/stats", dashboardAdminHandler.GetStats)

			// Article Management
			admin.Route("/articles", func(art chi.Router) {
				art.Get("/", articleAdminHandler.List)
				art.Post("/", articleAdminHandler.Create)
				art.Get("/{id}", articleAdminHandler.Get)
				art.Put("/{id}", articleAdminHandler.Update)
				art.Get("/{id}/revisions", articleAdminHandler.ListRevisions)
				art.Post("/{id}/restore/{revision_id}", articleAdminHandler.RestoreRevision)
				art.Post("/{id}/preview-token", articleAdminHandler.PreviewToken)

				// Restrict article deletion to admin and editor
				art.Group(func(del chi.Router) {
					del.Use(middleware.RequireRoles("admin", "editor"))
					del.Delete("/{id}", articleAdminHandler.Delete)
				})
			})

			// Categories Management
			admin.Route("/categories", func(cat chi.Router) {
				cat.Get("/", taxonomyAdminHandler.ListCategories)
				cat.Post("/", taxonomyAdminHandler.CreateCategory)
				cat.Put("/{id}", taxonomyAdminHandler.UpdateCategory)
				cat.Delete("/{id}", taxonomyAdminHandler.DeleteCategory)
			})

			// Tags Management
			admin.Route("/tags", func(tag chi.Router) {
				tag.Get("/", taxonomyAdminHandler.ListTags)
				tag.Post("/", taxonomyAdminHandler.CreateTag)
				tag.Delete("/{id}", taxonomyAdminHandler.DeleteTag)
			})

			// Authors Management
			admin.Route("/authors", func(author chi.Router) {
				author.Get("/", taxonomyAdminHandler.ListAuthors)
				author.Post("/", taxonomyAdminHandler.CreateAuthor)
				author.Put("/{id}", taxonomyAdminHandler.UpdateAuthor)
			})

			// Media Management
			admin.Route("/media", func(med chi.Router) {
				med.Post("/upload", mediaAdminHandler.Upload)
				med.Get("/", mediaAdminHandler.List)
				med.Delete("/{id}", mediaAdminHandler.Delete)
			})

			// Redirects Management
			admin.Route("/redirects", func(redir chi.Router) {
				redir.Get("/", redirectAdminHandler.List)
				redir.Post("/", redirectAdminHandler.CreateOrUpdate)
			})
		})
	})

	return r
}

// Compile-time check to verify string utilities import
var _ = strings.TrimSpace
