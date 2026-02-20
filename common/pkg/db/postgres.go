package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Host string
	Port int
	User string
	Pass string
	DB   string

	// Optional tuning
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

func Open(cfg Config) (*sql.DB, error) {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 20
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 30
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DB)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
