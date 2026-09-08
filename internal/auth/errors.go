package auth

import (
	"errors"
)

// Errors used by the HTTP server.
var (
	// errInvalidSubClaim is returned by DecodeAccessToken.
	errInvalidSubClaim = errors.New(
		"could not validate credentials: sub claim missing or invalid",
	)
)
