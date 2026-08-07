package api

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestSaveToDiskLeavesOpenReadersIntact(t *testing.T) {
	t.Setenv("STORE_PATH", t.TempDir())

	content := []byte("bytes a client is already streaming")
	dst, _, err := saveToDisk(bytes.NewReader(content), "clip.mp4", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	reader, err := os.Open(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Logf("Failed to close reader: %v", err)
		}
	}()
	held, err := reader.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// Same bytes hash to the same name, so this rewrites the path above.
	if _, _, err := saveToDisk(bytes.NewReader(content), "clip.mp4", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}

	current, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(held, current) {
		t.Fatal("Expected the upload to replace the file, a reader mid-stream sees truncation otherwise")
	}

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("Expected the open reader to still see %q, got %q", content, got)
	}
}
