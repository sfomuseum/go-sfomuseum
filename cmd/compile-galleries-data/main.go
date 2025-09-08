package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sfomuseum/go-sfomuseum/architecture/galleries"
)

func main() {

	default_target := fmt.Sprintf("data/%s", galleries.DATA_JSON)

	iterator_uri := flag.String("iterator-uri", "repo://?include=properties.sfomuseum:placetype=gallery&exclude=properties.edtf:deprecated=.*", "A valid whosonfirst/go-whosonfirst-iterate URI")
	iterator_source := flag.String("iterator-source", "/usr/local/data/sfomuseum-data-architecture", "The URI containing documents to iterate.")

	target := flag.String("target", default_target, "The path to write SFO Museum galleries data.")
	stdout := flag.Bool("stdout", false, "Emit SFO Museum galleries data to SDOUT.")

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

	lookup, err := galleries.CompileGalleriesData(ctx, *iterator_uri, *iterator_source)

	if err != nil {
		log.Fatalf("Failed to compile galleries data, %v", err)
	}

	enc := json.NewEncoder(wr)
	err = enc.Encode(lookup)

	if err != nil {
		log.Fatalf("Failed to marshal results, %v", err)
	}
}
