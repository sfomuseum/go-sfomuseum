package export

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/tidwall/gjson"
	wof_export "github.com/whosonfirst/go-whosonfirst/v4/export"
)

func TestSFOMuseumExport(t *testing.T) {

	ctx := context.Background()

	cwd, err := os.Getwd()

	if err != nil {
		t.Fatal(err)
	}

	fixtures := filepath.Join(cwd, "fixtures")
	feature_path := filepath.Join(fixtures, "1159159407.geojson")

	feature_fh, err := os.Open(feature_path)

	if err != nil {
		t.Fatal(err)
	}

	defer feature_fh.Close()

	body, err := ioutil.ReadAll(feature_fh)

	if err != nil {
		t.Fatal(err)
	}

	ex, err := wof_export.NewExporter(ctx, "sfomuseum://")

	if err != nil {
		t.Fatal(err)
	}

	_, body, err = ex.Export(ctx, body)

	if err != nil {
		t.Fatal(err)
	}

	ensure_props := []string{
		"properties.wof:id",
		"properties.geom:bbox",
		"properties.wof:depicts",
		"bbox",
	}

	for _, prop := range ensure_props {

		prop_rsp := gjson.GetBytes(body, prop)

		if !prop_rsp.Exists() {
			t.Fatalf("Missing property '%s'", prop)
		}

		fmt.Printf("%s: %s\n", prop, prop_rsp.String())
	}

	bbox_rsp := gjson.GetBytes(body, "properties.geom:bbox")
	bbox_str := bbox_rsp.String()

	if bbox_str != "-122.384119,37.615457,-122.384119,37.615457" {
		t.Fatal("Unexpected geom:bbox")
	}

}
