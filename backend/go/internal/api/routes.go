package api

import (
	"database/sql"
	"net/http"

	"indexarr/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRoutes(db *sql.DB, scheduler *services.Scheduler) *chi.Mux {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Movies
		r.Get("/movies", ListMovies(db))
		r.Get("/movies/{id}", GetMovie(db))

		// Series
		r.Get("/series", ListSeries(db))
		r.Get("/series/{id}", GetSeriesByID(db))

		// Stats
		r.Get("/stats", GetStats(db))

		// Scan (only if scheduler is provided)
		if scheduler != nil {
			r.Post("/scan", TriggerScan(scheduler))
			r.Get("/scan/status", GetScanStatus(scheduler))
			r.Post("/scan/stop", StopScan(scheduler))
		}
	})

	return r
}
