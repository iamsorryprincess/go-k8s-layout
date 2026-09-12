package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/iamsorryprincess/go-k8s-layout/internal/domain"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	pool *postgres.Pool
}

func NewUserRepository(pool *postgres.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) GetUser(ctx context.Context, id uint64) (domain.User, error) {
	const query = `SELECT id, name FROM users WHERE id = $1;`

	var user domain.User
	if err := r.pool.Querier(ctx).QueryRow(ctx, query, id).Scan(&user.ID, &user.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("postgres: get user %d: %w", id, domain.ErrNotFound)
		}

		return domain.User{}, fmt.Errorf("postgres: get user %d: %w", id, err)
	}

	return user, nil
}

func (r *UserRepository) GetUsers(ctx context.Context) ([]domain.User, error) {
	const query = `SELECT id, name FROM users ORDER BY id;`

	rows, err := r.pool.Querier(ctx).Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres: get users: %w", err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.User])
	if err != nil {
		return nil, fmt.Errorf("postgres: get users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (uint64, error) {
	const query = `INSERT INTO users (name) VALUES ($1) RETURNING id;`

	var id uint64
	if err := r.pool.Querier(ctx).QueryRow(ctx, query, user.Name).Scan(&id); err != nil {
		return 0, fmt.Errorf("postgres: create user: %w", err)
	}

	return id, nil
}
