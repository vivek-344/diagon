package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenManager struct {
	signingKey      []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewTokenManager(
	signingKey string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) *TokenManager {
	return &TokenManager{
		signingKey:      []byte(signingKey),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

func (m *TokenManager) GenerateTokenPair(
	userID string,
) (*TokenPair, error) {
	accessToken, err := m.generateToken(
		userID,
		m.accessDuration,
		"access",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"security: generate access token: %w",
			err,
		)
	}

	refreshToken, err := m.generateToken(
		userID,
		m.refreshDuration,
		"refresh",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"security: generate refresh token: %w",
			err,
		)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *TokenManager) generateToken(
	userID string,
	duration time.Duration,
	tokenType string,
) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":  userID,
		"type": tokenType,
		"iat":  now.Unix(),
		"exp":  now.Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(m.signingKey)
}

func (m *TokenManager) parseToken(
	tokenString string,
	expectedType string,
) (string, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf(
					"security: unexpected signing method",
				)
			}

			return m.signingKey, nil
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"security: parse token: %w",
			err,
		)
	}

	if !token.Valid {
		return "", fmt.Errorf(
			"security: invalid token",
		)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf(
			"security: invalid token claims",
		)
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != expectedType {
		return "", fmt.Errorf(
			"security: invalid token type",
		)
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf(
			"security: token missing subject",
		)
	}

	return userID, nil
}

func (m *TokenManager) ValidateAccessToken(
	tokenString string,
) (string, error) {
	return m.parseToken(tokenString, "access")
}

func (m *TokenManager) ValidateRefreshToken(
	tokenString string,
) (string, error) {
	return m.parseToken(tokenString, "refresh")
}
