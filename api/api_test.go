package api_test

import (
	"testing"

	"github.com/matheusfillipe/girafiles/api"
)

func TestIToString(t *testing.T) {
	t.Parallel()

	for i := int64(0); i < 1000000; i++ {
		s := api.IdxToString(i)
		if len(s) < api.MIN_LENGTH {
			t.Fatalf("Expected length of %d, got %d", api.MIN_LENGTH, len(s))
		}
		j, err := api.StringToIdx(s)
		if err != nil {
			t.Fatal(err)
		}
		if i != j {
			t.Fatalf("Expected %d, got %d", i, j)
		}
		if j < 0 {
			t.Fatalf("Expected positive number, got %d", j)
		}
	}
}
