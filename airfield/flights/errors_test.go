package flights

import (
	_ "fmt"
	"testing"
)

func TestNotFound(t *testing.T) {

	e := NotFound{"52E"}

	if !IsNotFound(e) {
		t.Fatalf("Expected NotFound error")
	}

	if e.String() != "Flight '52E' not found" {
		t.Fatalf("Invalid stringification")
	}
}

func TestMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{"52E"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected MultipleCandidates error")
	}

	if e.String() != "Multiple candidates for flight '52E'" {
		t.Fatalf("Invalid stringification")
	}
}
