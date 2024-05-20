package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Builds dockerfile and runs the container exposing the port 8585
func CreateApiContainer(ctx context.Context, env map[string]string) (testcontainers.Container, string, error) {
	const apiPort = 8585
	env["PORT"] = fmt.Sprint(apiPort)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "..",
			Dockerfile: "Dockerfile",
			KeepImage:  true,
		},
		Env:          env,
		ExposedPorts: []string{fmt.Sprint(apiPort) + "/tcp"},
		WaitingFor:   wait.ForListeningPort(nat.Port(fmt.Sprint(apiPort))),
	}
	apiContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	ip, err := apiContainer.Host(ctx)
	if err != nil {
		return nil, "", err
	}

	port, err := apiContainer.MappedPort(ctx, nat.Port(fmt.Sprint(apiPort)))
	if err != nil {
		return nil, "", err
	}

	uri := fmt.Sprintf("http://%s:%d", ip, port.Int())
	return apiContainer, uri, nil
}

func TestApi(t *testing.T) {
	ctx := context.Background()
	apiContainer, uri, err := CreateApiContainer(ctx, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", uri)

	resp, err := http.Get(uri)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, resp.StatusCode)
	}
}
