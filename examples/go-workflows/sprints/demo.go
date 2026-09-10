package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func main() {
	program.Main(Input{
		Goal: "Add a quote command and demonstrate its boundary cases.",
		Sprints: []Sprint{
			{Name: "Quote command", Command: "go run . --check-command"},
			{Name: "Price boundaries", Command: "go run . --check-boundaries"},
		},
		MaxIterations: 6,
		MaxRepairs:    2,
	}, func(ctx context.Context, input Input) error {
		ledger, err := SprintExecute(ctx, demoRuntime(), input)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(ledger)
	})
}

// Responses are canned; this does not implement or evaluate a sprint.
func demoRuntime() *program.Runtime {
	return &program.Runtime{Agent: func(_ context.Context, call program.Call) (any, error) {
		switch call.Name {
		case "implement", "repair":
			return Change{Summary: "canned change"}, nil
		case "review":
			return Review{Notes: "canned review"}, nil
		case "evaluate":
			return Judgment{Passed: true}, nil
		default:
			return nil, fmt.Errorf("unknown canned response %q", call.Name)
		}
	}}
}
