// assign-hash is a command line tool to ensure that all records in the sfomuseum-data-socialmedia-instagram
// repo have `instagram:post.perceptual_hash` properties. For records without the property the tool will fetch
// the corresponding "_o.jpg" image from the `sfomuseum-media` S3 bucket and use it to derive a new perceptual
// hash and then update the record accordingly. This was originally written to "backfill" properties and shouldn't
// be necessary going forward (hash are appended in publish/publish.go) but the tool is being kept around in
// case it's ever needed again.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"path/filepath"

	_ "github.com/aaronland/gocloud/blob/s3"

	"github.com/sfomuseum/go-sfomuseum/instagram/hash"
	"github.com/sfomuseum/go-sfomuseum/instagram/publish/secret"
	sfom_writer "github.com/sfomuseum/go-sfomuseum/writer"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst/v4/export"
	"github.com/whosonfirst/go-whosonfirst/v4/iterate"
	"github.com/whosonfirst/go-writer/v3"
	"gocloud.dev/blob"
)

func main() {

	media_bucket_uri := flag.String("media-bucket-uri", "s3blob://sfomuseum-media?prefix=media/instagram/&region=us-west-2&credentials=session", "...")

	iterator_uri := flag.String("iterator-uri", "repo://", "...")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-socialmedia-instagram", "...")

	writer_uri := flag.String("writer-uri", "repo:///usr/local/data/sfomuseum-data-socialmedia-instagram", "...")

	flag.Parse()

	ctx := context.Background()

	wr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	media_bucket, err := blob.OpenBucket(ctx, *media_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open media bucket, %v", err)
	}

	iter, err := iterate.NewIterator(ctx, *iterator_uri)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	for rec, err := range iter.Iterate(ctx, *iterator_source) {

		if err != nil {
			log.Fatalf("Failed to iterate, %v", err)
		}

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			log.Fatalf("Failed to read %s, %v", rec.Path, err)
		}

		phash_rsp := gjson.GetBytes(body, "properties.instagram:post.perceptual_hash")

		if phash_rsp.Exists() {
			// return nil
		}

		id_rsp := gjson.GetBytes(body, "properties.instagram:post.media_id")

		if !id_rsp.Exists() {
			log.Fatalf("%s is missing instagram:post.media_id property, %v", rec.Path, err)
		}

		media_id := id_rsp.String()
		media_secret := secret.DeriveSecret(media_id)

		im_fname := fmt.Sprintf("%s_%s_o.jpg", media_id, media_secret)
		im_path := filepath.Join(media_id, im_fname)

		im_r, err := media_bucket.NewReader(ctx, im_path, nil)

		// TBD: Mark as deprecated?

		if err != nil {
			log.Printf("Failed to open %s, %v", im_path, err)
			continue
		}

		defer im_r.Close()

		phash, err := hash.PerceptualHash(im_r)

		if err != nil {
			log.Fatalf("Failed to generate perceptual hash for %s (%s %s), %v", im_path, media_id, media_secret, err)
		}

		updates := map[string]any{
			"properties.instagram:post.perceptual_hash": phash,
		}

		has_changed, new_body, err := export.AssignPropertiesIfChanged(ctx, body, updates)

		if err != nil {
			log.Fatalf("Failed to assign perceptual hash to %s, %v", rec.Path, err)
		}

		if has_changed {

			_, err := sfom_writer.WriteBytes(ctx, wr, new_body)

			if err != nil {
				log.Fatalf("Failed to write %s, %v", rec.Path, err)
			}
		}
	}

}
