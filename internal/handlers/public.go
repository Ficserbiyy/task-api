package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ficserbiyy/task-api/internal/auth"
	"github.com/Ficserbiyy/task-api/internal/config"
	"github.com/Ficserbiyy/task-api/internal/models"
	"gorm.io/gorm"
)

// Either sets or deletes
// current_user_session Cookie.
func setSessionCookie(w http.ResponseWriter, accessToken string) {
	// Used for logout
	if accessToken == "" {
		http.SetCookie(w, &http.Cookie{
			Name:     auth.SessionCookieKey,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		return
	}

	// Used for login
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieKey,
		Value:    accessToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   config.TokenExpire * 60,
	})
}

// Register registers a new user
// @Summary 	Register a user
// @Description Create a new user
// @Tags 		public
// @Accept 		json
// @Produce 	json
// @Param 		user body models.RegisterRequest true "User"
// @Success 	201 "Successfully Created"
// @Failure 	400 {string} string "Bad Request. Possible causes:<br>• <b>Invalid JSON payload</b><br>• <b>Email registered</b>"
// @Failure 	500 {string} string "Internal Server Error. Possible causes:<br>• <b>Internal Server Error</b><br>• <b>Failed to hash password</b><br>• <b>Failed to create user</b>"
// @Router 		/auth/register [post]
func (s *TaskRepository) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		_, err := auth.GetUserByEmail(req.Email, s.DB, r.Context())
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

		user := models.User{
			Email:    req.Email,
			Hashed:   hashedPassword,
			IsActive: true,
		}

		if err := s.DB.WithContext(r.Context()).Create(&user).Error; err != nil {
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"detail": "Successfully registered",
		})
	}
}

// Login is used for the user login.
// @Summary 	Log a user in
// @Description Authenticate the user and set a Cookie
// @Tags 		public
// @Accept 		json
// @Produce 	json
// @Param 		user body models.LoginRequest true "User"
// @Success 	200 "Successfully logged in"
// @Failure		401 {string} string "Incorrect email address or password"
// @Failure 	400 {string} string "Invalid JSON payload"
// @Failure 	500 {string} string "Failed to create access token"
// @Router 		/auth/login  [post]
func (s *TaskRepository) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			config.ErrInvalidRequest.Raise(w)
			return
		}

		user, err := auth.GetUserByEmail(req.Email, s.DB, r.Context())
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

		setSessionCookie(w, accessToken)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"detail": "Successfully logged in",
		})
	}
}

// Logout is used for the user logout.
// @Summary 	Log a user out
// @Description Delete the current user session Cookie
// @Tags 		public
// @Produce 	json
// @Success 	200 "OK"
// @Router 		/auth/logout  [post]
func Logout(w http.ResponseWriter, r *http.Request) {
	setSessionCookie(w, "")

	_ = json.NewEncoder(w).Encode(map[string]string{
		"detail": "Successfully logged out",
	})
}
