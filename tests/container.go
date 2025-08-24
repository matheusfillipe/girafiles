package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"runtime/debug"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Builds dockerfile and runs the container exposing the port 8585
func CreateApiContainer(ctx context.Context, env map[string]string) (testcontainers.Container, string, error) {
	const apiPort = 8585
	env["PORT"] = fmt.Sprint(apiPort)
	env["DEBUG"] = "1"

	test := "true"
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "..",
			Dockerfile: "Dockerfile",
			KeepImage:  true,
			BuildArgs:  map[string]*string{"TESTING": &test},
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

func randomJpegBytes(size int) io.Reader {
	var buf bytes.Buffer
	buf.WriteString("\xFF\xD8\xFF")
	for i := 0; i < size; i++ {
		buf.WriteByte(byte(rand.Intn(256)))
	}
	return io.Reader(&buf)
}

func uploadFile(t *testing.T, url string, file io.Reader, expecServerError bool, headers map[string]string) map[string]string {
	formBody := &bytes.Buffer{}
	writer := multipart.NewWriter(formBody)
	part, err := writer.CreateFormFile("file", "file.jpg")
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", url, formBody)
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK && !expecServerError {
		t.Log(string(debug.Stack()))
		t.Fatalf("Expected status code %d but got %d: %s", http.StatusOK, resp.StatusCode, body)
	}

	var parsed map[string]string
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		if expecServerError {
			return nil
		}
		t.Fatalf("Failed to parse json: %s", body)
	}

	return parsed
}

func getFile(t *testing.T, url string) io.Reader {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Expected status code %d but got %d: %s", http.StatusOK, resp.StatusCode, body)
	}
	return resp.Body
}

func dumpContainerLogs(t *testing.T, container testcontainers.Container) {
	ctx := context.Background()
	// Wait a little bit for the logs to be written
	time.Sleep(5 * time.Second)
	logs, err := container.Logs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := logs.Close(); err != nil {
			t.Logf("Failed to close logs: %v", err)
		}
	}()
	bytes, err := io.ReadAll(logs)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}
