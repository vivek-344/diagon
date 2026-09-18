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
