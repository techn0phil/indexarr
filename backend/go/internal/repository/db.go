package repository

import (
	"context"
	"database/sql"
	"embed"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var db *sql.DB

func InitDB(dbPath string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", "file:"+dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	// Configure SQLite for concurrency and performance
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Enable WAL mode for concurrent reads during writes
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		return nil, err
	}

	// Set busy timeout to 5 seconds (5000ms) to handle lock contention
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout=5000"); err != nil {
		return nil, err
	}

	// Set cache size to 64MB for better performance
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA cache_size=-64000"); err != nil {
		return nil, err
	}

	// Use NORMAL synchronous mode (balance between speed and safety)
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA synchronous=NORMAL"); err != nil {
		return nil, err
	}

	// Run migrations
	if err := runMigrations(dbPath); err != nil {
		return nil, err
	}

	db = sqlDB
	return sqlDB, nil
}

func runMigrations(dbPath string) error {
	// Open the embedded filesystem for migrations
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// Create a new migration instance
	m, err := migrate.NewWithSourceInstance("iofs", d, "sqlite3://"+dbPath)
	if err != nil {
		return err
	}
	defer m.Close()

	// Run all pending migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}
