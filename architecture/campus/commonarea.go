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

// type CommonArea is a lightweight data structure to represent common areas at SFO with pointers its descendants.
type CommonArea struct {
	Element          `json:",omitempty"`
	WhosOnFirstId    int64              `json:"id"`
	SFOId            string             `json:"sfo:id"`
	Gates            []*Gate            `json:"gates,omitempty"`
	Checkpoints      []*Checkpoint      `json:"checkpoints,omitempty"`
	Galleries        []*Gallery         `json:"galleries,omitempty"`
	PublicArt        []*PublicArt       `json:"publicart,omitempty"`
	ObservationDecks []*ObservationDeck `json:"observationdecks,omitempty"` // for example T2
	Museums          []*Museum          `json:"museums,omitempty"`          // for example AML
}

func (c *CommonArea) Id() int64 {
	return c.WhosOnFirstId
}

func (c *CommonArea) AltId() string {
	return c.SFOId
}

func (c *CommonArea) Placetype() string {
	return "commonarea"
}

func (c *CommonArea) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	for _, g := range c.Gates {

		err := cb(ctx, g)

		if err != nil {
			return nil
		}
	}

	for _, cp := range c.Checkpoints {

		err := cb(ctx, cp)

		if err != nil {
			return nil
		}
	}

	for _, g := range c.Galleries {

		err := cb(ctx, g)

		if err != nil {
			return nil
		}
	}

	for _, pa := range c.PublicArt {

		err := cb(ctx, pa)

		if err != nil {
			return nil
		}
	}

	for _, od := range c.ObservationDecks {

		err := cb(ctx, od)

		if err != nil {
			return nil
		}
	}

	for _, m := range c.Museums {

		err := cb(ctx, m)

		if err != nil {
			return nil
		}
	}

	return nil
}

func (ca *CommonArea) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {

	return elementTree(ctx, ca, r, wr, indent)
}

func DeriveCommonAreas(ctx context.Context, db *sql.DB, parent_id int64) ([]*CommonArea, error) {

	slog.Debug("Derive common areas", "parent", parent_id)

	commonarea_ids, err := findChildIDs(ctx, db, parent_id, "commonarea")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (common areas) for %d, %v", parent_id, err)
	}

	commonareas := make([]*CommonArea, 0)

	for _, c_id := range commonarea_ids {

		gates, err := DeriveGates(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive gates for common area %d, %w", c_id, err)
		}

		checkpoints, err := DeriveCheckpoints(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive gates for check points %d, %w", c_id, err)
		}

		galleries, err := DeriveGalleries(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive gates for galleries %d, %w", c_id, err)
		}

		observation_decks, err := DeriveObservationDecks(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive observation decks for galleries %d, %w", c_id, err)
		}

		museums, err := DeriveMuseums(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive museums for common area %d, %w", c_id, err)
		}

		publicart, err := DerivePublicArt(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive public art for common area %d, %w", c_id, err)
		}

		c_body, err := loadFeatureWithDBAndChecks(ctx, db, c_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load feature %d, %w", c_id, err)
		}

		if c_body == nil {
			continue
		}

		var sfoid string

		rsp := gjson.GetBytes(c_body, "properties.sfo:id")

		if rsp.Exists() {

			sfoid = rsp.String()

		} else {

			rsp := gjson.GetBytes(c_body, "properties.sfo:building_id")

			if !rsp.Exists() {
				return nil, fmt.Errorf("Unable to find sfo:building_id for %d", c_id)
			}

			switch rsp.String() {
			case "ITB", "100":
				sfoid = "100CAD" // gis.COMMONAREA_ITB_DEPARTURES
			case "T1", "200":
				sfoid = "200CAD" // gis.COMMONAREA_T1_DEPARTURES
			case "T2", "300":
				sfoid = "300CAD" // gis.COMMONAREA_T2_DEPARTURES
			case "T3", "400":
				sfoid = "400CAD" // gis.COMMONAREA_T3_DEPARTURES
			default:
				return nil, fmt.Errorf("Unrecognized sfo:id '%s' for %d", rsp.String(), c_id)
			}
		}

		name_rsp := gjson.GetBytes(c_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(c_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(c_body, "properties.edtf:cessation")

		slog.Debug("Add common area", "sfo id", sfoid, "id", c_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		area := &CommonArea{
			WhosOnFirstId: c_id,
			SFOId:         sfoid,
		}

		if len(gates) > 0 {
			area.Gates = gates
		}

		if len(checkpoints) > 0 {
			area.Checkpoints = checkpoints
		}

		if len(galleries) > 0 {
			area.Galleries = galleries
		}

		if len(publicart) > 0 {
			area.PublicArt = publicart
		}

		if len(observation_decks) > 0 {
			area.ObservationDecks = observation_decks
		}

		if len(museums) > 0 {
			area.Museums = museums
		}

		commonareas = append(commonareas, area)
	}

	return commonareas, nil
}
