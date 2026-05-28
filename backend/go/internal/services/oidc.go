package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrOIDCNotConfigured = errors.New("OIDC is not configured")
	ErrInvalidState      = errors.New("invalid state parameter")
	ErrTokenExchange     = errors.New("failed to exchange token")
	ErrClaimsExtraction  = errors.New("failed to extract claims")
)

// OIDCService handles OIDC authentication
type OIDCService struct {
	cfg          *config.Config
	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	userRepo     *repository.UserRepository

	// State storage (in-memory, short-lived)
	states     map[string]time.Time
	statesMu   sync.RWMutex
	statesOnce sync.Once
}

// OIDCClaims represents the claims extracted from an OIDC token
type OIDCClaims struct {
	Subject           string   `json:"sub"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups"`
	Roles             []string `json:"roles"`
	// Custom claims map for flexible admin claim checking
	CustomClaims map[string]interface{} `json:"-"`
}

// NewOIDCService creates a new OIDC service
func NewOIDCService(cfg *config.Config, userRepo *repository.UserRepository) (*OIDCService, error) {
	if !cfg.IsOIDCAuth() || !cfg.HasOIDCConfig() {
		return nil, ErrOIDCNotConfigured
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Discover OIDC provider
	provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC provider: %w", err)
	}

	// Configure OAuth2
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.OIDCScopes,
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.OIDCClientID,
	})

	svc := &OIDCService{
		cfg:          cfg,
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		userRepo:     userRepo,
		states:       make(map[string]time.Time),
	}

	// Start state cleanup goroutine
	go svc.cleanupStates()

	// log.Printf("[OIDC] Service initialized with issuer: %s", cfg.OIDCIssuerURL)

	return svc, nil
}

// generateState creates a cryptographically secure random state string
func (s *OIDCService) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthorizationURL returns the URL to redirect users to for OIDC authentication
func (s *OIDCService) GetAuthorizationURL() (string, string, error) {
	state, err := s.generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state with expiration (10 minutes)
	s.statesMu.Lock()
	s.states[state] = time.Now().Add(10 * time.Minute)
	s.statesMu.Unlock()

	url := s.oauth2Config.AuthCodeURL(state)
	return url, state, nil
}

// ValidateState checks if the state parameter is valid
func (s *OIDCService) ValidateState(state string) bool {
	s.statesMu.Lock()
	defer s.statesMu.Unlock()

	expiry, exists := s.states[state]
	if !exists {
		return false
	}

	// Remove used state
	delete(s.states, state)

	// Check if expired
	return time.Now().Before(expiry)
}

// ExchangeCode exchanges the authorization code for tokens and extracts claims
func (s *OIDCService) ExchangeCode(ctx context.Context, code string) (*OIDCClaims, error) {
	// Exchange code for tokens
	token, err := s.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenExchange, err)
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	// Verify ID token
	_, err = s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Get access token
	rawAccessToken := token.AccessToken
	if rawAccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	// Verify access token
	accessToken, err := s.verifier.Verify(ctx, rawAccessToken)
	if err != nil {
		log.Printf("[OIDC] Warning: failed to verify access token: %v", err)
		return nil, fmt.Errorf("failed to verify access token: %w", err)
	}

	// Extract claims
	var claims OIDCClaims
	if err := accessToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClaimsExtraction, err)
	}

	// Also extract raw claims for custom admin claim checking
	var rawClaims map[string]interface{}
	if err := accessToken.Claims(&rawClaims); err == nil {
		claims.CustomClaims = rawClaims
	}

	return &claims, nil
}

// GetUserFromClaims creates or updates a user from OIDC claims
func (s *OIDCService) GetUserFromClaims(claims *OIDCClaims) (*models.User, error) {
	// Determine username from claims
	username := s.getUsernameFromClaims(claims)
	if username == "" {
		return nil, errors.New("unable to determine username from claims")
	}

	// Determine role from claims
	role := s.getRoleFromClaims(claims)

	// Check if user exists in database
	if s.userRepo != nil {
		user, err := s.userRepo.GetByUsername(username)
		if err == nil && user != nil {
			// User exists, update role if needed
			if user.Role != role {
				user, _ = s.userRepo.Update(user.ID, "", role, nil)
			}
			if !user.Enabled {
				return nil, ErrUserDisabled
			}
			return user, nil
		}

		// User doesn't exist, create if auto-create is enabled
		if s.cfg.OIDCAutoCreateUser {
			// OIDC users don't have a password hash
			user, err = s.userRepo.Create(username, "", role)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
			log.Printf("[OIDC] Created new user: %s with role: %s", username, role)
			return user, nil
		}

		return nil, errors.New("user not found and auto-create is disabled")
	}

	// No user repository, return a virtual user
	return &models.User{
		ID:       0,
		Username: username,
		Role:     role,
		Enabled:  true,
	}, nil
}

// getUsernameFromClaims extracts the username from claims based on config
func (s *OIDCService) getUsernameFromClaims(claims *OIDCClaims) string {
	switch s.cfg.OIDCUsernameClaim {
	case "email":
		return claims.Email
	case "sub":
		return claims.Subject
	case "name":
		return claims.Name
	case "preferred_username":
		fallthrough
	default:
		if claims.PreferredUsername != "" {
			return claims.PreferredUsername
		}
		// Fallback chain
		if claims.Email != "" {
			return claims.Email
		}
		if claims.Name != "" {
			return claims.Name
		}
		return claims.Subject
	}
}

// getRoleFromClaims determines the user role from claims
func (s *OIDCService) getRoleFromClaims(claims *OIDCClaims) string {
	// If no admin claim configured, everyone is a guest
	if s.cfg.OIDCRolesClaim == "" || s.cfg.OIDCAdminRoleValue == "" {
		return "guest"
	}

	adminValue := s.cfg.OIDCAdminRoleValue

	// Check known claim types
	switch s.cfg.OIDCRolesClaim {
	case "groups":
		for _, group := range claims.Groups {
			if group == adminValue {
				return "admin"
			}
		}
	case "roles":
		for _, role := range claims.Roles {
			if role == adminValue {
				return "admin"
			}
		}
	default:
		// Check custom claims
		if claims.CustomClaims != nil {
			// Handle structured claim case (e.g. "resource_access.indexarr.roles") by searching for the claim key in the raw claims map
			claimParts := strings.Split(s.cfg.OIDCRolesClaim, ".")

			var currentVal interface{} = claims.CustomClaims
			for _, part := range claimParts {
				if m, ok := currentVal.(map[string]interface{}); ok {
					if val, exists := m[part]; exists {
						currentVal = val
					} else {
						currentVal = nil
						break
					}
				} else {
					currentVal = nil
					break
				}
			}

			if currentVal != nil {
				// Check if the final value is a string that matches adminValue
				if strVal, ok := currentVal.(string); ok && strVal == adminValue {
					return "admin"
				}
				// Check if the final value is a slice of strings that contains adminValue
				if sliceVal, ok := currentVal.([]interface{}); ok {
					for _, v := range sliceVal {
						if strV, ok := v.(string); ok && strV == adminValue {
							return "admin"
						}
					}
				}
			}
		}
	}

	return "guest"
}

// cleanupStates periodically removes expired states
func (s *OIDCService) cleanupStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.statesMu.Lock()
		now := time.Now()
		for state, expiry := range s.states {
			if now.After(expiry) {
				delete(s.states, state)
			}
		}
		s.statesMu.Unlock()
	}
}

// IsConfigured returns true if OIDC is properly configured
func (s *OIDCService) IsConfigured() bool {
	return s.provider != nil && s.oauth2Config != nil
}

// GetConfig returns the OIDC-related configuration (safe for frontend)
func (s *OIDCService) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"issuer":   s.cfg.OIDCIssuerURL,
		"clientId": s.cfg.OIDCClientID,
	}
}
