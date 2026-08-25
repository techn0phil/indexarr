package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"indexarr/internal/models"
)

func TestAuthMiddleware_DisabledPassThrough(t *testing.T) {
	authService := newTestAuthService("none")

	called := false
	handler := AuthMiddleware(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected wrapped handler to be called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusNoContent)
	}
}

func TestAuthMiddleware_EnabledNoCookieUnauthorized(t *testing.T) {
	authService := newTestAuthService("simple")

	handler := AuthMiddleware(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_EnabledInvalidTokenUnauthorizedAndClearCookie(t *testing.T) {
	authService := newTestAuthService("simple")

	handler := AuthMiddleware(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "bad-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}

	cookies := rr.Result().Cookies()
	foundCleared := false
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" && cookie.MaxAge == -1 {
			foundCleared = true
			break
		}
	}
	if !foundCleared {
		t.Fatalf("expected auth_token clear cookie in response")
	}
}

func TestAuthMiddleware_EnabledValidTokenSetsContext(t *testing.T) {
	authService := newTestAuthService("simple")
	validToken := generateTestToken(t, authService, &models.User{ID: 7, Username: "john", Role: "guest"})

	handler := AuthMiddleware(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r)
		if claims == nil {
			t.Fatalf("expected user claims in context")
		}
		if claims.UserID != 7 || claims.Username != "john" || claims.Role != "guest" {
			t.Fatalf("unexpected claims: %+v", claims)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: validToken})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", rr.Code, http.StatusOK)
	}
}
