package vmdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-manager/internal/lib/vmerr"
)

type pgxPoolDbRunner pgxpool.Pool

func (p *pgxPoolDbRunner) exec(ctx context.Context, statement Statement) (pgconn.CommandTag, error) {
	sql, params := statement.query()
	asPool := (*pgxpool.Pool)(p)
	ct, err := asPool.Exec(ctx, sql, params...)
	return ct, vmerr.InternalError(err)
}

func (p *pgxPoolDbRunner) query(ctx context.Context, statement Statement) (pgx.Rows, error) {
	sql, params := statement.query()
	asPool := (*pgxpool.Pool)(p)
	rows, err := asPool.Query(ctx, sql, params...)
	return rows, vmerr.InternalError(err)
}

func (p *pgxPoolDbRunner) Begin(ctx context.Context) (TxRunner, error) {
	asPool := (*pgxpool.Pool)(p)
	tx, err := asPool.Begin(ctx)
	if err != nil {
		return nil, vmerr.InternalError(err)
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
	ct, err := t.tx.Exec(ctx, sql, params...)
	return ct, vmerr.InternalError(err)
}

func (t pgxTxRunner) query(ctx context.Context, statement Statement) (pgx.Rows, error) {
	sql, params := statement.query()
	rows, err := t.tx.Query(ctx, sql, params...)
	return rows, vmerr.InternalError(err)
}

func (t pgxTxRunner) Commit(ctx context.Context) error {
	err := t.tx.Commit(ctx)
	return vmerr.InternalError(err)
}

func (t pgxTxRunner) Rollback(ctx context.Context) {
	err := t.tx.Rollback(ctx)
	if err != nil && err != pgx.ErrTxClosed {
		panic(fmt.Errorf("failed to rollback transaction: %w", err))
	}
}

func New(url string, options ...Option) (DbRunner, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		// Not using functions from vmerr here because this should only be called during startup.
		return nil, fmt.Errorf("could not parse connection url %q: %w", url, err)
	}
	for _, opt := range options {
		opt.apply(cfg)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	// Not using functions from vmerr here because this should only be called during startup.
	if err != nil {
		return nil, err
	}
	return (*pgxPoolDbRunner)(pool), nil
}
