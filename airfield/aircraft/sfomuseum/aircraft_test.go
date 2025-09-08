package sfomuseum

import (
	"context"
	"testing"
)

func TestFindCurrentAircraft(t *testing.T) {

	// As of this writing most (all?) of the records in sfomuseum-data-aircraft
	// are still marked as mz:is_current=-1 and I still need to establish the rules
	// for which ones to mark as mz:is_current=1
	// (20211001/thisisaaronland)

	tests := map[string]int64{
		// "B744": 1159289915,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentAircraft(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current aircraft for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for aircraft %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}
