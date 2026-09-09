package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ficserbiyy/task-api/internal/auth"
	"github.com/Ficserbiyy/task-api/internal/config"
	"gorm.io/gorm"
)

func Register(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authenticationRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		_, err := auth.GetUserByEmail(req.Email, db, r.Context())
		if err == nil {
			http.Error(w, "email registered", http.StatusBadRequest)
			return
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			config.ErrInternal.Raise(w)
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}

		user := config.User{
			Email:    req.Email,
			Hashed:   hashedPassword,
			IsActive: true,
		}

		if err := db.WithContext(r.Context()).Create(&user).Error; err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"detail": "Successfully registered",
		})
	}
}
