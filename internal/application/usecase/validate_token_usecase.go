package usecase

import (
	"errors"
	"hub-user-service/internal/domain/service"
)

// ValidateTokenResult represents the result of token validation
type ValidateTokenResult struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"userId"`
	Email  string `json:"email"`
}

// IValidateTokenUseCase defines the interface for token validation
type IValidateTokenUseCase interface {
	Execute(token string) (*ValidateTokenResult, error)
}

// ValidateTokenUseCase implements token validation business logic
type ValidateTokenUseCase struct {
	authService service.IAuthService
}

// NewValidateTokenUseCase creates a new ValidateTokenUseCase instance
func NewValidateTokenUseCase(authService service.IAuthService) IValidateTokenUseCase {
	return &ValidateTokenUseCase{
		authService: authService,
	}
}

// Execute validates a JWT token
func (uc *ValidateTokenUseCase) Execute(token string) (*ValidateTokenResult, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	claims, err := uc.authService.ValidateToken(token)
	if err != nil {
		return &ValidateTokenResult{
			Valid: false,
		}, nil
	}

	return &ValidateTokenResult{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
