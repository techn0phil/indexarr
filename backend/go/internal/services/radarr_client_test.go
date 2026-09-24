package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRadarrClientConnectionAndMovieQueries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Api-Key"); got != "radarr-key" {
			t.Fatalf("unexpected api key header: %q", got)
		}

		switch r.URL.Path {
		case "/api/v3/system/status":
			fmt.Fprint(w, `{"appName":"Radarr","version":"5.0"}`)
		case "/api/v3/movie":
			if r.URL.RawQuery == "tmdbId=12345" {
				fmt.Fprint(w, `[{"id":10,"title":"Movie A","tmdbId":12345}]`)
				return
			}
			fmt.Fprint(w, `[{"id":1,"title":"Movie 1"},{"id":2,"title":"Movie 2"}]`)
		case "/api/v3/movie/10":
			fmt.Fprint(w, `{"id":10,"title":"Movie A","tmdbId":12345}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer srv.Close()

	client := NewRadarrClient(srv.URL, "radarr-key")

	if err := client.TestConnection(); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}

	movies, err := client.GetMovies()
	if err != nil {
		t.Fatalf("GetMovies() error = %v", err)
	}
	if len(movies) != 2 {
		t.Fatalf("expected 2 movies, got %d", len(movies))
	}

	movie, err := client.GetMovie(10)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}
	if movie.ID != 10 || movie.Title != "Movie A" {
		t.Fatalf("unexpected movie payload: %+v", movie)
	}

	byTMDB, err := client.GetMovieByTMDBId(12345)
	if err != nil {
		t.Fatalf("GetMovieByTMDBId() error = %v", err)
	}
	if byTMDB == nil || byTMDB.ID != 10 {
		t.Fatalf("unexpected TMDB lookup result: %+v", byTMDB)
	}
}

func TestRadarrClientRequestErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `{"error":"gateway"}`)
	}))
	defer srv.Close()

	client := NewRadarrClient(srv.URL, "radarr-key")
	if _, err := client.GetMovies(); err == nil {
		t.Fatalf("expected API error")
	}
}
