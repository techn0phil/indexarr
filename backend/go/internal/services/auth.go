package services

import (
	"crypto/subtle"
	"errors"
	"time"

	"indexarr/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

// UserClaims represents the JWT claims for authenticated users
type UserClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	cfg *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

// IsEnabled returns true if authentication is enabled
func (s *AuthService) IsEnabled() bool {
	return s.cfg.HasAuthEnabled()
}

// GetAuthMode returns the current authentication mode
func (s *AuthService) GetAuthMode() string {
	return s.cfg.AuthMode
}

// ValidateCredentials checks if the provided credentials are valid
func (s *AuthService) ValidateCredentials(username, password string) bool {
	if s.cfg.AuthMode != "simple" {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.cfg.AuthAdminUsername)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.AuthAdminPassword)) == 1

	return usernameMatch && passwordMatch
}

// GenerateToken creates a JWT token for the authenticated user
func (s *AuthService) GenerateToken(username string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.AuthSessionMaxAge) * time.Hour)

	claims := &UserClaims{
		Username: username,
		Role:     "admin", // For Step 1, all authenticated users are admin
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "indexarr",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.AuthSessionSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken parses and validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.AuthSessionSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
