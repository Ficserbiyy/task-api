package auth

import "errors"

var (
	errInvalidSubClaim = errors.New(
		"could not validate credentials: sub claim missing or invalid",
	)
)
