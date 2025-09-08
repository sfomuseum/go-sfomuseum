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

// type ObservationDeck is a lightweight data structure to represent observation decks at SFO with pointers its descendants.
type ObservationDeck struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64        `json:"id"`
	SFOId         string       `json:"sfo:id"`
	PublicArt     []*PublicArt `json:"publicart,omitempty"`
	Galleries     []*Gallery   `json:"galleries,omitempty"`
}

func (od *ObservationDeck) Id() int64 {
	return od.WhosOnFirstId
}

func (od *ObservationDeck) AltId() string {
	return od.SFOId
}

func (od *ObservationDeck) Placetype() string {
	return "observationdeck"
}

func (od *ObservationDeck) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	for _, pa := range od.PublicArt {

		err := cb(ctx, pa)

		if err != nil {
			return nil
		}
	}

	for _, g := range od.Galleries {

		err := cb(ctx, g)

		if err != nil {
			return nil
		}
	}

	return nil
}

func (od *ObservationDeck) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, od, r, wr, indent)
}

func DeriveObservationDecks(ctx context.Context, db *sql.DB, t_id int64) ([]*ObservationDeck, error) {

	slog.Debug("Derive observation decks", "parent id", t_id)

	deck_ids, err := findChildIDs(ctx, db, t_id, "observationdeck")

	if err != nil {
		return nil, fmt.Errorf("Failed to find any child records (observation decks) for %d, %v", t_id, err)
	}

	decks := make([]*ObservationDeck, 0)

	for _, d_id := range deck_ids {

		galleries, err := DeriveGalleries(ctx, db, d_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive galleries for observation deck %d, %w", d_id, err)
		}

		publicart, err := DerivePublicArt(ctx, db, d_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive public art for observation deck %d, %w", d_id, err)
		}

		d_body, err := loadFeatureWithDBAndChecks(ctx, db, d_id)

		if err != nil {
			return nil, fmt.Errorf("Failed to load feature for observation deck %d, %w", d_id, err)
		}

		if d_body == nil {
			continue
		}

		var sfoid string

		rsp := gjson.GetBytes(d_body, "properties.sfo:id")

		if !rsp.Exists() {
			return nil, fmt.Errorf("Unable to find sfo:id for WOF record, %d", d_id)
		}

		sfoid = rsp.String()

		name_rsp := gjson.GetBytes(d_body, "properties.wof:name")
		inception_rsp := gjson.GetBytes(d_body, "properties.edtf:inception")
		cessation_rsp := gjson.GetBytes(d_body, "properties.edtf:cessation")

		slog.Debug("Add observation deck", "sfo id", sfoid, "parent id", t_id, "id", d_id, "name", name_rsp.String(), "inception", inception_rsp.String(), "cessation", cessation_rsp.String())

		deck := &ObservationDeck{
			WhosOnFirstId: d_id,
			SFOId:         sfoid,
		}

		if len(galleries) > 0 {
			deck.Galleries = galleries
		}

		if len(publicart) > 0 {
			deck.PublicArt = publicart
		}

		decks = append(decks, deck)
	}

	return decks, nil
}
