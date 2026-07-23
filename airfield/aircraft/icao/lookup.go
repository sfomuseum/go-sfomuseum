package icao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	_ "log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sfomuseum/go-sfomuseum/airfield/aircraft"
	"github.com/sfomuseum/go-sfomuseum/airfield/data"
)

var lookup_table *sync.Map
var lookup_idx int64

var lookup_init sync.Once
var lookup_init_err error

type ICAOLookupFunc func(context.Context)

type ICAOLookup struct {
	aircraft.AircraftLookup
}

func init() {
	ctx := context.Background()
	aircraft.RegisterAircraftLookup(ctx, "icao", NewICAOLookup)

	lookup_idx = int64(0)
}

// NewICAOLookup will return an `aircraft.AircraftLookup` instance derived from precompiled (embedded) data in `data/icao.json`
func NewICAOLookup(ctx context.Context, uri string) (aircraft.AircraftLookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	// Account for both:
	// airfield.NewLookup(ctx, "aircraft://icao/github")
	// aircraft.NewAircraftLookup(ctx, "icao://github")

	var source string

	switch u.Host {
	case "icao":
		source = u.Path
	default:
		source = u.Host
	}

	switch source {

	case "github":

		rsp, err := http.Get(DATA_GITHUB)

		if err != nil {
			return nil, fmt.Errorf("Failed to load remote data from Github, %w", err)
		}

		lookup_func := NewICAOLookupFuncWithReader(ctx, rsp.Body)
		return NewICAOLookupWithLookupFunc(ctx, lookup_func)

	default:

		fs := data.FS
		fh, err := fs.Open(DATA_JSON)

		if err != nil {
			return nil, fmt.Errorf("Failed to load local precompiled data, %w", err)
		}

		lookup_func := NewICAOLookupFuncWithReader(ctx, fh)
		return NewICAOLookupWithLookupFunc(ctx, lookup_func)
	}
}

// NewICAOLookup will return an `ICAOLookupFunc` function instance that, when invoked, will populate an `aircraft.AircraftLookup` instance with data stored in `r`.
// `r` will be closed when the `ICAOLookupFunc` function instance is invoked.
// It is assumed that the data in `r` will be formatted in the same way as the procompiled (embedded) data stored in `data/icao.json`.
func NewICAOLookupFuncWithReader(ctx context.Context, r io.ReadCloser) ICAOLookupFunc {

	lookup_func := func(ctx context.Context) {

		defer r.Close()

		var aircraft []*Aircraft

		dec := json.NewDecoder(r)
		err := dec.Decode(&aircraft)

		if err != nil {
			lookup_init_err = err
			return
		}

		table := new(sync.Map)

		for _, data := range aircraft {

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			appendData(ctx, table, data)
		}

		lookup_table = table
	}

	return lookup_func
}

// NewICAOLookupWithLookupFunc will return an `aircraft.AircraftLookup` instance derived by data compiled using `lookup_func`.
func NewICAOLookupWithLookupFunc(ctx context.Context, lookup_func ICAOLookupFunc) (aircraft.AircraftLookup, error) {

	fn := func() {
		lookup_func(ctx)
	}

	lookup_init.Do(fn)

	if lookup_init_err != nil {
		return nil, lookup_init_err
	}

	l := ICAOLookup{}
	return &l, nil
}

func (l *ICAOLookup) Find(ctx context.Context, code string) ([]any, error) {

	pointers, ok := lookup_table.Load(code)

	if !ok {
		return nil, aircraft.NotFound{code}
	}

	aircraft := make([]any, 0)

	for _, p := range pointers.([]string) {

		if !strings.HasPrefix(p, "pointer:") {
			return nil, errors.New("Invalid pointer")
		}

		row, ok := lookup_table.Load(p)

		if !ok {
			return nil, errors.New("Invalid pointer")
		}

		aircraft = append(aircraft, row.(*Aircraft))
	}

	return aircraft, nil
}

func (l *ICAOLookup) Append(ctx context.Context, data any) error {
	return appendData(ctx, lookup_table, data.(*Aircraft))
}

func appendData(ctx context.Context, table *sync.Map, data *Aircraft) error {

	idx := atomic.AddInt64(&lookup_idx, 1)

	pointer := fmt.Sprintf("pointer:%d", idx)
	table.Store(pointer, data)

	possible_codes := []string{
		data.Designator,
		data.ManufacturerCode,
	}

	for _, code := range possible_codes {

		if code == "" {
			continue
		}

		pointers := make([]string, 0)
		has_pointer := false

		others, ok := table.Load(code)

		if ok {

			pointers = others.([]string)
		}

		if slices.Contains(pointers, pointer) {
			has_pointer = true
		}

		if has_pointer {
			continue
		}

		pointers = append(pointers, pointer)
		table.Store(code, pointers)
	}

	return nil
}
