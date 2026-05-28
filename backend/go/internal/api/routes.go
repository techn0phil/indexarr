package api

import (
	"database/sql"
	"net/http"

	"indexarr/internal/config"
	"indexarr/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRoutes(db *sql.DB, cfg *config.Config, scheduler *services.Scheduler, broadcaster *services.Broadcaster, authService *services.AuthService, oidcService *services.OIDCService) *chi.Mux {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // Required for cookies
		MaxAge:           300,
	}))

	// Health check (always public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public auth routes (no middleware)
		r.Route("/auth", func(r chi.Router) {
			r.Get("/config", HandleAuthConfig(authService, oidcService))
			r.Post("/login", HandleLogin(authService))
			r.Post("/logout", HandleLogout())

			// OIDC routes (public, only if OIDC is configured)
			if oidcService != nil {
				r.Get("/oidc/login", HandleOIDCLogin(oidcService))
				r.Get("/oidc/callback", HandleOIDCCallback(oidcService, authService))
			}
		})

		// Protected routes (require authentication if enabled)
		r.Group(func(r chi.Router) {
			// Apply auth middleware
			r.Use(AuthMiddleware(authService))

			// Auth - current user and password change
			r.Get("/auth/me", HandleMe(authService))
			r.Post("/auth/change-password", HandleChangePassword(authService))

			// User management (admin only, simple auth mode only)
			r.Route("/users", func(r chi.Router) {
				r.Get("/", HandleListUsers(authService))
				r.Post("/", HandleCreateUser(authService))
				r.Put("/{id}", HandleUpdateUser(authService))
				r.Delete("/{id}", HandleDeleteUser(authService))
				r.Post("/{id}/password", HandleAdminSetPassword(authService))
			})

			// Config
			r.Get("/config", GetConfig(cfg))

			// Movies
			r.Get("/movies", ListMovies(db))
			r.Get("/movies/{id}", GetMovie(db))

			// Series
			r.Get("/series", ListSeries(db))
			r.Get("/series/{id}", GetSeriesByID(db))

			// Stats
			r.Get("/stats", GetStats(db))

			// Purge
			r.Post("/purge", Purge(db, authService))

			// Scan (only if scheduler is provided)
			if scheduler != nil {
				r.Post("/scan", TriggerScan(scheduler, authService))
				r.Post("/scan/movies", TriggerMoviesScan(scheduler, authService))
				r.Post("/scan/series", TriggerSeriesScan(scheduler, authService))
				r.Get("/scan/status", GetScanStatus(scheduler))
				r.Post("/scan/stop", StopScan(scheduler, authService))

				r.Post("/movies/{id}/refresh", RefreshMovie(scheduler, authService))
				r.Post("/series/{id}/refresh", RefreshSeries(scheduler, authService))

				// WebSocket endpoint for real-time scan updates
				if broadcaster != nil {
					r.Get("/scan/ws", HandleWebSocket(db, broadcaster))
				}
			}
		})
	})

	return r
}
