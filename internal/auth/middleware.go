package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"gorm.io/gorm"
)

type (
	contextKey string

	User struct {
		ID       uint
		Email    string
		IsActive bool
	}
)

const (
	userIDContextKey contextKey = "user_id"

	sessionCookieKey = "current_user_session"
)

// getUserByEmail returns
// User if found in the database,
// otherwise gorm.ErrRecordNotFound.
func getUserByEmail(email string, db *gorm.DB, ctx context.Context) (User, error) {
	// Find the user
	var user User

	err := db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	return user, err
}

// The GetCurrentUser function either
// provides the current user's ID, or returns
// false if the user is not authenticated.
func GetCurrentUser(w http.ResponseWriter, ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uint)

	return userID, ok
}

// AuthMiddleware ensures user authentication,
// and puts the user id into context.
func AuthMiddleware(db *gorm.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT from cookie.
		cookie, err := r.Cookie(sessionCookieKey)
		if err != nil {
			ErrUnauthorized.Raise(w)
			return
		}

		// Validate JWT and extract subject.
		email, err := DecodeAccessToken(cookie.Value)
		if err != nil {
			log.Println(err)
			ErrUnauthorized.Raise(w)
			return
		}

		user, err := getUserByEmail(email, db, r.Context())

		if err != nil || !user.IsActive {
			if errors.Is(err, gorm.ErrRecordNotFound) || !user.IsActive {
				http.Error(w, "user not found", http.StatusUnauthorized)
				return
			}

			log.Println(err)
			ErrInternal.Raise(w)
			return
		}

		// Add authenticated user's ID to the request context.
		ctx := context.WithValue(
			r.Context(),
			userIDContextKey,
			user.ID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
