package grpc

import (
	"context"
	"hub-user-service/internal/application/usecase"
	pb "hub-user-service/proto/pb"
)

// UserGRPCServer implements the gRPC UserService
type UserGRPCServer struct {
	pb.UnimplementedUserServiceServer
	loginUseCase          usecase.ILoginUseCase
	registerUseCase       usecase.IRegisterUserUseCase
	getUserProfileUseCase usecase.IGetUserProfileUseCase
	validateTokenUseCase  usecase.IValidateTokenUseCase
}

// NewUserGRPCServer creates a new UserGRPCServer instance
func NewUserGRPCServer(
	loginUC usecase.ILoginUseCase,
	registerUC usecase.IRegisterUserUseCase,
	getUserProfileUC usecase.IGetUserProfileUseCase,
	validateTokenUC usecase.IValidateTokenUseCase,
) *UserGRPCServer {
	return &UserGRPCServer{
		loginUseCase:          loginUC,
		registerUseCase:       registerUC,
		getUserProfileUseCase: getUserProfileUC,
		validateTokenUseCase:  validateTokenUC,
	}
}

// UserLogin handles user authentication via gRPC
func (s *UserGRPCServer) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.UserLoginResponse, error) {
	cmd := &usecase.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := s.loginUseCase.Execute(cmd)
	if err != nil {
		return &pb.UserLoginResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.UserLoginResponse{
		Success:       true,
		Token:         result.Token,
		UserId:        result.UserID,
		Email:         result.Email,
		FirstName:     result.FirstName,
		LastName:      result.LastName,
		EmailVerified: result.EmailVerified,
	}, nil
}

// UserValidateToken validates a JWT token via gRPC
func (s *UserGRPCServer) UserValidateToken(ctx context.Context, req *pb.UserValidateTokenRequest) (*pb.UserValidateTokenResponse, error) {
	result, err := s.validateTokenUseCase.Execute(req.Token)
	if err != nil {
		return &pb.UserValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.UserValidateTokenResponse{
		Valid:  result.Valid,
		UserId: result.UserID,
		Email:  result.Email,
	}, nil
}

// RegisterUser handles user registration via gRPC
func (s *UserGRPCServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	cmd := &usecase.RegisterUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	result, err := s.registerUseCase.Execute(cmd)
	if err != nil {
		return &pb.RegisterUserResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.RegisterUserResponse{
		Success:   true,
		UserId:    result.UserID,
		Email:     result.Email,
		FirstName: result.FirstName,
		LastName:  result.LastName,
	}, nil
}

// GetUserProfile retrieves user profile via gRPC
func (s *UserGRPCServer) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	result, err := s.getUserProfileUseCase.Execute(req.UserId)
	if err != nil {
		return &pb.GetUserProfileResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &pb.GetUserProfileResponse{
		Success:       true,
		UserId:        result.UserID,
		Email:         result.Email,
		FirstName:     result.FirstName,
		LastName:      result.LastName,
		IsActive:      result.IsActive,
		EmailVerified: result.EmailVerified,
	}, nil
}

// HealthCheck implements health check endpoint
func (s *UserGRPCServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Healthy: true,
		Version: "1.0.0",
	}, nil
}
