package service

import (
	"context"
	"database/sql"
	"errors"
	"ticket-tix/common/pkg/jwt"
	"ticket-tix/service/auth/internal/infra/redis"
	"ticket-tix/service/auth/internal/model"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrLoginFailed       = errors.New("login failed")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrInvalidToken      = errors.New("invalid or revoked refresh token")
)

type userService struct {
	repo       model.UserRepo
	tokenCache redis.RefreshToken
	secretKey  string
}

func NewUserService(repo model.UserRepo, secretKey string, tokenCache redis.RefreshToken) model.UserService {
	return &userService{
		repo:       repo,
		secretKey:  secretKey,
		tokenCache: tokenCache,
	}
}

func (s *userService) RegisterUser(ctx context.Context, email, password string) (model.User, error) {
	userDB, err := s.repo.GetUserByEmail(ctx, email)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.User{}, err
	}

	if userDB.ID != 0 {
		return model.User{}, ErrUserAlreadyExists
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user, err := s.repo.InsertUser(ctx, email, string(hashPassword))
	if err != nil {
		return model.User{}, err
	}
	return model.User{ID: user.ID, Email: user.Email}, err
}

func (s *userService) Login(ctx context.Context, email, password string) (model.LoginResponse, error) {
	userDetail, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LoginResponse{}, ErrInvalidCredential
		}
		return model.LoginResponse{}, err
	}

	if userDetail.ID == 0 {
		return model.LoginResponse{}, ErrInvalidCredential
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(userDetail.PasswordHash), []byte(password)); compareErr != nil {
		return model.LoginResponse{}, ErrInvalidCredential
	}

	// Generate Tokens
	refreshToken, err := jwt.GenerateRefreshToken(userDetail.ID, s.secretKey)
	if err != nil {
		return model.LoginResponse{}, err
	}

	accessToken, err := jwt.GenerateAccessToken(userDetail.ID, s.secretKey)
	if err != nil {
		return model.LoginResponse{}, err
	}

	err = s.tokenCache.SaveRefreshToken(ctx, userDetail.ID, refreshToken)
	if err != nil {
		return model.LoginResponse{}, err
	}

	return model.LoginResponse{
		User: model.User{
			ID:    userDetail.ID,
			Email: userDetail.Email,
		},
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := jwt.ParseToken(refreshToken, s.secretKey)
	if err != nil {
		return "", err
	}

	if claims.Type != jwt.RefreshType {
		return "", ErrInvalidToken
	}

	isValid, err := s.tokenCache.ValidateRefreshToken(ctx, claims.UserID, refreshToken)
	if err != nil || !isValid {
		return "", ErrInvalidToken
	}

	return jwt.GenerateAccessToken(claims.UserID, s.secretKey)
}

func (s *userService) LogOut(ctx context.Context, userId int32, token string, allDevices bool) error {
	if allDevices {
		return s.tokenCache.RevokeUser(ctx, userId)
	}
	// single device — revoke only this token
	return s.tokenCache.RevokeRefreshToken(ctx, userId, token)
}
