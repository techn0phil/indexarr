package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTMDBSearchMovie_RetryWithoutYear(t *testing.T) {
	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/3/search/movie" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		requestCount++

		if got := r.URL.Query().Get("query"); got != "The Matrix" {
			t.Fatalf("query param mismatch: got %q", got)
		}
		if got := r.URL.Query().Get("api_key"); got != "test-key" {
			t.Fatalf("api_key mismatch: got %q", got)
		}

		if requestCount == 1 {
			if got := r.URL.Query().Get("primary_release_year"); got != "1999" {
				t.Fatalf("expected year filter on first call, got %q", got)
			}
			fmt.Fprint(w, `{"page":1,"total_results":0,"results":[]}`)
			return
		}

		if got := r.URL.Query().Get("primary_release_year"); got != "" {
			t.Fatalf("expected retry call without year, got %q", got)
		}
		fmt.Fprint(w, `{"page":1,"total_results":1,"results":[{"id":603,"title":"The Matrix"}]}`)
	}))
	defer srv.Close()

	client := NewTMDBClient("test-key", "en", "fr")
	client.httpClient = newRewrittenClient(srv)

	result, err := client.SearchMovie("The Matrix", 1999)
	if err != nil {
		t.Fatalf("SearchMovie() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected 2 requests due to retry, got %d", requestCount)
	}
	if result.TotalResults != 1 || len(result.Results) != 1 || result.Results[0].ID != 603 {
		t.Fatalf("unexpected search result: %+v", result)
	}
}

func TestTMDBGetMovieDetails_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"status_message":"rate limited"}`)
	}))
	defer srv.Close()

	client := NewTMDBClient("test-key", "en", "en")
	client.httpClient = newRewrittenClient(srv)

	if _, err := client.GetMovieDetails(42); err == nil {
		t.Fatalf("expected GetMovieDetails() to return API error")
	}
}

func TestTMDBClient_RequiresAPIKey(t *testing.T) {
	client := NewTMDBClient("", "en", "en")
	if _, err := client.SearchTV("Some Show", 0); err == nil {
		t.Fatalf("expected missing API key error")
	}
}
