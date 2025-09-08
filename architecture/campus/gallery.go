package campus

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"

	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader/v2"
)

// type Gallery is a lightweight data structure to represent SFO Museum galleries at SFO.
type Gallery struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64  `json:"id"`
	SFOId         string `json:"sfomuseum:id"`
}

func (g *Gallery) Id() int64 {
	return g.WhosOnFirstId
}

func (g *Gallery) AltId() string {
	return g.SFOId
}

func (g *Gallery) Placetype() string {
	return "gallery"
}

func (g *Gallery) Walk(ctx context.Context, cb ElementCallbackFunc) error {
	return nil
}

func (g *Gallery) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {

	return elementTree(ctx, g, r, wr, indent)
}

func DeriveGalleries(ctx context.Context, db *sql.DB, parent_id int64) ([]*Gallery, error) {

	slog.Debug("Derive galleries", "parent id", parent_id)

	gallery_ids, err := findChildIDs(ctx, db, parent_id, "gallery")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (galleries) for %d, %w", parent_id, err)
	}

	galleries := make([]*Gallery, 0)

	for _, g_id := range gallery_ids {

		g_body, err := loadFeatureWithDBAndChecks(ctx, db, g_id)

		if err != nil {
			return nil, fmt.Errorf("Failed load feature for gallery %d, %w", g_id, err)
		}

		if g_body == nil {
			continue
		}

		map_rsp := gjson.GetBytes(g_body, "properties.sfomuseum:map_id")
		gallery_rsp := gjson.GetBytes(g_body, "properties.sfomuseum:gallery_id")

		if !gallery_rsp.Exists() {
			return nil, fmt.Errorf("Missing sfomuseum:gallery_id property for gallery %d", g_id)
		}

		sfom_id := fmt.Sprintf("%s#%d", map_rsp.String(), gallery_rsp.Int())

		name_rsp := gjson.GetBytes(g_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(g_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(g_body, "properties.edtf:cessation")

		slog.Debug("Add gallery", "sfo id", sfom_id, "parent id", parent_id, "id", g_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		g := &Gallery{
			WhosOnFirstId: g_id,
			SFOId:         sfom_id,
		}

		galleries = append(galleries, g)
	}

	return galleries, nil
}
