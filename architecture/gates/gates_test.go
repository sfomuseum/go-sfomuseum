package gates

import (
	"context"
	"log/slog"
	"testing"
)

type gateTest struct {
	Id   int64
	Code string
	Date string
}

func TestFindCurrentGate(t *testing.T) {

	// slog.SetLogLoggerLevel(slog.LevelDebug)

	// "A9": 1763588417,

	tests := map[string]int64{
		"A9": 1914601013,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentGate(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current gate for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for gate %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}

func TestFindGateForDate(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	tests := []*gateTest{
		&gateTest{Id: 1745882165, Code: "B25", Date: "2021-11-09"},
		&gateTest{Id: 1914600935, Code: "E13", Date: "2024-07-23"},
		&gateTest{Id: 1914601021, Code: "A11", Date: "2024-07-25"},
		&gateTest{Id: 1763588293, Code: "F5", Date: "2022"},
		&gateTest{Id: 1930457221, Code: "F5", Date: "2024-07-25"},
	}

	ctx := context.Background()

	for _, gate := range tests {

		g, err := FindGateForDate(ctx, gate.Code, gate.Date)

		if err != nil {
			t.Fatalf("Failed to find current gate for %s, %v", gate.Code, err)
		}

		if g.WhosOnFirstId != gate.Id {
			t.Fatalf("Unexpected ID for gate %s. Got %d but expected %d", gate.Code, g.WhosOnFirstId, gate.Id)
		}
	}

}
