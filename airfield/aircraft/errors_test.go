package aircraft

import (
	_ "fmt"
	"testing"
)

func TestNotFound(t *testing.T) {

	e := NotFound{"B744"}

	if !IsNotFound(e) {
		t.Fatalf("Expected NotFound error")
	}

	if e.String() != "Aircraft 'B744' not found" {
		t.Fatalf("Invalid stringification")
	}
}

func TestMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{"B744"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected MultipleCandidates error")
	}

	if e.String() != "Multiple candidates for aircraft 'B744'" {
		t.Fatalf("Invalid stringification")
	}
}
