package update

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/whosonfirst/go-whosonfirst/v4/export"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial/database"
	"github.com/whosonfirst/go-whosonfirst/v4/spatial/hierarchy"
	"github.com/whosonfirst/go-writer/v3"
)

func Run(ctx context.Context) error {

	fs, err := DefaultFlagSet(ctx)

	if err != nil {
		return fmt.Errorf("Failed to create application flag set, %w", err)
	}

	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	opts, err := RunOptionsFromFlagSet(ctx, fs)

	if err != nil {
		return fmt.Errorf("Failed to derive run options, %w", err)
	}

	return RunWithOptions(ctx, opts)
}

func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	if opts.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	// Note that the bulk of this method is simply taking opts and using it to
	// instantiate all the different pieces necessary for the updateApplication
	// type to actually do the work of updating hierarchies.

	var ex export.Exporter
	var wr writer.Writer
	var spatial_db database.SpatialDatabase

	if opts.Exporter != nil {
		ex = opts.Exporter
	} else {

		_ex, err := export.NewExporter(ctx, opts.ExporterURI)

		if err != nil {
			return fmt.Errorf("Failed to create exporter for '%s', %v", opts.ExporterURI, err)
		}

		ex = _ex
	}

	if opts.Writer != nil {
		wr = opts.Writer
	} else {
		_wr, err := writer.NewWriter(ctx, opts.WriterURI)

		if err != nil {
			return fmt.Errorf("Failed to create writer for '%s', %v", opts.WriterURI, err)
		}

		wr = _wr
	}

	if opts.SpatialDatabase != nil {
		spatial_db = opts.SpatialDatabase
	} else {

		_db, err := database.NewSpatialDatabase(ctx, opts.SpatialDatabaseURI)

		if err != nil {
			return fmt.Errorf("Failed to create spatial database for '%s', %v", opts.SpatialDatabaseURI, err)
		}

		spatial_db = _db
	}

	update_cb := opts.PIPUpdateFunc

	if update_cb == nil {
		update_cb = hierarchy.DefaultPointInPolygonHierarchyResolverUpdateCallback()
	}

	resolver_opts := &hierarchy.PointInPolygonHierarchyResolverOptions{
		Database: spatial_db,
	}

	resolver, err := hierarchy.NewPointInPolygonHierarchyResolver(ctx, resolver_opts)

	if err != nil {
		return fmt.Errorf("Failed to create PIP tool, %v", err)
	}

	app := &updateApplication{
		spatial_db:          spatial_db,
		resolver:            resolver,
		exporter:            ex,
		writer:              wr,
		sprFilterInputs:     opts.SPRFilterInputs,
		sprResultsFunc:      opts.SPRResultsFunc,
		hierarchyUpdateFunc: update_cb,
	}

	return app.Run(ctx, opts.SourceIteratorSources, opts.TargetIteratorSources)
}
