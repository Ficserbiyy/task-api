// Package models defines the database, response,
// and input models used by the application.
package models

import (
	"time"
)

type (
	User struct {
		ID       uint   `gorm:"primaryKey" json:"id"`
		Username string `gorm:"uniqueIndex;not null" json:"username"`
		Email    string `gorm:"uniqueIndex;not null" json:"email"`
		Hashed   string `gorm:"not null" json:"hashed_password"`
		IsActive bool   `gorm:"not null;default:true" json:"is_active"`

		Tasks []Task `gorm:"foreignKey:OwnerID" json:"tasks"`
	}

	UserResponse struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
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

	Task struct {
		ID          uint      `gorm:"primaryKey" json:"id"`
		OwnerID     uint      `gorm:"not null;index" json:"owner_id"`
		Title       string    `gorm:"not null" json:"title"`
		Description string    `gorm:"not null" json:"description"`
		Owner       User      `gorm:"foreignKey:OwnerID" json:"owner" swaggerignore:"true"`
		CreatedAt   time.Time `json:"created_at"`
	}

	CreateTaskRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	TaskResponse struct {
		ID          uint      `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"created_at"`
	}

	Pagination struct {
		Limit      int   `json:"limit"`
		Page       int   `json:"page"`
		TotalRows  int64 `json:"total_rows"`
		TotalPages int   `json:"total_pages"`
		Rows       any   `json:"rows"`
	}
)

// Constructs response DTO from Task model.
func (task Task) ResponseModel() TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		CreatedAt:   task.CreatedAt,
	}
}

// Constructs response DTO from User model.
func (user User) ResponseModel() UserResponse {
	return UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}
}
