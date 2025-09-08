package flysfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sfomuseum/go-sfomuseum/airfield/airlines"
	"github.com/sfomuseum/go-sfomuseum/airfield/data"
)

var lookup_table *sync.Map
var lookup_idx int64

var lookup_init sync.Once
var lookup_init_err error

type FlySFOLookupFunc func(context.Context)

type FlySFOLookup struct {
	airlines.AirlinesLookup
}

func init() {
	ctx := context.Background()
	airlines.RegisterAirlinesLookup(ctx, "flysfo", NewFlySFOLookup)

	lookup_idx = int64(0)
}

// NewFlySFOLookup will return an `airlines.AirlinesLookup` instance derived from precompiled (embedded) data in `data/flysfo.json`
func NewFlySFOLookup(ctx context.Context, uri string) (airlines.AirlinesLookup, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	// account for both
	// airfield.NewLookup(ctx, "airlines://flysfo/github")
	// airlines.NewAirLinesLookup(ctx, "flysfo://github")

	var source string

	switch u.Host {
	case "sfomuseum":
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

		lookup_func := NewFlySFOLookupFuncWithReader(ctx, rsp.Body)
		return NewFlySFOLookupWithLookupFunc(ctx, lookup_func)

	default:

		fs := data.FS
		fh, err := fs.Open(DATA_JSON)

		if err != nil {
			return nil, fmt.Errorf("Failed to load local precompiled data, %w", err)
		}

		lookup_func := NewFlySFOLookupFuncWithReader(ctx, fh)
		return NewFlySFOLookupWithLookupFunc(ctx, lookup_func)
	}
}

// NewFlySFOLookup will return an `FlysfoLookupFunc` function instance that, when invoked, will populate an `airlines.AirlinesLookup` instance with data stored in `r`.
// `r` will be closed when the `FlysfoLookupFunc` function instance is invoked.
// It is assumed that the data in `r` will be formatted in the same way as the procompiled (embedded) data stored in `data/airlines-flysfo.json`.
func NewFlySFOLookupFuncWithReader(ctx context.Context, r io.ReadCloser) FlySFOLookupFunc {

	defer r.Close()

	var airlines_list []*Airline

	dec := json.NewDecoder(r)
	err := dec.Decode(&airlines_list)

	if err != nil {

		lookup_func := func(ctx context.Context) {
			lookup_init_err = err
		}

		return lookup_func
	}

	return NewFlySFOLookupFuncWithAirlines(ctx, airlines_list)
}

// NewLookup will return an `FlySFOLookupFunc` function instance that, when invoked, will populate an `airlines.Lookup` instance with data stored in `airlines_list`.
func NewFlySFOLookupFuncWithAirlines(ctx context.Context, airlines_list []*Airline) FlySFOLookupFunc {

	lookup_func := func(ctx context.Context) {

		table := new(sync.Map)

		for _, data := range airlines_list {

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

// NewFlySFOLookupWithLookupFunc will return an `airlines.AirlinesLookup` instance derived by data compiled using `lookup_func`.
func NewFlySFOLookupWithLookupFunc(ctx context.Context, lookup_func FlySFOLookupFunc) (airlines.AirlinesLookup, error) {

	fn := func() {
		lookup_func(ctx)
	}

	lookup_init.Do(fn)

	if lookup_init_err != nil {
		return nil, lookup_init_err
	}

	l := FlySFOLookup{}
	return &l, nil
}

func (l *FlySFOLookup) Find(ctx context.Context, code string) ([]interface{}, error) {

	pointers, ok := lookup_table.Load(code)

	if !ok {
		return nil, airlines.NotFound{code}
	}

	airline := make([]interface{}, 0)

	for _, p := range pointers.([]string) {

		if !strings.HasPrefix(p, "pointer:") {
			return nil, errors.New("Invalid pointer")
		}

		row, ok := lookup_table.Load(p)

		if !ok {
			return nil, errors.New("Invalid pointer")
		}

		airline = append(airline, row.(*Airline))
	}

	return airline, nil
}

func (l *FlySFOLookup) Append(ctx context.Context, data interface{}) error {
	return appendData(ctx, lookup_table, data.(*Airline))
}

func appendData(ctx context.Context, table *sync.Map, data *Airline) error {

	idx := atomic.AddInt64(&lookup_idx, 1)

	pointer := fmt.Sprintf("pointer:%d", idx)
	table.Store(pointer, data)

	str_wofid := strconv.FormatInt(data.WhosOnFirstId, 10)

	possible_codes := []string{
		data.IATACode,
		data.ICAOCode,
		str_wofid,
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

		for _, dupe := range pointers {

			if dupe == pointer {
				has_pointer = true
				break
			}
		}

		if has_pointer {
			continue
		}

		pointers = append(pointers, pointer)
		table.Store(code, pointers)
	}

	return nil
}
