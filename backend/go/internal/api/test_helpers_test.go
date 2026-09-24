package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/services"

	_ "github.com/mattn/go-sqlite3"
)

func newTestAuthService(authMode string) *services.AuthService {
	cfg := &config.Config{
		AuthMode:          authMode,
		AuthAdminUsername: "admin",
		AuthAdminPassword: "password",
		AuthSessionSecret: "0123456789abcdef0123456789abcdef",
		AuthSessionMaxAge: 24,
	}
	return services.NewAuthService(cfg, nil)
}

func generateTestToken(t *testing.T, authService *services.AuthService, user *models.User) string {
	t.Helper()
	token, _, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	return token
}

func decodeJSONMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}
	return body
}

func jsonBody(payload string) *bytes.Buffer {
	return bytes.NewBufferString(payload)
}

func setupAPITestDBWithMigrations(t *testing.T) *sql.DB {
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
