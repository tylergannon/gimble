package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

// Every response is canned. No agent, shell command, or Git operation runs.
func demoRuntime() *program.Runtime {
	return &program.Runtime{
		Agent: func(_ context.Context, call program.Call) (any, error) {
			switch call.Name {
			case "build":
				return fmt.Sprintf("Stub REPORT.md from %s: candidate prepared and demonstrated.", call.Role), nil
			case "judge":
				return Decision{Winner: 1, Reason: "Scripted choice; no candidates executed and WINNER.md is simulated."}, nil
			default:
				return nil, fmt.Errorf("unexpected stub agent call %q", call.Name)
			}
		},
		RunCommand: func(_ context.Context, _, _ string) (program.Check, error) {
			return program.Check{Passed: true, Output: "Stub acceptance passed; no command was executed."}, nil
		},
	}
}

func main() {
	input := Input{
		Goal: "Create fizzbuzz.sh: sh fizzbuzz.sh N prints classic FizzBuzz lines from 1 to N, one per line.",
		Acceptance: Acceptance{Command: `set -eu
test "$(sh fizzbuzz.sh 3 | tail -1)" = Fizz
test "$(sh fizzbuzz.sh 5 | tail -1)" = Buzz
test "$(sh fizzbuzz.sh 15 | tail -1)" = FizzBuzz
test -f WINNER.md`},
	}
	program.Main(input, func(ctx context.Context, input Input) error {
		result, err := BakeOff(ctx, demoRuntime(), input)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	})
}
