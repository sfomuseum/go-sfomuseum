package gates

import (
	_ "fmt"
	"testing"
)

func TestNotFound(t *testing.T) {

	e := NotFound{"A6"}

	if !IsNotFound(e) {
		t.Fatalf("Expected NotFound error")
	}

	if e.String() != "Gate 'A6' not found" {
		t.Fatalf("Invalid stringification")
	}
}

func TestMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{"A6"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected MultipleCandidates error")
	}

	if e.String() != "Multiple candidates for gate 'A6'" {
		t.Fatalf("Invalid stringification")
	}
}
