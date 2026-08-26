package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
	"indexarr/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func newDBBackedAuthService(t *testing.T) (*services.AuthService, *repository.UserRepository) {
	t.Helper()

	db := setupAPITestDBWithMigrations(t)
	userRepo := repository.NewUserRepository(db)
	cfg := &config.Config{
		AuthMode:          "simple",
		AuthAdminUsername: "admin",
		AuthAdminPassword: "password",
		AuthSessionSecret: "0123456789abcdef0123456789abcdef",
		AuthSessionMaxAge: 24,
	}

	return services.NewAuthService(cfg, userRepo), userRepo
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestHandleAuthConfig(t *testing.T) {
	authService := newTestAuthService("simple")
	h := HandleAuthConfig(authService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/config", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}

	var resp AuthConfigResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AuthMode != "simple" {
		t.Fatalf("unexpected auth mode: got %q want %q", resp.AuthMode, "simple")
	}
}

func TestHandleLogin_Success(t *testing.T) {
	authService := newTestAuthService("simple")
	h := HandleLogin(authService)

	body := jsonBody(`{"username":"admin","password":"password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}

	respBody := decodeJSONMap(t, rr)
	if success, ok := respBody["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got: %v", respBody["success"])
	}

	cookies := rr.Result().Cookies()
	foundAuthCookie := false
	for _, c := range cookies {
		if c.Name == "auth_token" && c.Value != "" && c.HttpOnly {
			foundAuthCookie = true
			break
		}
	}
	if !foundAuthCookie {
		t.Fatalf("expected auth_token cookie to be set")
	}
}

func TestHandleLogin_InvalidBody(t *testing.T) {
	authService := newTestAuthService("simple")
	h := HandleLogin(authService)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", jsonBody("{"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleLogin_InvalidCredentials(t *testing.T) {
	authService := newTestAuthService("simple")
	h := HandleLogin(authService)

	body := jsonBody(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestHandleLogout_ClearsCookie(t *testing.T) {
	h := HandleLogout()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}

	cookies := rr.Result().Cookies()
	foundCleared := false
	for _, c := range cookies {
		if c.Name == "auth_token" && c.MaxAge == -1 {
			foundCleared = true
			break
		}
	}
	if !foundCleared {
		t.Fatalf("expected auth_token clear cookie")
	}
}

func TestHandleMe_AuthDisabled(t *testing.T) {
	authService := newTestAuthService("none")
	h := HandleMe(authService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}

	respBody := decodeJSONMap(t, rr)
	if success, ok := respBody["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got: %v", respBody["success"])
	}
	if _, ok := respBody["user"]; ok {
		t.Fatalf("did not expect user key when user is nil and omitempty applies")
	}
}

func TestHandleMe_AuthEnabledNoContext(t *testing.T) {
	authService := newTestAuthService("simple")
	h := HandleMe(authService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestHandleMe_AuthEnabledWithContext(t *testing.T) {
	authService := newTestAuthService("simple")
	h := HandleMe(authService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	claims := &services.UserClaims{UserID: 9, Username: "jane", Role: "admin"}
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, claims))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}

	respBody := decodeJSONMap(t, rr)
	if success, ok := respBody["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got: %v", respBody["success"])
	}
}

func TestHandleChangePassword_Success(t *testing.T) {
	authService, userRepo := newDBBackedAuthService(t)
	h := HandleChangePassword(authService)

	hash, err := services.HashPassword("old-password")
	if err != nil {
		t.Fatalf("failed to hash initial password: %v", err)
	}
	user, err := userRepo.Create("jane", hash, "admin")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	body := jsonBody(`{"currentPassword":"old-password","newPassword":"new-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", body)
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, &services.UserClaims{UserID: user.ID, Username: user.Username, Role: user.Role}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}

	updated, err := userRepo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("failed to reload user: %v", err)
	}
	if !services.VerifyPassword(updated.PasswordHash, "new-password") {
		t.Fatalf("expected password to be updated")
	}
}

func TestHandleChangePassword_WrongCurrentPassword(t *testing.T) {
	authService, userRepo := newDBBackedAuthService(t)
	h := HandleChangePassword(authService)

	hash, err := services.HashPassword("old-password")
	if err != nil {
		t.Fatalf("failed to hash initial password: %v", err)
	}
	user, err := userRepo.Create("jane", hash, "admin")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	body := jsonBody(`{"currentPassword":"bad-password","newPassword":"new-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", body)
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, &services.UserClaims{UserID: user.ID, Username: user.Username, Role: user.Role}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestHandleListUsers_RequiresAdmin(t *testing.T) {
	authService, _ := newDBBackedAuthService(t)
	h := HandleListUsers(authService)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, &services.UserClaims{UserID: 5, Username: "guest", Role: "guest"}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusForbidden)
	}
}

func TestHandleCreateUpdateSetPasswordAndDeleteUser(t *testing.T) {
	authService, userRepo := newDBBackedAuthService(t)

	adminHash, err := services.HashPassword("admin-pass")
	if err != nil {
		t.Fatalf("failed to hash admin password: %v", err)
	}
	admin, err := userRepo.Create("admin-user", adminHash, "admin")
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	create := HandleCreateUser(authService)
	createReq := httptest.NewRequest(http.MethodPost, "/api/auth/users", jsonBody(`{"username":"new-user","password":"first-pass","role":"guest"}`))
	createReq = createReq.WithContext(context.WithValue(createReq.Context(), userContextKey, &services.UserClaims{UserID: admin.ID, Username: admin.Username, Role: "admin"}))
	createRR := httptest.NewRecorder()
	create.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("create user unexpected status: got %d want %d", createRR.Code, http.StatusOK)
	}

	created, err := userRepo.GetByUsername("new-user")
	if err != nil {
		t.Fatalf("failed to load created user: %v", err)
	}
	createdID := strconv.FormatInt(created.ID, 10)

	update := HandleUpdateUser(authService)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/auth/users/1", jsonBody(`{"username":"new-user-renamed","role":"admin","enabled":true}`))
	updateReq = withURLParam(updateReq, "id", createdID)
	updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), userContextKey, &services.UserClaims{UserID: admin.ID, Username: admin.Username, Role: "admin"}))
	updateRR := httptest.NewRecorder()
	update.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update user unexpected status: got %d want %d", updateRR.Code, http.StatusOK)
	}

	renamed, err := userRepo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("failed to reload updated user: %v", err)
	}
	if renamed.Username != "new-user-renamed" || renamed.Role != "admin" {
		t.Fatalf("unexpected updated user values: username=%q role=%q", renamed.Username, renamed.Role)
	}

	setPassword := HandleAdminSetPassword(authService)
	setPassReq := httptest.NewRequest(http.MethodPost, "/api/auth/users/1/password", jsonBody(`{"newPassword":"second-pass"}`))
	setPassReq = withURLParam(setPassReq, "id", createdID)
	setPassReq = setPassReq.WithContext(context.WithValue(setPassReq.Context(), userContextKey, &services.UserClaims{UserID: admin.ID, Username: admin.Username, Role: "admin"}))
	setPassRR := httptest.NewRecorder()
	setPassword.ServeHTTP(setPassRR, setPassReq)
	if setPassRR.Code != http.StatusOK {
		t.Fatalf("admin set password unexpected status: got %d want %d", setPassRR.Code, http.StatusOK)
	}

	changed, err := userRepo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("failed to reload password-updated user: %v", err)
	}
	if !services.VerifyPassword(changed.PasswordHash, "second-pass") {
		t.Fatalf("expected admin-set password to be applied")
	}

	deleteHandler := HandleDeleteUser(authService)
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/auth/users/1", nil)
	deleteReq = withURLParam(deleteReq, "id", createdID)
	deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), userContextKey, &services.UserClaims{UserID: admin.ID, Username: admin.Username, Role: "admin"}))
	deleteRR := httptest.NewRecorder()
	deleteHandler.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete user unexpected status: got %d want %d", deleteRR.Code, http.StatusOK)
	}

	if _, err := userRepo.GetByID(created.ID); !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf("expected deleted user to be missing, got %v", err)
	}
}

type scanTestMovieImporter struct {
	importCalled  chan struct{}
	importMovieID int64
	importErr     error
	stopCalled    bool
}

func (m *scanTestMovieImporter) Import(*models.ProgressContext) (*models.ScanResult, error) {
	if m.importCalled != nil {
		select {
		case m.importCalled <- struct{}{}:
		default:
		}
	}
	if m.importErr != nil {
		return nil, m.importErr
	}
	return &models.ScanResult{FilesProcessed: 1, MoviesAdded: 1}, nil
}

func (m *scanTestMovieImporter) ImportMovie(id int64) (*models.ScanResult, error) {
	m.importMovieID = id
	if m.importErr != nil {
		return nil, m.importErr
	}
	return &models.ScanResult{FilesProcessed: 1, MoviesAdded: 1}, nil
}

func (m *scanTestMovieImporter) GetPendingFileCount() (int, error) { return 1, nil }
func (m *scanTestMovieImporter) GetStatus() (*models.ScanStatus, error) {
	return &models.ScanStatus{Status: "idle"}, nil
}
func (m *scanTestMovieImporter) Stop()           { m.stopCalled = true }
func (m *scanTestMovieImporter) IsRunning() bool { return false }

type scanTestSeriesImporter struct {
	importCalled   chan struct{}
	importSeriesID int64
	importErr      error
	stopCalled     bool
}

func (s *scanTestSeriesImporter) Import(*models.ProgressContext) (*models.ScanResult, error) {
	if s.importCalled != nil {
		select {
		case s.importCalled <- struct{}{}:
		default:
		}
	}
	if s.importErr != nil {
		return nil, s.importErr
	}
	return &models.ScanResult{FilesProcessed: 1, EpisodesAdded: 1}, nil
}

func (s *scanTestSeriesImporter) ImportSeries(id int64) (*models.ScanResult, error) {
	s.importSeriesID = id
	if s.importErr != nil {
		return nil, s.importErr
	}
	return &models.ScanResult{FilesProcessed: 1, EpisodesAdded: 1}, nil
}

func (s *scanTestSeriesImporter) GetPendingFileCount() (int, error) { return 1, nil }
func (s *scanTestSeriesImporter) GetStatus() (*models.ScanStatus, error) {
	return &models.ScanStatus{Status: "idle"}, nil
}
func (s *scanTestSeriesImporter) Stop()           { s.stopCalled = true }
func (s *scanTestSeriesImporter) IsRunning() bool { return false }

func waitForScanCall(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for scan goroutine to invoke importer")
	}
}

func adminContext(req *http.Request) *http.Request {
	claims := &services.UserClaims{UserID: 1, Username: "admin", Role: "admin"}
	return req.WithContext(context.WithValue(req.Context(), userContextKey, claims))
}

func TestScanAndRefreshHandlers(t *testing.T) {
	logger := zerolog.Nop()
	config.GlobalLogger = &logger

	authNone := newTestAuthService("none")
	authSimple := newTestAuthService("simple")

	t.Run("trigger scan starts scheduler in goroutine", func(t *testing.T) {
		movie := &scanTestMovieImporter{importCalled: make(chan struct{}, 1)}
		scheduler := services.NewScheduler(setupAPITestDBWithMigrations(t), movie, nil, nil, 0)

		req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
		rr := httptest.NewRecorder()
		TriggerScan(scheduler, authNone).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}
		waitForScanCall(t, movie.importCalled)
	})

	t.Run("trigger movies scan starts scheduler in goroutine", func(t *testing.T) {
		movie := &scanTestMovieImporter{importCalled: make(chan struct{}, 1)}
		scheduler := services.NewScheduler(setupAPITestDBWithMigrations(t), movie, nil, nil, 0)

		req := httptest.NewRequest(http.MethodPost, "/api/scan/movies", nil)
		rr := httptest.NewRecorder()
		TriggerMoviesScan(scheduler, authNone).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}
		waitForScanCall(t, movie.importCalled)
	})

	t.Run("trigger series scan starts scheduler in goroutine", func(t *testing.T) {
		series := &scanTestSeriesImporter{importCalled: make(chan struct{}, 1)}
		scheduler := services.NewScheduler(setupAPITestDBWithMigrations(t), nil, series, nil, 0)

		req := httptest.NewRequest(http.MethodPost, "/api/scan/series", nil)
		rr := httptest.NewRecorder()
		TriggerSeriesScan(scheduler, authNone).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}
		waitForScanCall(t, series.importCalled)
	})

	t.Run("stop scan sends stop to importers", func(t *testing.T) {
		movie := &scanTestMovieImporter{}
		series := &scanTestSeriesImporter{}
		scheduler := services.NewScheduler(nil, movie, series, nil, 0)

		req := httptest.NewRequest(http.MethodPost, "/api/scan/stop", nil)
		rr := httptest.NewRecorder()
		StopScan(scheduler, authNone).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}
		if !movie.stopCalled || !series.stopCalled {
			t.Fatalf("expected stop to be called on both importers")
		}
	})

	t.Run("refresh movie passes id and returns success", func(t *testing.T) {
		movie := &scanTestMovieImporter{}
		scheduler := services.NewScheduler(nil, movie, nil, nil, 0)

		req := withURLParam(httptest.NewRequest(http.MethodPost, "/api/movies/42/refresh", nil), "id", "42")
		rr := httptest.NewRecorder()
		RefreshMovie(scheduler, authNone).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
		}
		if movie.importMovieID != 42 {
			t.Fatalf("expected refresh id 42, got %d", movie.importMovieID)
		}
	})

	t.Run("refresh series passes id and handles importer error", func(t *testing.T) {
		series := &scanTestSeriesImporter{importErr: errors.New("refresh failed")}
		scheduler := services.NewScheduler(nil, nil, series, nil, 0)

		req := withURLParam(httptest.NewRequest(http.MethodPost, "/api/series/77/refresh", nil), "id", "77")
		rr := httptest.NewRecorder()
		RefreshSeries(scheduler, authNone).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusInternalServerError)
		}
		if series.importSeriesID != 77 {
			t.Fatalf("expected refresh id 77, got %d", series.importSeriesID)
		}
	})

	t.Run("admin-only gating when auth enabled", func(t *testing.T) {
		movie := &scanTestMovieImporter{}
		scheduler := services.NewScheduler(setupAPITestDBWithMigrations(t), movie, nil, nil, 0)

		req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
		rr := httptest.NewRecorder()
		TriggerScan(scheduler, authSimple).ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status code without admin context: got %d want %d", rr.Code, http.StatusUnauthorized)
		}

		adminReq := adminContext(httptest.NewRequest(http.MethodPost, "/api/scan", nil))
		adminRR := httptest.NewRecorder()
		TriggerScan(scheduler, authSimple).ServeHTTP(adminRR, adminReq)
		if adminRR.Code != http.StatusOK {
			t.Fatalf("unexpected status code with admin context: got %d want %d", adminRR.Code, http.StatusOK)
		}
	})
}
