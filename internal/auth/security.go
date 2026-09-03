package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword securely hashes password
// using direct bcrypt library.
func HashPassword(password string) (string, error) {
	// Password bytes
	pwdBytes := []byte(password)

	hashed, err := bcrypt.GenerateFromPassword(
		pwdBytes,
		bcrypt.DefaultCost,
	)

	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", errors.New("password exceeds bcrypt's 72-byte limit")
		}

		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashed), nil
}

// VerifyPassword verifies password.
func VerifyPassword(plain, hashed string) (bool, error) {

	err := bcrypt.CompareHashAndPassword(
		[]byte(hashed),
		[]byte(plain),
	)

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify password: %w", err)
	}

	return true, nil
}
