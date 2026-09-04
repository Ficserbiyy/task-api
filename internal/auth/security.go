package auth

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/Ficserbiyy/task-api/internal/config"
	"github.com/golang-jwt/jwt/v5"
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

// CreateAccessToken generates a secure JWT string given custom payload data.
func CreateAccessToken(data map[string]any) (string, error) {
	// Create a new map to avoid mutating the input map
	claims := jwt.MapClaims{}
	maps.Copy(claims, data)

	// Add the standard "exp" claim as a Unix timestamp
	claims["exp"] = time.Now().Add(config.TokenExpire * time.Minute).Unix()

	// Token with the signing method (HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and return the encoded JWT token string
	return token.SignedString([]byte(config.SecretKey))
}
