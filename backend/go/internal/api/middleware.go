package api

import (
	"context"
	"net/http"

	"indexarr/internal/services"
)

type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware creates a middleware that validates JWT tokens from cookies
func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If auth is disabled, continue without checking
			if !authService.IsEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from cookie
			cookie, err := r.Cookie("auth_token")
			if err != nil || cookie.Value == "" {
				http.Error(w, `{"success":false,"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := authService.ValidateToken(cookie.Value)
			if err != nil {
				// Clear invalid cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				http.Error(w, `{"success":false,"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Add user claims to request context
			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves user claims from the request context
func GetUserFromContext(r *http.Request) *services.UserClaims {
	if claims, ok := r.Context().Value(userContextKey).(*services.UserClaims); ok {
		return claims
	}
	return nil
}
