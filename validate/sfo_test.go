package validate

import (
	"testing"
)

func TestValidateSFOLevel(t *testing.T) {

	tests_ok := []string{
		`{"properties": { "sfomuseum:placetype": "region" }}`,
		`{"properties": { "sfomuseum:placetype": "publicart", "sfo:level": 2 }}`,
		`{"properties": { "sfomuseum:placetype": "garage", "sfo:level": 4 }}`,
		`{"properties": { "sfomuseum:placetype": "exhibition", "sfo:level": 3 }}`,
	}

	tests_fail := []string{
		`{"properties": {  "sfomuseum:placetype": "publicart" }}`,
	}

	for idx, str_body := range tests_ok {

		err := ValidateSFOLevel([]byte(str_body))

		if err != nil {
			t.Fatalf("Failed to validate SFO level for test %d, %v", idx, err)
		}
	}

	for idx, str_body := range tests_fail {

		err := ValidateSFOLevel([]byte(str_body))

		if err == nil {
			t.Fatalf("Expected SFO level for test %d to fail", idx)
		}
	}

}
