package services

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupServicesTestDBWithMigrations(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	migrationFiles, err := filepath.Glob("../repository/migrations/*.up.sql")
	if err != nil {
		t.Fatalf("failed to list migration files: %v", err)
	}
	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", migrationFile, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("failed to execute migration %s: %v", migrationFile, err)
		}
	}

	return db
}

type rewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	if t.base == nil {
		t.base = http.DefaultTransport
	}
	return t.base.RoundTrip(clone)
}

func newRewrittenClient(server *httptest.Server) *http.Client {
	u, _ := url.Parse(server.URL)
	return &http.Client{Transport: &rewriteTransport{target: u, base: http.DefaultTransport}}
}
