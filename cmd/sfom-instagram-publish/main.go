// publish is a command-line tool to merge Instagram posts defined in one or more "media.json" files
// with the sfomuseum-data-socialmedia-instagram repository. For example:
//
//	$> ./bin/publish -media-bucket-uri file:///Volumes/Museum/_Public/_Social_Media/SM\ downloads/2022/sfomuseum_20220418/ file:///usr/local/data/media.json
//
// Important: As of April, 2022 Instagram no longer publishes "media.json" files with the export bundles.
// Use the sfomuseum/go-sfomuseum-instagram/cmd/derive-media-json tool to create a media.json file from
// the available data.
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"

	_ "github.com/aaronland/gocloud/blob/s3"
	_ "gocloud.dev/blob/fileblob"
	_ "image/jpeg"

	"github.com/sfomuseum/go-sfomuseum/instagram/media"
	"github.com/sfomuseum/go-sfomuseum/instagram/publish"
	"github.com/sfomuseum/go-sfomuseum/instagram/walk"
	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-writer/v3"
	"gocloud.dev/blob"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-socialmedia-instagram", "...")

	reader_uri := flag.String("reader-uri", "repo:///usr/local/data/sfomuseum-data-socialmedia-instagram", "A valid whosonfirst/go-reader URI")
	writer_uri := flag.String("writer-uri", "repo:///usr/local/data/sfomuseum-data-socialmedia-instagram", "A valid whosonfirst/go-writer URI")

	media_bucket_uri := flag.String("media-bucket-uri", "", "A valid gocloud.dev/blob URI where Instagram (export) media files are stored.")

	verbose := flag.Bool("verbose", false, "Enable verbose (debug) logging.")

	flag.Parse()

	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	rdr, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}

	wrtr, err := writer.NewWriter(ctx, *writer_uri)

	if err != nil {
		log.Fatalf("Failed to create writer, %v", err)
	}

	lookup, err := publish.BuildLookup(ctx, *iterator_uri, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to build lookup, %v", err)
	}

	media_bucket, err := blob.OpenBucket(ctx, *media_bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open media bucket, %v", err)
	}

	publish_opts := &publish.PublishOptions{
		Lookup:      lookup,
		Reader:      rdr,
		Writer:      wrtr,
		MediaBucket: media_bucket,
	}

	max_procs := 10
	throttle := make(chan bool, max_procs)

	for range max_procs {
		throttle <- true
	}

	cb := func(ctx context.Context, body []byte) error {

		<-throttle

		defer func() {
			throttle <- true
		}()

		err := publish.PublishMedia(ctx, publish_opts, body)

		if err != nil {
			// slog.Warn("Failed to publish media", "error", err)
			return err
		}

		return nil
	}

	walk_opts := &walk.WalkWithCallbackOptions{
		Callback: cb,
	}

	args := flag.Args()

	for _, media_uri := range args {

		media_fh, err := media.Open(ctx, media_uri)

		if err != nil {
			log.Fatalf("Failed to open %s, %v", media_uri, err)
		}

		defer media_fh.Close()

		err = walk.WalkMediaWithCallback(ctx, walk_opts, media_fh)

		if err != nil {
			log.Fatalf("Failed to walk media for %s, %v", media_uri, err)
		}

		log.Println(media_uri)
	}

}
