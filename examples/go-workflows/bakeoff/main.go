package main

import (
	"context"
	"fmt"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
	"golang.org/x/sync/errgroup"
)

const (
	sswe       program.Role = "sswe"
	alternate  program.Role = "sswe-alternative"
	challenger program.Role = "sswe-challenger"
	tester     program.Role = "tester"
)

type Candidate struct {
	Builder   program.Role `json:"builder"`
	Workspace string       `json:"workspace"`
	Report    string       `json:"report"`
}

type Decision struct {
	Winner int    `json:"winner"`
	Reason string `json:"reason"`
}

type Result struct {
	Winner   Candidate `json:"winner"`
	Reason   string    `json:"reason"`
	Verified bool      `json:"verified"`
}

func BakeOff(ctx context.Context, runtime *program.Runtime, input Input) (Result, error) {
	if err := input.validate(); err != nil {
		return Result{}, err
	}
	builders := []program.Role{sswe, alternate, challenger}
	candidates := make([]Candidate, len(builders))
	group, buildCtx := errgroup.WithContext(ctx)
	for i, builder := range builders {
		group.Go(func() error {
			branch, err := runtime.Worktree(buildCtx, program.Workspace(ctx), string(builder))
			if err != nil {
				return err
			}
			report, err := program.Codergen[string](branch, runtime, "build", builder, buildPrompt(input))
			candidates[i] = Candidate{builder, program.Workspace(branch), report}
			return err
		})
	}
	if err := group.Wait(); err != nil {
		return Result{}, err
	}

	decision, err := program.Codergen[Decision](ctx, runtime, "judge", tester, judgePrompt(input, candidates))
	if err != nil {
		return Result{}, err
	}
	if decision.Winner < 0 || decision.Winner >= len(candidates) {
		return Result{}, fmt.Errorf("judge selected unknown candidate %d", decision.Winner)
	}
	winner := candidates[decision.Winner]
	if err := runtime.Integrate(ctx, winner.Workspace); err != nil {
		return Result{}, err
	}
	check, err := runtime.Command(ctx, input.Acceptance.Command)
	if err != nil {
		return Result{}, err
	}
	if !check.Passed {
		return Result{}, fmt.Errorf("integrated winner failed acceptance: %s", check.Output)
	}
	return Result{winner, decision.Reason, true}, nil
}
