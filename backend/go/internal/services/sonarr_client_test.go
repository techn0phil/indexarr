package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSonarrClientConnectionAndSeriesQueries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Api-Key"); got != "sonarr-key" {
			t.Fatalf("unexpected api key header: %q", got)
		}

		switch r.URL.Path {
		case "/api/v3/system/status":
			fmt.Fprint(w, `{"appName":"Sonarr","version":"4.0"}`)
		case "/api/v3/series":
			if r.URL.RawQuery == "tvdbId=321" {
				fmt.Fprint(w, `[{"id":99,"title":"Show A","tvdbId":321}]`)
				return
			}
			fmt.Fprint(w, `[{"id":99,"title":"Show A"},{"id":100,"title":"Show B"}]`)
		case "/api/v3/series/99":
			fmt.Fprint(w, `{"id":99,"title":"Show A","tvdbId":321}`)
		case "/api/v3/episode":
			fmt.Fprint(w, `[{"id":501,"seriesId":99,"episodeNumber":1,"seasonNumber":1,"hasFile":true}]`)
		case "/api/v3/episodefile/700":
			fmt.Fprint(w, `{"id":700,"seriesId":99,"path":"/media/Show/S01E01.mkv"}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer srv.Close()

	client := NewSonarrClient(srv.URL, "sonarr-key")

	if err := client.TestConnection(); err != nil {
		t.Fatalf("TestConnection() error = %v", err)
	}

	series, err := client.GetSeries()
	if err != nil {
		t.Fatalf("GetSeries() error = %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(series))
	}

	one, err := client.GetSeriesByID(99)
	if err != nil {
		t.Fatalf("GetSeriesByID() error = %v", err)
	}
	if one.ID != 99 {
		t.Fatalf("unexpected series payload: %+v", one)
	}

	byTVDB, err := client.GetSeriesByTVDBId(321)
	if err != nil {
		t.Fatalf("GetSeriesByTVDBId() error = %v", err)
	}
	if byTVDB == nil || byTVDB.ID != 99 {
		t.Fatalf("unexpected TVDB lookup result: %+v", byTVDB)
	}

	episodes, err := client.GetEpisodes(99)
	if err != nil {
		t.Fatalf("GetEpisodes() error = %v", err)
	}
	if len(episodes) != 1 || episodes[0].ID != 501 {
		t.Fatalf("unexpected episodes payload: %+v", episodes)
	}

	episodeFile, err := client.GetEpisodeFile(700)
	if err != nil {
		t.Fatalf("GetEpisodeFile() error = %v", err)
	}
	if episodeFile.ID != 700 {
		t.Fatalf("unexpected episode file payload: %+v", episodeFile)
	}
}

func TestSonarrClientRequestErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"boom"}`)
	}))
	defer srv.Close()

	client := NewSonarrClient(srv.URL, "sonarr-key")
	if _, err := client.GetSeries(); err == nil {
		t.Fatalf("expected API error")
	}
}
