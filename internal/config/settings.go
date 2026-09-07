package config

import "os"

var (
	// Key used for JWT authentication.
	SecretKey = os.Getenv("SECRET_KEY")

	DBHost     = os.Getenv("DB_HOST")
	DBUser     = os.Getenv("DB_USER")
	DBName     = os.Getenv("DB_NAME")
	DBPassword = os.Getenv("DB_PASSWORD")
)

const (

	// JWT expiration time
	TokenExpire = 30

	// The Postgresql database port
	PostgresPort = "5432"
)
