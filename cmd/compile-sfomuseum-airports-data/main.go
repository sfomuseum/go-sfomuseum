package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	_ "github.com/whosonfirst/go-whosonfirst/v4/iterate/git"

	"github.com/sfomuseum/go-sfomuseum/airfield/airports/sfomuseum"
)

func main() {

	default_target := fmt.Sprintf("data/%s", sfomuseum.DATA_JSON)

	iterator_uri := flag.String("iterator-uri", "repo://?include=properties.sfomuseum:placetype=airport&exclude=properties.edtf:deprecated=.*", "...")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-whosonfirst", "...")

	target := flag.String("target", default_target, "The path to write SFO Museum airport data.")
	stdout := flag.Bool("stdout", false, "Emit SFO Museum airport data to SDOUT.")

	verbose := flag.Bool("verbose", false, "Enable verbose (debug) logging")
	flag.Parse()

	ctx := context.Background()

	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	writers := make([]io.Writer, 0)

	fh, err := os.OpenFile(*target, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalf("Failed to open '%s', %v", *target, err)
	}

	writers = append(writers, fh)

	if *stdout {
		writers = append(writers, os.Stdout)
	}

	wr := io.MultiWriter(writers...)

	slog.Debug("Compile airport data", "iterator uri", *iterator_uri, "iterator source", *iterator_source)
	lookup, err := sfomuseum.CompileAirportsData(ctx, *iterator_uri, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to compile data, %v", err)
	}

	enc := json.NewEncoder(wr)
	err = enc.Encode(lookup)

	if err != nil {
		log.Fatalf("Failed to marshal results, %v", err)
	}

}
