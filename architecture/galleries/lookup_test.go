package galleries

import (
	"context"
	"testing"

	"github.com/sfomuseum/go-sfomuseum/architecture"
)

func TestGalleriesLookup(t *testing.T) {

	/*

		> ./bin/lookup -lookup-uri galleries://sfomuseum 2D
		1729813699#80 2020~-2021-05-25 2D Sky Terrace Platform
		1729813701#81 2020~-2021-05-25 2D Sky Terrace Wall
		1745882461#81 2021-05-25-2021-11-09 2D Sky Terrace Wall
		1745882459#80 2021-05-25-2021-11-09 2D Sky Terrace Platform
		1763588491#80 2021-11-09-.. 2D Sky Terrace Platform
		1763588495#81 2021-11-09-.. 2D Sky Terrace Wall

	*/

	wofid_tests := map[string]int64{
		// "2D": 1745882459, // 2D Sky Terrace Platform
		"2D": 1729813701,
	}

	ctx := context.Background()

	lu, err := architecture.NewLookup(ctx, "galleries://")

	if err != nil {
		t.Fatalf("Failed to create lookup, %v", err)
	}

	for code, wofid := range wofid_tests {

		results, err := lu.Find(ctx, code)

		if err != nil {
			t.Fatalf("Unable to find '%s', %v", code, err)
		}

		a := results[0].(*Gallery)

		if a.WhosOnFirstId != wofid {
			t.Fatalf("Invalid match for '%s', expected %d but got %d", code, wofid, a.WhosOnFirstId)
		}
	}
}
