package api

import (
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
		user, err := authService.ValidateCredentials(req.Username, req.Password)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			errorMsg := "Identifiants invalides"
			if errors.Is(err, services.ErrUserDisabled) {
				errorMsg = "Compte désactivé"
			}
			json.NewEncoder(w).Encode(LoginResponse{
				Success: false,
				Error:   errorMsg,
			})
			return
		}

		// Generate token
		token, expiresAt, err := authService.GenerateToken(user)
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
				Success: false,
				Error:   "Not authenticated",
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
				"success": false,
				"error":   "Not authenticated",
			})
			return
		}

		// Env admin cannot change password (it's managed via env vars)
		if authService.IsEnvAdmin(claims) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Le mot de passe de l'administrateur principal est géré via les variables d'environnement",
			})
			return
		}

		var req models.ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Les mots de passe sont requis",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "User repository not available",
			})
			return
		}

		// Get user from database
		user, err := userRepo.GetByID(claims.UserID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Utilisateur non trouvé",
			})
			return
		}

		// Verify current password
		if !services.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Mot de passe actuel incorrect",
			})
			return
		}

		// Hash new password
		newHash, err := services.HashPassword(req.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to hash password",
			})
			return
		}

		// Update password
		if err := userRepo.UpdatePassword(claims.UserID, newHash); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to update password",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Mot de passe modifié avec succès",
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
			"success": false,
			"error":   "Not authenticated",
		})
		return nil
	}

	if claims.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Admin access required",
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
				"success": false,
				"error":   "User repository not available",
			})
			return
		}

		users, err := userRepo.List()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to list users",
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
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		// Validate request
		if req.Username == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Nom d'utilisateur et mot de passe requis",
			})
			return
		}

		if req.Role == "" {
			req.Role = "guest"
		}

		if req.Role != "admin" && req.Role != "guest" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Rôle invalide (admin ou guest)",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "User repository not available",
			})
			return
		}

		// Hash password
		passwordHash, err := services.HashPassword(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to hash password",
			})
			return
		}

		// Create user
		user, err := userRepo.Create(req.Username, passwordHash, req.Role)
		if err != nil {
			if errors.Is(err, repository.ErrUserAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Ce nom d'utilisateur existe déjà",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to create user",
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
				"success": false,
				"error":   "Invalid user ID",
			})
			return
		}

		var req models.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		if req.Role != "" && req.Role != "admin" && req.Role != "guest" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Rôle invalide (admin ou guest)",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "User repository not available",
			})
			return
		}

		user, err := userRepo.Update(id, req.Username, req.Role, req.Enabled)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Utilisateur non trouvé",
				})
				return
			}
			if errors.Is(err, repository.ErrUserAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Ce nom d'utilisateur existe déjà",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to update user",
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
				"success": false,
				"error":   "Invalid user ID",
			})
			return
		}

		// Prevent self-deletion
		if id == claims.UserID {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Vous ne pouvez pas supprimer votre propre compte",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "User repository not available",
			})
			return
		}

		if err := userRepo.Delete(id); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Utilisateur non trouvé",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to delete user",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Utilisateur supprimé",
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
				"success": false,
				"error":   "Invalid user ID",
			})
			return
		}

		var req models.AdminSetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		if req.NewPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Le nouveau mot de passe est requis",
			})
			return
		}

		userRepo := authService.GetUserRepo()
		if userRepo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "User repository not available",
			})
			return
		}

		// Hash new password
		passwordHash, err := services.HashPassword(req.NewPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to hash password",
			})
			return
		}

		if err := userRepo.UpdatePassword(id, passwordHash); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "Utilisateur non trouvé",
				})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to update password",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Mot de passe modifié",
		})
	}
}
