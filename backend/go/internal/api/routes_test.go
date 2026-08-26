package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"indexarr/internal/config"
	"indexarr/internal/services"
)

func TestSetupRoutes_HealthAndProtectedRoutesWithoutAuth(t *testing.T) {
	db := setupAPITestDBWithMigrations(t)
	authService := newTestAuthService("none")
	cfg := &config.Config{AuthMode: "none"}

	router := SetupRoutes(db, cfg, nil, nil, authService)

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRR := httptest.NewRecorder()
	router.ServeHTTP(healthRR, healthReq)
	if healthRR.Code != http.StatusOK {
		t.Fatalf("unexpected health status: got %d want %d", healthRR.Code, http.StatusOK)
	}

	moviesReq := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	moviesRR := httptest.NewRecorder()
	router.ServeHTTP(moviesRR, moviesReq)
	if moviesRR.Code != http.StatusOK {
		t.Fatalf("unexpected movies status: got %d want %d", moviesRR.Code, http.StatusOK)
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	statsRR := httptest.NewRecorder()
	router.ServeHTTP(statsRR, statsReq)
	if statsRR.Code != http.StatusOK {
		t.Fatalf("unexpected stats status: got %d want %d", statsRR.Code, http.StatusOK)
	}
}

func TestSetupRoutes_ScanStatusRouteRegistrationDependsOnScheduler(t *testing.T) {
	db := setupAPITestDBWithMigrations(t)
	authService := newTestAuthService("none")
	cfg := &config.Config{AuthMode: "none"}

	withoutScheduler := SetupRoutes(db, cfg, nil, nil, authService)
	withoutReq := httptest.NewRequest(http.MethodGet, "/api/scan/status", nil)
	withoutRR := httptest.NewRecorder()
	withoutScheduler.ServeHTTP(withoutRR, withoutReq)
	if withoutRR.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when scheduler is nil, got %d", withoutRR.Code)
	}

	scheduler := services.NewScheduler(db, nil, nil, nil, 0)
	withScheduler := SetupRoutes(db, cfg, scheduler, nil, authService)
	withReq := httptest.NewRequest(http.MethodGet, "/api/scan/status", nil)
	withRR := httptest.NewRecorder()
	withScheduler.ServeHTTP(withRR, withReq)
	if withRR.Code != http.StatusOK {
		t.Fatalf("expected 200 when scheduler is provided, got %d", withRR.Code)
	}
}
