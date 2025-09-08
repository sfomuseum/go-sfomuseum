// For example:
//
//	$> ./bin/create-airline -name 'Flair Airlines' -parent-id 890458485 -iata-code F8 -icao-code FLE -wikidata-id Q4038797 -inception 2005 -cessation ..
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/sfomuseum/go-flags/flagset"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-export/v3"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"github.com/whosonfirst/go-writer/v3"
)

//go:embed stub.geojson
var stub []byte

func main() {

	fs := flagset.NewFlagSet("create")

	name := fs.String("name", "", "The name of the new airline")
	sfom_id := fs.Int64("sfomuseum-id", -1, "The SFO Museum FileMaker ID for the new airline")
	parent_id := fs.Int64("parent-id", -1, "The parent (Who's On First) ID for the new airline")

	icao_code := fs.String("icao-code", "", "...")
	icao_callsign := fs.String("icao-callsign", "", "...")
	iata_code := fs.String("iata-code", "", "...")
	wikidata_id := fs.String("wikidata-id", "", "...")

	inception := fs.String("inception", "", "A valid EDTF date string")
	cessation := fs.String("cessation", "", "A valid EDTF date string")

	parent_reader_uri := fs.String("parent-reader-uri", "fs:///usr/local/data/sfomuseum-data-whosonfirst/data", "A valid whosonfirst/go-reader URI.")

	writer_uri := fs.String("writer-uri", "fs:///usr/local/data/sfomuseum-data-enterprise/data", "A valid whosonfirst/go-writer URI. If empty the value of the -s fs will be used in combination with the fs:// scheme.")

	flagset.Parse(fs)

	ctx := context.Background()

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	body := stub

	updates := map[string]interface{}{
		"properties.wof:name":             *name,
		"properties.wof:parent_id":        *parent_id,
		"properties.sfomuseum:airline_id": *sfom_id,
	}

	if *inception != "" {
		updates["properties.edtf:inception"] = *inception
	}

	if *cessation != "" {
		updates["properties.edtf:cessation"] = *cessation

		if *cessation == ".." {
			updates["properties.mz:is_current"] = 1
		} else {
			updates["properties.mz:is_current"] = 0
		}
	}

	concordances := make(map[string]interface{})

	if *iata_code != "" || *icao_code != "" || *icao_callsign != "" || *wikidata_id != "" {

		if *wikidata_id != "" {
			concordances["wk:id"] = *wikidata_id
		}

		// In principle we could derive icao/iata codes
		// from Wikipedia/Wikidata here

		if *iata_code != "" {
			concordances["iata:code"] = *iata_code
			updates["properties.flysfo:airline_code"] = *iata_code
			updates["properties.flysfo:is_sfo"] = 1
		}

		if *icao_code != "" {
			concordances["icao:code"] = *icao_code
		}

		if *icao_callsign != "" {
			concordances["icao:callsign"] = *icao_callsign
		}

		updates["properties.wof:concordances"] = concordances
	}

	if *parent_id != -1 {

		parent_r, err := reader.NewReader(ctx, *parent_reader_uri)

		if err != nil {
			log.Fatalf("Failed to create new (parent) reader, %v", err)
		}

		parent_body, err := wof_reader.LoadBytes(ctx, parent_r, *parent_id)

		if err != nil {
			log.Fatalf("Failed to load parent (%d) from reader, %v", *parent_id, err)
		}

		pt, _, err := properties.Centroid(parent_body)

		if err != nil {
			log.Fatalf("Failed to determine parent centroid, %v", err)
		}

		coords := []float64{pt.X(), pt.Y()}

		updates["geometry.type"] = "Point"
		updates["geometry.coordinates"] = coords
		updates["properties.mz:is_approximate"] = 1

		to_copy := []string{
			"properties.wof:hierarchy",
			"properties.src:geom",
		}

		for _, path := range to_copy {

			rsp := gjson.GetBytes(parent_body, path)

			if !rsp.Exists() {
				log.Fatalf("Parent record is missing path '%s'", path)
			}

			updates[path] = rsp.Value()
		}

	}

	_, body, err = export.AssignPropertiesIfChanged(ctx, body, updates)

	if err != nil {
		log.Fatalf("Failed to assign properties, %v", err)
	}

	new_id, err := sfom_writer.WriteBytes(ctx, wr, body)

	if err != nil {
		log.Fatalf("Failed to write body, %v", err)
	}

	rel_path, err := uri.Id2RelPath(new_id)

	if err != nil {
		log.Fatalf("Failed to derive relative path from ID (%d), %v", new_id, err)
	}

	fmt.Println(rel_path)
}
