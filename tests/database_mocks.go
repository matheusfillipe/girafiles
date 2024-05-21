package tests

import (
	"context"
	"fmt"
	"io"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/testcontainers/testcontainers-go"
)

func logReader(t *testing.T, reader io.Reader) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}

func dbTimeOffset(t *testing.T, dbPath string, container testcontainers.Container, offset int) error {
	_, reader, err := container.Exec(context.Background(), []string{"sqlite3", dbPath, fmt.Sprintf("UPDATE files SET timestamp = timestamp + %d", offset)})
	if err != nil {
		return err
	}
	logReader(t, reader)
	return nil
}

func dumpDatabase(t *testing.T, dbPath string, container testcontainers.Container) {
	_, reader, err := container.Exec(context.Background(), []string{"sqlite3", dbPath, "SELECT * FROM files;"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Database dump:")
	logReader(t, reader)

	_, reader, err = container.Exec(context.Background(), []string{"sqlite3", dbPath, "SELECT strftime('%s', DATETIME());"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Current timestamp:")
	logReader(t, reader)
}
