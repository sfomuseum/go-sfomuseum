package campus

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/whosonfirst/go-reader/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader/v2"
)

func treeLabel(ctx context.Context, r reader.Reader, el Element) string {

	id := el.Id()
	alt := el.AltId()
	pt := el.Placetype()

	return fmt.Sprintf("[%s] %d#%s %s", pt, id, alt, name(ctx, r, id))
}

func name(ctx context.Context, r reader.Reader, id int64) string {

	body, err := wof_reader.LoadBytes(ctx, r, id)

	if err != nil {
		slog.Warn("Failed to read bytes for ID", "id", id, "error", err)
		return ""
	}

	name, err := properties.Name(body)

	if err != nil {
		slog.Warn("Failed to read name", "id", id, "error", err)
		return ""
	}

	return name
}
