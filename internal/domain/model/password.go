package model

import (
	"errors"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Password represents a validated password value object
type Password struct {
	value string
}

// Password validation regex patterns
var (
	hasUppercase   = regexp.MustCompile(`[A-Z]`)
	hasSpecialChar = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':\"\\|,.<>\/?~` + "`" + `]`)
)

// NewPassword creates a new Password value object with validation
func NewPassword(password string) (*Password, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	return &Password{
		value: password,
	}, nil
}

// NewPasswordFromRepository creates a Password value object without validation
func NewPasswordFromRepository(password string) *Password {
	return &Password{
		value: password,
	}
}

// NewHashedPassword creates a Password with a pre-hashed value
func NewHashedPassword(hashedPassword string) *Password {
	return &Password{
		value: hashedPassword,
	}
}

// validatePassword performs comprehensive password validation
func validatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 72 {
		return errors.New("password is too long (maximum 72 characters for bcrypt)")
	}

	if !hasUppercase.MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	hasLowercase := false
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLowercase = true
			break
		}
	}
	if !hasLowercase {
		return errors.New("password must contain at least one lowercase letter")
	}

	hasDigit := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	if !hasSpecialChar.MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// Value returns the password as a string
func (p *Password) Value() string {
	return p.value
}

// Hash hashes the password using bcrypt
func (p *Password) Hash() (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(p.value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CompareWithHash compares the password with a bcrypt hash
func (p *Password) CompareWithHash(hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(p.value))
	return err == nil
}

// Equals checks if two Password value objects are equal
func (p *Password) Equals(other *Password) bool {
	if other == nil {
		return false
	}
	return p.value == other.value
}

// EqualsString checks if password equals a string value
func (p *Password) EqualsString(other string) bool {
	return p.value == other
}

// Length returns the length of the password
func (p *Password) Length() int {
	return len(p.value)
}

// Strength returns a strength score from 1-5 based on password complexity
func (p *Password) Strength() int {
	score := 0

	if len(p.value) >= 8 {
		score++
	}
	if len(p.value) >= 12 {
		score++
	}

	hasUpper := hasUppercase.MatchString(p.value)
	hasSpecial := hasSpecialChar.MatchString(p.value)

	var hasLower, hasDigit bool
	for _, char := range p.value {
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	characterTypes := 0
	if hasUpper {
		characterTypes++
	}
	if hasLower {
		characterTypes++
	}
	if hasDigit {
		characterTypes++
	}
	if hasSpecial {
		characterTypes++
	}

	if characterTypes >= 3 {
		score++
	}
	if characterTypes == 4 {
		score++
	}

	if len(p.value) >= 16 {
		score++
	}

	if score < 1 {
		score = 1
	}
	if score > 5 {
		score = 5
	}

	return score
}

// IsValid checks if the current password is still valid
func (p *Password) IsValid() bool {
	return validatePassword(p.value) == nil
}
