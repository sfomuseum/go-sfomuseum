package document

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sfomuseum/go-url-unshortener"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var cache unshortener.Unshortener

func init() {

	rate := time.Second / time.Duration(10)
	timeout := time.Second * time.Duration(30)

	worker, err := unshortener.NewThrottledUnshortener(rate, timeout)

	if err != nil {
		panic(err)
	}

	seed := make(map[string]string)

	cache, err = unshortener.NewCachedUnshortenerWithSeed(worker, seed)

	if err != nil {
		panic(err)
	}
}

func AppendUnshortenedURLs(ctx context.Context, body []byte) ([]byte, error) {

	urls_rsp := gjson.GetBytes(body, "entities.urls")

	if !urls_rsp.Exists() {
		return body, nil
	}

	for idx, u := range urls_rsp.Array() {

		expanded_rsp := u.Get("expanded_url")

		if !expanded_rsp.Exists() {
			continue
		}

		expanded_url := expanded_rsp.String()

		u, err := unshortener.UnshortenString(ctx, cache, expanded_url)

		if err != nil {
			log.Println(expanded_url, err)
			continue
		}

		path := fmt.Sprintf("entities.urls.%d.unshortened_url", idx)
		body, err = sjson.SetBytes(body, path, u.String())

		if err != nil {
			return nil, err
		}

		// log.Println(path, u.String())
		// log.Println(string(body))
	}

	return body, nil
}
