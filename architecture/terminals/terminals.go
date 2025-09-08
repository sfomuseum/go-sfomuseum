// package terminals provides methods for working with boarding terminals at SFO.
package terminals

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sfomuseum/go-edtf/cmp"
	"github.com/sfomuseum/go-sfomuseum/architecture"
)

// To do: Make these sortable by inception/cessation
type Terminals []*Terminal

// type Terminal is a struct representing a passenger terminal at SFO.
type Terminal struct {
	// The Who's On First ID associated with this terminal.
	WhosOnFirstId int64 `json:"wof:id"`
	// The SFO Museum name/label for this terminal
	SFOMuseumId string `json:"sfomuseum:terminal_id,omityempty"`
	// The name of this terminal.
	Name string `json:"wof:name"`
	// A Who's On First "existential" (`KnownUnknownFlag`) flag signaling the terminal's status
	IsCurrent int64 `json:"mz:is_current"`
	// The list of name:{LANG}_x_preferred names for this terminal
	PreferredNames []string `json:"name:preferred,omitempty"`
	// The list of name:{LANG}_x_variant names for this terminal
	VariantNames []string `json:"name:variant,omitempty"`
	// The (EDTF) inception date for the gallery
	Inception string `json:"edtf:inception"`
	// The (EDTF) cessation date for the gallery
	Cessation string `json:"edtf:cessation"`
}

// String() will return the name of the terminal.
func (g *Terminal) String() string {
	return fmt.Sprintf("%d %s %s-%s (%d)", g.WhosOnFirstId, g.Name, g.Inception, g.Cessation, g.IsCurrent)
}

// Return the Terminal matching 'code' that was active for 'date'. Multiple matches throw an error.
func FindTerminalForDate(ctx context.Context, code string, date string) (*Terminal, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindTerminalForDateWithLookup(ctx, lookup, code, date)
}

// Return all the Terminals matching 'code' that were active for 'date'.
func FindAllTerminalsForDate(ctx context.Context, code string, date string) ([]*Terminal, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindAllTerminalsForDateWithLookup(ctx, lookup, code, date)
}

// Return the current Terminal matching 'code'. Multiple matches throw an error.
func FindCurrentTerminal(ctx context.Context, code string) (*Terminal, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindCurrentTerminalWithLookup(ctx, lookup, code)
}

// Return the current Terminal matching 'code' with a custom architecture.Lookup instance. Multiple matches throw an error.
func FindCurrentTerminalWithLookup(ctx context.Context, lookup architecture.Lookup, code string) (*Terminal, error) {

	current, err := FindTerminalsCurrentWithLookup(ctx, lookup, code)

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

// Returns all Terminal instances matching 'code' that are marked as current.
func FindTerminalsCurrent(ctx context.Context, code string) ([]*Terminal, error) {

	lookup, err := NewLookup(ctx, "")

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lookup, %w", err)
	}

	return FindTerminalsCurrentWithLookup(ctx, lookup, code)
}

// Returns all Terminal instances matching 'code' that are marked as current with a custom architecture.Lookup instance.
func FindTerminalsCurrentWithLookup(ctx context.Context, lookup architecture.Lookup, code string) ([]*Terminal, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to find terminal '%s', %w", code, err)
	}

	current := make([]*Terminal, 0)

	for _, r := range rsp {

		g := r.(*Terminal)

		// if g.IsCurrent == 0 {
		if g.IsCurrent != 1 {
			continue
		}

		current = append(current, g)
	}

	return current, nil
}

// Return the Terminal matching 'code' that was active for 'date' using 'lookup'. Multiple matches throw an error.
func FindTerminalForDateWithLookup(ctx context.Context, lookup architecture.Lookup, code string, date string) (*Terminal, error) {

	terminals, err := FindAllTerminalsForDateWithLookup(ctx, lookup, code, date)

	if err != nil {
		return nil, err
	}

	switch len(terminals) {
	case 0:
		return nil, NotFound{code}
	case 1:
		return terminals[0], nil
	default:
		return nil, MultipleCandidates{code}
	}

}

// Return all the Terminals matching 'code' that were active for 'date' using 'lookup'.
func FindAllTerminalsForDateWithLookup(ctx context.Context, lookup architecture.Lookup, code string, date string) ([]*Terminal, error) {

	rsp, err := lookup.Find(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("Failed to find terminals for code, %w", err)
	}

	terminals := make([]*Terminal, 0)

	for _, r := range rsp {

		g := r.(*Terminal)

		inception := g.Inception
		cessation := g.Cessation

		is_between, err := cmp.IsBetween(date, inception, cessation)

		if err != nil {
			slog.Debug("Failed to determine whether terminal matches date conditions", "code", code, "date", date, "terminal", g.Name, "inception", inception, "cessation", cessation, "error", err)
			continue
		}

		if !is_between {
			slog.Debug("Terminal does not match date conditions", "code", code, "date", date, "terminal", g.Name, "inception", inception, "cessation", cessation)
			continue
		}

		slog.Debug("Terminal DOES match date conditions", "code", code, "date", date, "terminal", g.Name, "inception", inception, "cessation", cessation)
		terminals = append(terminals, g)
		break
	}

	slog.Debug("Return terminals", "code", code, "date", date, "count", len(terminals))
	return terminals, nil
}
