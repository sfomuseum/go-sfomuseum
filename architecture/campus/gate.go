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

// type Gate is a lightweight data structure to represent passenger gates at SFO.
type Gate struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64  `json:"id"`
	SFOId         string `json:"sfo:id"`
}

func (g *Gate) Id() int64 {
	return g.WhosOnFirstId
}

func (g *Gate) AltId() string {
	return g.SFOId
}

func (g *Gate) Placetype() string {
	return "gate"
}

func (g *Gate) Walk(ctx context.Context, cb ElementCallbackFunc) error {
	return nil
}

func (g *Gate) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, g, r, wr, indent)
}

func DeriveGates(ctx context.Context, db *sql.DB, parent_id int64) ([]*Gate, error) {

	slog.Debug("Derive gates", "parent", parent_id)

	gate_ids, err := findChildIDs(ctx, db, parent_id, "gate")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (gates) for %d, %w", parent_id, err)
	}

	gates := make([]*Gate, 0)

	for _, g_id := range gate_ids {

		g_body, err := loadFeatureWithDBAndChecks(ctx, db, g_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load feature for date %d, %w", g_id, err)
		}

		if g_body == nil {
			continue
		}

		var sfoid string

		rsp := gjson.GetBytes(g_body, "properties.sfo:id")

		if rsp.Exists() {

			sfoid = rsp.String()

		} else {

			rsp := gjson.GetBytes(g_body, "properties.wof:name")

			if !rsp.Exists() {
				return nil, fmt.Errorf("Missing wof:name for %d", g_id)
			}

			sfoid = rsp.String()
		}

		name_rsp := gjson.GetBytes(g_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(g_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(g_body, "properties.edtf:cessation")

		slog.Debug("Add gate", "sfo id", sfoid, "parent_id", parent_id, "id", g_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		g := &Gate{
			WhosOnFirstId: g_id,
			SFOId:         sfoid,
		}

		gates = append(gates, g)
	}

	return gates, nil
}
