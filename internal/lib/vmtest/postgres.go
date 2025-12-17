package vmtest

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager/internal/lib/config"
	"github.com/krelinga/video-manager/internal/lib/migrate"
	"github.com/krelinga/video-manager/internal/lib/vmdb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Postgres struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
	db       vmdb.DbRunner
	once     sync.Once
	network  *testcontainers.DockerNetwork
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

func (p *Postgres) Network() *testcontainers.DockerNetwork {
	return p.network
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

func (p *Postgres) DbRunner(e exam.E) vmdb.DbRunner {
	e.Helper()
	p.once.Do(func() {
		var err error
		p.db, err = vmdb.New(p.URL(), vmdb.WithSimpleProtocol())
		if err != nil {
			e.Fatalf("failed to create vmdb DbRunner: %v", err)
		}
	})
	return p.db
}

func (p *Postgres) Reset(e exam.E) {
	e.Helper()
	// Roll back all migrations.
	if err := migrate.Down(p.Config()); err != nil {
		e.Fatalf("failed to reset database: %v", err)
	}

	// Recreate the initial state.
	if err := migrate.Up(p.Config()); err != nil {
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

	network, err := network.New(ctx)
	if err != nil {
		e.Fatalf("failed to create docker network: %v", err)
	}

	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		Hostname:     "postgres",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     postgresUser,
			"POSTGRES_PASSWORD": postgresPassword,
			"POSTGRES_DB":       postgresDB,
		},
		Networks: 	[]string{network.Name},
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
		network:  network,
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
