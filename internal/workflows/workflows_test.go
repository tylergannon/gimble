package workflows_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/internal/workflows"
	"github.com/tylergannon/tractor/lint"
)

// Every built-in parses and lints, so a broken workflow cannot ship.
func TestBuiltinWorkflowsValidate(t *testing.T) {
	t.Parallel()

	for _, workflow := range workflows.List() {
		t.Run(workflow.Name, func(t *testing.T) {
			t.Parallel()
			raw, err := workflows.Read(workflow.Name)
			if err != nil {
				t.Fatal(err)
			}
			pipeline, err := graph.ParseYAML(raw)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lint.ValidateOrError(*pipeline); err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(workflow.When) == "" {
				t.Error("built-in workflow has no when line")
			}
		})
	}
}

// The catalogue and the embedded files hold each other: a workflow cannot
// ship undocumented, and cannot be documented without shipping.
func TestCatalogueMatchesEmbeddedFiles(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	onDisk := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		onDisk[strings.TrimSuffix(entry.Name(), ".yaml")] = true
	}
	for _, name := range workflows.Names() {
		if !onDisk[name] {
			t.Errorf("catalogue lists %q but %s.yaml does not exist", name, name)
		}
		delete(onDisk, name)
	}
	for name := range onDisk {
		t.Errorf("%s.yaml ships but the catalogue does not list it", name)
	}
}

// The pipeline name inside each file matches the name it is served under, so
// a run's records name the workflow the operator asked for.
func TestBuiltinNamesMatchPipelineNames(t *testing.T) {
	t.Parallel()

	for _, workflow := range workflows.List() {
		raw, err := workflows.Read(workflow.Name)
		if err != nil {
			t.Fatal(err)
		}
		pipeline, err := graph.ParseYAML(raw)
		if err != nil {
			t.Fatal(err)
		}
		if pipeline.Name != workflow.Name {
			t.Errorf("%s.yaml declares name %q", workflow.Name, pipeline.Name)
		}
	}
}

// --goal is only meaningful where a prompt expands it, so every workflow has
// to carry the run goal into at least one agent turn.
func TestEveryBuiltinCarriesTheRunGoalIntoAnAgentTurn(t *testing.T) {
	t.Parallel()

	for _, workflow := range workflows.List() {
		t.Run(workflow.Name, func(t *testing.T) {
			t.Parallel()
			raw, err := workflows.Read(workflow.Name)
			if err != nil {
				t.Fatal(err)
			}
			pipeline, err := graph.ParseYAML(raw)
			if err != nil {
				t.Fatal(err)
			}
			for _, node := range pipeline.Nodes {
				var prompt string
				switch typed := node.(type) {
				case *graph.AgentNode:
					prompt = typed.PromptValue("")
				case *graph.FanInNode:
					prompt = typed.PromptValue("")
				case *graph.FanOutNode:
					prompt = typed.PromptValue("")
				}
				if strings.Contains(prompt, "$goal") {
					return
				}
			}
			t.Error("no agent prompt expands $goal, so --goal cannot steer this workflow")
		})
	}
}

func TestReadRejectsUnknownName(t *testing.T) {
	t.Parallel()

	if _, err := workflows.Read("no-such-workflow"); err == nil {
		t.Fatal("expected an error for an unknown workflow")
	}
}

// A workflow that advertises a precondition must actually read that path, and
// one that says nothing must not be waiting on a file the operator never
// heard about. Needs and NeedsGoal are alternatives, never both.
func TestNeedsNamesAChecklistTheWorkflowReads(t *testing.T) {
	t.Parallel()
	for _, workflow := range workflows.List() {
		if workflow.Needs != "" && workflow.NeedsGoal {
			t.Errorf("%s declares both Needs and NeedsGoal", workflow.Name)
		}
		raw, err := workflows.Read(workflow.Name)
		if err != nil {
			t.Fatalf("read %s: %v", workflow.Name, err)
		}
		pipeline, err := graph.ParseYAML(raw)
		if err != nil {
			t.Fatalf("parse %s: %v", workflow.Name, err)
		}
		declared := map[string]bool{}
		for _, node := range pipeline.Nodes {
			loop, ok := node.(*graph.LoopNode)
			if !ok || !loop.Checklist.Present {
				continue
			}
			if path := strings.TrimSpace(loop.Checklist.Value); path != "" {
				declared[path] = true
			}
		}
		if workflow.Needs != "" && !declared[workflow.Needs] {
			t.Errorf("%s needs %q but no loop node reads it", workflow.Name, workflow.Needs)
		}
		if workflow.Needs == "" && len(declared) > 0 && !workflow.NeedsGoal {
			t.Errorf("%s reads %v but declares no precondition", workflow.Name, declared)
		}
	}
}

// An agent inside a loop body must carry a visit budget. The loop node's own
// max_visits guards arrivals at the loop, which a cycle between two body
// agents never reaches — so a reviewer that keeps routing back to the coder
// runs forever. v0.9.0 shipped three workflows that could do exactly that.
func TestEveryAgentInALoopBodyIsBounded(t *testing.T) {
	t.Parallel()
	for _, workflow := range workflows.List() {
		raw, err := workflows.Read(workflow.Name)
		if err != nil {
			t.Fatalf("read %s: %v", workflow.Name, err)
		}
		pipeline, err := graph.ParseYAML(raw)
		if err != nil {
			t.Fatalf("parse %s: %v", workflow.Name, err)
		}
		byID := map[string]graph.Node{}
		for _, node := range pipeline.Nodes {
			byID[node.Base().ID] = node
		}
		for _, node := range pipeline.Nodes {
			loop, ok := node.(*graph.LoopNode)
			if !ok {
				continue
			}
			body, found := lint.LoopBodyNodes(*pipeline, loop.ID)
			if !found {
				t.Errorf("%s: no body node-set for loop %q", workflow.Name, loop.ID)
				continue
			}
			for _, id := range body {
				agent, isAgent := byID[id].(*graph.AgentNode)
				if !isAgent {
					continue
				}
				if !agent.MaxVisits.Present || agent.MaxVisits.Value <= 0 {
					t.Errorf("%s: agent %q is in loop %q's body with no max_visits; a cycle between body agents never reaches the loop node's own budget",
						workflow.Name, id, loop.ID)
				}
			}
		}
	}
}

// A checklist reviewer decides only whether it found a concrete defect. The
// loop owns the item's command, inference, and done field, so uncertainty must
// route to that engine check rather than back to another coding turn.
func TestChecklistReviewersDeferDoneToTheEngine(t *testing.T) {
	t.Parallel()

	const (
		readyForCheck = "The work is ready for the engine's check, or that check is needed to settle uncertainty."
		defectFound   = "You found a specific material defect to fix before the engine's check."
	)
	cases := []struct {
		workflow string
		reviewer string
		loop     string
		coder    string
	}{
		{workflow: "sprint-execute", reviewer: "review", loop: "sprints", coder: "implement"},
		{workflow: "chapter-loop", reviewer: "review", loop: "sprints", coder: "implement"},
		{workflow: "delivery-loop", reviewer: "review", loop: "work", coder: "coding"},
	}

	for _, test := range cases {
		t.Run(test.workflow, func(t *testing.T) {
			raw, err := workflows.Read(test.workflow)
			if err != nil {
				t.Fatal(err)
			}
			pipeline, err := graph.ParseYAML(raw)
			if err != nil {
				t.Fatal(err)
			}

			var reviewer *graph.AgentNode
			for _, node := range pipeline.Nodes {
				if node.Base().ID != test.reviewer {
					continue
				}
				reviewer, _ = node.(*graph.AgentNode)
				break
			}
			if reviewer == nil {
				t.Fatalf("reviewer %q is not an agent", test.reviewer)
			}
			if prompt := reviewer.PromptValue(""); !strings.Contains(prompt, "The engine runs the item's check and owns done.") {
				t.Errorf("reviewer prompt gives no ownership of done to the engine: %q", prompt)
			}

			conditions := map[string]string{}
			for _, edge := range reviewer.Edges {
				conditions[edge.To] = edge.Condition
			}
			if got := conditions[test.loop]; got != readyForCheck {
				t.Errorf("edge to checklist loop %q has condition %q, want %q", test.loop, got, readyForCheck)
			}
			if got := conditions[test.coder]; got != defectFound {
				t.Errorf("edge to coder %q has condition %q, want %q", test.coder, got, defectFound)
			}
		})
	}
}
