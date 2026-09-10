package users

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ficserbiyy/task-api/internal/auth"
	"github.com/Ficserbiyy/task-api/internal/config"
	"gorm.io/gorm"
)

func (h *UserRepository) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authenticationRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		_, err := auth.GetUserByEmail(req.Email, h.DB, r.Context())
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

		if err := h.DB.WithContext(r.Context()).Create(&user).Error; err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"detail": "Successfully registered",
		})
	}
}

func (h *UserRepository) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authenticationRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		user, err := auth.GetUserByEmail(req.Email, h.DB, r.Context())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				config.ErrIncorectPassword.Raise(w)
				return
			}
			config.ErrInternal.Raise(w)
			return
		}

		ok, err := auth.VerifyPassword(
			req.Password,
			user.Hashed,
		)
		if err != nil {
			config.ErrInternal.Raise(w)
			return
		}

		if !ok {
			config.ErrIncorectPassword.Raise(w)
			return
		}

		accessToken, err := auth.CreateAccessToken(map[string]any{
			"sub": user.Email,
		})
		if err != nil {
			http.Error(w, "failed to create access token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     auth.SessionCookieKey,
			Value:    accessToken,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   config.TokenExpire * 60,
		})

		_ = json.NewEncoder(w).Encode(map[string]string{
			"detail": "Successfully logged in",
		})
	}
}
