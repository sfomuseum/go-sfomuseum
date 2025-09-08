package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sfomuseum/go-sfomuseum/architecture/gates"
)

func main() {

	default_target := fmt.Sprintf("data/%s", gates.DATA_JSON)

	iterator_uri := flag.String("iterator-uri", "repo://?include=properties.sfomuseum:placetype=gate&exclude=properties.edtf:deprecated=.*", "A valid whosonfirst/go-whosonfirst-iterate URI")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-architecture", "The URI containing documents to iterate.")

	target := flag.String("target", default_target, "The path to write SFO Museum gates data.")
	stdout := flag.Bool("stdout", false, "Emit SFO Museum gates data to SDOUT.")

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

	lookup, err := gates.CompileGatesData(ctx, *iterator_uri, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to compile gates data, %v", err)
	}

	enc := json.NewEncoder(wr)
	err = enc.Encode(lookup)

	if err != nil {
		log.Fatalf("Failed to marshal results, %v", err)
	}
}
