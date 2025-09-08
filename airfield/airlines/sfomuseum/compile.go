package sfomuseum

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func CompileAirlinesData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]*Airline, error) {

	lookup := make([]*Airline, 0)
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

		if pt_rsp.String() != "airline" {
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

		sfom_rsp := gjson.GetBytes(body, "properties.sfomuseum:airline_id")

		if sfom_rsp.Exists() {
			sfom_id = sfom_rsp.Int()
		}

		role_rsp := gjson.GetBytes(body, "properties.sfomuseum:airline_role")

		a := &Airline{
			WhosOnFirstId: wof_id,
			SFOMuseumId:   sfom_id,
			Name:          wof_name,
			Role:          role_rsp.String(),
			IsCurrent:     fl.Flag(),
		}

		concordances := properties.Concordances(body)

		if concordances != nil {

			iata_code, ok := concordances["iata:code"]

			if ok {
				a.IATACode = fmt.Sprintf("%s", iata_code)
			}

			icao_code, ok := concordances["icao:code"]

			if ok {
				a.ICAOCode = fmt.Sprintf("%s", icao_code)
			}

			callsign, ok := concordances["icao:callsign"]

			if ok {
				a.ICAOCallsign = fmt.Sprintf("%s", callsign)
			}

			id, ok := concordances["wd:id"]

			if ok {
				a.WikidataId = fmt.Sprintf("%s", id)
			}
		}

		mu.Lock()
		lookup = append(lookup, a)
		mu.Unlock()
	}

	return lookup, nil
}
