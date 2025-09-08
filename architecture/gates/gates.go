// package gates provides methods for working with boarding gates at SFO.
package gates

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfomuseum/go-edtf/cmp"
	"github.com/sfomuseum/go-sfomuseum/architecture"
)

// To do: Make these sortable by inception/cessation
type Gates []*Gate

// type Gate is a struct representing a passenger gate at SFO.
type Gate struct {
	// The Who's On First ID associated with this gate.
	WhosOnFirstId int64 `json:"wof:id"`
	// The name of this gate.
	Name string `json:"wof:name"`
	// A Who's On First "existential" (`KnownUnknownFlag`) flag signaling the gate's status
	IsCurrent int64 `json:"mz:is_current"`
	// The (EDTF) inception date for the gallery
	Inception string `json:"edtf:inception"`
	// The (EDTF) cessation date for the gallery
	Cessation string `json:"edtf:cessation"`
}

// String() will return the name of the gate.
func (g *Gate) String() string {
	return fmt.Sprintf("%d %s %s-%s (%d)", g.WhosOnFirstId, g.Name, g.Inception, g.Cessation, g.IsCurrent)
}

// Return the Gate matching 'code' that was active for 'date'. Multiple matches throw an error.
func FindGateForDate(ctx context.Context, code string, date string) (*Gate, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindGateForDateWithLookup(ctx, lookup, code, date)
}

// Return all the Gates matching 'code' that were active for 'date'.
func FindAllGatesForDate(ctx context.Context, code string, date string) ([]*Gate, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindAllGatesForDateWithLookup(ctx, lookup, code, date)
}

// Return the current Gate matching 'code'. Multiple matches throw an error.
func FindCurrentGate(ctx context.Context, code string) (*Gate, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindCurrentGateWithLookup(ctx, lookup, code)
}

// Return the current Gate matching 'code' with a custom architecture.Lookup instance. Multiple matches throw an error.
func FindCurrentGateWithLookup(ctx context.Context, lookup architecture.Lookup, code string) (*Gate, error) {

	current, err := FindGatesCurrentWithLookup(ctx, lookup, code)

	if err != nil {
		return nil, err
	}

	switch len(current) {
	case 0:
		return nil, NotFound{code}
	case 1:
		return current[0], nil
	default:
		return nil, MultipleCandidates{code}
	}

}

// Returns all Gate instances matching 'code' that are marked as current.
func FindGatesCurrent(ctx context.Context, code string) ([]*Gate, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindGatesCurrentWithLookup(ctx, lookup, code)
}

// Returns all Gate instances matching 'code' that are marked as current with a custom architecture.Lookup instance.
func FindGatesCurrentWithLookup(ctx context.Context, lookup architecture.Lookup, code string) ([]*Gate, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to find gate '%s', %w", code, err)
	}

	current := make([]*Gate, 0)

	for _, r := range rsp {

		g := r.(*Gate)

		// if g.IsCurrent == 0 {
		if g.IsCurrent != 1 {
			continue
		}

		current = append(current, g)
	}

	return current, nil
}

// Return the Gate matching 'code' that was active for 'date' using 'lookup'. Multiple matches throw an error.
func FindGateForDateWithLookup(ctx context.Context, lookup architecture.Lookup, code string, date string) (*Gate, error) {

	gates, err := FindAllGatesForDateWithLookup(ctx, lookup, code, date)

	if err != nil {
		return nil, err
	}

	switch len(gates) {
	case 0:
		return nil, NotFound{code}
	case 1:
		return gates[0], nil
	default:
		return nil, MultipleCandidates{code}
	}

}

// Return all the Gates matching 'code' that were active for 'date' using 'lookup'.
func FindAllGatesForDateWithLookup(ctx context.Context, lookup architecture.Lookup, code string, date string) ([]*Gate, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to find gates for code, %w", err)
	}

	gates := make([]*Gate, 0)

	for _, r := range rsp {

		g := r.(*Gate)

		inception := g.Inception
		cessation := g.Cessation

		is_between, err := cmp.IsBetween(date, inception, cessation)

		if err != nil {
			slog.Debug("Failed to determine whether gate matches date conditions", "code", code, "date", date, "gate", g.Name, "inception", inception, "cessation", cessation, "error", err)
			continue
		}

		if !is_between {
			slog.Debug("Gate does not match date conditions", "id", g.WhosOnFirstId, "code", code, "date", date, "gate", g.Name, "inception", inception, "cessation", cessation)
			continue
		}

		slog.Debug("Gate DOES match date conditions", "id", g.WhosOnFirstId, "code", code, "date", date, "gate", g.Name, "inception", inception, "cessation", cessation)

		gates = append(gates, g)
		break
	}

	slog.Debug("Return gates", "code", code, "date", date, "count", len(gates))
	return gates, nil
}
