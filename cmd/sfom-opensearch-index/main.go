// sfom-opensearch-index bulk indexes one or more whosonfirst/go-whosonfirst-iterate/v2 sources in an OpenSearch database.
package main

import (
	"context"
	"log"
	"log/slog"

	_ "github.com/whosonfirst/go-whosonfirst/v4/iterate/git"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/sfomuseum/go-sfomuseum/database/opensearch/document"
	iterwriter_app "github.com/whosonfirst/go-whosonfirst/v4/app/iterate/writer"
	os_document "github.com/whosonfirst/go-whosonfirst/v4/database/opensearch/document"
	os_writer "github.com/whosonfirst/go-whosonfirst/v4/database/opensearch/writer"
	"github.com/whosonfirst/go-writer/v3"
)

func main() {

	var sfom_writer_uri string
	var index_embeddings bool

	fs := iterwriter_app.DefaultFlagSet()

	fs.StringVar(&sfom_writer_uri, "sfomuseum-writer-uri", "", "...")
	fs.BoolVar(&index_embeddings, "index-embeddings", false, "...")

	flagset.Parse(fs)

	verbose, err := lookup.BoolVar(fs, "verbose")

	if err != nil {
		log.Fatal(err)
	}

	logger := slog.Default()
	log_level := slog.LevelInfo

	if verbose {
		log_level = slog.LevelDebug
		slog.SetLogLoggerLevel(log_level)
		slog.Debug("Verbose logging enabled")
	}

	log_logger := slog.NewLogLogger(logger.Handler(), log_level)

	ctx := context.Background()

	// logger := log.Default()

	wr, err := writer.NewWriter(ctx, sfom_writer_uri)

	if err != nil {
		log.Fatalf("Failed to create new writer, %v", err)
	}

	err = wr.SetLogger(ctx, log_logger)

	if err != nil {
		log.Fatalf("Failed to assign logger to writer, %v", err)
	}

	// To do: Some day we may have multiple prepare document funcs

	var sfom_prepare_func os_document.PrepareDocumentFunc

	if index_embeddings {
		sfom_prepare_func = document.SFOMuseumPrepareEmbeddingsDocumentFunc()
	} else {
		sfom_prepare_func = document.SFOMuseumPrepareDocumentFunc()
	}

	// To do: type/interface checking here...

	err = wr.(os_writer.DocumentWriter).AppendPrepareFunc(ctx, sfom_prepare_func)

	if err != nil {
		log.Fatalf("Failed to append SFOM prepare func, %v", err)
	}

	run_opts, err := iterwriter_app.DefaultOptionsFromFlagSet(fs, true)

	if err != nil {
		log.Fatalf("Failed to create run options, %v", err)
	}

	run_opts.Writer = wr

	err = iterwriter_app.RunWithOptions(ctx, run_opts)

	if err != nil {
		log.Fatalf("Failed to run iterwriter, %v", err)
	}

}
