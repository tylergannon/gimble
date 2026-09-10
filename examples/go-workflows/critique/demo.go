package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

// Returned values stand in for agent responses and unchanged file collection.
func demoRuntime() *program.Runtime {
	return &program.Runtime{
		Output: os.Stderr,
		Agent: func(_ context.Context, call program.Call) (any, error) {
			switch call.Name {
			case "propose":
				return fmt.Sprintf("Stub independent proposal by %s.", call.Role), nil
			case "critique":
				return fmt.Sprintf("Stub critique by %s: both peers have clear aims; their evidence needs detail; all agree on useful outcomes.", call.Role), nil
			default:
				return nil, fmt.Errorf("unexpected stub agent call %q", call.Name)
			}
		},
	}
}

func main() {
	input := Input{Topic: "Propose a three-sentence README tagline for a tool that runs multi-model agent pipelines. State your reasoning in one short paragraph."}
	program.Main(input, func(ctx context.Context, input Input) error {
		result, err := CritiqueCircle(ctx, demoRuntime(), input)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	})
}
