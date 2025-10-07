package usecase

import (
	"errors"
	"hub-user-service/internal/domain/repository"
)

// UserProfileResult represents user profile information
type UserProfileResult struct {
	UserID        string `json:"userId"`
	Email         string `json:"email"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	IsActive      bool   `json:"isActive"`
	EmailVerified bool   `json:"emailVerified"`
}

// IGetUserProfileUseCase defines the interface for getting user profile
type IGetUserProfileUseCase interface {
	Execute(userID string) (*UserProfileResult, error)
}

// GetUserProfileUseCase implements getting user profile business logic
type GetUserProfileUseCase struct {
	userRepo repository.IUserRepository
}

// NewGetUserProfileUseCase creates a new GetUserProfileUseCase instance
func NewGetUserProfileUseCase(userRepo repository.IUserRepository) IGetUserProfileUseCase {
	return &GetUserProfileUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves the user profile
func (uc *GetUserProfileUseCase) Execute(userID string) (*UserProfileResult, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return &UserProfileResult{
		UserID:        user.ID,
		Email:         user.GetEmailString(),
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
	}, nil
}
