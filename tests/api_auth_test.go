package tests

import (
	"context"
	"encoding/base64"
	"testing"
)

func TestAuthentication(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{
		"FILE_PERSISTANCE_TIME": "10",
		"FILE_SIZE_LIMIT":       "10",
		"STORE_PATH_SIZE_LIMIT": "50",
		"IP_MIN_RATE_LIMIT":     "3",
		"IP_HOUR_RATE_LIMIT":    "6",
		"IP_DAY_RATE_LIMIT":     "10",
		"USERS":                 "user1:pass1,user2:pass2",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	// Upload file without authentication
	jerr := uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*9), true, nil)
	if _, ok := jerr["error"]; !ok {
		t.Fatalf("Expected authentication error to exist. Response was: %v", jerr)
	}
	if jerr["error"] != "Unauthorized" {
		t.Fatalf("Expected error to be Unauthorized. Response was: %v", jerr)
	}

	// Upload file with wrong authentication
	headers := map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("user1:wrongpass"))}
	jerr = uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*9), true, headers)
	if _, ok := jerr["error"]; !ok {
		t.Fatalf("Expected authentication error to exist. Response was: %v", jerr)
	}
	if jerr["error"] != "Unauthorized" {
		t.Fatalf("Expected error to be Unauthorized. Response was: %v", jerr)
	}

	// Upload file with correct authentication
	headers = map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("user1:pass1"))}
	j := uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*9), false, headers)
	if _, ok := j["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}

	// Upload with second user
	headers = map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("user2:pass2"))}
	j = uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*9), false, headers)
	if _, ok := j["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}
}
