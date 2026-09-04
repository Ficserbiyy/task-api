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

var (
	errInvalidSubClaim = errors.New(
		"could not validate credentials: sub claim missing or invalid",
	)
)

// HashPassword securely hashes password
// using direct bcrypt library.
func HashPassword(password string) (string, error) {
	// Hash password bytes
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
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

// CreateAccessToken generates a secure JWT
// string given custom payload data.
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

// DecodeAccessToken parses, validates,
// and extracts the "sub" claim from the JWT.
func DecodeAccessToken(tokenStr string) (string, error) {
	// 1. Parse and validate the token signature and expiration
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		// Ensure the signing method matches what you expect (e.g., HMAC/HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.SecretKey), nil
	})

	// Handle invalid token or parsing error
	if err != nil || !token.Valid {
		return "", fmt.Errorf("token validation failed: %w", err)
	}

	// 2. Extract payload claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("could not parse claims")
	}

	// 3. Extract the "sub" field safely
	email, ok := claims["sub"].(string)
	if !ok || email == "" {
		return "", errInvalidSubClaim
	}

	return email, nil
}
