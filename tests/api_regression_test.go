package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Regression tests
func TestRegression(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{
		"FILE_PERSISTANCE_TIME": "2",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	// Upload file
	j := uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024*1024*9), false, nil)
	firstUrl := j["url"]
	if _, ok := j["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}

	// Travel 2 hours into the future
	if err := dbTimeOffset(t, apiContainer, -2*60*60); err != nil {
		t.Fatal(err)
	}

	// Upload file
	dupFile := randomJpegBytes(1024 * 1024 * 9)
	dupBuf := &bytes.Buffer{}
	if _, err = io.Copy(dupBuf, dupFile); err != nil {
		t.Fatal(err)
	}
	j = uploadFile(t, baseUrl+"/api/", bytes.NewReader(dupBuf.Bytes()), false, nil)
	if _, ok := j["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}
	dupUrl := j["url"]

	// First file should be deleted
	resp, err := http.Get(firstUrl)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected status code %d but got %d", http.StatusNotFound, resp.StatusCode)
	}

	// Travel 2 hours into the future
	if err := dbTimeOffset(t, apiContainer, -2*60*60); err != nil {
		t.Fatal(err)
	}

	// Duplicates should not be renewed
	j = uploadFile(t, baseUrl+"/api/", bytes.NewReader(dupBuf.Bytes()), false, nil)
	if _, ok := j["url"]; !ok {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}
	msg, ok := j["message"]
	if !ok {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected message to exist. Response was: %v", j)
	}
	if !strings.Contains(msg, "exists") {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected message to contain 'exists'. Response was: %v", j)
	}

	resp, err = http.Get(dupUrl)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected status code %d but got %d", http.StatusNotFound, resp.StatusCode)
	}
}
