package db

import (
	// Context manages cancellation and timeouts for database operations
	// Basically, it allows you to control how long a database operation should run before it is automatically canceled.
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, config)
}