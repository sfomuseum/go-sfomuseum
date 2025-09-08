package sfomuseum

import (
	"context"
	"testing"
)

func TestFindCurrentAirline(t *testing.T) {

	tests := map[string]int64{
		"AC":         1159283597,
		"ACA":        1159283597,
		"AIR CANADA": 1159283597,
		"YP":         1930448549,
		"W1":         1930448553,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentAirline(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current airline for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for airline %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}

func TestFindCurrentPassengerAirlineWithRoles(t *testing.T) {

	passenger_tests := map[string]int64{
		"AA": 1159283849,
		"AF": 1159283627,
	}

	cargo_tests := map[string]int64{
		"AA": 1159283857,
		"AF": 1159283629,
	}

	ctx := context.Background()

	for code, id := range passenger_tests {

		// See this? It's a trick in advance of updating all the airline
		// records to have a `airline_role=passenger` property. Basically
		// we're trying to filter out things that have a `airline_role=cargo`
		// property.

		g, err := FindCurrentAirline(ctx, code, "")

		if err != nil {
			t.Fatalf("Failed to find current airline for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for airline %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}

	for code, id := range cargo_tests {

		g, err := FindCurrentAirline(ctx, code, "cargo")

		if err != nil {
			t.Fatalf("Failed to find current airline for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for airline %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}

}
