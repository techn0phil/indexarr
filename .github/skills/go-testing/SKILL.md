---
name: go-testing
description: "Create and implement Go tests for Indexarr backend: table-driven tests, repository mocks, service integration tests, API handler tests. Use when: writing tests, creating test files, implementing test coverage, testing Go services, mocking database connections."
---

# Go Testing Skill — Indexarr Backend

Generate comprehensive tests for Indexarr's Go backend following established Go testing conventions and project patterns.

## When to Use

- Creating new `*_test.go` files
- Implementing unit tests for services, repositories, or API handlers
- Writing table-driven tests for complex logic
- Mocking database connections or external API clients
- Testing filter logic, pagination, or SQL query builders
- Integration testing with SQLite in-memory databases

## Project Context

**Test Status**: No test files currently exist. This skill helps bootstrap testing infrastructure.

**Key Testing Targets**:
- Repository layer: `internal/repository/queries.go`, `mutations.go`
- Services: `internal/services/scanner.go`, `tmdb.go`, `tvdb.go`, `extractor.go`
- API handlers: `internal/api/handlers.go`, `scan_handlers.go`
- Utilities: Filter builders, path mapping, normalization functions

## Testing Patterns

### 1. Table-Driven Tests

Use table-driven tests for testing multiple scenarios with different inputs:

```go
func TestBuildOrClause(t *testing.T) {
	tests := []struct {
		name       string
		fieldName  string
		filterValue string
		want       string
	}{
		{
			name:       "single value",
			fieldName:  "resolution",
			filterValue: "3840",
			want:       "(resolution LIKE '%3840%')",
		},
		{
			name:       "multiple values",
			fieldName:  "codec",
			filterValue: "H.265,H.264",
			want:       "(codec LIKE '%H.265%' OR codec LIKE '%H.264%')",
		},
		{
			name:       "empty value",
			fieldName:  "resolution",
			filterValue: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildOrClause(tt.fieldName, tt.filterValue)
			if got != tt.want {
				t.Errorf("buildOrClause() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

### 2. Repository Tests with In-Memory SQLite

Test database operations using in-memory SQLite:

```go
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	// Run migrations or create schema
	_, err = db.Exec(schemaSQL) // Load from migrations
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestInsertMovie(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	movie := &models.Movie{
		Title: "Test Movie",
		Year:  2024,
		Status: "available",
	}

	err := repository.InsertMovie(db, movie)
	if err != nil {
		t.Fatalf("InsertMovie() error = %v", err)
	}

	if movie.ID == 0 {
		t.Error("InsertMovie() did not set movie ID")
	}

	// Verify movie was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM movies WHERE id = ?", movie.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query movie: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 movie, got %d", count)
	}
}
```

### 3. Service Tests with Mocks

Mock external dependencies (TMDB, TVDB, filesystem) for service tests:

```go
type mockTMDBClient struct {
	searchFunc func(title string) (*models.Movie, error)
}

func (m *mockTMDBClient) SearchMovie(title string) (*models.Movie, error) {
	if m.searchFunc != nil {
		return m.searchFunc(title)
	}
	return nil, fmt.Errorf("not implemented")
}

func TestEnrichMovie(t *testing.T) {
	mockTMDB := &mockTMDBClient{
		searchFunc: func(title string) (*models.Movie, error) {
			return &models.Movie{
				Title:  title,
				Rating: 8.5,
				TMDBId: 12345,
			}, nil
		},
	}

	movie := &models.Movie{Title: "Test Movie"}
	err := enrichMovieWithTMDB(mockTMDB, movie)
	
	if err != nil {
		t.Fatalf("enrichMovieWithTMDB() error = %v", err)
	}
	if movie.TMDBId != 12345 {
		t.Errorf("expected TMDB ID 12345, got %d", movie.TMDBId)
	}
}
```

### 4. API Handler Tests

Test HTTP handlers using `httptest`:

```go
func TestGetMoviesHandler(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	repository.InsertMovie(db, &models.Movie{Title: "Movie 1", Year: 2023})
	repository.InsertMovie(db, &models.Movie{Title: "Movie 2", Year: 2024})

	req := httptest.NewRequest("GET", "/api/movies?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	handler := NewMovieHandler(db)
	handler.GetMovies(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	if !result["success"].(bool) {
		t.Error("expected success=true")
	}
}
```

### 5. Authentication & Security Tests

Test JWT tokens, password hashing, and auth handlers:

```go
func TestAuthService_LoginSuccess(t *testing.T) {
	t.Parallel()
	
	cfg := &config.Config{
		JWTSecret:     "test-secret-key-32-chars-minimum!",
		AuthMode:      "database",
		TokenDuration: time.Hour,
	}
	
	userRepo := &mockUserRepository{
		getByUsernameFunc: func(username string) (*models.User, error) {
			// Simulate a user with bcrypt-hashed password
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &models.User{
				ID:           1,
				Username:     username,
				PasswordHash: string(hashedPassword),
				Role:         "admin",
				Enabled:      true,
			}, nil
		},
	}
	
	authService := services.NewAuthService(cfg, userRepo)
	
	token, err := authService.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	
	if token == "" {
		t.Error("expected token to be non-empty")
	}
	
	// Verify token is valid
	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	
	if claims.Username != "testuser" {
		t.Errorf("expected username=testuser, got %s", claims.Username)
	}
}

func TestAuthService_LoginInvalidCredentials(t *testing.T) {
	t.Parallel()
	
	cfg := &config.Config{
		JWTSecret:     "test-secret-key-32-chars-minimum!",
		AuthMode:      "database",
		TokenDuration: time.Hour,
	}
	
	userRepo := &mockUserRepository{
		getByUsernameFunc: func(username string) (*models.User, error) {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			return &models.User{
				ID:           1,
				Username:     username,
				PasswordHash: string(hashedPassword),
				Role:         "admin",
				Enabled:      true,
			}, nil
		},
	}
	
	authService := services.NewAuthService(cfg, userRepo)
	
	_, err := authService.Login("testuser", "wrongpassword")
	if err == nil {
		t.Error("expected error for invalid credentials")
	}
	
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_TokenExpiration(t *testing.T) {
	t.Parallel()
	
	cfg := &config.Config{
		JWTSecret:     "test-secret-key-32-chars-minimum!",
		AuthMode:      "database",
		TokenDuration: -time.Hour, // Expired token
	}
	
	userRepo := &mockUserRepository{}
	authService := services.NewAuthService(cfg, userRepo)
	
	// Create a token with expired duration
	claims := &services.UserClaims{
		UserID:   1,
		Username: "testuser",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
	
	_, err := authService.ValidateToken(tokenString)
	if !errors.Is(err, services.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

// Mock user repository for auth testing
type mockUserRepository struct {
	getByUsernameFunc func(username string) (*models.User, error)
}

func (m *mockUserRepository) GetByUsername(username string) (*models.User, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(username)
	}
	return nil, repository.ErrUserNotFound
}
```

### 6. Protected Endpoint Tests

Test HTTP handlers with JWT authentication:

```go
func TestGetMoviesHandler_WithAuth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Create a test token
	token := generateTestJWT("testuser", "admin", "test-secret")
	
	req := httptest.NewRequest("GET", "/api/movies", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	
	handler := NewMovieHandler(db)
	authMiddleware := middleware.AuthMiddleware("test-secret")
	
	// Wrap handler with auth middleware
	authHandler := authMiddleware(handler.GetMovies)
	authHandler.ServeHTTP(w, req)
	
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestProtectedEndpoint_MissingToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/movies", nil)
	// No Authorization header
	w := httptest.NewRecorder()
	
	handler := NewMovieHandler(nil)
	authMiddleware := middleware.AuthMiddleware("test-secret")
	authHandler := authMiddleware(handler.GetMovies)
	authHandler.ServeHTTP(w, req)
	
	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}
```

### 7. Integration Tests

Test full workflows with real dependencies:

```go
func TestFullScanWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup
	db := setupTestDB(t)
	defer db.Close()
	
	cfg := &config.Config{
		MediaLibraryPaths: []string{"testdata/movies"},
		MediainfoPath:     "/usr/bin/mediainfo",
	}
	
	scanner := services.NewScanner(db, cfg, nil)
	
	// Run scan
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	
	if result.TotalFiles == 0 {
		t.Error("expected files to be scanned")
	}
}
```

## Test File Organization

```
backend/go/
├── internal/
│   ├── repository/
│   │   ├── queries.go
│   │   ├── queries_test.go          # Table-driven tests for filters
│   │   ├── mutations.go
│   │   ├── mutations_test.go        # Insert/update/delete tests
│   │   ├── user_repository.go
│   │   └── user_repository_test.go  # User CRUD operations
│   ├── services/
│   │   ├── scanner.go
│   │   ├── scanner_test.go          # Mock filesystem, DB
│   │   ├── tmdb.go
│   │   ├── tmdb_test.go             # Mock HTTP responses
│   │   ├── auth.go
│   │   ├── auth_test.go             # JWT, password hashing, login flows
│   │   ├── extractor.go
│   │   └── extractor_test.go        # Mock mediainfo output
│   └── api/
│       ├── handlers.go
│       ├── handlers_test.go         # httptest with mock DB
│       ├── auth_handlers.go
│       ├── auth_handlers_test.go    # Auth endpoints (login, token validation)
│       ├── middleware.go
│       └── middleware_test.go       # JWT validation middleware tests
└── testdata/                        # Test fixtures
    ├── movies/
    │   └── sample.mkv
    └── mediainfo/
        └── sample_output.json
```

## Running Tests

```bash
# Run all tests
cd backend/go
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/repository

# Run specific test
go test -run TestBuildOrClause ./internal/repository

# Skip integration tests (fast unit tests only)
go test -short ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Key Testing Principles

1. **Test Public APIs**: Focus on exported functions and types
2. **Use Table-Driven Tests**: For functions with multiple input scenarios
3. **Mock External Dependencies**: TMDB, TVDB, filesystem operations, HTTP clients
4. **In-Memory DB for Repository**: Use SQLite `:memory:` for fast tests
5. **Test Edge Cases**: Empty inputs, nil pointers, error conditions, invalid tokens
6. **Keep Tests Fast**: Unit tests should run in milliseconds
7. **Use `testing.Short()`**: Flag integration tests that require external resources
8. **Parallel Tests**: Use `t.Parallel()` for tests that can run concurrently

## Common Gotchas

- **SQLite Pragmas**: In-memory DBs may need WAL mode disabled or explicit transactions
- **Cleanup**: Always `defer db.Close()` and `defer rows.Close()`
- **Context Timeouts**: Test context cancellation for long-running operations
- **Race Conditions**: Run `go test -race` to detect concurrent access issues
- **Nil Checks**: Scanner services can be nil if not configured—test both cases
- **JWT Secret Management**: Test with different secret lengths; ensure 32+ chars for HS256
- **Token Expiration**: Use negative durations or manipulate `ExpiresAt` for testing expired tokens
- **Password Hashing**: Always use bcrypt for testing; never compare plain text passwords
- **Auth Modes**: Test both "env-admin" (no DB users) and "database" auth modes

## Example: Complete Test File

```go
package repository

import (
	"database/sql"
	"testing"

	"indexarr/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	
	// Load schema from migration
	schema := `CREATE TABLE movies (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		year INTEGER,
		status TEXT DEFAULT 'available'
	);`
	
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	
	return db
}

func TestGetMovies(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	// Insert test data
	_, err := db.Exec("INSERT INTO movies (title, year) VALUES (?, ?)", "Movie 1", 2023)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}
	
	// Test
	filters := &models.FilterCriteria{
		Page:     1,
		PageSize: 10,
	}
	
	movies, total, err := GetMovies(db, filters)
	if err != nil {
		t.Fatalf("GetMovies() error = %v", err)
	}
	
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	
	if len(movies) != 1 {
		t.Errorf("expected 1 movie, got %d", len(movies))
	}
	
	if movies[0].Title != "Movie 1" {
		t.Errorf("expected title='Movie 1', got '%s'", movies[0].Title)
	}
}
```

## Next Steps After Creating Tests

1. **Add to CI/CD**: Run `go test ./...` in GitHub Actions workflow
2. **Coverage Goals**: Aim for >70% coverage on critical paths (repository, services)
3. **Benchmark Tests**: Add benchmark tests for performance-critical functions
4. **Integration Suite**: Create separate integration test suite with Docker fixtures
5. **Test Documentation**: Document test setup requirements in `backend/go/README.md`
