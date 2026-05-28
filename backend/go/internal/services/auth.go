package services

import (
	"crypto/subtle"
	"errors"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserDisabled       = errors.New("user account is disabled")
)

// UserClaims represents the JWT claims for authenticated users
type UserClaims struct {
	UserID   int64  `json:"userId,omitempty"` // 0 for env admin
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
	}
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
// Returns the user info if valid, or an error if not
func (s *AuthService) ValidateCredentials(username, password string) (*models.User, error) {
	if s.cfg.AuthMode != "local" {
		return nil, ErrInvalidCredentials
	}

	// First, check env admin credentials (constant-time comparison)
	if s.cfg.AuthAdminUsername != "" && s.cfg.AuthAdminPassword != "" {
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.cfg.AuthAdminUsername)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.AuthAdminPassword)) == 1

		if usernameMatch && passwordMatch {
			// Return a virtual user for env admin
			return &models.User{
				ID:       0, // ID 0 indicates env admin
				Username: username,
				Role:     "admin",
				Enabled:  true,
			}, nil
		}
	}

	// Then, check database users
	if s.userRepo != nil {
		user, err := s.userRepo.GetByUsername(username)
		if err == nil && user != nil {
			// Check if user is enabled
			if !user.Enabled {
				return nil, ErrUserDisabled
			}

			// Verify password
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err == nil {
				return user, nil
			}
		}
	}

	return nil, ErrInvalidCredentials
}

// GenerateToken creates a JWT token for the authenticated user
func (s *AuthService) GenerateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.cfg.AuthSessionMaxAge) * time.Hour)

	claims := &UserClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
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

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches a hash
func VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsEnvAdmin checks if the user claims represent the env admin user
func (s *AuthService) IsEnvAdmin(claims *UserClaims) bool {
	return claims.UserID == 0 && claims.Username == s.cfg.AuthAdminUsername
}

// GetUserRepo returns the user repository (for use in handlers)
func (s *AuthService) GetUserRepo() *repository.UserRepository {
	return s.userRepo
}
