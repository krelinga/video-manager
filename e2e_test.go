package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestBinaryStartsWithPostgres verifies that the video-manager binary can start successfully
// with a PostgreSQL database connection configured via environment variables.
func TestBinaryStartsWithPostgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create a network for the containers
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           "video-manager-test",
			CheckDuplicate: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	defer func() {
		if err := network.Remove(ctx); err != nil {
			t.Logf("failed to remove network: %v", err)
		}
	}()

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
		Networks:   []string{network.(*testcontainers.DockerNetwork).ID},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}()

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
	tmpDir, err := os.MkdirTemp("", "video-manager-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.CommandContext(ctx, "docker", "build", "-t", "video-manager-test:latest", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build docker image: %v\noutput: %s", err, string(output))
	}
	defer func() {
		// Clean up the docker image
		exec.Command("docker", "rmi", "video-manager-test:latest").Run()
	}()

	// Extract binary from the docker image
	cmd = exec.CommandContext(ctx, "docker", "run", "--rm", "-v", fmt.Sprintf("%s:/out", tmpDir), "video-manager-test:latest", "cp", "/app/video-manager", "/out/video-manager")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to extract binary from docker image: %v\noutput: %s", err, string(output))
	}

	binaryPath := fmt.Sprintf("%s/video-manager", tmpDir)

	// Set up environment variables for the binary
	env := os.Environ()
	env = append(env, fmt.Sprintf("VIDEO_MANAGER_POSTGRES_HOST=%s", host))
	env = append(env, fmt.Sprintf("VIDEO_MANAGER_POSTGRES_PORT=%s", port.Port()))
	env = append(env, fmt.Sprintf("VIDEO_MANAGER_POSTGRES_DBNAME=%s", postgresDB))
	env = append(env, fmt.Sprintf("VIDEO_MANAGER_POSTGRES_USER=%s", postgresUser))
	env = append(env, fmt.Sprintf("VIDEO_MANAGER_POSTGRES_PASSWORD=%s", postgresPassword))

	// Start the binary in a separate goroutine with a timeout
	binaryCtx, binaryCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer binaryCancel()

	binaryCmd := exec.CommandContext(binaryCtx, binaryPath)
	binaryCmd.Env = env

	output, err := binaryCmd.CombinedOutput()
	if err != nil {
		// Check if the error is just from context timeout (expected behavior after startup)
		if binaryCtx.Err() == context.DeadlineExceeded {
			// This is fine - the binary started and ran until we cancelled the context
			return
		}
		t.Fatalf("binary failed to start or run: %v\noutput: %s", err, string(output))
	}
}
