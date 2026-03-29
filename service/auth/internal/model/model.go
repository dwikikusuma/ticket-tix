package model

type User struct {
	ID           int32
	Email        string
	PasswordHash string
}

type LoginResponse struct {
	User         User
	RefreshToken string
	AccessToken  string
}

type LoginRequest struct {
	Email    string
	Password string
}
