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
