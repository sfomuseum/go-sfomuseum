package validate

import (
	"fmt"

	"github.com/sfomuseum/go-sfomuseum/feature/properties"
	"github.com/tidwall/gjson"
)

func ValidateSFOLevel(body []byte) error {

	pt, err := properties.Placetype(body)

	if err != nil {
		return fmt.Errorf("Failed to derive sfomuseum:placetype from body, %w", err)
	}

	// This check is incomplete because it doesn't account for buildings on campus
	// or the campus itself, but it will do for now.

	if !isSFOMuseumPlacetype(pt) {
		return nil
	}

	rsp := gjson.GetBytes(body, "properties.sfo:level")

	if !rsp.Exists() {
		return fmt.Errorf("Missing sfo:level property")
	}

	return nil
}
