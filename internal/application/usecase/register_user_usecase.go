package usecase

import (
	"errors"
	"hub-user-service/internal/domain/model"
	"hub-user-service/internal/domain/repository"
)

// RegisterUserCommand represents the input for user registration
type RegisterUserCommand struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// RegisterUserResult represents the output of user registration
type RegisterUserResult struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// IRegisterUserUseCase defines the interface for user registration use case
type IRegisterUserUseCase interface {
	Execute(cmd *RegisterUserCommand) (*RegisterUserResult, error)
}

// RegisterUserUseCase implements user registration business logic
type RegisterUserUseCase struct {
	userRepo repository.IUserRepository
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase instance
func NewRegisterUserUseCase(userRepo repository.IUserRepository) IRegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo: userRepo,
	}
}

// Execute performs the user registration
func (uc *RegisterUserUseCase) Execute(cmd *RegisterUserCommand) (*RegisterUserResult, error) {
	if cmd == nil {
		return nil, errors.New("register command cannot be nil")
	}

	if cmd.Email == "" {
		return nil, errors.New("email is required")
	}

	if cmd.Password == "" {
		return nil, errors.New("password is required")
	}

	if cmd.FirstName == "" {
		return nil, errors.New("first name is required")
	}

	if cmd.LastName == "" {
		return nil, errors.New("last name is required")
	}

	exists, err := uc.userRepo.ExistsByEmail(cmd.Email)
	if err != nil {
		return nil, errors.New("failed to check if user exists")
	}

	if exists {
		return nil, errors.New("user with this email already exists")
	}

	user, err := model.NewUser(cmd.Email, cmd.Password, cmd.FirstName, cmd.LastName)
	if err != nil {
		return nil, err
	}

	if err := user.HashPassword(); err != nil {
		return nil, errors.New("failed to hash password")
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return &RegisterUserResult{
		UserID:    user.ID,
		Email:     user.GetEmailString(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
