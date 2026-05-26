package api

import (
	"net/http"
	"strconv"

	"indexarr/internal/services"

	"github.com/go-chi/chi/v5"
)

// TriggerScan starts a manual scan
func TriggerScan(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start scan in goroutine so we can return immediately
		go func() {
			scheduler.TriggerScan()
		}()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Scan started",
		})
	}
}

// TriggerMoviesScan starts a manual scan for movies only
func TriggerMoviesScan(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go func() {
			scheduler.TriggerMoviesScan()
		}()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Movies scan started",
		})
	}
}

// TriggerSeriesScan starts a manual scan for series only
func TriggerSeriesScan(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go func() {
			scheduler.TriggerSeriesScan()
		}()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Series scan started",
		})
	}
}

// GetScanStatus returns the current scan status
func GetScanStatus(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := scheduler.GetScanStatus()
		if err != nil {
			respondError(w, 500, "Failed to get scan status: "+err.Error())
			return
		}

		respond(w, status)
	}
}

// StopScan stops the currently running scan
func StopScan(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheduler.StopCurrentScan()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Stop signal sent",
		})
	}
}

func RefreshMovie(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			respondError(w, 400, "Invalid movie ID")
			return
		}

		result, err := scheduler.TriggerSingleMovieScan(id)
		if err != nil {
			respondError(w, 500, "Failed to refresh movie: "+err.Error())
			return
		}

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Movie refresh started",
			"result":  result,
		})
	}
}

func RefreshSeries(scheduler *services.Scheduler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			respondError(w, 400, "Invalid series ID")
			return
		}

		result, err := scheduler.TriggerSingleSeriesScan(id)
		if err != nil {
			respondError(w, 500, "Failed to refresh series: "+err.Error())
			return
		}

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Series refresh started",
			"result":  result,
		})
	}
}
