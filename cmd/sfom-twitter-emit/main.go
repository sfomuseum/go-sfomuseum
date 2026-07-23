package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aaronland/go-json-query"
	"github.com/sfomuseum/go-sfomuseum/twitter"
	"github.com/sfomuseum/go-sfomuseum/twitter/document"
	"github.com/sfomuseum/go-sfomuseum/twitter/walk"
	"github.com/tidwall/pretty"
)

func main() {

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

	to_stdout := flag.Bool("stdout", true, "Emit to STDOUT")
	to_devnull := flag.Bool("null", false, "Emit to /dev/null")
	as_json := flag.Bool("json", false, "Emit a JSON list.")
	format_json := flag.Bool("format-json", false, "Format JSON output for each record.")

	append_timestamp := flag.Bool("append-timestamp", false, "Append a `created` property containing a Unix timestamp derived from the `created_at` property.")
	append_urls := flag.Bool("append-urls", false, "Append a `unshortened_url` property for each `entities.urls.(n)` property.")
	append_all := flag.Bool("append-all", false, "Enable all the `-append-` flags.")

	var queries query.QueryFlags
	flag.Var(&queries, "query", "One or more {PATH}={REGEXP} parameters for filtering records.")

	valid_modes := strings.Join([]string{query.QUERYSET_MODE_ALL, query.QUERYSET_MODE_ANY}, ", ")
	desc_modes := fmt.Sprintf("Specify how query filtering should be evaluated. Valid modes are: %s", valid_modes)

	query_mode := flag.String("query-mode", query.QUERYSET_MODE_ALL, desc_modes)

	flag.Parse()

	if *append_all {
		*append_timestamp = true
		*append_urls = true
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	writers := make([]io.Writer, 0)

	if *to_stdout {
		writers = append(writers, os.Stdout)
	}

	if *to_devnull {
		writers = append(writers, ioutil.Discard)
	}

	if len(writers) == 0 {
		log.Fatal("Nothing to write to.")
	}

	wr := io.MultiWriter(writers...)

	open_opts := &twitter.OpenTweetsOptions{
		TrimPrefix: *trim_prefix,
	}

	tweets_fh, err := twitter.OpenTweets(ctx, *tweets_uri, open_opts)

	if err != nil {
		log.Fatalf("Failed to open %s, %v", *tweets_uri, err)
	}

	defer tweets_fh.Close()

	count := uint32(0)
	mu := new(sync.RWMutex)

	if *as_json {
		wr.Write([]byte("["))
	}

	cb := func(ctx context.Context, body []byte) error {

		if *append_timestamp {

			b, err := document.AppendCreatedAtTimestamp(ctx, body)

			if err != nil {
				return err
			}

			body = b
		}

		if *append_urls {

			b, err := document.AppendUnshortenedURLs(ctx, body)

			if err != nil {
				return err
			}

			body = b
		}

		mu.Lock()
		defer mu.Unlock()

		new_count := atomic.AddUint32(&count, 1)

		if new_count > 1 {

			if *as_json {
				wr.Write([]byte(","))
			}
		}

		if *as_json && *format_json {
			body = pretty.Pretty(body)
		}

		wr.Write(body)
		wr.Write([]byte("\n"))

		return nil
	}

	walk_opts := &walk.WalkWithCallbackOptions{
		Callback: cb,
	}

	if len(queries) > 0 {

		qs := &query.QuerySet{
			Queries: queries,
			Mode:    *query_mode,
		}

		walk_opts.QuerySet = qs
	}

	err = walk.WalkTweetsWithCallback(ctx, walk_opts, tweets_fh)

	if err != nil {
		log.Fatal(err)
	}

	if *as_json {
		wr.Write([]byte("]"))
	}

}
