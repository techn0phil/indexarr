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

### 5. Integration Tests

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
│   │   └── mutations_test.go        # Insert/update/delete tests
│   ├── services/
│   │   ├── scanner.go
│   │   ├── scanner_test.go          # Mock filesystem, DB
│   │   ├── tmdb.go
│   │   ├── tmdb_test.go             # Mock HTTP responses
│   │   ├── extractor.go
│   │   └── extractor_test.go        # Mock mediainfo output
│   └── api/
│       ├── handlers.go
│       └── handlers_test.go         # httptest with mock DB
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
3. **Mock External Dependencies**: TMDB, TVDB, filesystem operations
4. **In-Memory DB for Repository**: Use SQLite `:memory:` for fast tests
5. **Test Edge Cases**: Empty inputs, nil pointers, error conditions
6. **Keep Tests Fast**: Unit tests should run in milliseconds
7. **Use `testing.Short()`**: Flag integration tests that require external resources
8. **Parallel Tests**: Use `t.Parallel()` for tests that can run concurrently

## Common Gotchas

- **SQLite Pragmas**: In-memory DBs may need WAL mode disabled or explicit transactions
- **Cleanup**: Always `defer db.Close()` and `defer rows.Close()`
- **Context Timeouts**: Test context cancellation for long-running operations
- **Race Conditions**: Run `go test -race` to detect concurrent access issues
- **Nil Checks**: Scanner services can be nil if not configured—test both cases

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
