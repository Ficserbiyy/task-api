package config

type (
	User struct {
		ID       uint
		Email    string
		Hashed   string
		IsActive bool
	}
)
