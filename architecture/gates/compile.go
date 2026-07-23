package gates

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
	"github.com/whosonfirst/go-whosonfirst/v4/iterate"
	"github.com/whosonfirst/go-whosonfirst/v4/uri"
)

// CompileGatesData will generate a list of `Gate` struct to be used as the source data for an `SFOMuseumLookup` instance.
// The list of gate are compiled by iterating over one or more source. `iterator_uri` is a valid `whosonfirst/go-whosonfirst-iterate` URI
// and `iterator_sources` are one more (iterator) URIs to process.
func CompileGatesData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Gate, error) {

	lookup := make([]*Gate, 0)
	mu := new(sync.RWMutex)

	iter, err := iterate.NewIterator(ctx, iterator_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create iterator, %w", err)
	}

	for rec, err := range iter.Iterate(ctx, iterator_sources...) {

		if err != nil {
			return nil, err
		}

		defer rec.Body.Close()

		select {
		case <-ctx.Done():
			continue
		default:
			// pass
		}

		if strings.HasSuffix(rec.Path, "~") {
			continue
		}

		_, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse %s, %w", rec.Path, err)
		}

		if uri_args.IsAlternate {
			continue
		}

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			return nil, fmt.Errorf("Failed to read %s, %w", rec.Path, err)
		}

		wof_id, err := properties.Id(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive ID for %s, %w", rec.Path, err)
		}

		wof_name, err := properties.Name(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive name for %s, %w", rec.Path, err)
		}

		fl, err := properties.IsCurrent(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to determine is current for %s, %v", rec.Path, err)
		}

		inception := properties.Inception(body)
		cessation := properties.Cessation(body)

		g := &Gate{
			WhosOnFirstId: wof_id,
			Name:          wof_name,
			IsCurrent:     fl.Flag(),
			Inception:     inception,
			Cessation:     cessation,
		}

		mu.Lock()
		lookup = append(lookup, g)
		mu.Unlock()
	}

	return lookup, nil
}
