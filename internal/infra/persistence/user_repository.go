package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"hub-user-service/internal/domain/model"
	"hub-user-service/internal/domain/repository"
	"time"
)

// UserRepository implements the IUserRepository interface using PostgreSQL
type UserRepository struct {
	db *sql.DB
}

// userDTO represents the database structure for user data
type userDTO struct {
	ID                  string
	Email               string
	Password            string
	FirstName           string
	LastName            string
	IsActive            bool
	EmailVerified       bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastLoginAt         *time.Time
	LockedUntil         *time.Time
	FailedLoginAttempts int
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sql.DB) repository.IUserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	query := `
		INSERT INTO yanrodrigues.users (
			id, email, password, first_name, last_name, is_active, 
			email_verified, created_at, updated_at, failed_login_attempts
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.GetEmailString(),
		user.Password.Value(),
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
		user.FailedLoginAttempts,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by their ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	query := `
		SELECT id, email, password, first_name, last_name, is_active, 
		       email_verified, created_at, updated_at, last_login_at, 
		       locked_until, failed_login_attempts
		FROM yanrodrigues.users
		WHERE id = $1
	`

	var dto userDTO
	err := r.db.QueryRow(query, id).Scan(
		&dto.ID,
		&dto.Email,
		&dto.Password,
		&dto.FirstName,
		&dto.LastName,
		&dto.IsActive,
		&dto.EmailVerified,
		&dto.CreatedAt,
		&dto.UpdatedAt,
		&dto.LastLoginAt,
		&dto.LockedUntil,
		&dto.FailedLoginAttempts,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return r.dtoToModel(&dto), nil
}

// FindByEmail retrieves a user by their email address
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	query := `
		SELECT id, email, password, first_name, last_name, is_active, 
		       email_verified, created_at, updated_at, last_login_at, 
		       locked_until, failed_login_attempts
		FROM yanrodrigues.users
		WHERE email = $1
	`

	var dto userDTO
	err := r.db.QueryRow(query, email).Scan(
		&dto.ID,
		&dto.Email,
		&dto.Password,
		&dto.FirstName,
		&dto.LastName,
		&dto.IsActive,
		&dto.EmailVerified,
		&dto.CreatedAt,
		&dto.UpdatedAt,
		&dto.LastLoginAt,
		&dto.LockedUntil,
		&dto.FailedLoginAttempts,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return r.dtoToModel(&dto), nil
}

// Update updates an existing user
func (r *UserRepository) Update(user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	query := `
		UPDATE yanrodrigues.users
		SET email = $2, password = $3, first_name = $4, last_name = $5,
		    is_active = $6, email_verified = $7, updated_at = $8,
		    last_login_at = $9, locked_until = $10, failed_login_attempts = $11
		WHERE id = $1
	`

	result, err := r.db.Exec(
		query,
		user.ID,
		user.GetEmailString(),
		user.Password.Value(),
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.EmailVerified,
		user.UpdatedAt,
		user.LastLoginAt,
		user.LockedUntil,
		user.FailedLoginAttempts,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Delete deletes a user by their ID
func (r *UserRepository) Delete(id string) error {
	if id == "" {
		return errors.New("user ID cannot be empty")
	}

	query := `DELETE FROM yanrodrigues.users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	if email == "" {
		return false, errors.New("email cannot be empty")
	}

	query := `SELECT EXISTS(SELECT 1 FROM yanrodrigues.users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// FindAll retrieves all users with pagination
func (r *UserRepository) FindAll(limit, offset int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, email, password, first_name, last_name, is_active, 
		       email_verified, created_at, updated_at, last_login_at, 
		       locked_until, failed_login_attempts
		FROM yanrodrigues.users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var dto userDTO
		err := rows.Scan(
			&dto.ID,
			&dto.Email,
			&dto.Password,
			&dto.FirstName,
			&dto.LastName,
			&dto.IsActive,
			&dto.EmailVerified,
			&dto.CreatedAt,
			&dto.UpdatedAt,
			&dto.LastLoginAt,
			&dto.LockedUntil,
			&dto.FailedLoginAttempts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, r.dtoToModel(&dto))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// CountUsers returns the total number of users
func (r *UserRepository) CountUsers() (int, error) {
	query := `SELECT COUNT(*) FROM yanrodrigues.users`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// dtoToModel converts a userDTO to a domain model
func (r *UserRepository) dtoToModel(dto *userDTO) *model.User {
	return model.NewUserFromRepository(
		dto.ID,
		dto.Email,
		dto.Password,
		dto.FirstName,
		dto.LastName,
		dto.IsActive,
		dto.EmailVerified,
		dto.CreatedAt,
		dto.UpdatedAt,
		dto.LastLoginAt,
		dto.LockedUntil,
		dto.FailedLoginAttempts,
	)
}
