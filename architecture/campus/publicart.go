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

// type PublicArt is a lightweight data structure to represent public art works at SFO.
type PublicArt struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64  `json:"id"`
	SFOId         string `json:"sfomuseum:id"`
}

func (pa *PublicArt) Id() int64 {
	return pa.WhosOnFirstId
}

func (pa *PublicArt) AltId() string {
	return pa.SFOId
}

func (pa *PublicArt) Placetype() string {
	return "publicart"
}

func (pa *PublicArt) Walk(ctx context.Context, cb ElementCallbackFunc) error {
	return nil
}

func (pa *PublicArt) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, pa, r, wr, indent)
}

func DerivePublicArt(ctx context.Context, db *sql.DB, parent_id int64) ([]*PublicArt, error) {

	slog.Debug("Derive public art", "parent id", parent_id)

	publicart_ids, err := findChildIDs(ctx, db, parent_id, "publicart")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (public art) for %d, %w", parent_id, err)
	}

	publicarts := make([]*PublicArt, 0)

	for _, p_id := range publicart_ids {

		p_body, err := loadFeatureWithDBAndChecks(ctx, db, p_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load feature for public art %d, %w", p_id, err)
		}

		if p_body == nil {
			continue
		}

		map_rsp := gjson.GetBytes(p_body, "properties.sfomuseum:map_id")
		object_rsp := gjson.GetBytes(p_body, "properties.sfomuseum:object_id")

		if !object_rsp.Exists() {
			return nil, fmt.Errorf("Missing sfomuseum:object_id property for public art %d, %w", p_id, err)
		}

		sfom_id := fmt.Sprintf("%s#%d", map_rsp.String(), object_rsp.Int())

		name_rsp := gjson.GetBytes(p_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(p_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(p_body, "properties.edtf:cessation")

		slog.Debug("Add public art", "sfo id", sfom_id, "parent id", parent_id, "id", p_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		pa := &PublicArt{
			WhosOnFirstId: p_id,
			SFOId:         sfom_id,
		}

		publicarts = append(publicarts, pa)
	}

	return publicarts, nil
}
