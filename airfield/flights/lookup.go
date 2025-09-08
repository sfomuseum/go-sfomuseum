package flights

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aaronland/go-roster"
	"github.com/sfomuseum/go-sfomuseum/airfield"
)

type FlightsLookup interface {
	airfield.Lookup
	// FindFlight(context.Context, string) ([]*Flight, error)
}

func init() {
	ctx := context.Background()
	airfield.RegisterLookup(ctx, "flights", newLookup)
}

func newLookup(ctx context.Context, uri string) (airfield.Lookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	// Rewrite flights://sfomuseum/github as sfomuseum://github

	u.Scheme = u.Host
	u.Host = ""

	path := strings.TrimLeft(u.Path, "/")
	p := strings.Split(path, "/")

	if len(p) > 0 {
		u.Host = p[0]
		u.Path = strings.Join(p[1:], "/")
	}

	return NewFlightsLookup(ctx, u.String())
}

var flights_lookup_roster roster.Roster

type FlightsLookupInitializationFunc func(ctx context.Context, uri string) (FlightsLookup, error)

func RegisterFlightsLookup(ctx context.Context, scheme string, init_func FlightsLookupInitializationFunc) error {

	err := ensureFlightsLookupRoster()

	if err != nil {
		return err
	}

	return flights_lookup_roster.Register(ctx, scheme, init_func)
}

func ensureFlightsLookupRoster() error {

	if flights_lookup_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		flights_lookup_roster = r
	}

	return nil
}

func NewFlightsLookup(ctx context.Context, uri string) (FlightsLookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := flights_lookup_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(FlightsLookupInitializationFunc)
	return init_func(ctx, uri)
}
