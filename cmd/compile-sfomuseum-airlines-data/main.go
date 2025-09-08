package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v3"

	"github.com/sfomuseum/go-sfomuseum/airfield/airlines/sfomuseum"
)

func main() {

	default_target := fmt.Sprintf("data/%s", sfomuseum.DATA_JSON)

	iterator_uri := flag.String("iterator-uri", "repo://?exclude=properties.edtf:deprecated=.*", "...")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-enterprise", "...")

	target := flag.String("target", default_target, "The path to write SFO Museum airline data.")
	stdout := flag.Bool("stdout", false, "Emit SFO Museum aircraft data to SDOUT.")

	flag.Parse()

	ctx := context.Background()

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

	lookup, err := sfomuseum.CompileAirlinesData(ctx, *iterator_uri, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to compile data, %v", err)
	}

	enc := json.NewEncoder(wr)
	err = enc.Encode(lookup)

	if err != nil {
		log.Fatalf("Failed to marshal results, %v", err)
	}

}
