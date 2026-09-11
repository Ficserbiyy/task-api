// Package models defines the database, response,
// and input models used by the application.
package models

import (
	"time"
)

type (
	User struct {
		ID       uint   `gorm:"primaryKey"`
		Username string `gorm:"uniqueIndex;not null"`
		Email    string `gorm:"uniqueIndex;not null"`
		Hashed   string `gorm:"not null"`
		IsActive bool   `gorm:"not null;default:true"`

		Tasks []Task `gorm:"foreignKey:OwnerID"`
	}

	Task struct {
		ID          uint   `gorm:"primaryKey"`
		OwnerID     uint   `gorm:"not null;index"`
		Title       string `gorm:"not null"`
		Description string `gorm:"not null"`
		Owner       User   `gorm:"foreignKey:OwnerID"`
		CreatedAt   time.Time
	}

	RegisterRequest struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)
