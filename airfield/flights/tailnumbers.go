package flights

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
)

// TailNumbersFromIterator will return the unique set of `swim:tail_number` values from one or more whosonfirst/go-whosonfirst-iterate sources for SFO Museum flight data.
func TailNumbersFromIterator(ctx context.Context, iterator_uri string, iterator_sources ...string) ([]string, error) {

	catalog := new(sync.Map)

	iter, err := iterate.NewIterator(ctx, iterator_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create iterator, %v", err)
	}

	for rec, err := range iter.Iterate(ctx, iterator_sources...) {

		if err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done():
			continue
		default:
			// pass
		}
		body, err := io.ReadAll(rec.Body)

		if err != nil {
			return nil, fmt.Errorf("Failed to read %s, %w", rec.Path, err)
		}

		rsp := gjson.GetBytes(body, "properties.swim:tail_number")

		if !rsp.Exists() {
			continue
		}

		v := rsp.String()

		if v == "" {
			continue
		}

		catalog.Store(v, true)
	}

	values := make([]string, 0)

	catalog.Range(func(v interface{}, ignore interface{}) bool {
		values = append(values, v.(string))
		return true
	})

	sort.Strings(values)
	return values, nil
}
