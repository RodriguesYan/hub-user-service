package service

import (
	"errors"
	"hub-user-service/internal/domain/model"
	"hub-user-service/internal/domain/repository"
)

// IAuthService defines the interface for authentication business logic
type IAuthService interface {
	Authenticate(email, password string) (*model.User, string, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	CreateUserToken(user *model.User) (string, error)
}

// AuthService implements authentication business logic
type AuthService struct {
	userRepo     repository.IUserRepository
	tokenService ITokenService
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.IUserRepository, tokenService ITokenService) IAuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// Authenticate authenticates a user with email and password
func (s *AuthService) Authenticate(email, password string) (*model.User, string, error) {
	if email == "" {
		return nil, "", errors.New("email cannot be empty")
	}

	if password == "" {
		return nil, "", errors.New("password cannot be empty")
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if user == nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := user.CanLogin(); err != nil {
		user.RecordFailedLogin()
		_ = s.userRepo.Update(user)
		return nil, "", err
	}

	if !user.VerifyPasswordString(password) {
		user.RecordFailedLogin()
		_ = s.userRepo.Update(user)
		return nil, "", errors.New("invalid credentials")
	}

	user.RecordSuccessfulLogin()
	if err := s.userRepo.Update(user); err != nil {
		return nil, "", errors.New("failed to update user login information")
	}

	token, err := s.tokenService.CreateToken(user.ID, user.GetEmailString())
	if err != nil {
		return nil, "", errors.New("failed to create authentication token")
	}

	return user, token, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	return s.tokenService.ValidateToken(tokenString)
}

// CreateUserToken creates a JWT token for a user
func (s *AuthService) CreateUserToken(user *model.User) (string, error) {
	if user == nil {
		return "", errors.New("user cannot be nil")
	}

	return s.tokenService.CreateToken(user.ID, user.GetEmailString())
}
