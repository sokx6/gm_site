package service

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() *JWTService {
	return NewJWTService(
		"test-access-secret-key",
		"test-refresh-secret-key",
		15*time.Minute,
		168*time.Hour,
	)
}

func newShortLivedService() *JWTService {
	return NewJWTService(
		"test-access-secret-key",
		"test-refresh-secret-key",
		1*time.Second,
		1*time.Second,
	)
}

// TestJWTGenerateAndValidateAccessToken verifies that a valid access token
// can be generated and then validated successfully.
func TestJWTGenerateAndValidateAccessToken(t *testing.T) {
	svc := newTestService()

	token, expiresAt, err := svc.GenerateAccessToken(42, "admin")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	assert.False(t, expiresAt.IsZero())
	assert.True(t, expiresAt.After(time.Now()))

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "admin", claims.Role)
}

// TestJWTGenerateAndValidateRefreshToken verifies that a valid refresh token
// can be generated and then validated to extract the correct user ID.
func TestJWTGenerateAndValidateRefreshToken(t *testing.T) {
	svc := newTestService()

	token, err := svc.GenerateRefreshToken(99)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	userID, err := svc.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(99), userID)
}

// TestJWTExpiredAccessToken verifies that an expired access token is rejected.
func TestJWTExpiredAccessToken(t *testing.T) {
	svc := newShortLivedService()

	token, _, err := svc.GenerateAccessToken(1, "user")
	require.NoError(t, err)

	// Wait for the token to expire
	time.Sleep(1500 * time.Millisecond)

	_, err = svc.ValidateAccessToken(token)
	require.Error(t, err)
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired),
		"expected token expired error, got: %v", err)
}

// TestJWTExpiredRefreshToken verifies that an expired refresh token is rejected.
func TestJWTExpiredRefreshToken(t *testing.T) {
	svc := newShortLivedService()

	token, err := svc.GenerateRefreshToken(1)
	require.NoError(t, err)

	// Wait for the token to expire
	time.Sleep(1500 * time.Millisecond)

	_, err = svc.ValidateRefreshToken(token)
	require.Error(t, err)
	assert.True(t, errors.Is(err, jwt.ErrTokenExpired),
		"expected token expired error, got: %v", err)
}

// TestJWTInvalidToken verifies that a tampered or malformed token is rejected.
func TestJWTInvalidToken(t *testing.T) {
	svc := newTestService()

	t.Run("tampered signature", func(t *testing.T) {
		// Generate a valid token, then tamper with it
		token, _, err := svc.GenerateAccessToken(1, "user")
		require.NoError(t, err)

		// Tamper with the signature by appending garbage
		tamperedToken := token + "tampered"

		_, err = svc.ValidateAccessToken(tamperedToken)
		require.Error(t, err)
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := svc.ValidateAccessToken("")
		require.Error(t, err)
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := svc.ValidateAccessToken("not-a-valid-jwt")
		require.Error(t, err)
	})

	t.Run("refresh token with access secret", func(t *testing.T) {
		// A refresh token presented as an access token should fail
		refreshToken, err := svc.GenerateRefreshToken(1)
		require.NoError(t, err)

		_, err = svc.ValidateAccessToken(refreshToken)
		require.Error(t, err)
	})

	t.Run("access token as refresh token", func(t *testing.T) {
		accessToken, _, err := svc.GenerateAccessToken(1, "user")
		require.NoError(t, err)

		_, err = svc.ValidateRefreshToken(accessToken)
		require.Error(t, err)
	})
}

// TestJWTTokenPair verifies that GenerateTokenPair produces a valid pair of tokens.
func TestJWTTokenPair(t *testing.T) {
	svc := newTestService()

	pair, err := svc.GenerateTokenPair(7, "moderator")
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Greater(t, pair.ExpiresIn, int64(0))

	// Validate access token
	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(7), claims.UserID)
	assert.Equal(t, "moderator", claims.Role)

	// Validate refresh token
	userID, err := svc.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, int64(7), userID)

	// ExpiresIn should match the access token's expiry timestamp
	assert.InDelta(t, time.Now().Add(15*time.Minute).Unix(), pair.ExpiresIn, 2,
		"ExpiresIn should be within 2 seconds of expected")
}

// TestJWTClaimsValues verifies that the claims contain the correct UserID and Role.
func TestJWTClaimsValues(t *testing.T) {
	svc := newTestService()

	testCases := []struct {
		name   string
		userID int64
		role   string
	}{
		{"admin user", 1, "admin"},
		{"regular user", 100, "user"},
		{"moderator user", 55, "moderator"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, _, err := svc.GenerateAccessToken(tc.userID, tc.role)
			require.NoError(t, err)

			claims, err := svc.ValidateAccessToken(token)
			require.NoError(t, err)
			require.NotNil(t, claims)
			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.role, claims.Role)
		})
	}
}
