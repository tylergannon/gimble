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
		directory, err := os.MkdirTemp("", "gimble-sprints-")
		if err != nil {
			return err
		}
		ctx, err = program.NewContext(ctx, directory, program.ContextLimits{ValueBytes: 600, PromptBytes: 2400})
		if err != nil {
			return err
		}
		if err := program.SetContext(ctx, "goal", input.Goal); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Context files retained under %s\n", directory)
		ledger, err := SprintExecute(ctx, demoRuntime(input), input)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(ledger)
	})
}

// This script forces a material-defect repair and a failed command after that
// repair. Only subsequent successful checks close the item.
// It is a fixed scenario, not an evaluator of the supplied task or commands.
func demoRuntime(input Input) *program.Runtime {
	thresholds := make(map[string]int)
	for i, sprint := range input.Sprints {
		thresholds[sprint.Command] = i + 2
	}
	implementations := 0
	reviews := 0
	return &program.Runtime{
		Output: os.Stderr,
		Agent: func(_ context.Context, call program.Call) (any, error) {
			switch call.Name {
			case "implement":
				implementations++
				return Change{Summary: "scripted edit"}, nil
			case "repair":
				return Change{Summary: "scripted edit"}, nil
			case "review":
				reviews++
				return Review{MaterialDefect: reviews == 1, Notes: "scripted boundary defect"}, nil
			case "evaluate":
				return Judgment{Passed: true}, nil
			default:
				return nil, fmt.Errorf("unscripted call %q", call.Name)
			}
		},
		RunCommand: func(_ context.Context, _, command string) (program.Check, error) {
			return program.Check{
				Passed: implementations >= thresholds[command],
				Output: "scripted boundary observations",
			}, nil
		},
	}
}
