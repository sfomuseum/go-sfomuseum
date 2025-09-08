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

// type Terminal is a lightweight data structure to represent terminals at SFO with pointers its descendants.
type Terminal struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64           `json:"id"`
	SFOId         string          `json:"sfo:id"`
	CommonAreas   []*CommonArea   `json:"commonareas,omitempty"`
	BoardingAreas []*BoardingArea `json:"boardingareas,omitempty"`
}

func (t *Terminal) Id() int64 {
	return t.WhosOnFirstId
}

func (t *Terminal) AltId() string {
	return t.SFOId
}

func (t *Terminal) Placetype() string {
	return "terminal"
}

func (t *Terminal) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	for _, ba := range t.BoardingAreas {

		err := cb(ctx, ba)

		if err != nil {
			return err
		}
	}

	for _, ca := range t.CommonAreas {

		err := cb(ctx, ca)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Terminal) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, t, r, wr, indent)
}

func DeriveTerminals(ctx context.Context, db *sql.DB, sfo_id int64) ([]*Terminal, error) {

	slog.Debug("Derive terminals", "parent id", sfo_id)

	terminal_ids, err := findChildIDs(ctx, db, sfo_id, "terminal")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (terminals) for %d, %v", sfo_id, err)
	}

	terminals := make([]*Terminal, 0)

	for _, t_id := range terminal_ids {

		commonareas, err := DeriveCommonAreas(ctx, db, t_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive common areas for %d, %w", t_id, err)
		}

		boardingareas, err := DeriveBoardingAreas(ctx, db, t_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive boarding areas for %d, %w", t_id, err)
		}

		t_body, err := loadFeatureWithDBAndChecks(ctx, db, t_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load feature for %d, %w", t_id, err)
		}

		if t_body == nil {
			continue
		}

		var sfoid string

		rsp := gjson.GetBytes(t_body, "properties.sfo:id")

		if rsp.Exists() {

			sfoid = rsp.String()

		} else {

			rsp := gjson.GetBytes(t_body, "properties.sfomuseum:terminal_id")

			if !rsp.Exists() {
				return nil, fmt.Errorf("Missing properties.sfomuseum:terminal_id property for terminal %d", t_id)
			}

			switch rsp.String() {
			case "ITB":
				sfoid = "100" // gis.BUILDING_ITB
			case "T1":
				sfoid = "200" // gis.BUILDING_T1
			case "T2":
				sfoid = "300" // gis.BUILDING_T2
			case "T3":
				sfoid = "400" // gis.BUILDING_T3
			default:
				return nil, fmt.Errorf("Unrecognized terminal_id '%s' for %d", rsp.String(), t_id)
			}
		}

		name_rsp := gjson.GetBytes(t_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(t_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(t_body, "properties.edtf:cessation")

		slog.Debug("Add terminal", "sfo id", sfoid, "id", t_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		terminal := &Terminal{
			WhosOnFirstId: t_id,
			SFOId:         sfoid,
		}

		if len(commonareas) > 0 {
			terminal.CommonAreas = commonareas
		}

		if len(boardingareas) > 0 {
			terminal.BoardingAreas = boardingareas
		}

		terminals = append(terminals, terminal)
	}

	return terminals, nil
}
