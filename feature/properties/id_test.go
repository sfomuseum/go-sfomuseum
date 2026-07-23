package properties

import (
	"io"
	"strings"
	"testing"
)

var aircraft_id = strings.NewReader(`{"properties":{"sfomuseum:aircraft_id": 1234, "sfomuseum:placetype":"aircraft" }}`)
var airline_id = strings.NewReader(`{"properties":{"sfomuseum:airline_id": 456, "sfomuseum:placetype":"airline" }}`)
var airport_id = strings.NewReader(`{"properties":{"sfomuseum:airport_id": 789, "sfomuseum:placetype":"airport" }}`)

var object_id = strings.NewReader(`{"properties":{"sfomuseum:object_id": 101, "sfomuseum:placetype":"object" }}`)
var publicart_id = strings.NewReader(`{"properties":{"sfomuseum:object_id": 102, "sfomuseum:placetype":"publicart" }}`)

var missing_id = strings.NewReader(`{"properties":{ }}`)

func TestValidId(t *testing.T) {

	tests := map[int64]io.Reader{
		1234: aircraft_id,
		456:  airline_id,
		789:  airport_id,
		101:  object_id,
		102:  publicart_id,
	}

	for expected, r := range tests {

		body, err := io.ReadAll(r)

		if err != nil {
			t.Fatalf("Failed to read data (valid), %v", err)
		}

		id, err := Id(body)

		if err != nil {
			t.Fatalf("Expect data (valid) failed, %v", err)
		}

		if id != expected {
			t.Fatal("Invalid ID (valid)")
		}
	}
}

func TestMissingId(t *testing.T) {

	body, err := io.ReadAll(missing_id)

	if err != nil {
		t.Fatalf("Failed to read data (missing), %v", err)
	}

	_, err = Id(body)

	if err == nil {
		t.Fatalf("Expect data (missing) to fail")
	}
}
