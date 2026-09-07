package auth

import (
	"errors"
	"net/http"
)

type (
	// Custom error message.
	ErrorMessage string

	// HTTPException is a custom error type
	// that holds HTTP status codes.
	HTTPException struct {
		Code    int
		Message ErrorMessage
	}
)

// Errors used by the HTTP server.
var (

	// 401 Unauthorized
	ErrUnauthorized = HTTPException{
		Code:    http.StatusUnauthorized,
		Message: "not authorized",
	}

	// 500 Internal Server Error
	ErrInternal = HTTPException{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	}

	// errInvalidSubClaim is returned by DecodeAccessToken.
	errInvalidSubClaim = errors.New(
		"could not validate credentials: sub claim missing or invalid",
	)
)

func (e HTTPException) Error() string {
	return string(e.Message)
}

// Raise method throws an http error
// with provided status code and error message.
func (e HTTPException) Raise(w http.ResponseWriter) {
	http.Error(w, string(e.Message), e.Code)
}
