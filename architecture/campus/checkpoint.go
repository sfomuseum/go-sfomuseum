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

// type Checkpoint is a lightweight data structure to represent security checkpoints at SFO.
type Checkpoint struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64  `json:"id"`
	SFOId         string `json:"sfo:id"`
}

func (c *Checkpoint) Id() int64 {
	return c.WhosOnFirstId
}

func (c *Checkpoint) AltId() string {
	return c.SFOId
}

func (c *Checkpoint) Placetype() string {
	return "checkpoint"
}

func (c *Checkpoint) Walk(ctx context.Context, cb ElementCallbackFunc) error {
	return nil
}

func (cp *Checkpoint) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {

	return elementTree(ctx, cp, r, wr, indent)
}

func DeriveCheckpoints(ctx context.Context, db *sql.DB, parent_id int64) ([]*Checkpoint, error) {

	slog.Debug("Derive check points", "parent id", parent_id)

	checkpoint_ids, err := findChildIDs(ctx, db, parent_id, "checkpoint")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (checkpoints) for %d, %w", parent_id, err)
	}

	checkpoints := make([]*Checkpoint, 0)

	for _, cp_id := range checkpoint_ids {

		cp_body, err := loadFeatureWithDBAndChecks(ctx, db, cp_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load feature for %d, %w", cp_id, err)
		}

		if cp_body == nil {
			continue
		}

		var sfoid string

		rsp := gjson.GetBytes(cp_body, "properties.sfo:id")

		if !rsp.Exists() {
			return nil, fmt.Errorf("Missing sfo:id for %d", cp_id)
		}

		sfoid = rsp.String()

		name_rsp := gjson.GetBytes(cp_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(cp_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(cp_body, "properties.edtf:cessation")

		slog.Debug("Add checkpoint", "sfo id", sfoid, "parent id", parent_id, "id", cp_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		cp := &Checkpoint{
			WhosOnFirstId: cp_id,
			SFOId:         sfoid,
		}

		checkpoints = append(checkpoints, cp)
	}

	return checkpoints, nil
}
