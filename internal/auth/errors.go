package auth

import (
	"errors"
	"net/http"
)

var (
	errInvalidSubClaim = errors.New(
		"could not validate credentials: sub claim missing or invalid",
	)
)

const (
	messageStatus401 = "not authorized"

	messageStatus500 = "internal server error"
)

// raiseUnauthorized raises 401 Unauthorized.
func raiseUnauthorized(w http.ResponseWriter) {
	http.Error(w, messageStatus401, http.StatusUnauthorized)
}
