package main

import (
	"fmt"
	"strings"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

// Input is this workflow's JSON argument, independent of the runtime API.
type Input struct {
	Goal          string   `json:"goal"`
	Sprints       []Sprint `json:"sprints"`
	MaxIterations int      `json:"max_iterations"`
	MaxRepairs    int      `json:"max_repairs"`
}

type Sprint struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func (in Input) ledger() ([]program.Item, error) {
	if strings.TrimSpace(in.Goal) == "" || len(in.Sprints) == 0 {
		return nil, fmt.Errorf("goal and at least one sprint are required")
	}
	if in.MaxIterations < 1 || in.MaxRepairs < 0 {
		return nil, fmt.Errorf("max_iterations must be positive; max_repairs cannot be negative")
	}
	items := make([]program.Item, len(in.Sprints))
	for i, sprint := range in.Sprints {
		if strings.TrimSpace(sprint.Name) == "" || strings.TrimSpace(sprint.Command) == "" {
			return nil, fmt.Errorf("every sprint needs a name and command")
		}
		items[i] = program.Item{Name: sprint.Name, Command: sprint.Command}
	}
	return items, nil
}

type Change struct{ Summary string }

type Review struct {
	MaterialDefect bool
	Notes          string
}

type Judgment struct{ Passed bool }
