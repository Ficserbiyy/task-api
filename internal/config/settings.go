package config

import "os"

var (
	SecretKey = os.Getenv("SECRET_KEY")
)

const (

	// JWT expiration time
	TokenExpire = 30
)
