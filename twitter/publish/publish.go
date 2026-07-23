package publish

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/sfomuseum/go-sfomuseum/twitter/document"
	sfom_writer "github.com/sfomuseum/go-sfomuseum/writer"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst/v4/export"
	wof_reader "github.com/whosonfirst/go-whosonfirst/v4/reader"
	"github.com/whosonfirst/go-writer/v3"
)

type PublishOptions struct {
	Lookup   *sync.Map
	Reader   reader.Reader
	Writer   writer.Writer
	Exporter export.Exporter
}

func PublishTweet(ctx context.Context, opts *PublishOptions, body []byte) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	id_rsp := gjson.GetBytes(body, "id")

	if !id_rsp.Exists() {
		return errors.New("Missing 'id' property")
	}

	text_rsp := gjson.GetBytes(body, "full_text")

	if !text_rsp.Exists() {
		return errors.New("Missing 'full_text' property")
	}

	tweet_id := id_rsp.Int()
	full_text := text_rsp.String()

	tweet_body, err := document.AppendCreatedAtTimestamp(ctx, body)

	if err != nil {
		return nil
	}

	pointer, ok := opts.Lookup.Load(tweet_id)

	var wof_record []byte

	if ok {

		wof_id := pointer.(int64)

		wof_body, err := wof_reader.LoadBytes(ctx, opts.Reader, wof_id)

		if err != nil {
			return err
		}

		wof_record = wof_body

	} else {

		new_record, err := newWOFRecord(ctx)

		if err != nil {
			return err
		}

		wof_record = new_record

		created_rsp := gjson.GetBytes(tweet_body, "created")

		if !created_rsp.Exists() {
			return errors.New("Missing created timestamp")
		}

		created_ts := created_rsp.Int()
		created_t := time.Unix(created_ts, 0)

		edtf_t := created_t.Format(time.RFC3339)

		wof_record, err = sjson.SetBytes(wof_record, "properties.edtf:inception", edtf_t)

		if err != nil {
			return err
		}

		wof_record, err = sjson.SetBytes(wof_record, "properties.edtf:cessation", edtf_t)

		if err != nil {
			return err
		}
	}

	wof_name := full_text // fmt.Sprintf("Twitter message #%d", tweet_id)
	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:name", wof_name)

	if err != nil {
		return err
	}

	// why does this keep getting set incorrectly...

	/*
		wof_record, err = sjson.SetBytes(wof_record, "properties.wof:concordances.twitter:id", tweet_id)

		if err != nil {
			return err
		}
	*/

	var tw any

	err = json.Unmarshal(tweet_body, &tw)

	if err != nil {
		return err
	}

	updates := map[string]any{
		"properties.twitter:tweet": tw,
	}

	// Denormalize hash tags in to something that easier to facet

	hashtags_lookup := new(sync.Map)
	hashtags := make([]string, 0)

	for _, r := range gjson.GetBytes(tweet_body, "entities.hashtags").Array() {

		t := r.Get("text")
		hashtags_lookup.Store(t.String(), true)
	}

	hashtags_lookup.Range(func(k any, v any) bool {
		hashtags = append(hashtags, k.(string))
		return true
	})

	if len(hashtags) > 0 {
		updates["properties.twitter:hashtags"] = hashtags
	}

	// Denormalize user mentions in to something that easier to facet

	mentions_lookup := new(sync.Map)
	mentions := make([]string, 0)

	for _, r := range gjson.GetBytes(tweet_body, "entities.user_mentions").Array() {

		n := r.Get("screen_name")
		mentions_lookup.Store(n.String(), true)
	}

	mentions_lookup.Range(func(k any, v any) bool {
		mentions = append(mentions, k.(string))
		return true
	})

	if len(mentions) > 0 {
		updates["properties.twitter:user_mentions"] = mentions
	}

	//

	has_changed, new_body, err := export.AssignPropertiesIfChanged(ctx, wof_record, updates)

	if err != nil {
		return err
	}

	if !has_changed {
		return nil
	}

	if err != nil {
		return err
	}

	id, err := sfom_writer.WriteBytes(ctx, opts.Writer, new_body)

	if err != nil {
		return err
	}

	log.Printf("Wrote %d\n", id)
	return nil
}

func newWOFRecord(ctx context.Context) ([]byte, error) {

	// Null Terminal - please read these details from source...
	// https://raw.githubusercontent.com/sfomuseum-data/sfomuseum-data-architecture/master/data/115/916/086/9/1159160869.geojson

	parent_id := 1159160869

	lat := 37.616356
	lon := -122.386166

	hier := []map[string]any{
		{
			"building_id":      1159160869,
			"campus_id":        102527513,
			"continent_id":     102191575,
			"country_id":       85633793,
			"county_id":        102087579,
			"locality_id":      85922583,
			"neighbourhood_id": -1,
			"region_id":        85688637,
		},
	}

	geom := map[string]any{
		"type":        "Point",
		"coordinates": [2]float64{lon, lat},
	}

	feature := map[string]any{
		"type": "Feature",
		"properties": map[string]any{
			"sfomuseum:placetype": "tweet",
			"src:geom":            "sfomuseum",
			"wof:country":         "US",
			"wof:parent_id":       parent_id,
			"wof:placetype":       "custom",
			"wof:repo":            "sfomuseum-data-socialmedia-twitter",
			"wof:hierarchy":       hier,
		},
		"geometry": geom,
	}

	return json.Marshal(feature)
}
