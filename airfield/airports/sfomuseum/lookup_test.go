package sfomuseum

import (
	"context"
	"testing"

	"github.com/sfomuseum/go-sfomuseum/airfield"
)

func TestSFOMuseumLookup(t *testing.T) {

	wofid_tests := map[string]int64{
		"YUL":                      102554351,
		"EGLL":                     102556703,
		"162":                      1360695653,
		"260":                      102525431,
		"icao:code=KSFO":           102527513,
		"icao:code=CYVR":           102555307,
		"iata:code=YVR":            102555307,
		"sfomuseum:airport_id=174": 102525799,
		"wof:id=102533269":         102533269,
		"COS":                      102526247,
		"CEB":                      102553239,
		"KRIC":                     102532729,
		"YBBN":                     102557775,
	}

	ctx := context.Background()

	schemes := []string{
		"airports://sfomuseum",
		"airports://sfomuseum/github",
	}

	for _, s := range schemes {

		lu, err := airfield.NewLookup(ctx, s)

		if err != nil {
			t.Fatalf("Failed to create lookup using scheme '%s', %v", s, err)
		}

		for code, wofid := range wofid_tests {

			results, err := lu.Find(ctx, code)

			if err != nil {
				t.Fatalf("Unable to find '%s' using scheme '%s', %v", code, s, err)
			}

			if len(results) != 1 {
				t.Fatalf("Invalid results for '%s' using scheme '%s'", s, code)
			}

			a := results[0].(*Airport)

			if a.WhosOnFirstId != wofid {
				t.Fatalf("Invalid match for '%s' using scheme '%s', expected %d but got %d", code, s, wofid, a.WhosOnFirstId)
			}
		}
	}
}
