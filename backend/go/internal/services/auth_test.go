package services

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

func setupAuthTestRepo(t *testing.T) *repository.UserRepository {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite memory db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	return repository.NewUserRepository(db)
}

func newSimpleAuthConfig() *config.Config {
	return &config.Config{
		AuthMode:          "simple",
		AuthAdminUsername: "env-admin",
		AuthAdminPassword: "env-password",
		AuthSessionSecret: "0123456789abcdef0123456789abcdef",
		AuthSessionMaxAge: 1,
	}
}

func TestAuthServiceValidateCredentials_EnvAdminSuccess(t *testing.T) {
	authService := NewAuthService(newSimpleAuthConfig(), nil)

	user, err := authService.ValidateCredentials("env-admin", "env-password")
	if err != nil {
		t.Fatalf("ValidateCredentials returned error: %v", err)
	}
	if user.ID != 0 || user.Role != "admin" || user.Username != "env-admin" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestAuthServiceValidateCredentials_DBUserSuccess(t *testing.T) {
	repo := setupAuthTestRepo(t)
	hash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if _, err := repo.Create("db-user", hash, "guest"); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	authService := NewAuthService(newSimpleAuthConfig(), repo)

	user, err := authService.ValidateCredentials("db-user", "secret-pass")
	if err != nil {
		t.Fatalf("ValidateCredentials returned error: %v", err)
	}
	if user.Username != "db-user" || user.Role != "guest" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestAuthServiceValidateCredentials_UserDisabled(t *testing.T) {
	repo := setupAuthTestRepo(t)
	hash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	user, err := repo.Create("disabled-user", hash, "guest")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	enabled := false
	if _, err := repo.Update(user.ID, "", "", &enabled); err != nil {
		t.Fatalf("failed to disable user: %v", err)
	}

	authService := NewAuthService(newSimpleAuthConfig(), repo)

	_, err = authService.ValidateCredentials("disabled-user", "secret-pass")
	if !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("expected ErrUserDisabled, got %v", err)
	}
}

func TestAuthServiceValidateCredentials_InvalidCredentials(t *testing.T) {
	repo := setupAuthTestRepo(t)
	hash, err := HashPassword("secret-pass")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if _, err := repo.Create("db-user", hash, "guest"); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	authService := NewAuthService(newSimpleAuthConfig(), repo)

	_, err = authService.ValidateCredentials("db-user", "wrong-pass")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceGenerateAndValidateToken(t *testing.T) {
	authService := NewAuthService(newSimpleAuthConfig(), nil)
	user := &models.User{ID: 42, Username: "alice", Role: "admin"}

	token, expiresAt, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatalf("expected future expiration")
	}

	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "alice" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestAuthServiceValidateToken_Expired(t *testing.T) {
	cfg := newSimpleAuthConfig()
	authService := NewAuthService(cfg, nil)

	expiredClaims := &UserClaims{
		UserID:   1,
		Username: "expired-user",
		Role:     "guest",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "indexarr",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, err := token.SignedString([]byte(cfg.AuthSessionSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = authService.ValidateToken(tokenString)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestAuthServiceValidateToken_Invalid(t *testing.T) {
	authService := NewAuthService(newSimpleAuthConfig(), nil)

	_, err := authService.ValidateToken("not-a-jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("my-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "my-password" {
		t.Fatalf("hash should not equal plain password")
	}
	if !VerifyPassword(hash, "my-password") {
		t.Fatalf("expected password to verify")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestAuthServiceIsEnvAdmin(t *testing.T) {
	authService := NewAuthService(newSimpleAuthConfig(), nil)

	if !authService.IsEnvAdmin(&UserClaims{UserID: 0, Username: "env-admin"}) {
		t.Fatalf("expected env-admin claims to match")
	}
	if authService.IsEnvAdmin(&UserClaims{UserID: 1, Username: "env-admin"}) {
		t.Fatalf("expected non-zero user ID not to match env admin")
	}
}
