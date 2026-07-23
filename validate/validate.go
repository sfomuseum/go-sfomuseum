package validate

import (
	"fmt"
)

type Options struct {
	ValidatePlacetype bool
	ValidateRepo      bool
	ValidateSFOLevel  bool
}

func DefaultValidateOptions() *Options {

	return &Options{
		ValidatePlacetype: true,
		ValidateRepo:      true,
		ValidateSFOLevel:  true,
	}
}

func Validate(body []byte) error {

	opts := DefaultValidateOptions()
	return ValidateWithOptions(body, opts)
}

func ValidateWithOptions(body []byte, options *Options) error {

	if options.ValidatePlacetype {

		err := ValidatePlacetype(body)

		if err != nil {
			return fmt.Errorf("Failed to validate sfomuseum:placetype, %w", err)
		}
	}

	if options.ValidateRepo {

		err := ValidateRepo(body)

		if err != nil {
			return fmt.Errorf("Failed to validate wof:repo, %w", err)
		}
	}

	if options.ValidateSFOLevel {

		err := ValidateSFOLevel(body)

		if err != nil {
			return fmt.Errorf("Failed to validate sfo:level, %w", err)
		}
	}

	return nil
}
