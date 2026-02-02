package tests

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWebUIAuthentication(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{
		"USERS": "testuser:testpass",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	t.Run("GET / without auth returns 401", func(t *testing.T) {
		resp, err := http.Get(baseUrl + "/")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() // nolint: errcheck

		if resp.StatusCode != 401 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 401 Unauthorized, got %d: %s", resp.StatusCode, body)
		}

		wwwAuth := resp.Header.Get("WWW-Authenticate")
		if !strings.Contains(wwwAuth, "Basic") {
			t.Fatalf("Expected WWW-Authenticate header with Basic realm, got: %s", wwwAuth)
		}
	})

	t.Run("GET / with valid auth returns 200", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseUrl+"/", nil)
		if err != nil {
			t.Fatal(err)
		}
		auth := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
		req.Header.Set("Authorization", "Basic "+auth)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() // nolint: errcheck

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 200 OK, got %d: %s", resp.StatusCode, body)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "File Upload") {
			t.Fatalf("Expected upload interface to be visible, but page content missing upload UI")
		}
		if strings.Contains(bodyStr, "Only API uploads are accepted") {
			t.Fatalf("Expected upload interface to be shown, but found 'Only API uploads are accepted' message")
		}
	})

	t.Run("GET / with invalid auth returns 401", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseUrl+"/", nil)
		if err != nil {
			t.Fatal(err)
		}
		auth := base64.StdEncoding.EncodeToString([]byte("testuser:wrongpass"))
		req.Header.Set("Authorization", "Basic "+auth)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() // nolint: errcheck

		if resp.StatusCode != 401 {
			t.Fatalf("Expected 401 Unauthorized with wrong password, got %d", resp.StatusCode)
		}
	})

	t.Run("POST / without auth returns 401", func(t *testing.T) {
		resp, err := http.Post(baseUrl+"/", "application/octet-stream", strings.NewReader("test content"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() // nolint: errcheck

		if resp.StatusCode != 401 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 401 Unauthorized, got %d: %s", resp.StatusCode, body)
		}
	})

	t.Run("POST / with valid auth succeeds", func(t *testing.T) {
		auth := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
		j := uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024), false, map[string]string{"Authorization": "Basic " + auth})

		if _, ok := j["url"]; !ok {
			t.Fatalf("Expected url to exist after authenticated upload. Response was: %v", j)
		}

		fileUrl := j["url"]

		t.Run("File viewing remains public", func(t *testing.T) {
			resp, err := http.Get(fileUrl)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close() // nolint: errcheck

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("Expected file to be publicly accessible without auth, got %d: %s", resp.StatusCode, body)
			}
		})
	})

	t.Run("Static files accessible without auth", func(t *testing.T) {
		resp, err := http.Get(baseUrl + "/static/style.css")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() // nolint: errcheck

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Static files should be public, got %d: %s", resp.StatusCode, body)
		}
	})
}

func TestWebUINoAuth(t *testing.T) {
	ctx := context.Background()
	apiContainer, baseUrl, err := CreateApiContainer(ctx, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	defer apiContainer.Terminate(ctx) // nolint: errcheck
	t.Log("Running tests on", baseUrl)

	t.Run("GET / without USERS works", func(t *testing.T) {
		resp, err := http.Get(baseUrl + "/")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close() // nolint: errcheck

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected 200 OK without auth when USERS not configured, got %d: %s", resp.StatusCode, body)
		}
	})

	t.Run("POST / without USERS works", func(t *testing.T) {
		j := uploadFile(t, baseUrl+"/api/", randomJpegBytes(1024), false, nil)
		if _, ok := j["url"]; !ok {
			t.Fatalf("Expected upload to work without auth when USERS not configured. Response was: %v", j)
		}
	})
}
