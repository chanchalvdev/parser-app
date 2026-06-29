package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type PostgresHealthRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresHealthRepository(pool *pgxpool.Pool) *PostgresHealthRepository {
	return &PostgresHealthRepository{pool: pool}
}

func (r *PostgresHealthRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return fmt.Errorf("postgres pool is not initialized")
	}
	return r.pool.Ping(ctx)
}
