package vmdb

import (
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
