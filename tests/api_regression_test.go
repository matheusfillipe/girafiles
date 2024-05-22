package tests

import (
	"context"
	"net/http"
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
	j := uploadFile(t, baseUrl+"/api/files/", randomJpegBytes(1024*1024*9), false, nil)
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
	j = uploadFile(t, baseUrl+"/api/files/", dupFile, false, nil)
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

	// Do not delete a duplicate upload. Regression test for duplicated files being deleted
	// right after the upload. The timestamps should be updated to prevent this.
	j = uploadFile(t, baseUrl+"/api/files/", dupFile, false, nil)
	if _, ok := j["url"]; !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
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
