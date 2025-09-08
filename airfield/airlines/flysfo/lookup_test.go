package flysfo

import (
	"context"
	"testing"

	"github.com/sfomuseum/go-sfomuseum/airfield"
)

func TestSFOMuseumLookup(t *testing.T) {

	wofid_tests := map[string]int64{
		"VOI": 1360700749,
		"Y4":  1360700749,
	}

	ctx := context.Background()

	lu, err := airfield.NewLookup(ctx, "airlines://flysfo")

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

		a := results[0].(*Airline)

		if a.WhosOnFirstId != wofid {
			t.Fatalf("Invalid match for '%s', expected %d but got %d", code, wofid, a.WhosOnFirstId)
		}
	}
}
