package users

import "gorm.io/gorm"

type authenticationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRepository struct {
	DB *gorm.DB
}
