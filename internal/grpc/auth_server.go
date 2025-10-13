package grpc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hub-user-service/internal/auth"
	"hub-user-service/internal/grpc/proto"
	"hub-user-service/internal/login/application/usecase"
)

// AuthServer implements the gRPC AuthService interface
type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	loginUsecase usecase.IDoLoginUsecase
	authService  auth.IAuthService
}

// NewAuthServer creates a new AuthServer instance
func NewAuthServer(loginUsecase usecase.IDoLoginUsecase, authService auth.IAuthService) *AuthServer {
	return &AuthServer{
		loginUsecase: loginUsecase,
		authService:  authService,
	}
}

// Login handles user authentication and returns a JWT token
func (s *AuthServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// Validate request
	if req.Email == "" {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "email is required",
				Code:      http.StatusBadRequest,
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	if req.Password == "" {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "password is required",
				Code:      http.StatusBadRequest,
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	// Execute login use case (existing business logic)
	user, err := s.loginUsecase.Execute(req.Email, req.Password)
	if err != nil {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   fmt.Sprintf("login failed: %v", err),
				Code:      http.StatusUnauthorized,
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	// Create JWT token using existing auth service
	token, err := s.authService.CreateToken(user.GetEmailString(), user.ID)
	if err != nil {
		return &proto.LoginResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   fmt.Sprintf("failed to create token: %v", err),
				Code:      http.StatusInternalServerError,
				Timestamp: time.Now().Unix(),
			},
		}, nil
	}

	// Return successful response
	return &proto.LoginResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "login successful",
			Code:      http.StatusOK,
			Timestamp: time.Now().Unix(),
		},
		Token: token,
		UserInfo: &proto.UserInfo{
			UserId: user.ID,
			Email:  user.GetEmailString(),
		},
	}, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	// Validate request
	if req.Token == "" {
		return &proto.ValidateTokenResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   "token is required",
				Code:      http.StatusBadRequest,
				Timestamp: time.Now().Unix(),
			},
			IsValid: false,
		}, nil
	}

	// Verify token using existing auth service
	// Note: VerifyToken expects http.ResponseWriter, but we don't need it for gRPC
	// We'll pass nil and handle the response differently
	userId, err := s.authService.VerifyToken(req.Token, nil)
	if err != nil {
		return &proto.ValidateTokenResponse{
			ApiResponse: &proto.APIResponse{
				Success:   false,
				Message:   fmt.Sprintf("token validation failed: %v", err),
				Code:      http.StatusUnauthorized,
				Timestamp: time.Now().Unix(),
			},
			IsValid: false,
		}, nil
	}

	// Token is valid
	return &proto.ValidateTokenResponse{
		ApiResponse: &proto.APIResponse{
			Success:   true,
			Message:   "token is valid",
			Code:      http.StatusOK,
			Timestamp: time.Now().Unix(),
		},
		IsValid: true,
		UserInfo: &proto.UserInfo{
			UserId: userId,
		},
		ExpiresAt: 0, // TODO: Extract expiration from token if needed
	}, nil
}
