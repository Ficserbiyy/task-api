package auth

import (
	"context"
	"net/http"
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

func getUserByEmail(email string) (User, error) {
	// Find the user

	return User{Email: email}, nil
}

// AuthMiddleware ensures user authentication,
// and puts user.ID into context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT from cookie.
		cookie, err := r.Cookie(sessionCookieKey)
		if err != nil {
			raiseUnauthorized(w)
			return
		}

		// Validate JWT and extract subject.
		email, err := DecodeAccessToken(cookie.Value)
		if err != nil {
			raiseUnauthorized(w)
			return
		}

		user, err := getUserByEmail(email)

		if err != nil || !user.IsActive {
			if !user.IsActive {
				raiseUnauthorized(w)
				return
			}

			http.Error(w, messageStatus500, http.StatusInternalServerError)
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
