package repository

import (
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

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

	// Execute embedded schema
	if _, err := sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	db = sqlDB
	return sqlDB, nil
}

func GetDB() *sql.DB {
	return db
}
