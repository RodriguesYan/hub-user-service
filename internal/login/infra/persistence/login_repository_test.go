package persistence

import (
	"context"
	"database/sql"
	"errors"
	"hub-user-service/internal/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a simplified mock that only implements the methods we need for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Get(dest interface{}, query string, args ...interface{}) error {
	mockArgs := m.Called(dest, query, args)

	// Check if we have a second argument with data to return
	if len(mockArgs) > 1 && mockArgs.Get(1) != nil {
		if dto, ok := mockArgs.Get(1).(userDTO); ok {
			if destDTO, ok := dest.(*userDTO); ok {
				*destDTO = dto
			}
		}
	}

	return mockArgs.Error(0)
}

// Minimal stubs for unused methods - not tested
func (m *MockDatabase) Query(query string, args ...interface{}) (database.Rows, error) {
	return nil, nil
}
func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	return nil, nil
}
func (m *MockDatabase) QueryRow(query string, args ...interface{}) database.Row { return nil }
func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) database.Row {
	return nil
}
func (m *MockDatabase) Exec(query string, args ...interface{}) (database.Result, error) {
	return nil, nil
}
func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	return nil, nil
}
func (m *MockDatabase) Begin() (database.Transaction, error) { return nil, nil }
func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Transaction, error) {
	return nil, nil
}
func (m *MockDatabase) Select(dest interface{}, query string, args ...interface{}) error { return nil }
func (m *MockDatabase) Ping() error                                                      { return nil }
func (m *MockDatabase) Close() error                                                     { return nil }

func TestNewLoginRepository(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}

	// Act
	repo := NewLoginRepository(mockDB)

	// Assert
	assert.NotNil(t, repo)
	assert.IsType(t, &LoginRepository{}, repo)
}

func TestLoginRepository_GetUserByEmail_Success(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := "test@example.com"
	expectedDTO := userDTO{
		ID:       "user123",
		Email:    email,
		Password: "hashedpassword123",
	}

	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock successful database query
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(nil, expectedDTO)

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user123", result.ID)
	assert.Equal(t, email, result.GetEmailString())
	assert.NotNil(t, result.Password)
}

func TestLoginRepository_GetUserByEmail_UserNotFound(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := "nonexistent@example.com"
	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock database returning no rows (user not found)
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(errors.New("sql: no rows in result set"))

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestLoginRepository_GetUserByEmail_DatabaseError(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := "test@example.com"
	databaseError := errors.New("connection timeout")
	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock database error
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(databaseError)

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found or database error")
	assert.Contains(t, err.Error(), "connection timeout")
}

func TestLoginRepository_GetUserByEmail_EmptyEmail(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := ""
	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock database returning no rows for empty email
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(errors.New("sql: no rows in result set"))

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestLoginRepository_GetUserByEmail_InvalidEmailFormat(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := "invalid-email-format"
	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock database returning no rows for invalid email
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(errors.New("sql: no rows in result set"))

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestLoginRepository_GetUserByEmail_DTOToModelMapping(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := "mapping@test.com"
	expectedDTO := userDTO{
		ID:       "mapping123",
		Email:    email,
		Password: "hashed_password_value",
	}

	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock successful database query
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(nil, expectedDTO)

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify mapping from DTO to domain model
	assert.Equal(t, "mapping123", result.ID)
	assert.Equal(t, email, result.GetEmailString())
	assert.NotNil(t, result.Email)
	assert.NotNil(t, result.Password)

	// Verify that the password value object was created correctly
	assert.Equal(t, "hashed_password_value", result.GetPasswordString())
}

func TestLoginRepository_GetUserByEmail_QueryStructure(t *testing.T) {
	// Arrange
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	email := "query@test.com"
	expectedDTO := userDTO{
		ID:       "query123",
		Email:    email,
		Password: "password123",
	}

	// Verify the exact query structure
	expectedQuery := "SELECT id, email, password FROM users WHERE email = $1"
	expectedArgs := []interface{}{email}

	// Mock successful database query with exact expectations
	mockDB.On("Get",
		mock.AnythingOfType("*persistence.userDTO"),
		expectedQuery,
		expectedArgs,
	).Return(nil, expectedDTO)

	repo := NewLoginRepository(mockDB)

	// Act
	result, err := repo.GetUserByEmail(email)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "query123", result.ID)
}
