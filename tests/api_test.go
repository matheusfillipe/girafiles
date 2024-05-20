package tests

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestApi(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{
		"FILE_PERSISTANCE_TIME": "10",
		"FILE_SIZE_LIMIT":       "10",
		"STORE_PATH_SIZE_LIMIT": "50",
		"IP_MIN_RATE_LIMIT":     "3",
		"IP_HOUR_RATE_LIMIT":    "6",
		"IP_DAY_RATE_LIMIT":     "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	// File too large
	jerr := uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*11), true, nil)
	if _, ok := jerr["error"]; !ok {
		t.Fatalf("Expected size limit error to exist. Response was: %v", jerr)
	}

	// Minute rate limit
	for i := 0; i < 3; i++ {
		// Upload random file
		file := randomJpegBytes(1024 * 1024 * 9)
		fileBytes := &bytes.Buffer{}
		j := uploadFile(t, baseUrl+"/api/files/", io.TeeReader(file, fileBytes), false, nil)
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
	jerr = uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*9), true, nil)
	if _, ok := jerr["error"]; !ok {
		t.Fatalf("Expected rate limit error to exist. Response was: %v", jerr)
	}
}
