package main

import (
	"context"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

const (
	engMgr program.Role = "eng-mgr"
	sswe   program.Role = "sswe"
	tester program.Role = "tester"
)

// ContextWalkthrough leaves prompt assembly and index readiness to Codergen.
func ContextWalkthrough(ctx context.Context, runtime *program.Runtime, input Input) error {
	if err := input.validate(); err != nil {
		return err
	}
	if err := program.SetContext(ctx, "goal", input.Goal); err != nil {
		return err
	}
	if err := program.SetContext(ctx, "constraints", pricingConstraints()); err != nil {
		return err
	}
	if _, err := program.Codergen[string](ctx, runtime, "plan", engMgr, "Outline the implementation and its proof."); err != nil {
		return err
	}

	if err := program.SetContext(ctx, "research", researchNotes()); err != nil {
		return err
	}
	if _, err := program.Codergen[string](ctx, runtime, "build", sswe, "Implement the quote command."); err != nil {
		return err
	}

	for _, note := range acceptanceNotes() {
		if err := program.SetContext(ctx, note.Key, note.Value); err != nil {
			return err
		}
	}
	_, err := program.Codergen[string](ctx, runtime, "verify", tester, "Exercise the boundary cases and judge the recorded evidence.")
	return err
}
