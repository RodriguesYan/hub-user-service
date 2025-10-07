package model

import (
	"errors"
	"regexp"
	"strings"
)

// Email represents a validated email address value object
type Email struct {
	value string
}

// emailRegex validates email format according to RFC 5322 specification
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (*Email, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	return &Email{
		value: strings.ToLower(strings.TrimSpace(email)),
	}, nil
}

// NewEmailFromRepository creates an Email value object without validation
func NewEmailFromRepository(email string) *Email {
	return &Email{
		value: email,
	}
}

// validateEmail performs comprehensive email validation
func validateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	email = strings.TrimSpace(email)

	if email == "" {
		return errors.New("email cannot be empty")
	}

	if len(email) > 254 {
		return errors.New("email address is too long (maximum 254 characters)")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("email must contain exactly one @ symbol")
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) < 1 || len(localPart) > 64 {
		return errors.New("email local part must be between 1 and 64 characters")
	}

	if len(domainPart) < 1 || len(domainPart) > 253 {
		return errors.New("email domain part must be between 1 and 253 characters")
	}

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	if strings.Contains(email, "..") {
		return errors.New("email cannot contain consecutive dots")
	}

	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return errors.New("email local part cannot start or end with a dot")
	}

	return nil
}

// Value returns the email address as a string
func (e *Email) Value() string {
	return e.value
}

// Equals checks if two Email value objects are equal
func (e *Email) Equals(other *Email) bool {
	if other == nil {
		return false
	}
	return e.value == other.value
}

// Domain returns the domain part of the email address
func (e *Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// LocalPart returns the local part of the email address
func (e *Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// IsValid checks if the current email is still valid
func (e *Email) IsValid() bool {
	return validateEmail(e.value) == nil
}
