// Package workflows holds the pipelines that ship inside the binary. Each one
// is a complete graph an operator can run by name, read with `tractor
// workflows show`, or copy out and edit.
package workflows

import (
	"embed"
	"fmt"
	"slices"
)

//go:embed *.yaml
var files embed.FS

// Workflow is one built-in pipeline and the situation it is for.
type Workflow struct {
	// Name is the identifier `tractor run` and `tractor workflows show` accept.
	Name string
	// When describes the situation that calls for this workflow, in one line.
	When string
	// NeedsGoal reports whether the workflow is meaningless without a goal
	// supplied at the command line. Workflows that take their input from a
	// file in the workspace, such as a ledger, do not need one.
	NeedsGoal bool
	// GoalHint shows what a usable goal looks like. It is set only when
	// NeedsGoal is.
	GoalHint string
}

// builtin is the catalogue. A test holds it and the embedded files to each
// other, so a workflow cannot ship undocumented or be documented without
// shipping.
var builtin = []Workflow{
	{
		Name: "sprint-execute",
		When: "Work a planned sprint ledger to done, one sprint per lap, each demonstrated before it closes.",
	},
	{
		Name: "chapter-loop",
		When: "Carry a chapter's worth of work in one run: chapters holding sprints holding coding and validation.",
	},
	{
		Name:      "sprint-plan",
		When:      "Plan the next sprint properly: three independent drafts, mutual critique, your answers, one merge.",
		NeedsGoal: true,
		GoalHint:  `for example: --goal "Replace the polling status endpoint with server-sent events"`,
	},
	{
		Name:      "delivery-loop",
		When:      "Turn a written specification into working software while you are away.",
		NeedsGoal: true,
		GoalHint:  `for example: --goal "Build what docs/SPEC.md describes"`,
	},
	{
		Name: "promise-loop",
		When: "Advance one repository promise by a bounded run and finish with a verdict.",
	},
}

// List returns every built-in workflow in catalogue order.
func List() []Workflow {
	return slices.Clone(builtin)
}

// Lookup returns the workflow named name.
func Lookup(name string) (Workflow, bool) {
	for _, workflow := range builtin {
		if workflow.Name == name {
			return workflow, true
		}
	}
	return Workflow{}, false
}

// Names returns every built-in name in catalogue order.
func Names() []string {
	names := make([]string, len(builtin))
	for index, workflow := range builtin {
		names[index] = workflow.Name
	}
	return names
}

// Read returns the pipeline source for name.
func Read(name string) ([]byte, error) {
	if _, ok := Lookup(name); !ok {
		return nil, fmt.Errorf("no built-in workflow named %q", name)
	}
	raw, err := files.ReadFile(name + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("read built-in workflow %q: %w", name, err)
	}
	return raw, nil
}
