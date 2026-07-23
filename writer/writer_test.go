package writer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
	go_writer "github.com/whosonfirst/go-writer/v3"
)

func TestWriteFeature(t *testing.T) {

	ctx := context.Background()

	body, err := read_feature(ctx)

	if err != nil {
		t.Fatalf("Failed to read feature, %v", err)
	}

	f, err := geojson.UnmarshalFeature(body)

	if err != nil {
		t.Fatalf("Failed to unmarshal feature, %v", err)
	}

	var buf bytes.Buffer
	buf_wr := bufio.NewWriter(&buf)

	wr, err := go_writer.NewWriter(ctx, "io://")

	if err != nil {
		t.Fatalf("Failed to create new writer, %v", err)
	}

	ctx, err = go_writer.SetIOWriterWithContext(ctx, buf_wr)

	if err != nil {
		t.Fatalf("Failed to set IO writer context, %v", err)
	}

	_, err = WriteFeature(ctx, wr, f)

	if err != nil {
		t.Fatalf("Failed to write feature, %v", err)
	}

	buf_wr.Flush()

	id, err := properties.Id(buf.Bytes())

	if err != nil {
		t.Fatalf("Failed to derive ID, %v", err)
	}

	if id != 1159160649 {
		t.Fatalf("Unexpected ID returned: %d", id)
	}
}

func TestWriteBytes(t *testing.T) {

	ctx := context.Background()

	body, err := read_feature(ctx)

	if err != nil {
		t.Fatalf("Failed to read feature, %v", err)
	}

	wr, err := go_writer.NewWriter(ctx, "null://")

	if err != nil {
		t.Fatalf("Failed to create new writer, %v", err)
	}

	id, err := WriteBytes(ctx, wr, body)

	if err != nil {
		t.Fatalf("Failed to write feature, %v", err)
	}

	if id != 1159160649 {
		t.Fatalf("Unexpected ID returned: %d", id)
	}

}

func read_feature(ctx context.Context) ([]byte, error) {

	cwd, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("Failed to determine current working directory, %v", err)
	}

	fixtures := filepath.Join(cwd, "fixtures")
	feature_path := filepath.Join(fixtures, "1159160649.geojson")

	fh, err := os.Open(feature_path)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s, %v", feature_path, err)
	}

	defer fh.Close()

	body, err := io.ReadAll(fh)

	if err != nil {
		return nil, fmt.Errorf("Failed to read %s, %v", feature_path, err)
	}

	return body, nil
}
