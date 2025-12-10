package vmdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option interface {
	updatePoolConfig(*pgxpool.Config)
}

type optionFunc func(*pgxpool.Config)

func (f optionFunc) updatePoolConfig(cfg *pgxpool.Config) {
	f(cfg)
}

func WithSimpleProtocol() Option {
	return optionFunc(func(cfg *pgxpool.Config) {
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	})
}

type TxOption interface {
	Option
	updateTxOptions(*pgx.TxOptions)
}

type withReadCommitted struct{}

func (w withReadCommitted) updatePoolConfig(cfg *pgxpool.Config) {
	oldAfterConnect := cfg.AfterConnect
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL READ COMMITTED;")
		if err != nil {
			return err
		}
		if oldAfterConnect != nil {
			return oldAfterConnect(ctx, conn)
		}
		return nil
	}
}

func (w withReadCommitted) updateTxOptions(txOptions *pgx.TxOptions) {
	txOptions.IsoLevel = pgx.ReadCommitted
}
func WithReadCommitted() TxOption {
	return withReadCommitted{}
}

type withRepeatableRead struct{}

func (w withRepeatableRead) updatePoolConfig(cfg *pgxpool.Config) {
	oldAfterConnect := cfg.AfterConnect
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL REPEATABLE READ;")
		if err != nil {
			return err
		}
		if oldAfterConnect != nil {
			return oldAfterConnect(ctx, conn)
		}
		return nil
	}
}

func (w withRepeatableRead) updateTxOptions(txOptions *pgx.TxOptions) {
	txOptions.IsoLevel = pgx.RepeatableRead
}

func WithRepeatableRead() TxOption {
	return withRepeatableRead{}
}

type withSerializable struct{}

func (w withSerializable) updatePoolConfig(cfg *pgxpool.Config) {
	oldAfterConnect := cfg.AfterConnect
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL SERIALIZABLE;")
		if err != nil {
			return err
		}
		if oldAfterConnect != nil {
			return oldAfterConnect(ctx, conn)
		}
		return nil
	}
}

func (w withSerializable) updateTxOptions(txOptions *pgx.TxOptions) {
	txOptions.IsoLevel = pgx.Serializable
}

func WithSerializable() TxOption {
	return withSerializable{}
}
