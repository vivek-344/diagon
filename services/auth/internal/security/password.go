package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	saltLength   uint32 = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("security: generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("security: invalid password hash format")
	}

	params := strings.Split(parts[3], ",")

	if len(params) != 3 {
		return false, fmt.Errorf("security: invalid password hash parameters")
	}

	memory, err := parseArgonParameter(params[0], "m")
	if err != nil {
		return false, err
	}

	timeCost, err := parseArgonParameter(params[1], "t")
	if err != nil {
		return false, err
	}

	threads, err := parseArgonParameter(params[2], "p")
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("security: decode password salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("security: decode password hash: %w", err)
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(timeCost),
		uint32(memory),
		uint8(threads),
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}

func parseArgonParameter(value, name string) (uint64, error) {
	parts := strings.SplitN(value, "=", 2)

	if len(parts) != 2 || parts[0] != name {
		return 0, fmt.Errorf(
			"security: invalid argon2 parameter %q",
			value,
		)
	}

	result, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"security: invalid argon2 parameter %q: %w",
			value,
			err,
		)
	}

	return result, nil
}
