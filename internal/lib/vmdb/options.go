package vmdb

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option interface {
	apply(*pgxpool.Config)
}

type optionFunc func(*pgxpool.Config)

func (f optionFunc) apply(cfg *pgxpool.Config) {
	f(cfg)
}

func WithSimpleProtocol() Option {
	return optionFunc(func(cfg *pgxpool.Config) {
		cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	})
}