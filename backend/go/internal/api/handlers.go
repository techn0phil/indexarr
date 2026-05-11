package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"

	"github.com/go-chi/chi/v5"
)

func ListMovies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize == 0 {
			pageSize = 50
		}
		if page == 0 {
			page = 1
		}

		filters := &models.FilterCriteria{
			Status:     r.URL.Query().Get("status"),
			Resolution: r.URL.Query().Get("resolution"),
			Codec:      r.URL.Query().Get("codec"),
			Audio:      r.URL.Query().Get("audio"),
			HDR:        r.URL.Query().Get("hdr"),
			Search:     r.URL.Query().Get("search"),
			Page:       page,
			PageSize:   pageSize,
		}

		movies, total, err := repository.GetMovies(db, filters)
		if err != nil {
			respondError(w, 500, err.Error())
			return
		}

		respond(w, &models.PaginatedResponse{
			Success:  true,
			Data:     movies,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

func GetMovie(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			respondError(w, 400, "Invalid movie ID")
			return
		}

		movie, err := repository.GetMovieByID(db, id)
		if err != nil {
			respondError(w, 404, "Movie not found")
			return
		}

		respond(w, movie)
	}
}

func ListSeries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize == 0 {
			pageSize = 50
		}
		if page == 0 {
			page = 1
		}

		filters := &models.FilterCriteria{
			Status:     r.URL.Query().Get("status"),
			Resolution: r.URL.Query().Get("resolution"),
			Codec:      r.URL.Query().Get("codec"),
			Audio:      r.URL.Query().Get("audio"),
			HDR:        r.URL.Query().Get("hdr"),
			Search:     r.URL.Query().Get("search"),
			Page:       page,
			PageSize:   pageSize,
		}

		series, total, err := repository.GetSeries(db, filters)
		if err != nil {
			respondError(w, 500, err.Error())
			return
		}

		respond(w, &models.PaginatedResponse{
			Success:  true,
			Data:     series,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

func GetSeriesByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			respondError(w, 400, "Invalid series ID")
			return
		}

		series, err := repository.GetSeriesByID(db, id)
		if err != nil {
			respondError(w, 404, "Series not found")
			return
		}

		respond(w, series)
	}
}

func GetStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := repository.GetStats(db)
		if err != nil {
			respondError(w, 500, err.Error())
			return
		}

		respond(w, stats)
	}
}

// GetConfig returns application configuration
func GetConfig(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"radarrUrl": cfg.RadarrURL,
			"sonarrUrl": cfg.SonarrURL,
		})
	}
}

func respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func Purge(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := repository.PurgeDatabase(db)
		if err != nil {
			respondError(w, 500, "Failed to purge database: "+err.Error())
			return
		}

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Database purged successfully",
		})
	}
}
