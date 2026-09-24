package repository

import (
	"database/sql"
	"errors"
	"time"

	"indexarr/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("username already exists")
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	query := `SELECT id, username, password_hash, role, enabled, created_at, updated_at 
			  FROM users WHERE id = ?`

	user := &models.User{}
	var enabled int
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.Enabled = enabled == 1
	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, role, enabled, created_at, updated_at 
			  FROM users WHERE username = ?`

	user := &models.User{}
	var enabled int
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	user.Enabled = enabled == 1
	return user, nil
}

// List retrieves all users
func (r *UserRepository) List() ([]*models.User, error) {
	query := `SELECT id, username, password_hash, role, enabled, created_at, updated_at 
			  FROM users ORDER BY username ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var enabled int
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&enabled,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		user.Enabled = enabled == 1
		users = append(users, user)
	}

	return users, rows.Err()
}

// Create creates a new user
func (r *UserRepository) Create(username, passwordHash, role string) (*models.User, error) {
	// Check if username already exists
	existing, err := r.GetByUsername(username)
	if err == nil && existing != nil {
		return nil, ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	now := time.Now()
	query := `INSERT INTO users (username, password_hash, role, enabled, created_at, updated_at) 
			  VALUES (?, ?, ?, 1, ?, ?)`

	result, err := r.db.Exec(query, username, passwordHash, role, now, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Update updates a user's details (not password)
func (r *UserRepository) Update(id int64, username, role string, enabled *bool) (*models.User, error) {
	// Get existing user
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check if new username conflicts with existing user
	if username != "" && username != user.Username {
		existing, err := r.GetByUsername(username)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		user.Username = username
	}

	if role != "" {
		user.Role = role
	}

	if enabled != nil {
		user.Enabled = *enabled
	}

	user.UpdatedAt = time.Now()

	enabledInt := 0
	if user.Enabled {
		enabledInt = 1
	}

	query := `UPDATE users SET username = ?, role = ?, enabled = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.Exec(query, user.Username, user.Role, enabledInt, user.UpdatedAt, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(id int64, passwordHash string) error {
	now := time.Now()
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	result, err := r.db.Exec(query, passwordHash, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// CountAdmins returns the number of admin users
func (r *UserRepository) CountAdmins() (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = 'admin' AND enabled = 1`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
