package repository

import (
	"errors"
	"reflect"
	"testing"

	"indexarr/internal/models"
)

func reversed(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[len(in)-1-i] = v
	}
	return out
}

func TestGetMovies_Sort(t *testing.T) {
	db := setupTestDB(t)

	rows := []struct {
		title    string
		year     int
		duration int
		rating   float64
		size     int64
		added    string
	}{
		{"Alpha", 2000, 90, 5, 300, "2026-01-03"},
		{"bravo", 2010, 120, 9, 100, "2026-01-01"},
		{"charlie", 1990, 60, 7, 200, "2026-01-02"},
	}
	for i, r := range rows {
		id := insertTestMovie(t, db, r.title, "available", int64(100+i), "/movies/"+r.title+".mkv")
		if _, err := db.Exec(`UPDATE movies SET year=?, duration=?, rating=?, file_size=?, date_added=? WHERE id=?`,
			r.year, r.duration, r.rating, r.size, r.added, id); err != nil {
			t.Fatalf("failed to update movie: %v", err)
		}
	}

	tests := []struct {
		name string
		sort string
		want []string
	}{
		{"default is title ascending, case-insensitive", "", []string{"Alpha", "bravo", "charlie"}},
		{"title", "title", []string{"Alpha", "bravo", "charlie"}},
		{"year", "year", []string{"charlie", "Alpha", "bravo"}},
		{"duration", "duration", []string{"charlie", "Alpha", "bravo"}},
		{"rating", "rating", []string{"Alpha", "charlie", "bravo"}},
		{"size", "size", []string{"bravo", "charlie", "Alpha"}},
		{"added", "added", []string{"bravo", "charlie", "Alpha"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movies, _, err := GetMovies(db, &models.FilterCriteria{Sort: tt.sort})
			if err != nil {
				t.Fatalf("GetMovies returned error: %v", err)
			}
			var got []string
			for _, m := range movies {
				got = append(got, m.Title)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ascending: got %v want %v", got, tt.want)
			}

			if tt.sort == "" {
				return
			}
			movies, _, err = GetMovies(db, &models.FilterCriteria{Sort: "-" + tt.sort})
			if err != nil {
				t.Fatalf("GetMovies descending returned error: %v", err)
			}
			got = nil
			for _, m := range movies {
				got = append(got, m.Title)
			}
			if want := reversed(tt.want); !reflect.DeepEqual(got, want) {
				t.Fatalf("descending: got %v want %v", got, want)
			}
		})
	}

	t.Run("equal values fall back to title ascending", func(t *testing.T) {
		if _, err := db.Exec(`UPDATE movies SET year=2000`); err != nil {
			t.Fatalf("failed to update movies: %v", err)
		}
		movies, _, err := GetMovies(db, &models.FilterCriteria{Sort: "-year"})
		if err != nil {
			t.Fatalf("GetMovies returned error: %v", err)
		}
		var got []string
		for _, m := range movies {
			got = append(got, m.Title)
		}
		if want := []string{"Alpha", "bravo", "charlie"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	})

	t.Run("invalid sort returns ErrInvalidSort", func(t *testing.T) {
		for _, s := range []string{"bogus", "-bogus", "seasons", "title; DROP TABLE movies", "--year"} {
			_, _, err := GetMovies(db, &models.FilterCriteria{Sort: s})
			if !errors.Is(err, ErrInvalidSort) {
				t.Fatalf("sort %q: expected ErrInvalidSort, got %v", s, err)
			}
		}
	})
}

func TestGetSeries_Sort(t *testing.T) {
	db := setupTestDBWithMigrations(t)

	rows := []struct {
		title    string
		year     int
		seasons  int
		episodes int
		rating   float64
		size     int64
		added    string
	}{
		{"Alpha", 2000, 3, 30, 5, 300, "2026-01-03"},
		{"bravo", 2010, 1, 10, 9, 100, "2026-01-01"},
		{"charlie", 1990, 2, 20, 7, 200, "2026-01-02"},
	}
	for i, r := range rows {
		id := insertTestSeries(t, db, r.title, "ongoing", int64(200+i), int64(300+i))
		if _, err := db.Exec(`UPDATE series SET year_start=?, total_season_count=?, total_episode_count=?, rating=?, file_size=?, date_added=? WHERE id=?`,
			r.year, r.seasons, r.episodes, r.rating, r.size, r.added, id); err != nil {
			t.Fatalf("failed to update series: %v", err)
		}
	}

	tests := []struct {
		name string
		sort string
		want []string
	}{
		{"default is title ascending, case-insensitive", "", []string{"Alpha", "bravo", "charlie"}},
		{"title", "title", []string{"Alpha", "bravo", "charlie"}},
		{"year", "year", []string{"charlie", "Alpha", "bravo"}},
		{"seasons", "seasons", []string{"bravo", "charlie", "Alpha"}},
		{"episodes", "episodes", []string{"bravo", "charlie", "Alpha"}},
		{"rating", "rating", []string{"Alpha", "charlie", "bravo"}},
		{"size", "size", []string{"bravo", "charlie", "Alpha"}},
		{"added", "added", []string{"bravo", "charlie", "Alpha"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			series, _, err := GetSeries(db, &models.FilterCriteria{Sort: tt.sort})
			if err != nil {
				t.Fatalf("GetSeries returned error: %v", err)
			}
			var got []string
			for _, s := range series {
				got = append(got, s.Title)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ascending: got %v want %v", got, tt.want)
			}

			if tt.sort == "" {
				return
			}
			series, _, err = GetSeries(db, &models.FilterCriteria{Sort: "-" + tt.sort})
			if err != nil {
				t.Fatalf("GetSeries descending returned error: %v", err)
			}
			got = nil
			for _, s := range series {
				got = append(got, s.Title)
			}
			if want := reversed(tt.want); !reflect.DeepEqual(got, want) {
				t.Fatalf("descending: got %v want %v", got, want)
			}
		})
	}

	t.Run("invalid sort returns ErrInvalidSort", func(t *testing.T) {
		for _, s := range []string{"bogus", "-bogus", "duration", "title; DROP TABLE series"} {
			_, _, err := GetSeries(db, &models.FilterCriteria{Sort: s})
			if !errors.Is(err, ErrInvalidSort) {
				t.Fatalf("sort %q: expected ErrInvalidSort, got %v", s, err)
			}
		}
	})
}
