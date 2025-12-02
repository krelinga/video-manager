package vmdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxPoolDbRunner pgxpool.Pool

func (p *pgxPoolDbRunner) exec(ctx context.Context, statement Statement) (pgconn.CommandTag, error) {
	sql, params := statement.query()
	asPool := (*pgxpool.Pool)(p)
	return asPool.Exec(ctx, sql, params...)
}

func (p *pgxPoolDbRunner) query(ctx context.Context, statement Statement) (pgx.Rows, error) {
	sql, params := statement.query()
	asPool := (*pgxpool.Pool)(p)
	return asPool.Query(ctx, sql, params...)
}

func (p *pgxPoolDbRunner) Begin(ctx context.Context) (TxRunner, error) {
	asPool := (*pgxpool.Pool)(p)
	tx, err := asPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return pgxTxRunner{tx: tx}, nil
}

func (p *pgxPoolDbRunner) Close() {
	asPool := (*pgxpool.Pool)(p)
	asPool.Close()
}

type pgxTxRunner struct {
	tx pgx.Tx
}

func (t pgxTxRunner) exec(ctx context.Context, statement Statement) (pgconn.CommandTag, error) {
	sql, params := statement.query()
	return t.tx.Exec(ctx, sql, params...)
}

func (t pgxTxRunner) query(ctx context.Context, statement Statement) (pgx.Rows, error) {
	sql, params := statement.query()
	return t.tx.Query(ctx, sql, params...)
}

func (t pgxTxRunner) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t pgxTxRunner) Rollback(ctx context.Context) {
	err := t.tx.Rollback(ctx)
	if err != nil && err != pgx.ErrTxClosed {
		panic(fmt.Errorf("failed to rollback transaction: %w", err))
	}
}

func New(url string) (DbRunner, error) {
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return (*pgxPoolDbRunner)(pool), nil
}