package export

import (
	"context"

	wof_export "github.com/whosonfirst/go-whosonfirst/v4/export"
)

type SFOMuseumExporter struct {
	wof_export.Exporter
}

func init() {

	ctx := context.Background()

	err := wof_export.RegisterExporter(ctx, "sfomuseum", NewSFOMuseumExporter)

	if err != nil {
		panic(err)
	}
}

func NewSFOMuseumExporter(ctx context.Context, uri string) (wof_export.Exporter, error) {
	ex := &SFOMuseumExporter{}
	return ex, nil
}

func (ex *SFOMuseumExporter) Export(ctx context.Context, feature []byte) (bool, []byte, error) {

	feature, err := PrepareFeature(ctx, feature)

	if err != nil {
		return false, nil, err
	}

	return wof_export.Export(ctx, feature)
}
