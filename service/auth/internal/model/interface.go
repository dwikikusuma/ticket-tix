package model

import "context"

type UserRepo interface {
	InsertUser(ctx context.Context, email string, password string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}

type UserService interface {
	RegisterUser(ctx context.Context, email, password string) (User, error)
	Login(ctx context.Context, email, password string) (LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	LogOut(ctx context.Context, userId int32, token string, allDevices bool) error
}
