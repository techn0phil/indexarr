package api

import (
	"encoding/json"
	"net/http"
	"time"

	"indexarr/internal/services"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserResponse represents the user info in responses
type UserResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool          `json:"success"`
	User    *UserResponse `json:"user,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// AuthConfigResponse represents the auth configuration response
type AuthConfigResponse struct {
	AuthMode string `json:"authMode"`
}

// HandleAuthConfig returns the current authentication configuration (public endpoint)
func HandleAuthConfig(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := AuthConfigResponse{
			AuthMode: authService.GetAuthMode(),
		}

		json.NewEncoder(w).Encode(response)
	}
}

// HandleLogin authenticates a user and sets the auth cookie
func HandleLogin(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Parse request body
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   "Invalid request body",
			})
			return
		}

		// Validate credentials
		if !authService.ValidateCredentials(req.Username, req.Password) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   "Identifiants invalides",
			})
			return
		}

		// Generate token
		token, expiresAt, err := authService.GenerateToken(req.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   "Failed to generate token",
			})
			return
		}

		// Set httpOnly cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   r.TLS != nil, // Set Secure flag only for HTTPS
		})

		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			User: &UserResponse{
				Username: req.Username,
				Role:     "admin",
			},
		})
	}
}

// HandleLogout clears the auth cookie
func HandleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Clear the auth cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}
}

// HandleMe returns the current authenticated user's info
func HandleMe(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// If auth is disabled, return a default response
		if !authService.IsEnabled() {
			json.NewEncoder(w).Encode(LoginResponse{
				Success: true,
				User:    nil, // No user when auth is disabled
			})
			return
		}

		// Get user from context (set by middleware)
		claims := GetUserFromContext(r)
		if claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   "Not authenticated",
			})
			return
		}

		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			User: &UserResponse{
				Username: claims.Username,
				Role:     claims.Role,
			},
		})
	}
}
