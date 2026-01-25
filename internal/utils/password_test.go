package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword_Success(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	// Bcrypt hashes start with $2a$ or $2b$
	assert.Contains(t, hash, "$2a$")
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	password := "SamePassword123!"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	// Due to salt, same password produces different hashes
	assert.NotEqual(t, hash1, hash2)
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	// bcrypt can hash empty strings
	hash, err := HashPassword("")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestCheckPassword_Success(t *testing.T) {
	password := "CorrectPassword123!"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	result := CheckPassword(password, hash)

	assert.True(t, result)
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	password := "CorrectPassword123!"
	wrongPassword := "WrongPassword456!"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	result := CheckPassword(wrongPassword, hash)

	assert.False(t, result)
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	result := CheckPassword("anypassword", "not-a-valid-bcrypt-hash")

	assert.False(t, result)
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, _ := HashPassword("SomePassword123!")

	result := CheckPassword("", hash)

	assert.False(t, result)
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	result := CheckPassword("SomePassword123!", "")

	assert.False(t, result)
}

func TestValidatePasswordStrength_StrongPassword(t *testing.T) {
	password := "StrongPass123!"

	errors := ValidatePasswordStrength(password)

	assert.Empty(t, errors)
}

func TestValidatePasswordStrength_TooShort(t *testing.T) {
	password := "Short1A"

	errors := ValidatePasswordStrength(password)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Password must be at least 8 characters long")
}

func TestValidatePasswordStrength_NoUppercase(t *testing.T) {
	password := "nouppercase123"

	errors := ValidatePasswordStrength(password)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Password must contain at least one uppercase letter")
}

func TestValidatePasswordStrength_NoLowercase(t *testing.T) {
	password := "NOLOWERCASE123"

	errors := ValidatePasswordStrength(password)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Password must contain at least one lowercase letter")
}

func TestValidatePasswordStrength_NoNumber(t *testing.T) {
	password := "NoNumberHere"

	errors := ValidatePasswordStrength(password)

	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Password must contain at least one number")
}

func TestValidatePasswordStrength_MultipleErrors(t *testing.T) {
	password := "abc" // Too short, no uppercase, no number

	errors := ValidatePasswordStrength(password)

	assert.Len(t, errors, 3)
	assert.Contains(t, errors, "Password must be at least 8 characters long")
	assert.Contains(t, errors, "Password must contain at least one uppercase letter")
	assert.Contains(t, errors, "Password must contain at least one number")
}

func TestValidatePasswordStrength_OnlyNumbers(t *testing.T) {
	password := "12345678"

	errors := ValidatePasswordStrength(password)

	assert.Len(t, errors, 2)
	assert.Contains(t, errors, "Password must contain at least one uppercase letter")
	assert.Contains(t, errors, "Password must contain at least one lowercase letter")
}

func TestValidatePasswordStrength_SpecialCharactersAllowed(t *testing.T) {
	password := "Pass@123!"

	errors := ValidatePasswordStrength(password)

	assert.Empty(t, errors)
}

func TestValidatePasswordStrength_MinimumValid(t *testing.T) {
	// Exactly 8 characters with upper, lower, and number
	password := "Passwo1d"

	errors := ValidatePasswordStrength(password)

	assert.Empty(t, errors)
}

func TestValidatePasswordStrength_EmptyPassword(t *testing.T) {
	password := ""

	errors := ValidatePasswordStrength(password)

	// Should have all errors: too short, no upper, no lower, no number
	assert.Len(t, errors, 4)
}

func TestValidatePasswordStrength_UnicodeCharacters(t *testing.T) {
	// Unicode characters should work (bcrypt handles them)
	password := "Password123号"

	errors := ValidatePasswordStrength(password)

	assert.Empty(t, errors)
}
