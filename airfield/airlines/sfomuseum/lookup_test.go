package sfomuseum

import (
	"context"
	"testing"

	"github.com/sfomuseum/go-sfomuseum/airfield"
)

func TestSFOMuseumLookup(t *testing.T) {

	wofid_tests := map[string]int64{
		"AC":                       1159283597,
		"ACA":                      1159283597,
		"AIR CANADA":               1159283597,
		"MOV":                      1360700753,
		"NN":                       1360700753,
		"77":                       1159283643,
		"AHC":                      1159283643,
		"icao:callsign=SUNSTATE":   1159285043,
		"icao:callsign=DELTA":      1159284261,
		"icao:callsign=KLM":        1159284613,
		"wof:id=1159284389":        1159284389,
		"sfomuseum:airline_id=412": 1159284389,
		"Q174769":                  1159285413,
		"wikidata:id=Q174769":      1159285413,
		"TZP":                      1813375131,
		"ZG":                       1813375131,
	}

	passenger_tests := map[string]int64{
		"CI": 1159284153,
	}

	schemes := []string{
		"airlines://sfomuseum",
		"airlines://sfomuseum/github",
	}

	ctx := context.Background()

	for _, s := range schemes {

		lu, err := airfield.NewLookup(ctx, s)

		if err != nil {
			t.Fatalf("Failed to create lookup for '%s', %v", s, err)
		}

		for code, wofid := range wofid_tests {

			results, err := lu.Find(ctx, code)

			if err != nil {
				t.Fatalf("Unable to find '%s' using scheme '%s', %v", code, s, err)
			}

			if len(results) != 1 {
				t.Fatalf("Invalid results for '%s' using scheme '%s'", code, s)
			}

			a := results[0].(*Airline)

			if a.WhosOnFirstId != wofid {
				t.Fatalf("Invalid match for '%s', expected %d but got %d using scheme '%s'", code, wofid, a.WhosOnFirstId, s)
			}
		}

		for code, wofid := range passenger_tests {

			airline_roles := []string{
				"",
			}

			a, err := FindCurrentAirlineWithLookup(ctx, lu, code, airline_roles...)

			if err != nil {
				t.Fatalf("Failed to find current airline %s, %v", code, err)
			}

			if a.WhosOnFirstId != wofid {
				t.Fatalf("Invalid match for '%s', expected %d but got %d using scheme '%s'", code, wofid, a.WhosOnFirstId, s)
			}

		}
	}

}
