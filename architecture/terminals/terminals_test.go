package terminals

import (
	"context"
	"log/slog"
	"testing"
)

type terminalTest struct {
	Id   int64
	Code string
	Date string
}

func TestFindCurrentTerminal(t *testing.T) {

	tests := map[string]int64{
		"T2": 1914601345,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentTerminal(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current terminal for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for terminal %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}

func TestFindTerminalForDate(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	tests := []*terminalTest{
		&terminalTest{Id: 1763588123, Code: "T2", Date: "2023"},
		&terminalTest{Id: 1914601197, Code: "T1", Date: "2024-07-23"},
	}

	ctx := context.Background()

	for _, terminal := range tests {

		g, err := FindTerminalForDate(ctx, terminal.Code, terminal.Date)

		if err != nil {
			t.Fatalf("Failed to find current terminal for %s, %v", terminal.Code, err)
		}

		if g.WhosOnFirstId != terminal.Id {
			t.Fatalf("Unexpected ID for terminal %s. Got %d but expected %d", terminal.Code, g.WhosOnFirstId, terminal.Id)
		}
	}

}
