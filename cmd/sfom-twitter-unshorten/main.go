package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sfomuseum/go-sfomuseum/twitter"
	"github.com/sfomuseum/go-sfomuseum/twitter/walk"
	"github.com/sfomuseum/go-url-unshortener"
	"github.com/tidwall/gjson"	
)

func main() {

	progress := flag.Bool("progress", false, "Display progress information")
	qps := flag.Int("qps", 10, "Number of (unshortening) queries per second")
	to := flag.Int("timeout", 30, "Maximum number of seconds of for an unshorterning request")
	seed_file := flag.String("seed", "", "Pre-fill the unshortening cache with data in this file")

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

	flag.Parse()

	rate := time.Second / time.Duration(*qps)
	timeout := time.Second * time.Duration(*to)

	seed := make(map[string]string)

	if *seed_file != "" {

		fh, err := os.Open(*seed_file)

		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, &seed)

		if err != nil {
			log.Fatal(err)
		}
	}

	worker, err := unshortener.NewThrottledUnshortener(rate, timeout)

	if err != nil {
		log.Fatal(err)
	}

	cache, err := unshortener.NewCachedUnshortenerWithSeed(worker, seed)

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signal_ch := make(chan os.Signal)
	signal.Notify(signal_ch, os.Interrupt, syscall.SIGTERM)

	go func(c chan os.Signal) {
		<-c
		cancel()
		os.Exit(0)
	}(signal_ch)

	open_opts := &twitter.OpenTweetsOptions{
		TrimPrefix: *trim_prefix,
	}

	tweets_fh, err := twitter.OpenTweets(ctx, *tweets_uri, open_opts)

	if err != nil {
		log.Fatalf("Failed to open %s, %v", *tweets_uri, err)
	}

	defer tweets_fh.Close()

	lookup := new(sync.Map)

	count_urls := int32(0)
	remaining_urls := count_urls

	cb := func(ctx context.Context, body []byte) error {

		urls_rsp := gjson.GetBytes(body, "entities.urls")

		if !urls_rsp.Exists() {
			return errors.New("Missing URLs")
		}

		to_fetch := make([]string, 0)

		for _, u := range urls_rsp.Array() {

			short_rsp := u.Get("expanded_url")

			if !short_rsp.Exists() {
				return errors.New("Missing expanded_url property")
			}

			short_url := short_rsp.String()

			_, ok := lookup.LoadOrStore(short_url, "...")

			if ok {
				continue
			}

			atomic.AddInt32(&count_urls, 1)
			atomic.AddInt32(&remaining_urls, 1)

			to_fetch = append(to_fetch, short_url)
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		for _, short_url := range to_fetch {

			url, err := unshortener.UnshortenString(ctx, cache, short_url)

			atomic.AddInt32(&remaining_urls, -1)

			if err != nil {
				lookup.Store(short_url, "?")
				return err
			}

			long_url := url.String()

			if short_url == long_url {
				long_url = "-"
			}

			lookup.Store(short_url, long_url)
		}

		return nil
	}

	completed_ch := make(chan bool)

	if *progress {

		go func() {

			for {
				select {
				case <-completed_ch:
					break
				case <-time.After(10 * time.Second):

					count := atomic.LoadInt32(&count_urls)
					remaining := atomic.LoadInt32(&remaining_urls)

					log.Printf("%d of %d URLs left to unshorten\n", remaining, count)
				}
			}
		}()
	}

	opts := &walk.WalkWithCallbackOptions{
		Callback: cb,
	}

	err = walk.WalkTweetsWithCallback(ctx, opts, tweets_fh)

	completed_ch <- true

	if err != nil {
		log.Fatal(err)
	}

	report := make(map[string]string)

	lookup.Range(func(k interface{}, v interface{}) bool {
		shortened_url := k.(string)
		unshortened_url := v.(string)
		report[shortened_url] = unshortened_url
		return true
	})

	writers := make([]io.Writer, 0)
	writers = append(writers, os.Stdout)

	out := io.MultiWriter(writers...)

	enc := json.NewEncoder(out)
	err = enc.Encode(report)

	if err != nil {
		log.Fatal(err)
	}

}
