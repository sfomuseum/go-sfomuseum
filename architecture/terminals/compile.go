package terminals

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

// CompileTerminalsData will generate a list of `Terminal` struct to be used as the source data for an `SFOMuseumLookup` instance.
// The list of terminal are compiled by iterating over one or more source. `iterator_uri` is a valid `whosonfirst/go-whosonfirst-iterate` URI
// and `iterator_sources` are one more (iterator) URIs to process.
func CompileTerminalsData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Terminal, error) {

	lookup := make([]*Terminal, 0)
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

		preferred_names := make([]string, 0)
		variant_names := make([]string, 0)

		names := properties.Names(body)

		for k, k_names := range names {

			if strings.HasSuffix(k, "_preferred") {

				for _, n := range k_names {
					preferred_names = append(preferred_names, n)
				}
			} else if strings.HasSuffix(k, "_variant") {

				for _, n := range k_names {
					variant_names = append(variant_names, n)
				}
			} else {
			}

		}

		inception := properties.Inception(body)
		cessation := properties.Cessation(body)

		g := &Terminal{
			WhosOnFirstId:  wof_id,
			Name:           wof_name,
			IsCurrent:      fl.Flag(),
			PreferredNames: preferred_names,
			VariantNames:   variant_names,
			Inception:      inception,
			Cessation:      cessation,
		}

		sfom_rsp := gjson.GetBytes(body, "properties.sfomuseum:terminal_id")

		if sfom_rsp.Exists() {
			g.SFOMuseumId = sfom_rsp.String()
		}

		mu.Lock()
		lookup = append(lookup, g)
		mu.Unlock()
	}

	return lookup, nil
}
