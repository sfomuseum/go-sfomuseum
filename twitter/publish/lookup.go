package publish

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	_ "github.com/whosonfirst/go-whosonfirst/v4/iterate/git"
	
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst/v4/iterate"
)

func BuildLookup(ctx context.Context, indexer_uri string, indexer_path string) (*sync.Map, error) {

	lookup := new(sync.Map)
	count := int32(0)

	iter, err := iterate.NewIterator(ctx, indexer_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new iterator for %s, %w", indexer_uri, err)
	}

	for rec, err := range iter.Iterate(ctx, indexer_path) {
		
		if err != nil {
			return nil, fmt.Errorf("Failed to iterate URIs, %w", err)
		}

		defer rec.Body.Close()
		
		body, err := io.ReadAll(rec.Body)

		if err != nil {
			return nil, fmt.Errorf("Failed to read %s, %w", rec.Path, err)
		}

		wof_rsp := gjson.GetBytes(body, "properties.wof:id")

		if !wof_rsp.Exists() {
			return nil, fmt.Errorf("Missing WOF ID")
		}

		wof_id := wof_rsp.Int()

		tweet_rsp := gjson.GetBytes(body, "properties.twitter:tweet.id")

		if !tweet_rsp.Exists() {
			return nil, fmt.Errorf("Missing Twitter ID for record %d", wof_id)
		}

		tweet_id := tweet_rsp.Int()

		lookup.Store(tweet_id, wof_id)

		atomic.AddInt32(&count, 1)
	}

	return lookup, nil
}
