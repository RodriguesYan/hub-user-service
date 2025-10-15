package grpc

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"hub-user-service/internal/grpc/proto"
	"hub-user-service/internal/login/domain/model"
	"hub-user-service/internal/login/domain/valueobject"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Step 3.2: gRPC Integration Testing
// ============================================================================

// MockLoginUsecase mocks the login use case
type MockLoginUsecase struct {
	mock.Mock
}

func (m *MockLoginUsecase) Execute(email string, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockAuthService mocks the auth service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
	args := m.Called(tokenString, w)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) CreateToken(userName string, userId string) (string, error) {
	args := m.Called(userName, userId)
	return args.String(0), args.Error(1)
}

// Helper function to create a test user
func createTestUserForGRPC() *model.User {
	email := valueobject.NewEmailFromRepository("test@example.com")
	password := valueobject.NewPasswordFromRepository("password123")
	return &model.User{
		ID:       "user123",
		Email:    email,
		Password: password,
	}
}

// ============================================================================
// Login RPC Method Tests
// ============================================================================

func TestAuthServer_Login_Success(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUserForGRPC()

	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return("mock-jwt-token-123", nil)

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Act
	resp, err := server.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ApiResponse)
	assert.True(t, resp.ApiResponse.Success)
	assert.Equal(t, "mock-jwt-token-123", resp.Token)
	assert.NotNil(t, resp.UserInfo)
	assert.Equal(t, "user123", resp.UserInfo.UserId)
	assert.Equal(t, "test@example.com", resp.UserInfo.Email)

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestAuthServer_Login_EmptyEmail(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.LoginRequest{
		Email:    "",
		Password: "password123",
	}

	// Act
	resp, err := server.Login(ctx, req)

	// Assert
	assert.NoError(t, err) // gRPC doesn't return error, puts error in response
	assert.NotNil(t, resp)
	assert.False(t, resp.ApiResponse.Success)
	assert.Contains(t, resp.ApiResponse.Message, "email")
	assert.Equal(t, int32(http.StatusBadRequest), resp.ApiResponse.Code)

	// Ensure usecase was not called
	mockLoginUsecase.AssertNotCalled(t, "Execute")
	mockAuthService.AssertNotCalled(t, "CreateToken")
}

func TestAuthServer_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	mockLoginUsecase.On("Execute", "test@example.com", "wrongpassword").Return(nil, errors.New("invalid password"))

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	// Act
	resp, err := server.Login(ctx, req)

	// Assert
	assert.NoError(t, err) // gRPC doesn't return error, puts error in response
	assert.NotNil(t, resp)
	assert.False(t, resp.ApiResponse.Success)
	assert.Contains(t, resp.ApiResponse.Message, "login failed")
	assert.Equal(t, int32(http.StatusUnauthorized), resp.ApiResponse.Code)

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertNotCalled(t, "CreateToken")
}

func TestAuthServer_Login_TokenGenerationFailure(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUserForGRPC()

	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return("", errors.New("failed to sign token"))

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Act
	resp, err := server.Login(ctx, req)

	// Assert
	assert.NoError(t, err) // gRPC doesn't return error, puts error in response
	assert.NotNil(t, resp)
	assert.False(t, resp.ApiResponse.Success)
	assert.Contains(t, resp.ApiResponse.Message, "failed to create token")
	assert.Equal(t, int32(http.StatusInternalServerError), resp.ApiResponse.Code)

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

// ============================================================================
// ValidateToken RPC Method Tests
// ============================================================================

func TestAuthServer_ValidateToken_Success(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	mockAuthService.On("VerifyToken", "valid-token-123", mock.Anything).Return("user456", nil)

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.ValidateTokenRequest{
		Token: "valid-token-123",
	}

	// Act
	resp, err := server.ValidateToken(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ApiResponse)
	assert.True(t, resp.ApiResponse.Success)
	assert.True(t, resp.IsValid)
	assert.NotNil(t, resp.UserInfo)
	assert.Equal(t, "user456", resp.UserInfo.UserId)

	mockAuthService.AssertExpectations(t)
}

func TestAuthServer_ValidateToken_EmptyToken(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.ValidateTokenRequest{
		Token: "",
	}

	// Act
	resp, err := server.ValidateToken(ctx, req)

	// Assert
	assert.NoError(t, err) // gRPC doesn't return error, puts error in response
	assert.NotNil(t, resp)
	assert.False(t, resp.ApiResponse.Success)
	assert.Contains(t, resp.ApiResponse.Message, "token")
	assert.Equal(t, int32(http.StatusBadRequest), resp.ApiResponse.Code)

	mockAuthService.AssertNotCalled(t, "VerifyToken")
}

func TestAuthServer_ValidateToken_InvalidToken(t *testing.T) {
	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	mockAuthService.On("VerifyToken", "invalid-token", mock.Anything).Return("", errors.New("invalid token signature"))

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	req := &proto.ValidateTokenRequest{
		Token: "invalid-token",
	}

	// Act
	resp, err := server.ValidateToken(ctx, req)

	// Assert
	assert.NoError(t, err) // gRPC doesn't return error, puts error in response
	assert.NotNil(t, resp)
	assert.False(t, resp.ApiResponse.Success)
	assert.Contains(t, resp.ApiResponse.Message, "token validation failed")
	assert.Equal(t, int32(http.StatusUnauthorized), resp.ApiResponse.Code)

	mockAuthService.AssertExpectations(t)
}

// ============================================================================
// Integration Scenario Testing
// ============================================================================

func TestAuthServer_CompleteAuthenticationFlow(t *testing.T) {
	// This test simulates a complete auth flow: login then validate token

	// Arrange
	mockLoginUsecase := new(MockLoginUsecase)
	mockAuthService := new(MockAuthService)

	testUser := createTestUserForGRPC()
	generatedToken := "complete-flow-token-123"

	// Setup mocks for login
	mockLoginUsecase.On("Execute", "test@example.com", "password123").Return(testUser, nil)
	mockAuthService.On("CreateToken", "test@example.com", "user123").Return(generatedToken, nil)

	// Setup mocks for token validation
	mockAuthService.On("VerifyToken", generatedToken, mock.Anything).Return("user123", nil)

	server := NewAuthServer(mockLoginUsecase, mockAuthService)
	ctx := context.Background()

	// Act - Step 1: Login
	loginReq := &proto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	loginResp, loginErr := server.Login(ctx, loginReq)

	// Assert login
	assert.NoError(t, loginErr)
	assert.NotNil(t, loginResp)
	assert.True(t, loginResp.ApiResponse.Success)
	assert.Equal(t, generatedToken, loginResp.Token)

	// Act - Step 2: Validate the received token
	validateReq := &proto.ValidateTokenRequest{
		Token: loginResp.Token,
	}

	validateResp, validateErr := server.ValidateToken(ctx, validateReq)

	// Assert validation
	assert.NoError(t, validateErr)
	assert.NotNil(t, validateResp)
	assert.True(t, validateResp.ApiResponse.Success)
	assert.True(t, validateResp.IsValid)
	assert.Equal(t, "user123", validateResp.UserInfo.UserId)

	mockLoginUsecase.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}
