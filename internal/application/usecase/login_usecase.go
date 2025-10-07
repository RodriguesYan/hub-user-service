package usecase

import (
	"errors"
	"hub-user-service/internal/domain/service"
)

// LoginCommand represents the input for login use case
type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResult represents the output of login use case
type LoginResult struct {
	UserID        string `json:"userId"`
	Email         string `json:"email"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Token         string `json:"token"`
	EmailVerified bool   `json:"emailVerified"`
}

// ILoginUseCase defines the interface for login use case
type ILoginUseCase interface {
	Execute(cmd *LoginCommand) (*LoginResult, error)
}

// LoginUseCase implements user login business logic
type LoginUseCase struct {
	authService service.IAuthService
}

// NewLoginUseCase creates a new LoginUseCase instance
func NewLoginUseCase(authService service.IAuthService) ILoginUseCase {
	return &LoginUseCase{
		authService: authService,
	}
}

// Execute performs the login operation
func (uc *LoginUseCase) Execute(cmd *LoginCommand) (*LoginResult, error) {
	if cmd == nil {
		return nil, errors.New("login command cannot be nil")
	}

	if cmd.Email == "" {
		return nil, errors.New("email is required")
	}

	if cmd.Password == "" {
		return nil, errors.New("password is required")
	}

	user, token, err := uc.authService.Authenticate(cmd.Email, cmd.Password)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:        user.ID,
		Email:         user.GetEmailString(),
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Token:         token,
		EmailVerified: user.EmailVerified,
	}, nil
}
