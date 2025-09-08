package campus

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/paulmach/orb/geojson"
	"github.com/whosonfirst/go-reader/v2"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader/v2"
)

// SFO Terminal Complex (1954~ to 1963~)
// https://millsfield.sfomuseum.org/buildings/1159396329/
const FIRST_SFO_COMPLEX int64 = 1159396329

// type Complex is a lightweight data structure to represent the terminal complex at SFO with pointers its descendants.
type Complex struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64       `json:"id"`
	SFOId         string      `json:"sfo:id"`
	Terminals     []*Terminal `json:"terminals"`
}

func (c *Complex) Id() int64 {
	return c.WhosOnFirstId
}

func (c *Complex) AltId() string {
	return c.SFOId
}

func (c *Complex) Placetype() string {
	return "complex"
}

func (c *Complex) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	for _, t := range c.Terminals {

		err := cb(ctx, t)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Complex) AsJSON(ctx context.Context, wr io.Writer) error {

	enc := json.NewEncoder(wr)
	return enc.Encode(c)
}

func (c *Complex) AsGeoJSONLayers(ctx context.Context, r reader.Reader) (map[string]*geojson.FeatureCollection, error) {

	layers_map := make(map[string]*geojson.FeatureCollection)

	var cb func(ctx context.Context, el Element) error

	cb = func(ctx context.Context, el Element) error {

		id := el.Id()
		pt := el.Placetype()

		body, err := wof_reader.LoadBytes(ctx, r, id)

		if err != nil {
			return fmt.Errorf("Failed to load record '%d', %w", id, err)
		}

		f, err := geojson.UnmarshalFeature(body)

		if err != nil {
			return fmt.Errorf("Failed to unmarshal feature '%d', %w", id, err)
		}

		fc, exists := layers_map[pt]

		if !exists {
			fc = geojson.NewFeatureCollection()
		}

		fc.Append(f)
		layers_map[pt] = fc

		el.Walk(ctx, cb)
		return nil
	}

	err := c.Walk(ctx, cb)

	if err != nil {
		return nil, err
	}

	return layers_map, nil
}

func (c *Complex) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, c, r, wr, indent)
}

func (c *Complex) DeriveAltLookup(ctx context.Context) (map[string]int64, error) {

	lookup_map := make(map[string]int64)

	var cb func(ctx context.Context, el Element) error

	cb = func(ctx context.Context, el Element) error {

		alt := el.AltId()
		id := el.Id()

		other_id, exists := lookup_map[alt]

		if exists {
			slog.Warn("Duplicate key", "placetype", el.Placetype(), "alt", alt, "id", id, "other", other_id)
		} else {
			lookup_map[alt] = id
			// slog.Info("Current", "alt", alt, "id", id)
		}

		el.Walk(ctx, cb)
		return nil
	}

	err := c.Walk(ctx, cb)

	if err != nil {
		return nil, err
	}

	return lookup_map, nil
}

func DeriveComplex(ctx context.Context, db *sql.DB, complex_id int64) (*Complex, error) {

	if complex_id == 0 {

		id, err := findMostRecentComplexID(ctx, db, FIRST_SFO_COMPLEX)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive most recent complex ID, %w", err)
		}

		complex_id = id
	}

	terminals, err := DeriveTerminals(ctx, db, complex_id)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive terminals for complex %d, %w", complex_id, err)
	}

	c := &Complex{
		WhosOnFirstId: complex_id,
		SFOId:         "SFO",
		Terminals:     terminals,
	}

	return c, nil
}

func findMostRecentComplexID(ctx context.Context, db *sql.DB, id int64) (int64, error) {

	q := "SELECT DISTINCT(superseded_by_id) FROM supersedes WHERE superseded_id = ?"

	rows, err := db.QueryContext(ctx, q, id)

	if err != nil {

		if err == sql.ErrNoRows {
			return id, nil
		}

		return -1, err
	}

	defer rows.Close()
	possible := make([]int64, 0)

	for rows.Next() {

		var superseded_by int64
		err := rows.Scan(&superseded_by)

		if err != nil {
			return -1, fmt.Errorf("Failed to scan row, %w", err)
		}

		possible = append(possible, superseded_by)
	}

	err = rows.Close()

	if err != nil {
		return -1, err
	}

	err = rows.Err()

	if err != nil {
		return -1, err
	}

	switch len(possible) {
	case 0:
		return id, nil
	case 1:
		return findMostRecentComplexID(ctx, db, possible[0])
	default:
		return -1, fmt.Errorf("Multiple results for '%d', not implemented", id)
	}
}
