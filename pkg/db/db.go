package db

import (
	"context"

	"github.com/pkg/errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) *DB {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(errors.Wrap(err, "error initiating pool"))
	}

	config.MinConns = 1
	config.MaxConns = 1

	db, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		panic(errors.Wrap(err, "error initiating pool"))
	}

	return &DB{
		pool: db,
	}
}

func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

func (d *DB) Close() {
	d.pool.Close()
}
