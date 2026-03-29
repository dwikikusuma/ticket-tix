package repository

import (
	"context"
	"database/sql"
	authDB "ticket-tix/service/auth/internal/infra/postgres"
	"ticket-tix/service/auth/internal/model"
)

type userRepo struct {
	db *authDB.Queries
}

func NewUserRepo(db *sql.DB) model.UserRepo {
	return &userRepo{
		db: authDB.New(db),
	}
}

func (r *userRepo) InsertUser(ctx context.Context, email string, password string) (model.User, error) {
	user, err := r.db.InsertUser(ctx, authDB.InsertUserParams{
		Email:        email,
		PasswordHash: password,
	})

	if err != nil {
		return model.User{}, err
	}
	return model.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	user, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return model.User{}, err
	}

	return model.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}
