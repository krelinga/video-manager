package vmtest

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/migrate"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Postgres struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
	pool     *pgxpool.Pool
	poolOnce sync.Once
}

func (p *Postgres) Host() string {
	return p.host
}

func (p *Postgres) Port() int {
	return p.port
}

func (p *Postgres) PortString() string {
	return fmt.Sprintf("%d", p.port)
}

func (p *Postgres) User() string {
	return p.user
}

func (p *Postgres) Password() string {
	return p.password
}

func (p *Postgres) DBName() string {
	return p.dbName
}

func (p *Postgres) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		url.QueryEscape(p.user),
		url.QueryEscape(p.password),
		url.QueryEscape(p.host),
		p.port,
		url.QueryEscape(p.dbName),
	)
}

func (p *Postgres) Config() *config.Postgres {
	return &config.Postgres{
		Host:     p.host,
		Port:     p.port,
		User:     p.user,
		Password: p.password,
		DBName:   p.dbName,
	}
}

func (p *Postgres) Pool(e exam.E) *pgxpool.Pool {
	e.Helper()
	p.poolOnce.Do(func() {
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, p.URL())
		if err != nil {
			e.Fatalf("failed to create pgx pool: %v", err)
		}
		p.pool = pool
	})
	return p.pool
}

func (p *Postgres) Reset(e exam.E) {
	e.Helper()
	pool := p.Pool(e)
	ctx := context.Background()

	txn, err := pool.Begin(ctx)
	if err != nil {
		e.Fatalf("failed to begin transaction: %v", err)
	}
	defer txn.Rollback(ctx)
	// Get all table names
	rows, err := txn.Query(ctx, `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		e.Fatalf("failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tablename string
		if err := rows.Scan(&tablename); err != nil {
			e.Fatalf("failed to scan table name: %v", err)
		}
		tables = append(tables, tablename)
	}

	if err := rows.Err(); err != nil {
		e.Fatalf("failed to iterate over table rows: %v", err)
	}

	// Drop all tables
	for _, table := range tables {
		_, err := txn.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", pgx.Identifier{table}.Sanitize()))
		if err != nil {
			e.Fatalf("failed to drop table %s: %v", table, err)
		}
	}

	if err := txn.Commit(ctx); err != nil {
		e.Fatalf("failed to commit transaction: %v", err)
	}

	// Recreate the initial state.
	if err := migrate.Migrate(p.Config()); err != nil {
		e.Fatalf("failed to migrate database: %v", err)
	}
}

func newPostgres(e exam.E) *Postgres {
	e.Helper()
	ctx := context.Background()

	// Set up PostgreSQL test container
	postgresPassword := "testpassword"
	postgresUser := "testuser"
	postgresDB := "testdb"

	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		Hostname:     "postgres",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     postgresUser,
			"POSTGRES_PASSWORD": postgresPassword,
			"POSTGRES_DB":       postgresDB,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		e.Fatalf("failed to start postgres container: %v", err)
	}

	// Get the PostgreSQL container's address
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		e.Fatalf("failed to get postgres container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		e.Fatalf("failed to get postgres container port: %v", err)
	}

	pg := &Postgres{
		host:     host,
		port:     port.Int(),
		user:     postgresUser,
		password: postgresPassword,
		dbName:   postgresDB,
	}

	pg.Reset(e)

	return pg
}

var pgOnce sync.Once
var pgInstance *Postgres

func PostgresOnce(e exam.E) *Postgres {
	pgOnce.Do(func() {
		pgInstance = newPostgres(e)
	})
	return pgInstance
}
