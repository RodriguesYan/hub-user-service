package repository

import "hub-user-service/internal/domain/model"

// IUserRepository defines the interface for user persistence operations
type IUserRepository interface {
	// Create creates a new user in the database
	Create(user *model.User) error

	// FindByID retrieves a user by their ID
	FindByID(id string) (*model.User, error)

	// FindByEmail retrieves a user by their email address
	FindByEmail(email string) (*model.User, error)

	// Update updates an existing user
	Update(user *model.User) error

	// Delete deletes a user by their ID
	Delete(id string) error

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(email string) (bool, error)

	// FindAll retrieves all users (for admin purposes)
	FindAll(limit, offset int) ([]*model.User, error)

	// CountUsers returns the total number of users
	CountUsers() (int, error)
}
