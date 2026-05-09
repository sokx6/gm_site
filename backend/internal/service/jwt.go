package service

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gm_site/internal/model"
)

// Claims represents the JWT claims for access tokens.
// It embeds jwt.RegisteredClaims to satisfy the jwt.Claims interface.
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation.
type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

// NewJWTService creates a new JWTService with the given secrets and expiration durations.
func NewJWTService(accessSecret, refreshSecret string, accessExpire, refreshExpire time.Duration) *JWTService {
	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
	}
}

// GenerateAccessToken generates an access token for the given user and role.
// Returns the token string, expiration time, and any error.
func (s *JWTService) GenerateAccessToken(userID int64, role string) (tokenStr string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(s.accessExpire)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatInt(userID, 10),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err = token.SignedString(s.accessSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiresAt, nil
}

// GenerateRefreshToken generates a refresh token for the given user.
// Returns the token string and any error.
func (s *JWTService) GenerateRefreshToken(userID int64) (tokenStr string, err error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpire)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   strconv.FormatInt(userID, 10),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecret)
}

// ValidateAccessToken parses and validates an access token string.
// Returns the claims if the token is valid, or an error otherwise.
func (s *JWTService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// ValidateRefreshToken parses and validates a refresh token string.
// Returns the user ID if the token is valid, or an error otherwise.
func (s *JWTService) ValidateRefreshToken(tokenStr string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.refreshSecret, nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// GenerateTokenPair generates both access and refresh tokens for the given user.
// Returns a TokenPair containing both tokens and the access token expiry timestamp.
func (s *JWTService) GenerateTokenPair(userID int64, role string) (*model.TokenPair, error) {
	accessToken, expiresAt, err := s.GenerateAccessToken(userID, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt.Unix(),
	}, nil
}
