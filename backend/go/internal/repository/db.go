package repository

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var db *sql.DB

func InitDB(dbPath string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
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
