package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinat1313/source_html/internal/logging"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	var err error
	defer logging.Timed("NewPool")(&err)
	if databaseURL == "" {
		err = errors.New("DATABASE_URL is required")
		return nil, err
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = pool.Ping(pctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	var err error
	defer logging.Timed("Ping")(&err)
	pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = pool.Ping(pctx)
	return err
}
