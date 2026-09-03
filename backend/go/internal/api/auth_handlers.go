package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"indexarr/internal/models"
	"indexarr/internal/repository"
	"indexarr/internal/services"

	"github.com/go-chi/chi/v5"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserResponse represents the user info in responses
type UserResponse struct {
	ID       int64  `json:"id,omitempty"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success     bool          `json:"success"`
	User        *UserResponse `json:"user,omitempty"`
	Error       string        `json:"error,omitempty"`
	Description string        `json:"description,omitempty"`
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
				Success:     false,
				Error:       "invalidRequestBody",
				Description: "Invalid request body",
			})
			return
		}

		// Validate credentials
		user, err := authService.ValidateCredentials(req.Username, req.Password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			errorMsg := "invalidCredentials"
			descriptionMsg := "Invalid credentials"
			if errors.Is(err, services.ErrUserDisabled) {
				errorMsg = "userDisabled"
				descriptionMsg = "User account is disabled"
			}
			json.NewEncoder(w).Encode(LoginResponse{
				Success:     false,
				Error:       errorMsg,
				Description: descriptionMsg,
			})
			return
		}

		// Generate token
		token, expiresAt, err := authService.GenerateToken(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LoginResponse{
				Success:     false,
				Error:       "failedToGenerateToken",
				Description: "Failed to generate token",
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
				ID:       user.ID,
				Username: user.Username,
				Role:     user.Role,
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
				Success:     false,
				Error:       "notAuthenticated",
				Description: "User is not authenticated",
			})
			return
		}

		json.NewEncoder(w).Encode(LoginResponse{
			Success: true,
			User: &UserResponse{
				ID:       claims.UserID,
				Username: claims.Username,
				Role:     claims.Role,
			},
		})
	}
}

// HandleChangePassword allows users to change their own password
func HandleChangePassword(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := GetUserFromContext(r)
		if claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "notAuthenticated",
				"description": "Not authenticated",
			})
			return
		}

		// Env admin cannot change password (it's managed via env vars)
		if authService.IsEnvAdmin(claims) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "adminPasswordManagedViaEnv",
				"description": "Password of the main administrator is managed via environment variables",
			})
			return
		}

		var req models.ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidRequestBody",
				"description": "Invalid request body",
			})
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "passwordsRequired",
				"description": "Passwords are required",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userRepositoryNotAvailable",
				"description": "User repository not available",
			})
			return
		}

		// Get user from database
		user, err := userRepo.GetByID(claims.UserID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userNotFound",
				"description": "User not found",
			})
			return
		}

		// Verify current password
		if !services.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "wrongPassword",
				"description": "Current password is incorrect",
			})
			return
		}

		// Hash new password
		newHash, err := services.HashPassword(req.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "hashingPasswordFailed",
				"description": "Failed to hash password",
			})
			return
		}

		// Update password
		if err := userRepo.UpdatePassword(claims.UserID, newHash); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "updatePasswordFailed",
				"description": "Failed to update password",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"message":     "passwordUpdated",
			"description": "Password updated successfully",
		})
	}
}

// ============================================================================
// User Management Handlers (Admin only)
// ============================================================================

// requireAdmin is a helper that checks if the user is an admin
func requireAdmin(w http.ResponseWriter, r *http.Request) *services.UserClaims {
	claims := GetUserFromContext(r)
	if claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     false,
			"error":       "notAuthenticated",
			"description": "User is not authenticated",
		})
		return nil
	}

	if claims.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     false,
			"error":       "adminAccessRequired",
			"description": "Admin access required",
		})
		return nil
	}

	return claims
}

// HandleListUsers returns all users (admin only)
func HandleListUsers(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if requireAdmin(w, r) == nil {
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userRepositoryNotAvailable",
				"description": "User repository not available",
			})
			return
		}

		users, err := userRepo.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToFetchUsers",
				"description": "Failed to retrieve users from the database",
			})
			return
		}

		// Convert to response format
		var userResponses []*models.UserResponse
		for _, user := range users {
			userResponses = append(userResponses, user.ToResponse())
		}

		if userResponses == nil {
			userResponses = []*models.UserResponse{}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    userResponses,
		})
	}
}

// HandleCreateUser creates a new user (admin only)
func HandleCreateUser(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if requireAdmin(w, r) == nil {
			return
		}

		var req models.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidRequestBody",
				"description": "Invalid request body",
			})
			return
		}

		// Validate request
		if req.Username == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "usernameAndPasswordRequired",
				"description": "Username and password are required",
			})
			return
		}

		if req.Role == "" {
			req.Role = "guest"
		}

		if req.Role != "admin" && req.Role != "guest" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidRole",
				"description": "Invalid role",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userRepositoryNotAvailable",
				"description": "User repository not available",
			})
			return
		}

		// Hash password
		passwordHash, err := services.HashPassword(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToHashPassword",
				"description": "Failed to hash password",
			})
			return
		}

		// Create user
		user, err := userRepo.Create(req.Username, passwordHash, req.Role)
		if err != nil {
			if errors.Is(err, repository.ErrUserAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"error":       "userAlreadyExists",
					"description": "This username already exists",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToCreateUser",
				"description": "Failed to create user",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    user.ToResponse(),
		})
	}
}

// HandleUpdateUser updates a user (admin only)
func HandleUpdateUser(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if requireAdmin(w, r) == nil {
			return
		}

		// Get user ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidUserID",
				"description": "Invalid user ID",
			})
			return
		}

		var req models.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidRequestBody",
				"description": "Invalid request body",
			})
			return
		}

		if req.Role != "" && req.Role != "admin" && req.Role != "guest" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidRole",
				"description": "Invalid role",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userRepositoryNotAvailable",
				"description": "User repository not available",
			})
			return
		}

		user, err := userRepo.Update(id, req.Username, req.Role, req.Enabled)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"error":       "userNotFound",
					"description": "User not found",
				})
				return
			}
			if errors.Is(err, repository.ErrUserAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"error":       "userAlreadyExists",
					"description": "This username already exists",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToUpdateUser",
				"description": "Failed to update user",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    user.ToResponse(),
		})
	}
}

// HandleDeleteUser deletes a user (admin only)
func HandleDeleteUser(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		claims := requireAdmin(w, r)
		if claims == nil {
			return
		}

		// Get user ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidUserID",
				"description": "Invalid user ID",
			})
			return
		}

		// Prevent self-deletion
		if id == claims.UserID {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "cannotDeleteSelf",
				"description": "You cannot delete your own account",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userRepositoryNotAvailable",
				"description": "User repository not available",
			})
			return
		}

		if err := userRepo.Delete(id); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"error":       "userNotFound",
					"description": "User not found",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToDeleteUser",
				"description": "Failed to delete user",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"message":     "userDeleted",
			"description": "User deleted successfully",
		})
	}
}

// HandleAdminSetPassword allows admin to set a user's password (admin only)
func HandleAdminSetPassword(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if requireAdmin(w, r) == nil {
			return
		}

		// Get user ID from URL
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidUserID",
				"description": "Invalid user ID",
			})
			return
		}

		var req models.AdminSetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "invalidRequestBody",
				"description": "Invalid request body",
			})
			return
		}

		if req.NewPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "newPasswordRequired",
				"description": "New password is required",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "userRepositoryNotAvailable",
				"description": "User repository not available",
			})
			return
		}

		// Hash new password
		passwordHash, err := services.HashPassword(req.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToHashPassword",
				"description": "Failed to hash password",
			})
			return
		}

		if err := userRepo.UpdatePassword(id, passwordHash); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":     false,
					"error":       "userNotFound",
					"description": "User not found",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     false,
				"error":       "failedToUpdatePassword",
				"description": "Failed to update password",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"message":     "passwordUpdated",
			"description": "Password updated successfully",
		})
	}
}

// ============================================================================
// Scan and Refresh Handlers (Admin only)
// ============================================================================

// TriggerScan starts a manual scan
func TriggerScan(scheduler *services.Scheduler, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		// Start scan in goroutine so we can return immediately
		go func() {
			scheduler.TriggerScan()
		}()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Scan started",
		})
	}
}

// TriggerMoviesScan starts a manual scan for movies only
func TriggerMoviesScan(scheduler *services.Scheduler, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		go func() {
			scheduler.TriggerMoviesScan()
		}()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Movies scan started",
		})
	}
}

// TriggerSeriesScan starts a manual scan for series only
func TriggerSeriesScan(scheduler *services.Scheduler, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		go func() {
			scheduler.TriggerSeriesScan()
		}()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Series scan started",
		})
	}
}

// StopScan stops the currently running scan
func StopScan(scheduler *services.Scheduler, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		scheduler.StopCurrentScan()

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Stop signal sent",
		})
	}
}

func RefreshMovie(scheduler *services.Scheduler, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			respondError(w, 400, "Invalid movie ID")
			return
		}

		result, err := scheduler.TriggerSingleMovieScan(id)
		if err != nil {
			respondError(w, 500, "Failed to refresh movie: "+err.Error())
			return
		}

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Movie refresh started",
			"result":  result,
		})
	}
}

func RefreshSeries(scheduler *services.Scheduler, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			respondError(w, 400, "Invalid series ID")
			return
		}

		result, err := scheduler.TriggerSingleSeriesScan(id)
		if err != nil {
			respondError(w, 500, "Failed to refresh series: "+err.Error())
			return
		}

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Series refresh started",
			"result":  result,
		})
	}
}

// Purge deletes all data from the database (admin only)
func Purge(db *sql.DB, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if authService.GetAuthMode() != "none" && requireAdmin(w, r) == nil {
			return
		}

		err := repository.PurgeDatabase(db)
		if err != nil {
			respondError(w, 500, "Failed to purge database: "+err.Error())
			return
		}

		respond(w, map[string]interface{}{
			"success": true,
			"message": "Database purged successfully",
		})
	}
}
