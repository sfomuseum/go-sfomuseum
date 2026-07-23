package main

import (
	"context"
	"flag"
	"log"

	_ "github.com/sfomuseum/go-sfomuseum/export"
	_ "gocloud.dev/blob/fileblob"
	
	"github.com/sfomuseum/go-sfomuseum/twitter"
	"github.com/sfomuseum/go-sfomuseum/twitter/publish"
	"github.com/sfomuseum/go-sfomuseum/twitter/walk"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst/v4/export"
	"github.com/whosonfirst/go-writer/v3"
)

func main() {

	tweets_uri := flag.String("tweets-uri", "", "A valid gocloud.dev/blob URI to your `tweets.js` file.")
	trim_prefix := flag.Bool("trim-prefix", true, "Trim default tweet.js JavaScript prefix.")

	iterator_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-socialmedia-twitter", "...")

	reader_uri := flag.String("reader-uri", "fs:///usr/local/data/sfomuseum-data-socialmedia-twitter/data", "A valid whosonfirst/go-reader URI")
	writer_uri := flag.String("writer-uri", "fs:///usr/local/data/sfomuseum-data-socialmedia-twitter/data", "A valid whosonfirst/go-writer URI")

	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	open_opts := &twitter.OpenTweetsOptions{
		TrimPrefix: *trim_prefix,
	}

	tweets_fh, err := twitter.OpenTweets(ctx, *tweets_uri, open_opts)

	if err != nil {
		log.Fatalf("Failed to open %s, %v", *tweets_uri, err)
	}

	defer tweets_fh.Close()

	rdr, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}

	wrtr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create writer, %v", err)
	}

	exprtr, err := export.NewExporter(ctx, "sfomuseum://")

	if err != nil {
		log.Fatalf("Failed to create exported, %v", err)
	}

	lookup, err := publish.BuildLookup(ctx, *iterator_uri, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to build lookup for '%s' (%s), %v", *iterator_uri, *iterator_source, err)
	}

	publish_opts := &publish.PublishOptions{
		Lookup:   lookup,
		Reader:   rdr,
		Writer:   wrtr,
		Exporter: exprtr,
	}

	max_procs := 10
	throttle := make(chan bool, max_procs)

	for i := 0; i < max_procs; i++ {
		throttle <- true
	}

	cb := func(ctx context.Context, body []byte) error {

		<-throttle

		defer func() {
			throttle <- true
		}()

		return publish.PublishTweet(ctx, publish_opts, body)
	}

	opts := &walk.WalkWithCallbackOptions{
		Callback: cb,
	}

	err = walk.WalkTweetsWithCallback(ctx, opts, tweets_fh)

	if err != nil {
		log.Fatalf("Failed to walk tweets, %v", err)
	}

}
