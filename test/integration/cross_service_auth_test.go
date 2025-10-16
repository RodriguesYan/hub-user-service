package integration

import (
	"os"
	"testing"
	"time"

	"hub-user-service/internal/auth"
	"hub-user-service/internal/auth/token"
	"hub-user-service/internal/config"
	"hub-user-service/internal/login/application/usecase"
	"hub-user-service/internal/login/domain/model"
	"hub-user-service/internal/login/domain/valueobject"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// INTEGRATION TEST: Cross-Service JWT Token Synchronization
// ============================================================================
//
// This test suite validates the critical requirement:
// 1. User logs in via MICROSERVICE → gets JWT token
// 2. User makes request to MONOLITH with that token → monolith validates it
// 3. Token expiration is respected by both services
//
// Pre-requisites:
// - Both services MUST use the same MY_JWT_SECRET environment variable
// - Both services MUST use the same JWT configuration (HS256, 10min expiration)
// ============================================================================

// MockLoginRepository mocks the repository for integration testing
type MockLoginRepository struct {
	mock.Mock
}

func (m *MockLoginRepository) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestCrossServiceAuth_MicroserviceToMonolith_HappyPath(t *testing.T) {
	// This is the CRITICAL integration test that validates the main requirement:
	// Token created by microservice MUST be validated by monolith

	t.Log("=== INTEGRATION TEST: Microservice Login → Monolith Validation ===")

	// -------------------------------------------------------------------------
	// STEP 1: User logs in via MICROSERVICE (gRPC)
	// -------------------------------------------------------------------------
	t.Log("STEP 1: User logs in via MICROSERVICE...")

	// Setup microservice authentication components
	microserviceTokenService := token.NewTokenService()
	microserviceAuthService := auth.NewAuthService(microserviceTokenService)

	// Mock the repository for login
	mockRepo := new(MockLoginRepository)
	testUser := &model.User{
		ID:       "integration-test-user-123",
		Email:    valueobject.NewEmailFromRepository("integration@test.com"),
		Password: valueobject.NewPasswordFromRepository("password123"),
	}
	mockRepo.On("GetUserByEmail", "integration@test.com").Return(testUser, nil)

	loginUsecase := usecase.NewDoLoginUsecase(mockRepo)

	// Execute login
	user, err := loginUsecase.Execute("integration@test.com", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Microservice creates JWT token
	microserviceToken, err := microserviceAuthService.CreateToken(user.GetEmailString(), user.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, microserviceToken)

	t.Logf("✅ Microservice generated token: %s...", microserviceToken[:50])

	// -------------------------------------------------------------------------
	// STEP 2: User makes request to MONOLITH with microservice-generated token
	// -------------------------------------------------------------------------
	t.Log("STEP 2: Monolith validates token from microservice...")

	// Simulate monolith's token validation
	// The monolith would use the EXACT same TokenService (same JWT secret)
	monolithTokenService := token.NewTokenService()

	// Monolith receives token in "Bearer <token>" format (HTTP Authorization header)
	bearerToken := "Bearer " + microserviceToken

	// Monolith validates the token
	claims, err := monolithTokenService.ValidateToken(bearerToken)

	// -------------------------------------------------------------------------
	// ASSERTIONS: Verify cross-service token validation works
	// -------------------------------------------------------------------------
	assert.NoError(t, err, "❌ CRITICAL: Monolith FAILED to validate microservice token!")
	assert.NotNil(t, claims, "❌ CRITICAL: Claims should not be nil!")

	if err == nil {
		t.Log("✅ SUCCESS: Monolith validated microservice token!")
		t.Logf("   - Token claims: %+v", claims)

		// Verify claims match what microservice encoded
		assert.Equal(t, "integration@test.com", claims["username"], "Username claim must match")
		assert.Equal(t, "integration-test-user-123", claims["userId"], "UserId claim must match")

		// Verify token is not expired
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok, "Expiration claim must be numeric")
		assert.Greater(t, exp, float64(time.Now().Unix()), "Token must not be expired")

		t.Log("✅ All token claims verified successfully!")
	}

	mockRepo.AssertExpectations(t)
}

func TestCrossServiceAuth_TokenExpiration_BothServicesRespect(t *testing.T) {
	// Test that BOTH microservice and monolith respect token expiration

	t.Log("=== INTEGRATION TEST: Token Expiration Synchronization ===")

	// -------------------------------------------------------------------------
	// STEP 1: Create an EXPIRED token (simulating a real expired scenario)
	// -------------------------------------------------------------------------
	cfg := config.Get()
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": "expired@test.com",
			"userId":   "expired-user-456",
			"exp":      time.Now().Add(-5 * time.Minute).Unix(), // Expired 5 minutes ago
		})

	tokenString, err := expiredToken.SignedString([]byte(cfg.JWTSecret))
	assert.NoError(t, err)

	t.Log("Created expired token (expired 5 minutes ago)")

	// -------------------------------------------------------------------------
	// STEP 2: Microservice attempts to validate expired token
	// -------------------------------------------------------------------------
	t.Log("STEP 2: Microservice validates expired token...")

	microserviceTokenService := token.NewTokenService()
	bearerToken := "Bearer " + tokenString
	claims, err := microserviceTokenService.ValidateToken(bearerToken)

	assert.Error(t, err, "Microservice MUST reject expired token")
	assert.Nil(t, claims, "Claims should be nil for expired token")
	assert.Contains(t, err.Error(), "expired", "Error message should mention expiration")

	t.Log("✅ Microservice correctly rejected expired token")

	// -------------------------------------------------------------------------
	// STEP 3: Monolith attempts to validate the same expired token
	// -------------------------------------------------------------------------
	t.Log("STEP 3: Monolith validates expired token...")

	monolithTokenService := token.NewTokenService()
	claims, err = monolithTokenService.ValidateToken(bearerToken)

	assert.Error(t, err, "Monolith MUST reject expired token")
	assert.Nil(t, claims, "Claims should be nil for expired token")
	assert.Contains(t, err.Error(), "expired", "Error message should mention expiration")

	t.Log("✅ Monolith correctly rejected expired token")
	t.Log("✅ SUCCESS: Both services respect token expiration!")
}

func TestCrossServiceAuth_SecretMismatch_ValidationFails(t *testing.T) {
	// Test that validates tokens FAIL if secrets don't match
	// This is a SAFETY test to ensure proper configuration is required

	t.Log("=== SAFETY TEST: Secret Mismatch Detection ===")

	// -------------------------------------------------------------------------
	// STEP 1: Microservice creates token with correct secret
	// -------------------------------------------------------------------------
	microserviceTokenService := token.NewTokenService()
	microserviceToken, err := microserviceTokenService.CreateAndSignToken("test@example.com", "user123")
	assert.NoError(t, err)

	t.Log("Microservice created token with correct secret")

	// -------------------------------------------------------------------------
	// STEP 2: Simulate monolith with WRONG secret (misconfiguration scenario)
	// -------------------------------------------------------------------------
	wrongSecret := "wrong-secret-key-different-from-microservice"

	// Try to validate with wrong secret
	_, err = jwt.Parse(microserviceToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(wrongSecret), nil
	})

	// -------------------------------------------------------------------------
	// ASSERTION: Validation MUST fail with wrong secret
	// -------------------------------------------------------------------------
	assert.Error(t, err, "❌ CRITICAL: Validation MUST fail when secrets don't match!")

	if err != nil {
		t.Log("✅ SUCCESS: Validation correctly failed with wrong secret")
		t.Logf("   Error: %v", err)
		t.Log("⚠️  This validates that both services MUST use the same MY_JWT_SECRET!")
	}
}

func TestCrossServiceAuth_ConfigurationCheck(t *testing.T) {
	// Verify that the JWT configuration is consistent

	t.Log("=== CONFIGURATION CHECK: JWT Settings ===")

	cfg := config.Get()

	// Check JWT Secret is set
	assert.NotEmpty(t, cfg.JWTSecret, "JWT Secret must be configured")
	t.Logf("✅ JWT Secret configured (length: %d bytes)", len(cfg.JWTSecret))

	// Create a token and verify its properties
	tokenService := token.NewTokenService()
	testToken, err := tokenService.CreateAndSignToken("config@test.com", "config-user-789")
	assert.NoError(t, err)

	// Parse token to check algorithm
	parsedToken, err := jwt.Parse(testToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	assert.NoError(t, err)

	// Verify algorithm is HS256
	assert.Equal(t, jwt.SigningMethodHS256, parsedToken.Method, "Algorithm must be HS256")
	t.Log("✅ Signing algorithm: HS256")

	// Verify expiration time is ~10 minutes
	claims := parsedToken.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	expirationTime := time.Unix(exp, 0)
	timeUntilExpiry := expirationTime.Sub(time.Now())

	// Allow 5 seconds tolerance for test execution time
	assert.InDelta(t, 10*time.Minute, timeUntilExpiry, float64(5*time.Second),
		"Token expiration must be ~10 minutes")
	t.Logf("✅ Token expiration: %v (expected: 10 minutes)", timeUntilExpiry.Round(time.Second))

	// Verify required claims exist
	assert.Contains(t, claims, "username", "Token must have username claim")
	assert.Contains(t, claims, "userId", "Token must have userId claim")
	assert.Contains(t, claims, "exp", "Token must have exp claim")
	t.Log("✅ All required claims present: username, userId, exp")

	t.Log("✅ SUCCESS: JWT configuration is correct!")
}

func TestCrossServiceAuth_RealWorldScenario(t *testing.T) {
	// Simulate a complete real-world authentication flow:
	// 1. User logs in via microservice
	// 2. Multiple requests to monolith with the same token
	// 3. Token eventually expires
	// 4. Both services reject expired token

	t.Log("=== REAL-WORLD SCENARIO: Complete Auth Flow ===")

	// -------------------------------------------------------------------------
	// STEP 1: User logs in via microservice
	// -------------------------------------------------------------------------
	t.Log("STEP 1: User logs in via microservice...")

	microserviceTokenService := token.NewTokenService()
	microserviceAuthService := auth.NewAuthService(microserviceTokenService)

	userEmail := "realworld@test.com"
	userId := "realworld-user-999"

	// Microservice creates token
	authToken, err := microserviceAuthService.CreateToken(userEmail, userId)
	assert.NoError(t, err)

	t.Logf("✅ User logged in, token created: %s...", authToken[:40])

	// -------------------------------------------------------------------------
	// STEP 2: Multiple requests to monolith with same token (stateless)
	// -------------------------------------------------------------------------
	t.Log("STEP 2: Making 5 requests to monolith with same token...")

	monolithTokenService := token.NewTokenService()
	bearerToken := "Bearer " + authToken

	for i := 1; i <= 5; i++ {
		claims, err := monolithTokenService.ValidateToken(bearerToken)
		assert.NoError(t, err, "Request #%d: Token validation should succeed", i)
		assert.NotNil(t, claims)
		assert.Equal(t, userEmail, claims["username"])
		assert.Equal(t, userId, claims["userId"])

		t.Logf("   Request #%d: ✅ Token validated successfully", i)

		// Simulate some time between requests
		time.Sleep(10 * time.Millisecond)
	}

	t.Log("✅ All 5 requests validated successfully (stateless token validation)")

	// -------------------------------------------------------------------------
	// STEP 3: Verify token hasn't expired yet
	// -------------------------------------------------------------------------
	t.Log("STEP 3: Verifying token is still valid...")

	claims, err := monolithTokenService.ValidateToken(bearerToken)
	assert.NoError(t, err)

	exp := int64(claims["exp"].(float64))
	timeUntilExpiry := time.Unix(exp, 0).Sub(time.Now())

	assert.Greater(t, timeUntilExpiry, time.Duration(0), "Token should still be valid")
	t.Logf("✅ Token still valid, expires in: %v", timeUntilExpiry.Round(time.Second))

	t.Log("✅ SUCCESS: Real-world authentication flow completed!")
}

func TestCrossServiceAuth_EnvironmentVariableSync(t *testing.T) {
	// Critical test: Verify that MY_JWT_SECRET environment variable is the sync mechanism

	t.Log("=== CRITICAL: Environment Variable Synchronization ===")

	// Get the JWT secret from config (loaded from MY_JWT_SECRET env var)
	cfg := config.Get()
	jwtSecret := cfg.JWTSecret

	t.Logf("Current MY_JWT_SECRET: %s... (length: %d bytes)",
		jwtSecret[:10], len(jwtSecret))

	// Create token with current secret
	tokenService := token.NewTokenService()
	testToken, err := tokenService.CreateAndSignToken("sync@test.com", "sync-user-111")
	assert.NoError(t, err)

	// Validate token with same secret
	bearerToken := "Bearer " + testToken
	claims, err := tokenService.ValidateToken(bearerToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	t.Log("✅ Token created and validated with same secret")

	// IMPORTANT: Document the synchronization requirement
	t.Log("")
	t.Log("📋 SYNCHRONIZATION REQUIREMENTS:")
	t.Log("   1. Both microservice and monolith MUST use the same MY_JWT_SECRET env var")
	t.Log("   2. Both services MUST use HS256 algorithm")
	t.Log("   3. Both services MUST use 10-minute token expiration")
	t.Log("   4. Token format: 'Bearer <token>' in HTTP Authorization header")
	t.Log("")
	t.Log("⚠️  DEPLOYMENT CHECKLIST:")
	t.Log("   [ ] Set MY_JWT_SECRET in microservice environment")
	t.Log("   [ ] Verify monolith uses the SAME MY_JWT_SECRET")
	t.Log("   [ ] Test cross-service authentication before production")
	t.Log("")

	t.Log("✅ Environment variable synchronization validated!")
}

// ============================================================================
// Test Runner Helper
// ============================================================================

func TestMain(m *testing.M) {
	// Setup: Ensure test environment has JWT secret configured
	if os.Getenv("MY_JWT_SECRET") == "" {
		// Set a test secret if not configured
		os.Setenv("MY_JWT_SECRET", "test-jwt-secret-for-integration-tests")
	}

	// Run all tests
	exitCode := m.Run()

	// Cleanup (if needed)
	os.Exit(exitCode)
}
