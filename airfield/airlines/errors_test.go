package airlines

import (
	_ "fmt"
	"testing"
)

func TestNotFound(t *testing.T) {

	e := NotFound{"ACA"}

	if !IsNotFound(e) {
		t.Fatalf("Expected NotFound error")
	}

	if e.String() != "Airline 'ACA' not found" {
		t.Fatalf("Invalid stringification")
	}
}

func TestMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{"ACA"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected MultipleCandidates error")
	}

	if e.String() != "Multiple candidates for airline 'ACA'" {
		t.Fatalf("Invalid stringification")
	}
}
