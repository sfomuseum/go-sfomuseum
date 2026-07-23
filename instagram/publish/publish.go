package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/sfomuseum/go-sfomuseum/instagram/media"
	wof_reader "github.com/whosonfirst/go-whosonfirst/v4/reader"	
	sfom_writer "github.com/sfomuseum/go-sfomuseum/writer"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-writer/v3"
	"gocloud.dev/blob"
)

type PublishOptions struct {
	Lookup      *sync.Map
	Reader      reader.Reader
	Writer      writer.Writer
	MediaBucket *blob.Bucket
}

func PublishMedia(ctx context.Context, opts *PublishOptions, body []byte) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	path_rsp := gjson.GetBytes(body, "path")
	path := path_rsp.String()

	logger := slog.Default()
	logger = logger.With("path", path)

	is_video := false

	// This should be a little more sophisticated
	if filepath.Ext(path) == ".mp4" {
		is_video = true
	}

	body, err := media.AppendTakenAtTimestamp(ctx, body)

	if err != nil {
		return fmt.Errorf("Failed to append taken at timestamp, %w", err)
	}

	append_opts := &media.AppendHashesOptions{
		Bucket: opts.MediaBucket,
	}

	if is_video {
		append_opts.FileHash = true
	} else {
		append_opts.PerceptualHash = true
	}

	body, err = media.AppendHashes(ctx, append_opts, body)

	if err != nil {
		logger.Error("Failed to append hashes", "error", err)
		return fmt.Errorf("Failed to append hashes, %w", err)
	}

	body, err = media.ExpandCaption(ctx, body)

	if err != nil {
		logger.Error("Failed to expand caption", "error", err)
		return fmt.Errorf("Failed to expand caption, %w", err)
	}

	// We used to use media_id which is derived from the media file path.
	// However between Oct 2020 and April 2022 those paths changed from
	// being something like {HASH}.jpg to {SOME}-{THING}-{SOME}-{THING}.jpg
	// The former allows us to use {HASH} in the mf.sfom URL.

	// You might be asking yourself: Do we really need media ID? The answer
	// is yes. More specifically we need something that we can for reliably
	// de-depuplicating IG posts we've already imported. As stated we originally
	// thought we could rely on the path of the media file associated with a
	// post but apparently not (they seem to change).

	// Unfortunately for SFO Museum we can't use the body of the caption either
	// since we sometimes use the same caption for multiple posts. Nor can we
	// use caption + taken (or taken at) since many of these posts are posted
	// automatically by tools like hootsuite so they end up with the same timestamps.
	// For example:

	// https://raw.githubusercontent.com/sfomuseum-data/sfomuseum-data-socialmedia-instagram/main/data/172/935/502/5/1729355025.geojson?token={TOKEN}
	// https://raw.githubusercontent.com/sfomuseum-data/sfomuseum-data-socialmedia-instagram/main/data/172/935/502/3/1729355023.geojson?token={TOKEN}

	media_id, err := DeriveMediaId(body, "")

	if err != nil {
		logger.Error("Failed to derive media ID", "error", err)
		return fmt.Errorf("Failed to derive media ID, %w", err)
	}

	body, err = sjson.SetBytes(body, "media_id", media_id)

	if err != nil {
		logger.Error("Failed to assign media ID", "error", err)
		return fmt.Errorf("Failed to assign media_id to post, %w", err)
	}

	// lookup.go

	pointer, ok := opts.Lookup.Load(media_id)

	// Add path to the file as a fallback because apparently IG does stuff to the
	// photos between archive runs that causes the percaptual hash to change. Good
	// times...

	if !ok {
		pointer, ok = opts.Lookup.Load(path)
	}

	var wof_record []byte

	if ok {

		wof_id := pointer.(int64)

		wof_body, err := wof_reader.LoadBytes(ctx, opts.Reader, wof_id)

		if err != nil {
			return err
		}

		wof_record = wof_body

		// See this? We are going to ensure we don't accidentally overwrite an
		// existing media ID. For example the inputs for deriving a media ID
		// changed between 202010 and 202204 to reflect changes IG made to their
		// exports.

		id_rsp := gjson.GetBytes(wof_body, "properties.instagram:post.media_id")

		body, err = sjson.SetBytes(body, "media_id", id_rsp.String())

		if err != nil {
			logger.Error("Failed to assign media ID", "error", err)
			return fmt.Errorf("Failed to assign media_id to post, %w", err)
		}

	} else {

		new_record, err := newWOFRecord(ctx)

		if err != nil {
			logger.Error("Failed to create new record", "error", err)
			return err
		}

		wof_record = new_record
	}

	taken_rsp := gjson.GetBytes(body, "taken")

	if !taken_rsp.Exists() {
		logger.Error("Missing taken property", "error", err)
		return fmt.Errorf("Missing created timestamp")
	}

	taken := taken_rsp.Int()

	taken_t := time.Unix(taken, 0)

	if err != nil {
		logger.Error("Failed to parse taken timestamp", "timestamp", taken, "error", err)
		return err
	}

	taken_str := taken_t.Format(time.RFC3339)

	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:created", taken_t.Unix())

	if err != nil {
		return err
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.edtf:inception", taken_str)

	if err != nil {
		logger.Error("Failed to assign inception", "error", err)
		return err
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.edtf:cessation", taken_str)

	if err != nil {
		logger.Error("Failed to assign cessation", "error", err)
		return err
	}

	excerpt_rsp := gjson.GetBytes(body, "caption.excerpt")

	if !excerpt_rsp.Exists() {
		logger.Error("Failed to assign caption", "error", err)
		return fmt.Errorf("Missing caption.excerpt")
	}

	wof_name := fmt.Sprintf("%s..", excerpt_rsp.String())
	wof_record, err = sjson.SetBytes(wof_record, "properties.wof:name", wof_name)

	if err != nil {
		logger.Error("Failed to assign name", "error", err)
		return err
	}

	var post interface{}

	err = json.Unmarshal(body, &post)

	if err != nil {
		logger.Error("Failed to unmarshal post", "error", err)
		return fmt.Errorf("Failed to unmarshal record, %w", err)
	}

	wof_record, err = sjson.SetBytes(wof_record, "properties.instagram:post", post)

	if err != nil {
		logger.Error("Failed to assign post properties", "error", err)
		return fmt.Errorf("Failed to append post, %w", err)
	}

	_, err = sfom_writer.WriteBytes(ctx, opts.Writer, wof_record)

	if err != nil {
		logger.Error("Failed to write new record", "error", err)
		return fmt.Errorf("Failed to write record, %w", err)
	}

	return nil
}

func newWOFRecord(ctx context.Context) ([]byte, error) {

	// Null Terminal - please read these details from source...
	// https://raw.githubusercontent.com/sfomuseum-data/sfomuseum-data-architecture/master/data/115/916/086/9/1159160869.geojson

	parent_id := 1159160869

	lat := 37.616356
	lon := -122.386166

	hier := []map[string]interface{}{
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

	geom := map[string]interface{}{
		"type":        "Point",
		"coordinates": [2]float64{lon, lat},
	}

	feature := map[string]interface{}{
		"type": "Feature",
		"properties": map[string]interface{}{
			"sfomuseum:placetype": "instagram",
			"src:geom":            "sfomuseum",
			"wof:country":         "US",
			"wof:parent_id":       parent_id,
			"wof:placetype":       "custom",
			"wof:repo":            "sfomuseum-data-socialmedia-instagram",
			"wof:hierarchy":       hier,
		},
		"geometry": geom,
	}

	return json.Marshal(feature)
}
