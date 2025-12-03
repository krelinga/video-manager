package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"testing"

	"buf.build/gen/go/krelinga/proto/connectrpc/go/krelinga/video_manager/catalog/v1/catalogv1connect"
	catalogv1 "buf.build/gen/go/krelinga/proto/protocolbuffers/go/krelinga/video_manager/catalog/v1"
	"connectrpc.com/connect"
	"github.com/krelinga/go-libs/deep"
	"github.com/krelinga/go-libs/exam"
	"github.com/krelinga/video-manager-api/go/vmapi"
	"github.com/krelinga/video-manager/internal/lib/vmtest"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEndToEnd verifies that the video-manager binary can start successfully
// with a PostgreSQL database connection configured via environment variables.
func TestEndToEnd(t *testing.T) {
	e := exam.New(t)
	env := deep.NewEnv()
	ctx := context.Background()

	// Docker multistage builds leave unnamed images behind by default, this cleans them up.
	// This only works because we labeled the builder stage in the Dockerfile.
	e.Cleanup(func() {
		cmd := exec.Command("docker", "image", "prune", "--filter", "label=stage=builder", "-f")
		if err := cmd.Run(); err != nil {
			e.Fatalf("failed to prune docker images: %v", err)
		}
	})

	pg := vmtest.PostgresOnce(e)

	// Build the binary using Docker
	videoManagerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",          // Path to the directory containing the Dockerfile
			Dockerfile: "Dockerfile", // Name of the Dockerfile
		},
		ExposedPorts: []string{"25009/tcp"},
		Env: map[string]string{
			"VIDEO_MANAGER_POSTGRES_HOST":     pg.Host(),
			"VIDEO_MANAGER_POSTGRES_PORT":     pg.PortString(),
			"VIDEO_MANAGER_POSTGRES_DBNAME":   pg.DBName(),
			"VIDEO_MANAGER_POSTGRES_USER":     pg.User(),
			"VIDEO_MANAGER_POSTGRES_PASSWORD": pg.Password(),
		},
		WaitingFor: wait.ForHTTP("/health").WithPort("25009/tcp"),
	}

	videoManagerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: videoManagerReq,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	vcHost, err := videoManagerContainer.Host(ctx)
	if err != nil {
		e.Fatalf("failed to get video manager container host: %v", err)
	}

	vcPort, err := videoManagerContainer.MappedPort(ctx, "25009")
	if err != nil {
		e.Fatalf("failed to get video manager container port: %v", err)
	}

	// Grab the logs when this test function ends.
	defer func() {
		// Print logs from the video manager container
		logs, err := videoManagerContainer.Logs(ctx)
		if err != nil {
			e.Fatalf("failed to get video manager container logs: %v", err)
		}
		defer logs.Close()

		// Read and log the container output
		buf := make([]byte, 1024)
		for {
			n, err := logs.Read(buf)
			if n > 0 {
				e.Logf("video manager container: %s", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	vsConString := fmt.Sprintf("http://%s:%s", vcHost, vcPort.Port())

	e.Run("catalogproto", func(e exam.E) {
		catalogClient := catalogv1connect.NewCatalogServiceClient(http.DefaultClient, vsConString)
		e.Run("movie edition kind", func(e exam.E) {
			e.Run("list empty editions", func(e exam.E) {
				listReq := &catalogv1.ListMovieEditionKindRequest{}
				listResp, err := catalogClient.ListMovieEditionKind(ctx, connect.NewRequest(listReq))
				exam.Nil(e, env, err).Log(err).Must()
				wantResp := &catalogv1.ListMovieEditionKindResponse{}
				exam.Equal(e, env, wantResp, listResp.Msg).Log(listResp.Msg)
			})
		})
	})

	e.Run("catalog", func(e exam.E) {
		urlBase := vsConString + "/api/v1/catalog"
		client, err := vmapi.NewClientWithResponses(urlBase)
		exam.Nil(e, env, err).Log(err).Must()
		params := &vmapi.ListCardsParams{}
		response, err := client.ListCardsWithResponse(ctx, params)
		exam.Nil(e, env, err).Log(err).Must()
		exam.Equal(e, env, 200, response.StatusCode()).Log(response).Must()
		e.Log("ListCards response:", deep.Format(env, response))
	})
}
