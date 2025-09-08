// package campus provides methods for working with the SFO airport campus.
package campus

import (
	"context"
)

// type Campus is a lightweight data structure to represent the SFO campus with pointers its descendants.
type Campus struct {
	Element       `json:",omitempty"`
	WhosOnFirstId int64        `json:"id"`
	SFOId         string       `json:"sfo:id"`
	Complex       *Complex     `json:"complex"`
	Garages       []*Garage    `json:"garages"`
	Hotels        []*Hotel     `json:"hotels"`
	PublicArt     []*PublicArt `json:"buildings,omitempty"`
}

func (c *Campus) Id() int64 {
	return c.WhosOnFirstId
}

func (c *Campus) AltId() string {
	return c.SFOId
}

func (c *Campus) Placetype() string {
	return "campus"
}

func (c *Campus) Walk(ctx context.Context, cb ElementCallbackFunc) error {

	err := cb(ctx, c.Complex)

	if err != nil {
		return err
	}

	for _, g := range c.Garages {

		err := cb(ctx, g)

		if err != nil {
			return err
		}
	}

	for _, h := range c.Hotels {

		err := cb(ctx, h)

		if err != nil {
			return err
		}
	}

	for _, pa := range c.PublicArt {

		err := cb(ctx, pa)

		if err != nil {
			return err
		}
	}

	return nil
}
