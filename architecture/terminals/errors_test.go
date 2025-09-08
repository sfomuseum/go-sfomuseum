package terminals

import (
	_ "fmt"
	"testing"
)

func TestNotFound(t *testing.T) {

	e := NotFound{"T2"}

	if !IsNotFound(e) {
		t.Fatalf("Expected NotFound error")
	}

	if e.String() != "Terminal 'T2' not found" {
		t.Fatalf("Invalid stringification")
	}
}

func TestMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{"T2"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected MultipleCandidates error")
	}

	if e.String() != "Multiple candidates for terminal 'T2'" {
		t.Fatalf("Invalid stringification")
	}
}
