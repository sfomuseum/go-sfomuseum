package publish

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"

	_ "github.com/whosonfirst/go-whosonfirst/v4/iterate/git"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst/v4/iterate"
)

func BuildLookup(ctx context.Context, iterator_uri string, iterator_path string) (*sync.Map, error) {

	lookup := new(sync.Map)
	count := int32(0)

	iter, err := iterate.NewIterator(ctx, iterator_uri)

	if err != nil {
		return nil, err
	}

	for rec, err := range iter.Iterate(ctx, iterator_path) {

		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			return nil, err
		}

		rec.Body.Close()

		wof_rsp := gjson.GetBytes(body, "properties.wof:id")

		if !wof_rsp.Exists() {
			return nil, fmt.Errorf("Missing WOF ID")
		}

		wof_id := wof_rsp.Int()

		// See notes about lookup_keys (and media_id) in publish.go

		var media_id string

		phash_rsp := gjson.GetBytes(body, "properties.instagram:post.perceptual_hash")

		if phash_rsp.Exists() {

			m, err := DeriveMediaId(body, "properties.instagram:post")

			if err != nil {
				return nil, fmt.Errorf("Failed to derive media ID for %s, %w", rec.Path, err)
			}

			media_id = m

		} else {

			log.Printf("%s is missing hash\n", rec.Path)
			continue
		}

		lookup.Store(media_id, wof_id)

		// Add path to the file as a fallback because apparently IG does stuff to the
		// photos between archive runs that causes the percaptual hash to change. Good
		// times...

		path_rsp := gjson.GetBytes(body, "properties.instagram:post.media_id")

		if path_rsp.Exists() {

			path := path_rsp.String()

			v, exists := lookup.Load(path)

			if exists && v.(int64) != wof_id {
				return nil, fmt.Errorf("Failed to store path (%s) for %d because there is already an entry for %d", rec.Path, wof_id, v.(int64))
			}

			lookup.Store(path, wof_id)
		}

		atomic.AddInt32(&count, 1)
	}

	return lookup, nil
}
