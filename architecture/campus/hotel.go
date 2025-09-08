package campus

import (
	"context"
	"io"

	"github.com/whosonfirst/go-reader/v2"
)

// type Hotel is a lightweight data structure to represent garages at SFO with pointers its descendants.
type Hotel struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64        `json:"id"`
	SFOId         string       `json:"sfo:id"`
	PublicArt     []*PublicArt `json:"publicart,omitempty"`
}

func (h *Hotel) Id() int64 {
	return h.WhosOnFirstId
}

func (h *Hotel) AltId() string {
	return h.SFOId
}

func (h *Hotel) Placetype() string {
	return "hotel"
}

func (h *Hotel) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	for _, pa := range h.PublicArt {

		err := cb(ctx, pa)

		if err != nil {
			return nil
		}
	}

	return nil
}

func (h *Hotel) AsTree(ctx context.Context, r reader.Reader, wr io.Writer, indent int) error {
	return elementTree(ctx, h, r, wr, indent)
}
