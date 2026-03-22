package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knnedy/projectflow/internal/domain"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewDB(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Queries() *Queries {
	return New(db.Pool)
}

func (db *DB) WithTransaction(ctx context.Context, fn func(q *Queries) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return domain.ErrDatabase
	}

	q := New(tx)

	if err := fn(q); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
