package utils

import (
	"net/mail"

	"github.com/vivek-344/diagon/sigil/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsStrongPassword(password string) error {
	if len(password) < 8 {
		return domain.ErrShortPassword
	}

	hasLetter := false
	hasSymbol := false

	for _, ch := range password {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
			hasLetter = true
		} else if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			hasSymbol = true
		}
	}

	if hasLetter && hasSymbol {
		return nil
	}
	return domain.ErrWeakPassword
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
