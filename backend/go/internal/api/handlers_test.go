package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
	"indexarr/internal/services"
)

type handlerTestMovieImporter struct {
	status *models.ScanStatus
	err    error
}

func (m *handlerTestMovieImporter) Import(*models.ProgressContext) (*models.ScanResult, error) {
	return nil, nil
}

func (m *handlerTestMovieImporter) ImportMovie(int64) (*models.ScanResult, error) {
	return nil, nil
}

func (m *handlerTestMovieImporter) GetPendingFileCount() (int, error) {
	return 0, nil
}

func (m *handlerTestMovieImporter) GetStatus() (*models.ScanStatus, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.status != nil {
		return m.status, nil
	}
	return &models.ScanStatus{Status: "idle"}, nil
}

func (m *handlerTestMovieImporter) Stop() {}

func (m *handlerTestMovieImporter) IsRunning() bool { return false }

func TestListMoviesAndGetMovieHandlers(t *testing.T) {
	db := setupAPITestDBWithMigrations(t)

	movieID, err := repository.InsertMovie(db, &models.Movie{
		Title:     "Movie Handler Test",
		Year:      2025,
		Status:    "available",
		DateAdded: "2026-01-01T00:00:00Z",
		FilePath:  "/library/movies/handler-test.mkv",
		Container: "mkv",
		TMDBId:    5555,
		IMDbId:    "tt5555",
		MediaInfo: &models.MediaInfo{},
	})
	if err != nil {
		t.Fatalf("failed to insert movie: %v", err)
	}

	t.Run("list movies returns paginated success payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
		rr := httptest.NewRecorder()

		ListMovies(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}

		body := decodeJSONMap(t, rr)
		if success, ok := body["success"].(bool); !ok || !success {
			t.Fatalf("expected success=true, got: %v", body["success"])
		}
		if page, ok := body["page"].(float64); !ok || int(page) != 1 {
			t.Fatalf("expected page=1, got %v", body["page"])
		}
		if pageSize, ok := body["pageSize"].(float64); !ok || int(pageSize) != 50 {
			t.Fatalf("expected pageSize=50, got %v", body["pageSize"])
		}
		items, ok := body["data"].([]interface{})
		if !ok || len(items) == 0 {
			t.Fatalf("expected at least one movie in data payload")
		}
	})

	t.Run("get movie invalid id returns bad request", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/movies/not-a-number", nil), "id", "not-a-number")
		rr := httptest.NewRecorder()

		GetMovie(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("get movie not found returns 404", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/movies/999999", nil), "id", "999999")
		rr := httptest.NewRecorder()

		GetMovie(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("get movie success returns selected movie", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/movies/1", nil), "id", "1")
		rr := httptest.NewRecorder()

		GetMovie(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var movie models.Movie
		if err := json.Unmarshal(rr.Body.Bytes(), &movie); err != nil {
			t.Fatalf("failed to decode movie response: %v", err)
		}
		if movie.ID != movieID || movie.Title != "Movie Handler Test" {
			t.Fatalf("unexpected movie payload: %+v", movie)
		}
	})
}

func TestListSeriesAndGetSeriesHandlers(t *testing.T) {
	db := setupAPITestDBWithMigrations(t)

	seriesID, err := repository.InsertSeries(db, &models.Series{
		Title:             "Series Handler Test",
		YearStart:         2024,
		YearEnd:           2026,
		SeasonCount:       1,
		EpisodeCount:      8,
		TotalSeasonCount:  1,
		TotalEpisodeCount: 8,
		Status:            "ongoing",
		DateAdded:         "2026-01-01T00:00:00Z",
		TMDBId:            7777,
		TVDBId:            8888,
		IMDbId:            "tt7777",
		Slug:              "series-handler-test",
	})
	if err != nil {
		t.Fatalf("failed to insert series: %v", err)
	}
	if _, err := db.Exec(`UPDATE series SET missing_episode_count = 0 WHERE id = ?`, seriesID); err != nil {
		t.Fatalf("failed to set missing episode count: %v", err)
	}

	t.Run("list series returns paginated success payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/series?page=1&page_size=10", nil)
		rr := httptest.NewRecorder()

		ListSeries(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
		}

		body := decodeJSONMap(t, rr)
		if success, ok := body["success"].(bool); !ok || !success {
			t.Fatalf("expected success=true, got: %v", body["success"])
		}
		if total, ok := body["total"].(float64); !ok || int(total) < 1 {
			t.Fatalf("expected total >= 1, got %v", body["total"])
		}
	})

	t.Run("get series invalid id returns bad request", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/series/not-a-number", nil), "id", "not-a-number")
		rr := httptest.NewRecorder()

		GetSeriesByID(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("get series not found returns 404", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/series/999999", nil), "id", "999999")
		rr := httptest.NewRecorder()

		GetSeriesByID(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("get series success returns selected series", func(t *testing.T) {
		req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/series/1", nil), "id", strconv.FormatInt(seriesID, 10))
		rr := httptest.NewRecorder()

		GetSeriesByID(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var series models.Series
		if err := json.Unmarshal(rr.Body.Bytes(), &series); err != nil {
			t.Fatalf("failed to decode series response: %v", err)
		}
		if series.ID != seriesID || series.Title != "Series Handler Test" {
			t.Fatalf("unexpected series payload: %+v", series)
		}
	})
}

func TestGetStatsAndConfigHandlers(t *testing.T) {
	db := setupAPITestDBWithMigrations(t)

	movieID, err := repository.InsertMovie(db, &models.Movie{
		Title:     "Stats Test Movie",
		Year:      2025,
		Status:    "available",
		FileSize:  1024,
		DateAdded: "2026-01-01T00:00:00Z",
		FilePath:  "/library/movies/stats-test.mkv",
		Container: "mkv",
		TMDBId:    9090,
		IMDbId:    "tt9090",
		MediaInfo: &models.MediaInfo{VideoTracks: []models.VideoTrack{{
			Codec:      "H.265",
			Resolution: "3840x2160",
		}}},
	})
	if err != nil {
		t.Fatalf("failed to insert movie for stats: %v", err)
	}
	if movieID == 0 {
		t.Fatalf("expected non-zero movie id")
	}

	t.Run("get stats returns success payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
		rr := httptest.NewRecorder()

		GetStats(db).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var stats models.StatsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
			t.Fatalf("failed to decode stats response: %v", err)
		}
		if !stats.Success {
			t.Fatalf("expected success=true")
		}
		if stats.TotalMovies < 1 {
			t.Fatalf("expected at least one movie, got %d", stats.TotalMovies)
		}
	})

	t.Run("get config returns configured import modes", func(t *testing.T) {
		cfg := &config.Config{
			RadarrURL:          "http://radarr.local",
			RadarrAPIKey:       "test-key",
			MoviesLibraryPaths: []string{"/movies"},
			SeriesLibraryPaths: []string{"/series"},
		}

		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rr := httptest.NewRecorder()

		GetConfig(cfg).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}

		body := decodeJSONMap(t, rr)
		if got, ok := body["moviesImportMode"].(string); !ok || got != "radarr" {
			t.Fatalf("expected moviesImportMode=radarr, got %v", body["moviesImportMode"])
		}
		if got, ok := body["seriesImportMode"].(string); !ok || got != "filesystem" {
			t.Fatalf("expected seriesImportMode=filesystem, got %v", body["seriesImportMode"])
		}
	})
}

func TestGetScanStatusHandler(t *testing.T) {
	t.Run("returns disabled when no importers are configured", func(t *testing.T) {
		scheduler := services.NewScheduler(nil, nil, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/scan/status", nil)
		rr := httptest.NewRecorder()

		GetScanStatus(scheduler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var status models.ScanStatus
		if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
			t.Fatalf("failed to decode scan status response: %v", err)
		}
		if status.Status != "disabled" {
			t.Fatalf("expected status=disabled, got %q", status.Status)
		}
	})

	t.Run("returns internal server error when importer status fails", func(t *testing.T) {
		scheduler := services.NewScheduler(nil, &handlerTestMovieImporter{err: errors.New("status failure")}, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/api/scan/status", nil)
		rr := httptest.NewRecorder()

		GetScanStatus(scheduler).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusInternalServerError)
		}
	})
}
