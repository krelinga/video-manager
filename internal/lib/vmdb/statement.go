package vmdb

import "github.com/jackc/pgx/v5"

type constantSql string

func (c constantSql) query() (string, []any) {
	return string(c), nil
}

type positionalSql struct {
	Sql  string
	Args []any
}

func (p positionalSql) query() (string, []any) {
	return p.Sql, p.Args
}

type namedSql struct {
	Sql  string
	Args map[string]any
}

func (n namedSql) query() (string, []any) {
	return n.Sql, []any{pgx.NamedArgs(n.Args)}
}

func Constant(sql string) Statement {
	return constantSql(sql)
}

func Positional(sql string, args ...any) Statement {
	return positionalSql{Sql: sql, Args: args}
}

func Named(sql string, args map[string]any) Statement {
	return namedSql{Sql: sql, Args: args}
}
