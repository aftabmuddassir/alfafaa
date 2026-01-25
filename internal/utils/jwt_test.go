package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test-secret-key-for-jwt-testing"

func TestGenerateToken_Success(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "author"
	expiration := 24 * time.Hour

	token, expiresAt, err := GenerateToken(userID, email, role, testSecret, expiration, AccessToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))
	assert.True(t, expiresAt.Before(time.Now().Add(25*time.Hour)))
}

func TestGenerateToken_RefreshToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "reader"
	expiration := 7 * 24 * time.Hour

	token, expiresAt, err := GenerateToken(userID, email, role, testSecret, expiration, RefreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now().Add(6*24*time.Hour)))
}

func TestValidateToken_Success(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "editor"

	token, _, err := GenerateToken(userID, email, role, testSecret, time.Hour, AccessToken)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, AccessToken, claims.TokenType)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "reader"

	// Generate token with negative expiration (already expired)
	token, _, err := GenerateToken(userID, email, role, testSecret, -time.Hour, AccessToken)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "reader"

	// Generate token with one secret
	token, _, err := GenerateToken(userID, email, role, testSecret, time.Hour, AccessToken)
	assert.NoError(t, err)

	// Validate with different secret
	claims, err := ValidateToken(token, "wrong-secret")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_MalformedToken(t *testing.T) {
	claims, err := ValidateToken("not.a.valid.jwt.token", testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	claims, err := ValidateToken("", testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateAccessToken_Success(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "admin"

	token, _, err := GenerateToken(userID, email, role, testSecret, time.Hour, AccessToken)
	assert.NoError(t, err)

	claims, err := ValidateAccessToken(token, testSecret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, AccessToken, claims.TokenType)
}

func TestValidateAccessToken_WithRefreshToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "reader"

	// Generate a refresh token but try to validate as access token
	token, _, err := GenerateToken(userID, email, role, testSecret, time.Hour, RefreshToken)
	assert.NoError(t, err)

	claims, err := ValidateAccessToken(token, testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, "invalid token type", err.Error())
}

func TestValidateRefreshToken_Success(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "author"

	token, _, err := GenerateToken(userID, email, role, testSecret, 7*24*time.Hour, RefreshToken)
	assert.NoError(t, err)

	claims, err := ValidateRefreshToken(token, testSecret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, RefreshToken, claims.TokenType)
}

func TestValidateRefreshToken_WithAccessToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "reader"

	// Generate an access token but try to validate as refresh token
	token, _, err := GenerateToken(userID, email, role, testSecret, time.Hour, AccessToken)
	assert.NoError(t, err)

	claims, err := ValidateRefreshToken(token, testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, "invalid token type", err.Error())
}

func TestGenerateTokenPair_Success(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	role := "editor"
	accessExp := 24 * time.Hour
	refreshExp := 7 * 24 * time.Hour

	pair, err := GenerateTokenPair(userID, email, role, testSecret, accessExp, refreshExp)

	assert.NoError(t, err)
	assert.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Greater(t, pair.ExpiresAt, time.Now().Unix())

	// Validate both tokens
	accessClaims, err := ValidateAccessToken(pair.AccessToken, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, AccessToken, accessClaims.TokenType)

	refreshClaims, err := ValidateRefreshToken(pair.RefreshToken, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, RefreshToken, refreshClaims.TokenType)
}

func TestJWTClaims_Issuer(t *testing.T) {
	userID := uuid.New()
	token, _, err := GenerateToken(userID, "test@example.com", "reader", testSecret, time.Hour, AccessToken)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, "alfafaa-blog", claims.Issuer)
}

func TestJWTClaims_Subject(t *testing.T) {
	userID := uuid.New()
	token, _, err := GenerateToken(userID, "test@example.com", "reader", testSecret, time.Hour, AccessToken)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
}
