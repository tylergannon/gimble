package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

type Stage struct {
	Name    string                  `json:"name"`
	Code    string                  `json:"code"`
	Prompt  string                  `json:"prompt"`
	Context program.ContextSnapshot `json:"context"`
}

type Report struct {
	Directory string  `json:"directory"`
	Stages    []Stage `json:"stages"`
}

var stageDescriptions = map[string]struct{ name, code string }{
	"plan": {
		"Small goal and constraints",
		`goal, err := program.DeclareContext(ctx, "goal")
constraints, err := program.DeclareContext(ctx, "constraints")
research, err := program.DeclareContext(ctx, "research")
program.SetContext(ctx, goal, input.Goal)
program.SetContext(ctx, constraints, pricingConstraints())
program.Codergen[string](ctx, runtime, "plan", engMgr, "Outline the implementation and its proof.")`,
	},
	"build": {
		"Oversized research moves to the index",
		`program.SetContext(ctx, research, researchNotes())
program.Codergen[string](ctx, runtime, "build", sswe, "Implement the quote command.")`,
	},
	"verify": {
		"Small values together exceed the budget",
		`for _, note := range acceptanceNotes() {
    key, err := program.DeclareContext(ctx, note.Key)
    program.SetContext(ctx, key, note.Value)
}
program.Codergen[string](ctx, runtime, "verify", tester, "Exercise the boundary cases and judge the recorded evidence.")`,
	},
}

func captureRuntime(report *Report) *program.Runtime {
	return &program.Runtime{
		Output: os.Stderr,
		Agent: func(_ context.Context, call program.Call) (any, error) {
			// Codergen must have built this projection before invoking the agent.
			// This is the fixed view supplied to this call, even if a parent changes.
			snapshot := call.Context
			if snapshot.Prompt == "" || !strings.HasPrefix(call.Prompt, snapshot.Prompt) {
				return nil, fmt.Errorf("%s: agent did not receive the ready context projection", call.Name)
			}
			stage, ok := stageDescriptions[call.Name]
			if !ok {
				return nil, fmt.Errorf("unexpected agent stage %q", call.Name)
			}
			report.Stages = append(report.Stages, Stage{stage.name, stage.code, call.Prompt, snapshot})
			return "Canned response; no agent was launched.", nil
		},
	}
}

func main() {
	directory := flag.String("dir", "", "retain context files under this directory; defaults to a new temporary directory")
	input := Input{Goal: "Implement a quote CLI: free shipping starts at a 50.00 subtotal."}
	program.Main(input, func(ctx context.Context, input Input) error {
		if err := input.validate(); err != nil {
			return err
		}
		root := *directory
		if root == "" {
			var err error
			root, err = os.MkdirTemp("", "gimble-context-demo-")
			if err != nil {
				return err
			}
		}
		ctx, err := program.NewContext(ctx, root, program.ContextLimits{ValueBytes: 240, PromptBytes: 800})
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Context files retained under %s\n", root)
		report := Report{Directory: root}
		if err := ContextWalkthrough(ctx, captureRuntime(&report), input); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	})
}
