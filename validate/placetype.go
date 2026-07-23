package validate

import (
	"fmt"

	"github.com/sfomuseum/go-sfomuseum/feature/properties"
	wof_validate "github.com/whosonfirst/go-whosonfirst/v4/validate"
)

// This is hopefully temporary in advance of a proper sfomuseum-placetypes spec
var placetypes []string

func init() {

	placetypes = []string{
		"boardingarea",
		"commonarea",
		"exhibition",
		"hotel",
		"garage",
		"gate",
		"publicart",
		"terminal",
		"waypoint",
	}
}

func ValidatePlacetype(body []byte) error {

	pt, err := properties.Placetype(body)

	if err != nil {
		return fmt.Errorf("Failed to derive wof:placetype from body, %w", err)
	}

	if pt == "" {
		return fmt.Errorf("Empty wof:placetype string")
	}

	if !isSFOMuseumPlacetype(pt) {

		err := wof_validate.ValidatePlacetype(body)

		if err != nil {
			return fmt.Errorf("Invalid placetype, %w", err)
		}
	}

	return nil
}

func isSFOMuseumPlacetype(pt string) bool {

	for _, this_pt := range placetypes {

		if this_pt == pt {
			return true
		}
	}

	return false
}
