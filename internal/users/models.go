package users

type authenticationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
