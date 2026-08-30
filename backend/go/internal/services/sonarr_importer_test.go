package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"

	"github.com/rs/zerolog"
)

func TestSonarrImporterMapHelpers(t *testing.T) {
	importer := &SonarrImporter{
		client:   &SonarrClient{baseURL: "http://sonarr.local"},
		pathFrom: "/sonarr",
		pathTo:   "/local",
	}

	series := importer.mapSonarrSeries(&SonarrSeries{
		Title:     "My Show",
		TitleSlug: "my-show",
		Year:      2020,
		Status:    "ended",
		Overview:  "overview",
		Genres:    []string{"Drama", "Mystery"},
		Ratings:   SonarrRatings{Value: 8.1},
		Images:    []SonarrImage{{CoverType: "poster", URL: "/MediaCover/77/poster.jpg"}},
		TmdbId:    0,
		TvdbId:    777,
		ImdbId:    "tt77",
		Added:     "",
		LastAired: "2024-10-01T00:00:00Z",
		Statistics: SonarrStats{
			SeasonCount:       3,
			EpisodeCount:      20,
			TotalEpisodeCount: 25,
			SizeOnDisk:        1234,
		},
		Seasons: []SonarrSeason{
			{SeasonNumber: 0, Statistics: SonarrStats{TotalEpisodeCount: 2}},
		},
	})

	if got := importer.mapPath("/sonarr/Shows/My Show/S01E01.mkv"); got != "/local/Shows/My Show/S01E01.mkv" {
		t.Fatalf("mapPath() = %q", got)
	}
	if series.Status != "complete" {
		t.Fatalf("Status = %q, want complete", series.Status)
	}
	if series.TMDBId != -777 {
		t.Fatalf("TMDB fallback = %d, want -777", series.TMDBId)
	}
	if series.Poster == nil || *series.Poster != "http://sonarr.local/MediaCover/77/poster.jpg" {
		t.Fatalf("Poster = %v, want local URL expanded", series.Poster)
	}
	if series.DateAdded == "" {
		t.Fatalf("expected non-empty DateAdded fallback")
	}
	if series.YearEnd != 2024 {
		t.Fatalf("YearEnd = %d, want 2024", series.YearEnd)
	}
	if series.TotalEpisodeCount != 23 {
		t.Fatalf("TotalEpisodeCount = %d, want 23", series.TotalEpisodeCount)
	}

	episode := importer.mapSonarrEpisode(42, &SonarrEpisode{
		SeasonNumber:  1,
		EpisodeNumber: 5,
		Title:         "",
		HasFile:       true,
		Runtime:       50,
		EpisodeFile:   &SonarrEpisodeFile{Path: "/p.mkv", Size: 888},
	})
	if episode.Title != "Episode 5" {
		t.Fatalf("Title fallback = %q", episode.Title)
	}
	if episode.Duration != 3000 {
		t.Fatalf("Duration = %d, want 3000", episode.Duration)
	}
	if episode.Status != "available" {
		t.Fatalf("Status = %q, want available", episode.Status)
	}
}

func TestSonarrImporterImportUsesCacheProcessesEpisodesAndRemovesStale(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	db := setupServicesTestDBWithMigrations(t)

	staleSeries := &models.Series{
		Title:             "Stale Show",
		Slug:              "stale-show",
		YearStart:         2010,
		SeasonCount:       1,
		EpisodeCount:      1,
		TotalSeasonCount:  1,
		TotalEpisodeCount: 1,
		Status:            "ongoing",
		DateAdded:         "2020-01-01T00:00:00Z",
		SonarrID:          777,
	}
	if _, err := repository.InsertSeries(db, staleSeries); err != nil {
		t.Fatalf("InsertSeries(stale) error = %v", err)
	}

	seriesCalls := 0
	episodeCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/series":
			seriesCalls++
			fmt.Fprint(w, `[
				{
					"id":101,
					"title":"Imported Show",
					"titleSlug":"imported-show",
					"year":2022,
					"overview":"desc",
					"status":"continuing",
					"seasonCount":2,
					"genres":["Drama"],
					"ratings":{"value":7.9},
					"images":[{"coverType":"poster","remoteUrl":"https://img.example/show.jpg"}],
					"tmdbId":500,
					"tvdbId":600,
					"imdbId":"tt500",
					"added":"2025-06-01T12:00:00Z",
					"lastAired":"2025-01-01T00:00:00Z",
					"statistics":{"episodeFileCount":1,"episodeCount":2,"totalEpisodeCount":3,"sizeOnDisk":12345},
					"seasons":[{"seasonNumber":0,"statistics":{"totalEpisodeCount":1}}]
				}
			]`)
		case "/api/v3/episode":
			episodeCalls++
			fmt.Fprint(w, `[
				{"id":201,"seriesId":101,"seasonNumber":0,"episodeNumber":1,"title":"Special","hasFile":true,"runtime":30,"episodeFile":{"path":"","size":1}},
				{"id":202,"seriesId":101,"seasonNumber":1,"episodeNumber":1,"title":"Pilot","hasFile":true,"runtime":45,"episodeFile":{"path":"","size":999}}
			]`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer srv.Close()

	client := NewSonarrClient(srv.URL, "sonarr-key")
	cfg := &config.Config{MediainfoPath: "mediainfo", ScanTimeout: 1}
	importer := NewSonarrImporter(db, cfg, client, nil)

	count, err := importer.GetPendingFileCount()
	if err != nil {
		t.Fatalf("GetPendingFileCount() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("GetPendingFileCount() = %d, want 1", count)
	}

	result, err := importer.Import(nil)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.FilesFound != 1 {
		t.Fatalf("FilesFound = %d, want 1", result.FilesFound)
	}
	if result.FilesProcessed != 1 {
		t.Fatalf("FilesProcessed = %d, want 1", result.FilesProcessed)
	}
	if result.EpisodesAdded != 1 {
		t.Fatalf("EpisodesAdded = %d, want 1", result.EpisodesAdded)
	}

	if seriesCalls != 1 {
		t.Fatalf("expected one /series API call due to cache use, got %d", seriesCalls)
	}
	if episodeCalls != 1 {
		t.Fatalf("expected one /episode API call, got %d", episodeCalls)
	}

	importedSeries, err := repository.GetSeriesBySonarrID(db, 101)
	if err != nil {
		t.Fatalf("GetSeriesBySonarrID(imported) error = %v", err)
	}
	if importedSeries == nil {
		t.Fatalf("expected imported series")
	}
	if importedSeries.Poster == nil || *importedSeries.Poster != "https://img.example/show.jpg" {
		t.Fatalf("imported poster = %v", importedSeries.Poster)
	}

	deletedStale, err := repository.GetSeriesBySonarrID(db, 777)
	if err != nil {
		t.Fatalf("GetSeriesBySonarrID(stale) error = %v", err)
	}
	if deletedStale != nil {
		t.Fatalf("expected stale series to be deleted, got %+v", deletedStale)
	}
}
