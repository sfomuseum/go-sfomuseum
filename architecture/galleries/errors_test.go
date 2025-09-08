package galleries

import (
	_ "fmt"
	"testing"
)

func TestGalleriesNotFound(t *testing.T) {

	e := NotFound{"D16"}

	if !IsNotFound(e) {
		t.Fatalf("Expected NotFound error")
	}

	if e.String() != "Gallery 'D16' not found" {
		t.Fatalf("Invalid stringification")
	}
}

func TestGalleriesMultipleCandidates(t *testing.T) {

	e := MultipleCandidates{"D16"}

	if !IsMultipleCandidates(e) {
		t.Fatalf("Expected MultipleCandidates error")
	}

	if e.String() != "Multiple candidates for gallery 'D16'" {
		t.Fatalf("Invalid stringification")
	}
}
