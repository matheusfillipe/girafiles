package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// Media players seek by asking for byte ranges, so file delivery must answer
// with 206 and the requested slice.
func TestRangeRequests(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	content := &bytes.Buffer{}
	j := uploadFile(t, baseUrl+"/api/", io.TeeReader(randomJpegBytes(1024*100), content), false, nil)
	url, ok := j["url"]
	if !ok {
		t.Fatalf("Expected url to exist. Response was: %v", j)
	}

	for _, method := range []string{"GET", "HEAD"} {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if resp.Header.Get("Accept-Ranges") != "bytes" {
			dumpContainerLogs(t, apiContainer)
			t.Fatalf("Expected %s to advertise Accept-Ranges: bytes, got %q", method, resp.Header.Get("Accept-Ranges"))
		}
	}

	const start, end = 1000, 1999
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusPartialContent {
		dumpContainerLogs(t, apiContainer)
		t.Fatalf("Expected status code %d but got %d", http.StatusPartialContent, resp.StatusCode)
	}

	expectedRange := fmt.Sprintf("bytes %d-%d/%d", start, end, content.Len())
	if got := resp.Header.Get("Content-Range"); got != expectedRange {
		t.Fatalf("Expected Content-Range %q but got %q", expectedRange, got)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, content.Bytes()[start:end+1]) {
		t.Fatalf("Expected the requested slice, got %d bytes that do not match", len(body))
	}
}
