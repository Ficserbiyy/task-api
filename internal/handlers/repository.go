package handlers

import "gorm.io/gorm"

type TaskRepository struct {
	DB *gorm.DB
}
