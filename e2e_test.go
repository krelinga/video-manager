package main

import (
	"context"
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestBinaryStartsWithPostgres verifies that the video-manager binary can start successfully
// with a PostgreSQL database connection configured via environment variables.
func TestBinaryStartsWithPostgres(t *testing.T) {
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		t.Log("test deadline in", timeRemaining)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeRemaining)
		defer cancel()
	} else {
		t.Log("no test deadline set")
	}

	// Docker multistage builds leave unnamed images behind by default, this cleans them up.
	// This only works because we labeled the builder stage in the Dockerfile.
	defer func() {
		cmd := exec.Command("docker", "image", "prune", "--filter", "label=stage=builder", "-f")
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to prune docker images: %v", err)
		}
	}()

	// Create a network for the containers
	net, err := network.New(ctx)
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

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
		Networks:   []string{net.Name},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Get the PostgreSQL container's address
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get postgres container port: %v", err)
	}

	// Build the binary using Docker
	videoManagerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",          // Path to the directory containing the Dockerfile
			Dockerfile: "Dockerfile", // Name of the Dockerfile
			BuildArgs: map[string]*string{
				"--no-cache": nil,
			},
			// BuildOptionsModifier: func(buildOptions *build.ImageBuildOptions) {
			// 	buildOptions.BuildArgs
			// },
		},
		ExposedPorts: []string{"25009/tcp"},
		Env: map[string]string{
			"VIDEO_MANAGER_POSTGRES_HOST":     host,
			"VIDEO_MANAGER_POSTGRES_PORT":     port.Port(),
			"VIDEO_MANAGER_POSTGRES_DBNAME":   postgresDB,
			"VIDEO_MANAGER_POSTGRES_USER":     postgresUser,
			"VIDEO_MANAGER_POSTGRES_PASSWORD": postgresPassword,
		},
		Networks:   []string{net.Name},
		WaitingFor: wait.ForHTTP("/health").WithPort("25009/tcp"),
	}

	_, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: videoManagerReq,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
}
