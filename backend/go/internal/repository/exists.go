package repository

import (
	"database/sql"
)

// MovieExistsByFilePath returns true if a movie with the given file path exists
func MovieExistsByFilePath(db *sql.DB, filePath string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(1) FROM movies WHERE file_path = ?", filePath).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
