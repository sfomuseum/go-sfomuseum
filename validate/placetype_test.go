package validate

import (
	"testing"
)

func TestValidatePlacetype(t *testing.T) {

	tests_ok := []string{
		`{"properties": { "sfomuseum:placetype": "region", "wof:placetype": "region" }}`,
		`{"properties": { "sfomuseum:placetype": "publicart" }}`,
		`{"properties": { "sfomuseum:placetype": "garage" }}`,
		`{"properties": { "sfomuseum:placetype": "exhibition" }}`,
	}

	tests_fail := []string{
		`{"properties": {  "sfomuseum:placetype": "" }}`,
		`{"properties": { "sfomuseum:placetype": "foodtruck", "wof:placetype": "other" }}`,
	}

	for idx, str_body := range tests_ok {

		err := ValidatePlacetype([]byte(str_body))

		if err != nil {
			t.Fatalf("Failed to validate placetype for test %d, %v", idx, err)
		}
	}

	for idx, str_body := range tests_fail {

		err := ValidatePlacetype([]byte(str_body))

		if err == nil {
			t.Fatalf("Expected placetype for test %d to fail", idx)
		}
	}

}
