package sfomuseum

import (
	"context"
	"testing"
)

func TestFindCurrentAirport(t *testing.T) {

	tests := map[string]int64{
		"YUL":  102554351,
		"EGLL": 102556703,
		"KRIC": 102532729,
		"YBBN": 102557775,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentAirport(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current airport for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for airport %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}
