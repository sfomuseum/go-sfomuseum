package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sfomuseum/go-sfomuseum/architecture/campus"
	"github.com/whosonfirst/go-reader/v2"
)

func main() {

	var architecture_reader_uri string
	var publicart_reader_uri string

	var iterator_uri string
	var output_mode string
	var verbose bool
	var complex_id int64
	var dsn string

	flag.StringVar(&iterator_uri, "iterator-uri", "repo://", "...")
	flag.StringVar(&output_mode, "output-mode", "json", "...")
	flag.StringVar(&architecture_reader_uri, "architecture-reader-uri", "repo:///usr/local/data/sfomuseum-data-architecture", "...")
	flag.StringVar(&publicart_reader_uri, "publicart-reader-uri", "repo:///usr/local/data/sfomuseum-data-publicart", "...")
	flag.BoolVar(&verbose, "verbose", false, "...")
	flag.Int64Var(&complex_id, "complex-id", 0, "If 0 then the most recent (current) complex ID will be used.")
	flag.StringVar(&dsn, "dsn", ":memory:", "...")

	flag.Parse()

	ctx := context.Background()

	if verbose {
		// set slog.LogLevel here
		// why is this so hard?
	}

	paths := flag.Args()

	db, err := campus.NewDatabaseWithIterator(ctx, dsn, iterator_uri, paths...)

	if err != nil {
		log.Fatalf("Failed to create database, %v", err)
	}

	if complex_id != 0 {
		campus.WARN_IS_CURRENT = false
	}

	c, err := campus.DeriveComplex(ctx, db, complex_id)

	if err != nil {
		log.Fatalf("Failed to derive complex, %v", err)
	}

	writers := make([]io.Writer, 0)
	writers = append(writers, os.Stdout)
	wr := io.MultiWriter(writers...)

	mk_reader := func(ctx context.Context) (reader.Reader, error) {

		architecture_r, err := reader.NewReader(ctx, architecture_reader_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new architecture reader, %w", err)
		}

		publicart_r, err := reader.NewReader(ctx, publicart_reader_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new public art reader, %w", err)
		}

		multi_r, err := reader.NewMultiReader(ctx, architecture_r, publicart_r)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new multi reader, %w", err)
		}

		return multi_r, nil
	}

	switch output_mode {
	case "json":

		err := c.AsJSON(ctx, wr)

		if err != nil {
			log.Fatalf("Failed to encode complex, %v", err)
		}

	case "tree":

		r, err := mk_reader(ctx)

		if err != nil {
			log.Fatalf("Failed to create reader, %v", err)
		}

		err = c.AsTree(ctx, r, wr, 0)

		if err != nil {
			log.Fatalf("Failed to render complex as tree, %v", err)
		}

	default:
		log.Fatalf("Invalid or unsupported output mode, %v", err)
	}
}
