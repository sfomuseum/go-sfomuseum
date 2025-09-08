package flysfo

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func CompileAirlinesData(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]Airline, error) {

	lookup := make([]Airline, 0)
	mu := new(sync.RWMutex)

	seen := new(sync.Map)

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

		concordances := properties.Concordances(body)

		if concordances == nil {
			continue
		}

		iata_code, ok := concordances["flysfo:code"]

		if !ok {
			continue
		}

		fl, err := properties.IsCurrent(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to determine whether %s is current, %v", rec.Path, err)
		}

		if !fl.IsTrue() || !fl.IsKnown() {
			continue
		}

		wof_id, err := properties.Id(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive ID, %w", err)
		}

		wof_name, err := properties.Name(body)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive name, %w", err)
		}

		a := Airline{
			WhosOnFirstId: wof_id,
			FlysfoID:      fmt.Sprintf("%s", iata_code),
			Name:          wof_name,
			IATACode:      fmt.Sprintf("%s", iata_code),
		}

		icao_code, ok := concordances["icao:code"]

		if ok {
			w, ok := seen.Load(icao_code)

			if ok {
				log.Println("WARNING", icao_code, w, wof_id)

			} else {
				seen.Store(icao_code, wof_id)
			}

			a.ICAOCode = fmt.Sprintf("%s", icao_code)

		} else {
			log.Println("WARNING", "Missing ICAO code", wof_id)
		}

		mu.Lock()
		lookup = append(lookup, a)
		mu.Unlock()
	}

	return lookup, nil
}
