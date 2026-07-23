package properties

import (
	"io"
	"strings"
	"testing"
)

var has_level = strings.NewReader(`{"properties":{"sfomuseum:gallery_id": 101, "sfo:level":2 }}`)
var missing_level = strings.NewReader(`{"properties":{"sfomuseum:object_id": 102, "sfomuseum:post_security": 0 }}`)

func TestLevel(t *testing.T) {

	valid := map[int]io.Reader{
		2: has_level,
	}

	missing := []io.Reader{
		missing_level,
	}

	for expected, r := range valid {

		body, err := io.ReadAll(r)

		if err != nil {
			t.Fatalf("Failed to read data (valid), %v", err)
		}

		v, err := Level(body)

		if err != nil {
			t.Fatalf("Failed to derive sfo:level, %v", err)
		}

		if v != expected {
			t.Fatalf("Invalid data for %d, got %d", expected, v)
		}
	}

	for _, r := range missing {

		body, err := io.ReadAll(r)

		if err != nil {
			t.Fatalf("Failed to read data (valid), %v", err)
		}

		_, err = Level(body)

		if err == nil {
			t.Fatalf("Expected to fail deriving sfo:level")
		}

	}
}
