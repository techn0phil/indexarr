package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"indexarr/internal/models"
	"indexarr/internal/repository"
)

func listedTitles(t *testing.T, rr *httptest.ResponseRecorder) []string {
	t.Helper()

	body := decodeJSONMap(t, rr)
	items, _ := body["data"].([]interface{})
	var titles []string
	for _, item := range items {
		titles = append(titles, item.(map[string]interface{})["title"].(string))
	}
	return titles
}

func TestListMoviesAndSeries_SortParam(t *testing.T) {
	db := setupAPITestDBWithMigrations(t)

	for i, y := range []int{1990, 2000, 2010} {
		title := string(rune('A' + i))
		if _, err := repository.InsertMovie(db, &models.Movie{
			Title: title, Year: y, Status: "available", DateAdded: "2026-01-01T00:00:00Z",
			FilePath: "/m/" + title + ".mkv", Container: "mkv", TMDBId: int64(10 + i), IMDbId: "tt" + title,
			MediaInfo: &models.MediaInfo{},
		}); err != nil {
			t.Fatalf("failed to insert movie: %v", err)
		}
		if _, err := repository.InsertSeries(db, &models.Series{
			Title: title, YearStart: y, Status: "ongoing", DateAdded: "2026-01-01T00:00:00Z",
			TMDBId: int64(20 + i), TVDBId: int64(30 + i), IMDbId: "tt" + title, Slug: title,
		}); err != nil {
			t.Fatalf("failed to insert series: %v", err)
		}
	}
	if _, err := db.Exec(`UPDATE series SET missing_episode_count = 0`); err != nil {
		t.Fatalf("failed to set missing episode count: %v", err)
	}

	handlers := map[string]http.HandlerFunc{
		"movies": ListMovies(db),
		"series": ListSeries(db),
	}

	for name, handler := range handlers {
		t.Run(name+" descending sort", func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/"+name+"?sort=-year", nil))

			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status code: got %d body=%s", rr.Code, rr.Body.String())
			}
			if got, want := listedTitles(t, rr), []string{"C", "B", "A"}; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v want %v", got, want)
			}
		})

		t.Run(name+" invalid sort returns bad request", func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/"+name+"?sort=bogus", nil))

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}
