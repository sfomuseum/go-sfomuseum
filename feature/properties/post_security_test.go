package properties

import (
	"io"
	"strings"
	"testing"
)

var is_postsecurity = strings.NewReader(`{"properties":{"sfomuseum:gallery_id": 101, "sfomuseum:post_security":1 }}`)
var not_postsecurity = strings.NewReader(`{"properties":{"sfomuseum:object_id": 102, "sfomuseum:post_security": 0 }}`)
var unknown_postsecurity = strings.NewReader(`{"properties":{ }}`)

func TestPostSecurity(t *testing.T) {

	tests := map[int]io.Reader{
		1:  is_postsecurity,
		0:  not_postsecurity,
		-1: unknown_postsecurity,
	}

	for expected, r := range tests {

		body, err := io.ReadAll(r)

		if err != nil {
			t.Fatalf("Failed to read data (valid), %v", err)
		}

		v := PostSecurity(body)

		if v != expected {
			t.Fatalf("Invalid data for %d, got %d", expected, v)
		}
	}
}
