package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"sync"

	_ "gocloud.dev/blob/fileblob"
	
	"github.com/sfomuseum/go-sfomuseum/twitter"
	"github.com/sfomuseum/go-sfomuseum/twitter/walk"
	"github.com/tidwall/gjson"
)

func main() {

	dump_hashtags := flag.Bool("hashtags", true, "Export hash tags in tweets.")
	dump_mentions := flag.Bool("mentions", true, "Export users mentioned in tweets.")

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

	flag.Parse()

	ctx := context.Background()

	open_opts := &twitter.OpenTweetsOptions{
		TrimPrefix: *trim_prefix,
	}

	tweets_fh, err := twitter.OpenTweets(ctx, *tweets_uri, open_opts)

	if err != nil {
		log.Fatalf("Failed to open %s, %v", *tweets_uri, err)
	}

	defer tweets_fh.Close()

	mentions := new(sync.Map)
	hashtags := new(sync.Map)

	cb := func(ctx context.Context, body []byte) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		if *dump_mentions {
			mentions_rsp := gjson.GetBytes(body, "entities.user_mentions")

			if !mentions_rsp.Exists() {
				return errors.New("Missing mentions")
			}

			for _, m := range mentions_rsp.Array() {
				user_rsp := m.Get("screen_name")
				user := user_rsp.String()
				mentions.Store(user, true)
			}
		}

		if *dump_hashtags {
			hashtags_rsp := gjson.GetBytes(body, "entities.hashtags")

			if !hashtags_rsp.Exists() {
				return errors.New("Missing hashtags")
			}

			for _, h := range hashtags_rsp.Array() {
				tag_rsp := h.Get("text")
				tag := tag_rsp.String()
				hashtags.Store(tag, true)
			}
		}

		return nil
	}

	opts := &walk.WalkWithCallbackOptions{
		Callback: cb,
	}

	err = walk.WalkTweetsWithCallback(ctx, opts, tweets_fh)

	if err != nil {
		log.Fatal(err)
	}

	writers := []io.Writer{
		os.Stdout,
	}

	multi := io.MultiWriter(writers...)
	wr := csv.NewWriter(multi)

	write_row := func(row []string) bool {

		err := wr.Write(row)

		if err != nil {
			log.Println(err)
			return false
		}

		return true
	}

	write_row([]string{"property", "value"})

	mentions.Range(func(key interface{}, value interface{}) bool {

		user := key.(string)

		out := []string{
			"user",
			user,
		}

		return write_row(out)
	})

	hashtags.Range(func(key interface{}, value interface{}) bool {

		tag := key.(string)

		out := []string{
			"tag",
			tag,
		}

		return write_row(out)
	})

}
