package gates

import (
	"context"
	"testing"

	"github.com/sfomuseum/go-sfomuseum/architecture"
)

func TestGatesLookup(t *testing.T) {

	wofid_tests := map[string]int64{}

	ctx := context.Background()

	lu, err := architecture.NewLookup(ctx, "gates://")

	if err != nil {
		t.Fatalf("Failed to create lookup, %v", err)
	}

	for code, wofid := range wofid_tests {

		results, err := lu.Find(ctx, code)

		if err != nil {
			t.Fatalf("Unable to find '%s', %v", code, err)
		}

		if len(results) != 1 {
			t.Fatalf("Invalid results for '%s'", code)
		}

		a := results[0].(*Gate)

		if a.WhosOnFirstId != wofid {
			t.Fatalf("Invalid match for '%s', expected %d but got %d", code, wofid, a.WhosOnFirstId)
		}
	}
}
