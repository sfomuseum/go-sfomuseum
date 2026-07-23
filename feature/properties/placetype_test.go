package properties

import (
	"io"
	"strings"
	"testing"
)

var valid_placetype = strings.NewReader(`{"properties":{"sfomuseum:placetype": "terminal" }}`)

var missing_placetype = strings.NewReader(`{"properties":{ }}`)

func TestValidPlacetype(t *testing.T) {

	body, err := io.ReadAll(valid_placetype)

	if err != nil {
		t.Fatalf("Failed to read data (valid), %v", err)
	}

	placetype, err := Placetype(body)

	if err != nil {
		t.Fatalf("Expect data (valid) failed")
	}

	if placetype != "terminal" {
		t.Fatal("Invalid PLACETYPE (valid)")
	}

}

func TestMissingPlacetype(t *testing.T) {

	body, err := io.ReadAll(missing_placetype)

	if err != nil {
		t.Fatalf("Failed to read data (missing), %v", err)
	}

	_, err = Placetype(body)

	if err == nil {
		t.Fatalf("Expect data (missing) to fail")
	}
}
