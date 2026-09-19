package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenManager_GenerateAndValidate(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	tokens, err := manager.GenerateTokenPair("user-123")
	if err != nil {
		t.Fatalf("generate tokens: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}

	if tokens.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}

	userID, err := manager.ValidateAccessToken(
		tokens.AccessToken,
	)
	if err != nil {
		t.Fatalf(
			"validate access token: %v",
			err,
		)
	}

	if userID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			userID,
		)
	}

	userID, err = manager.ValidateRefreshToken(
		tokens.RefreshToken,
	)
	if err != nil {
		t.Fatalf(
			"validate refresh token: %v",
			err,
		)
	}

	if userID != "user-123" {
		t.Fatalf(
			"expected user-123, got %s",
			userID,
		)
	}
}

func TestTokenManager_TokenTypeIsolation(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	tokens, err := manager.GenerateTokenPair("user-123")
	if err != nil {
		t.Fatalf("generate tokens: %v", err)
	}

	if _, err := manager.ValidateRefreshToken(
		tokens.AccessToken,
	); err == nil {
		t.Fatal("access token must not validate as refresh token")
	}

	if _, err := manager.ValidateAccessToken(
		tokens.RefreshToken,
	); err == nil {
		t.Fatal("refresh token must not validate as access token")
	}
}

func TestTokenManager_InvalidSignature(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	tokens, err := manager.GenerateTokenPair("user-123")
	if err != nil {
		t.Fatalf("generate tokens: %v", err)
	}

	otherManager := NewTokenManager(
		"abcdefghijklmnopqrstuvwxyz123456",
		15*time.Minute,
		7*24*time.Hour,
	)

	if _, err := otherManager.ValidateAccessToken(
		tokens.AccessToken,
	); err == nil {
		t.Fatal("expected invalid signature error")
	}
}

func TestTokenManager_MalformedToken(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	_, err := manager.ValidateAccessToken(
		"not-a-jwt",
	)

	if err == nil {
		t.Fatal("expected malformed token error")
	}
}

func TestTokenManager_UnexpectedSigningMethod(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		jwt.MapClaims{
			"sub":  "user-123",
			"type": "access",
			"iat":  time.Now().Unix(),
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
		},
	)

	tokenString, err := token.SignedString(
		[]byte("12345678901234567890123456789012"),
	)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := manager.ValidateAccessToken(tokenString); err == nil {
		t.Fatal("expected unexpected signing method error")
	}
}

func TestTokenManager_ExpiredToken(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  "user-123",
			"type": "access",
			"iat":  time.Now().Add(-2 * time.Hour).Unix(),
			"exp":  time.Now().Add(-time.Hour).Unix(),
		},
	)

	tokenString, err := token.SignedString(
		[]byte("12345678901234567890123456789012"),
	)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := manager.ValidateAccessToken(tokenString); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestTokenManager_MissingSubject(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"type": "access",
			"iat":  time.Now().Unix(),
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
		},
	)

	tokenString, err := token.SignedString(
		[]byte("12345678901234567890123456789012"),
	)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := manager.ValidateAccessToken(tokenString); err == nil {
		t.Fatal("expected missing subject error")
	}
}

func TestTokenManager_MissingTokenType(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": "user-123",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(15 * time.Minute).Unix(),
		},
	)

	tokenString, err := token.SignedString(
		[]byte("12345678901234567890123456789012"),
	)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := manager.ValidateAccessToken(tokenString); err == nil {
		t.Fatal("expected missing token type error")
	}
}

func TestTokenManager_TokenDurations(t *testing.T) {
	accessDuration := 15 * time.Minute
	refreshDuration := 7 * 24 * time.Hour

	manager := NewTokenManager(
		"12345678901234567890123456789012",
		accessDuration,
		refreshDuration,
	)

	before := time.Now()

	tokens, err := manager.GenerateTokenPair("user-123")
	if err != nil {
		t.Fatalf("generate tokens: %v", err)
	}

	accessToken, err := jwt.Parse(
		tokens.AccessToken,
		func(token *jwt.Token) (any, error) {
			return []byte("12345678901234567890123456789012"), nil
		},
	)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}

	accessClaims := accessToken.Claims.(jwt.MapClaims)

	accessExp, ok := accessClaims["exp"].(float64)
	if !ok {
		t.Fatal("expected access exp claim")
	}

	accessExpiry := time.Unix(int64(accessExp), 0)

	if accessExpiry.Before(
		before.Add(accessDuration - time.Second),
	) {
		t.Fatal("access token expires too early")
	}

	refreshToken, err := jwt.Parse(
		tokens.RefreshToken,
		func(token *jwt.Token) (any, error) {
			return []byte("12345678901234567890123456789012"), nil
		},
	)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}

	refreshClaims := refreshToken.Claims.(jwt.MapClaims)

	refreshExp, ok := refreshClaims["exp"].(float64)
	if !ok {
		t.Fatal("expected refresh exp claim")
	}

	refreshExpiry := time.Unix(int64(refreshExp), 0)

	if refreshExpiry.Before(
		before.Add(refreshDuration - time.Second),
	) {
		t.Fatal("refresh token expires too early")
	}
}

func TestTokenManager_EmptySubject(t *testing.T) {
	manager := NewTokenManager(
		"12345678901234567890123456789012",
		15*time.Minute,
		7*24*time.Hour,
	)

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  "",
			"type": "access",
			"iat":  time.Now().Unix(),
			"exp":  time.Now().Add(15 * time.Minute).Unix(),
		},
	)

	tokenString, err := token.SignedString(
		[]byte("12345678901234567890123456789012"),
	)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := manager.ValidateAccessToken(tokenString); err == nil {
		t.Fatal("expected empty subject to be rejected")
	}
}
