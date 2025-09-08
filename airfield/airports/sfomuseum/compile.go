package sfomuseum

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func CompileAirportsData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Airport, error) {

	lookup := make([]*Airport, 0)
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

		logger := slog.Default()
		logger = logger.With("path", rec.Path)

		select {
		case <-ctx.Done():
			continue
		default:
			// pass
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

		pt_rsp := gjson.GetBytes(body, "properties.sfomuseum:placetype")

		if pt_rsp.String() != "airport" {
			slog.Info("Skipping record because it is not an airport", "placetype", pt_rsp.String())
			continue
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

		sfom_id := int64(-1)

		sfom_rsp := gjson.GetBytes(body, "properties.sfomuseum:airport_id")

		if sfom_rsp.Exists() {
			sfom_id = sfom_rsp.Int()
		}

		a := &Airport{
			WhosOnFirstId: wof_id,
			SFOMuseumId:   sfom_id,
			Name:          wof_name,
			IsCurrent:     fl.Flag(),
		}

		concordances := properties.Concordances(body)

		if concordances != nil {

			iata_code, ok := concordances["iata:code"]

			if ok && iata_code != "" {
				a.IATACode = fmt.Sprintf("%s", iata_code)
			}

			icao_code, ok := concordances["icao:code"]

			if ok && icao_code != "" {
				a.ICAOCode = fmt.Sprintf("%s", icao_code)
			}

			id, ok := concordances["wd:id"]

			if ok {
				a.WikidataId = fmt.Sprintf("%s", id)
			}
		}

		mu.Lock()
		lookup = append(lookup, a)
		mu.Unlock()

		logger.Debug("Add record", "wof id", wof_id, "iata code", a.IATACode, "icao code", a.ICAOCode, "is current", a.IsCurrent)
	}

	return lookup, nil
}
