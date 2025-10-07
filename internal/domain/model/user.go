package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the domain
type User struct {
	ID                  string
	Email               *Email
	Password            *Password
	FirstName           string
	LastName            string
	IsActive            bool
	EmailVerified       bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastLoginAt         *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
}

// NewUser creates a new User with validated email and password
func NewUser(email, password, firstName, lastName string) (*User, error) {
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := NewPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:                  uuid.New().String(),
		Email:               emailVO,
		Password:            passwordVO,
		FirstName:           firstName,
		LastName:            lastName,
		IsActive:            true,
		EmailVerified:       false,
		CreatedAt:           now,
		UpdatedAt:           now,
		FailedLoginAttempts: 0,
	}, nil
}

// NewUserFromRepository creates a User from database data without validation
func NewUserFromRepository(id, email, hashedPassword, firstName, lastName string, isActive, emailVerified bool, createdAt, updatedAt time.Time, lastLoginAt, lockedUntil *time.Time, failedAttempts int) *User {
	emailVO := NewEmailFromRepository(email)
	passwordVO := NewHashedPassword(hashedPassword)

	return &User{
		ID:                  id,
		Email:               emailVO,
		Password:            passwordVO,
		FirstName:           firstName,
		LastName:            lastName,
		IsActive:            isActive,
		EmailVerified:       emailVerified,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		LastLoginAt:         lastLoginAt,
		LockedUntil:         lockedUntil,
		FailedLoginAttempts: failedAttempts,
	}
}

// GetEmailString returns the email as a string
func (u *User) GetEmailString() string {
	if u.Email == nil {
		return ""
	}
	return u.Email.Value()
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// ChangeEmail updates the user's email after validation
func (u *User) ChangeEmail(newEmail string) error {
	emailVO, err := NewEmail(newEmail)
	if err != nil {
		return err
	}
	u.Email = emailVO
	u.EmailVerified = false
	u.UpdatedAt = time.Now()
	return nil
}

// ChangePassword updates the user's password after validation
func (u *User) ChangePassword(newPassword string) error {
	passwordVO, err := NewPassword(newPassword)
	if err != nil {
		return err
	}
	u.Password = passwordVO
	u.UpdatedAt = time.Now()
	return nil
}

// VerifyPassword checks if the provided password matches the user's password
func (u *User) VerifyPassword(password string) bool {
	if u.Password == nil {
		return false
	}
	return u.Password.CompareWithHash(u.Password.Value())
}

// VerifyPasswordString checks if the provided plain text password matches
func (u *User) VerifyPasswordString(plainPassword string) bool {
	if u.Password == nil {
		return false
	}
	return u.Password.CompareWithHash(plainPassword)
}

// RecordSuccessfulLogin updates login tracking after successful authentication
func (u *User) RecordSuccessfulLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = now
}

// RecordFailedLogin increments failed login attempts and locks account if needed
func (u *User) RecordFailedLogin() {
	u.FailedLoginAttempts++
	u.UpdatedAt = time.Now()

	if u.FailedLoginAttempts >= 5 {
		lockTime := time.Now().Add(30 * time.Minute)
		u.LockedUntil = &lockTime
	}
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// Unlock manually unlocks the user account
func (u *User) Unlock() {
	u.LockedUntil = nil
	u.FailedLoginAttempts = 0
	u.UpdatedAt = time.Now()
}

// Activate activates the user account
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// VerifyEmail marks the user's email as verified
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}

// CanLogin checks if the user can login
func (u *User) CanLogin() error {
	if !u.IsActive {
		return errors.New("account is inactive")
	}

	if u.IsLocked() {
		return errors.New("account is temporarily locked due to failed login attempts")
	}

	return nil
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	if u.Password == nil {
		return errors.New("password is not set")
	}

	hashedPassword, err := u.Password.Hash()
	if err != nil {
		return err
	}

	u.Password = NewHashedPassword(hashedPassword)
	return nil
}
