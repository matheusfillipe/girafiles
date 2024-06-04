package tests

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestApi(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{
		"FILE_PERSISTANCE_TIME": "1",
		"FILE_SIZE_LIMIT":       "10",
		"STORE_PATH_SIZE_LIMIT": "1024",
		"IP_MIN_RATE_LIMIT":     "4",
		"IP_HOUR_RATE_LIMIT":    "6",
		"IP_DAY_RATE_LIMIT":     "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	// File too large
	jerr := uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024*1024*11), true, nil)
	if _, ok := jerr["error"]; !ok {
		t.Fatalf("Expected size limit error to exist. Response was: %v", jerr)
	}

	// File we keep track of
	first := uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024*1024*9), false, nil)
	if _, ok := first["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", first)
	}
	firstUrl := first["url"]

	// Minute rate limit
	for i := 0; i < 3; i++ {
		// Upload random file
		file := randomJpegBytes(1024 * 1024 * 9)
		fileBytes := &bytes.Buffer{}
		j := uploadFile(t, baseUrl+"/api/", io.TeeReader(file, fileBytes), false, nil)
		if _, ok := j["url"]; !ok {
			t.Fatalf("Expected url to exist. Response was: %v", j)
		}

		// Check if server responds with the same
		resp := getFile(t, j["url"])
		respBytes := &bytes.Buffer{}
		_, err := io.Copy(respBytes, resp)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fileBytes.Bytes(), respBytes.Bytes()) {
			t.Fatalf("Expected file to be the same")
		}
	}

	// Upload file after minute rate limit
	jerr = uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024*1024*9), true, nil)
	if _, ok := jerr["error"]; !ok {
		t.Fatalf("Expected rate limit error to exist. Response was: %v", jerr)
	}

	// Change time in database to simulate 5 minutes later
	if err := dbTimeOffset(t, apiContainer, -5*60); err != nil {
		t.Fatal(err)
	}

	// Now upload should work 2 times
	for i := 0; i < 2; i++ {
		// Upload random file
		file := randomJpegBytes(1024 * 1024 * 9)
		fileBytes := &bytes.Buffer{}
		j := uploadFile(t, baseUrl+"/api/", io.TeeReader(file, fileBytes), true, nil)
		if _, ok := j["url"]; !ok {
			dumpDatabase(t, apiContainer)
			t.Fatalf("Expected url to exist. Response was: %v", j)
		}

		// Check if server responds with the same
		resp := getFile(t, j["url"])
		respBytes := &bytes.Buffer{}
		_, err := io.Copy(respBytes, resp)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fileBytes.Bytes(), respBytes.Bytes()) {
			t.Fatalf("Expected file to be the same")
		}
	}

	// Hour rate limit.
	jerr = uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024*1024*9), true, nil)
	if _, ok := jerr["error"]; !ok {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected rate limit error to exist. Response was: %v", jerr)
	}

	// Change time in database to simulate 1 hour later
	if err := dbTimeOffset(t, apiContainer, -120*60); err != nil {
		t.Fatal(err)
	}

	// Now upload should work 4 times
	for i := 0; i < 4; i++ {
		// Upload random file
		file := randomJpegBytes(1024 * 1024 * 9)
		fileBytes := &bytes.Buffer{}
		j := uploadFile(t, baseUrl+"/api/", io.TeeReader(file, fileBytes), true, nil)
		if _, ok := j["url"]; !ok {
			dumpDatabase(t, apiContainer)
			t.Fatalf("Expected url to exist. Response was: %v", j)
		}

		// Check if server responds with the same
		resp := getFile(t, j["url"])
		respBytes := &bytes.Buffer{}
		_, err := io.Copy(respBytes, resp)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fileBytes.Bytes(), respBytes.Bytes()) {
			t.Fatalf("Expected file to be the same")
		}
	}

	// The first file should be deleted now for time limit
	resp, err := http.Get(firstUrl)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		dumpContainerLogs(t, apiContainer)
		dumpDatabase(t, apiContainer)
		t.Fatalf("Expected status code %d but got %d", http.StatusNotFound, resp.StatusCode)
	}
}

// Ensure that storage limit is enforced
// We upload 2 files with 9MB each, but the limit is 10MB
// So the first file should be deleted
func TestStorageLimit(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{
		"STORE_PATH_SIZE_LIMIT": "10",
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

	// Move 1 min into the future
	if err := dbTimeOffset(t, apiContainer, -60); err != nil {
		t.Fatal(err)
	}

	// Upload file
	// Here the first should be deleted
	j = uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024*1024*9), false, nil)
	if _, ok := j["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}

	// The first file should be deleted now for storage limit
	resp, err := http.Get(firstUrl)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		dumpContainerLogs(t, apiContainer)
		dumpDatabase(t, apiContainer)
		t.Fatalf("Expected status code %d but got %d", http.StatusNotFound, resp.StatusCode)
	}
}
