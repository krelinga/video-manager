package pgxcatalog_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-manager/go/lib/catalog"
	"github.com/krelinga/video-manager/go/lib/catalog/pgxcatalog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// runPostgres is a helper function that starts a Postgres instance for testing and returns the connection string.
// The instance will be torn down when the test ends.
func runPostgres(t *testing.T) string {
	t.Helper()
	ctx := t.Context()

	// Database configuration
	const (
		dbHost = "postgres"
		dbPort = "5432"
		dbName = "videocatalog"
		dbUser = "videocataloguser"
		dbPass = "videocatalogpass"
 	)

	// Start Postgres container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPass,
		},
		WaitingFor:     wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		postgresContainer.Terminate(ctx)
	})

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	// Build connection string
	connStr := "postgres://" + dbUser + ":" + dbPass + "@" + host + ":" + mappedPort.Port() + "/" + dbName + "?sslmode=disable"

	return connStr
}

func clearAndConnect(t *testing.T, connStr string) catalog.Client {
	t.Helper()

	migrator := &pgxcatalog.Migrator{
		ConnStr: connStr,
	}
	if err := migrator.MigrateDown(); err != nil {
		t.Fatalf("failed to migrate down: %v", err)
	}
	if err := migrator.MigrateUp(); err != nil {
		t.Fatalf("failed to migrate up: %v", err)
	}

	pool, err := pgxpool.New(t.Context(), connStr)
	if err != nil {
		t.Fatalf("failed to create pgx pool: %v", err)
	}

	return &pgxcatalog.Client{
		Pool: pool,
	}
}