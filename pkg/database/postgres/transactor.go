package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type txKey struct{}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func (p *Pool) Querier(ctx context.Context) Querier {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}

	return p.Pool
}

type Transactor struct {
	pool *Pool
}

func NewTransactor(pool *Pool) *Transactor {
	return &Transactor{
		pool: pool,
	}
}

func (t *Transactor) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.do(ctx, pgx.TxOptions{}, fn)
}

func (t *Transactor) DoWithOptions(ctx context.Context, options pgx.TxOptions, fn func(ctx context.Context) error) error {
	return t.do(ctx, options, fn)
}

func (t *Transactor) do(ctx context.Context, options pgx.TxOptions, fn func(ctx context.Context) error) error {
	tx, err := t.begin(ctx, options)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = rollback(ctx, tx)
			panic(r)
		}
	}()

	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rollbackErr := rollback(ctx, tx); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}

		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres: commit transaction: %w", err)
	}

	return nil
}

func (t *Transactor) begin(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	if tx, ok := TxFromContext(ctx); ok {
		nested, err := tx.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("postgres: begin nested transaction: %w", err)
		}

		return nested, nil
	}

	tx, err := t.pool.BeginTx(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("postgres: begin transaction: %w", err)
	}

	return tx, nil
}

func rollback(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Rollback(context.WithoutCancel(ctx)); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("postgres: rollback transaction: %w", err)
	}

	return nil
}
