// Package services defines the application database.
package services

import (
	"fmt"

	"github.com/Ficserbiyy/task-api/internal/config"
	"github.com/Ficserbiyy/task-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The ConnectToDatabase function
// establishes a database connection.
func ConnectToDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		config.DBHost,
		config.DBUser,
		config.DBPassword,
		config.DBName,
		config.PostgresPort,
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to connect to database: %w",
			err,
		)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Task{},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to make migrations: %w",
			err,
		)
	}

	return db, nil
}
