package main

import (
	"context"
	"fmt"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

const (
	sswe   program.Role = "sswe"
	tester program.Role = "tester"
)

// SprintExecute leaves selection and done reconciliation to the iterator.
// Implementation, review, and bounded repair are ordinary Go control flow.
func SprintExecute(ctx context.Context, rt *program.Runtime, input Input) ([]program.Item, error) {
	ledger, err := input.ledger()
	if err != nil {
		return nil, err
	}
	for sprint, err := range program.Loop(ctx, ledger, program.LoopOptions{
		Validate:      validate(rt),
		MaxIterations: input.MaxIterations,
	}) {
		if err != nil {
			return nil, err
		}
		if _, err := program.Codergen[Change](ctx, rt, "implement", sswe, implementPrompt(input.Goal, sprint.Item)); err != nil {
			return nil, err
		}
		for repairs := 0; ; repairs++ {
			review, err := program.Codergen[Review](ctx, rt, "review", tester, reviewPrompt(sprint.Item))
			if err != nil {
				return nil, err
			}
			if !review.MaterialDefect {
				break
			}
			if repairs >= input.MaxRepairs {
				return nil, fmt.Errorf("sprint %q needs a decision: %s", sprint.Item.Name, review.Notes)
			}
			if _, err := program.Codergen[Change](ctx, rt, "repair", sswe, repairPrompt(sprint.Item, review)); err != nil {
				return nil, err
			}
		}
	}
	return ledger, nil
}

// A check observes behavior, then a separate role judges what that observation
// establishes. Neither a review nor an implementation response sets Done.
func validate(rt *program.Runtime) func(context.Context, program.Item) (bool, error) {
	return func(ctx context.Context, item program.Item) (bool, error) {
		check, err := rt.Command(ctx, item.Command)
		if err != nil || !check.Passed {
			return false, err
		}
		judgment, err := program.Codergen[Judgment](ctx, rt, "evaluate", tester, evaluatePrompt(item, check))
		return judgment.Passed, err
	}
}
