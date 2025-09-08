package campus

import (
	"context"
	"io"

	"github.com/whosonfirst/go-reader/v2"
)

// type Garage is a lightweight data structure to represent garages at SFO with pointers its descendants.
type Garage struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64        `json:"id"`
	SFOId         string       `json:"sfo:id"`
	PublicArt     []*PublicArt `json:"publicart,omitempty"`
}

func (g *Garage) Id() int64 {
	return g.WhosOnFirstId
}

func (g *Garage) AltId() string {
	return g.SFOId
}

func (g *Garage) Placetype() string {
	return "garage"
}

func (g *Garage) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	for _, pa := range g.PublicArt {

		err := cb(ctx, pa)

		if err != nil {
			return nil
		}
	}

	return nil
}

func (g *Garage) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, g, r, wr, indent)
}
