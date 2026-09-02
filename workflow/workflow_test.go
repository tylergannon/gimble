package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/lint"
)

func TestBuiltInPlan(t *testing.T) {
	listed := List()
	if len(listed) != 1 || listed[0].Name != PlanName || listed[0].Description == "" {
		t.Fatalf("List = %#v", listed)
	}

	params := Parameters{
		Project:    "quote-test",
		Seed:       "seed's brief.md",
		Workdir:    "/tmp/work dir",
		Executable: "/tmp/tractor's binary",
	}
	pipeline, err := Build(PlanName, params)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if pipeline.Name != PlanName || pipeline.Start != "planner" || len(pipeline.Nodes) != 2 {
		t.Fatalf("graph identity/shape = %#v", pipeline)
	}

	planner := mustNode[*graph.CodergenNode](t, pipeline, "planner")
	if len(planner.Edges) != 1 || planner.Edges[0].To != "validate" {
		t.Fatalf("planner edges = %#v", planner.Edges)
	}
	if planner.Fidelity.Value != "full" || planner.ThreadID.Value != "plan" {
		t.Fatalf("planner context = fidelity %q, thread %q", planner.Fidelity.Value, planner.ThreadID.Value)
	}
	timeout, err := planner.Timeout.Value.Parse()
	if err != nil || timeout < 12*time.Hour {
		t.Fatalf("planner timeout = %q, %v", planner.Timeout.Value, err)
	}
	prompt := planner.Prompt.Value
	for _, required := range []string{
		"tractor's binary", "seed's brief.md", "Ask one Markdown or HTML question at a time",
		"intent, scope and non-goals, constraints, and the Definition of success",
		"two passes", "brief.md", "checklist.md", "recommendation.md", "Never write a done field",
		"Size: SIMPLE|MEDIUM|LARGE", "tractor workflow run medium", "tractor workflow run large",
	} {
		if !strings.Contains(prompt, required) {
			t.Errorf("planner prompt does not contain %q", required)
		}
	}

	validator := mustNode[*graph.ToolNode](t, pipeline, "validate")
	if validator.OnSuccess != graph.Success || !validator.OnError.Present || validator.OnError.Value != "planner" {
		t.Fatalf("validator routes = success %q, error %#v", validator.OnSuccess, validator.OnError)
	}
	if strings.Contains(validator.ToolCommand, params.Seed) {
		t.Fatal("validator command unexpectedly contains the seed")
	}
	for _, required := range []string{"'/tmp/tractor'\\''s binary'", "--project 'quote-test'", "--workdir '/tmp/work dir'"} {
		if !strings.Contains(validator.ToolCommand, required) {
			t.Errorf("validator command %q does not contain %q", validator.ToolCommand, required)
		}
	}
	if diagnostics, err := lint.ValidateOrError(*pipeline); err != nil {
		t.Fatalf("embedded graph lint: %v (%#v)", err, diagnostics)
	}

	if _, err := Build("missing", params); err == nil {
		t.Fatal("Build accepted unknown workflow")
	}
	unsafe := params
	unsafe.Project = "../escape"
	if _, err := Build(PlanName, unsafe); err == nil {
		t.Fatal("Build accepted unsafe project")
	}

	originalDefinition := planDefinition
	planDefinition = []byte("not a pipeline")
	t.Cleanup(func() { planDefinition = originalDefinition })
	if _, err := Build(PlanName, params); err == nil {
		t.Fatal("Build accepted an invalid embedded definition")
	}
}

func TestPlanArtifacts(t *testing.T) {
	workdir := t.TempDir()
	root := filepath.Join(workdir, "ephemeral", "projects", "demo")
	writeFile(t, filepath.Join(root, BriefFile), "# Brief\n\nBuild the thing.\n")
	writeFile(t, filepath.Join(root, ChecklistFile), validChecklist)
	writeFile(t, filepath.Join(root, RecommendationFile), recommendationText("SIMPLE", "One bounded sprint.", "Execute the plan yourself."))

	result, err := ValidatePlanArtifacts(workdir, "demo")
	if err != nil {
		t.Fatalf("ValidatePlanArtifacts: %v", err)
	}
	if result.Size != "SIMPLE" {
		t.Fatalf("recommendation = %#v", result)
	}

	tests := map[string]func(root string){
		"missing brief":     func(root string) { removeFile(t, filepath.Join(root, BriefFile)) },
		"empty brief":       func(root string) { writeFile(t, filepath.Join(root, BriefFile), "") },
		"missing checklist": func(root string) { removeFile(t, filepath.Join(root, ChecklistFile)) },
		"empty checklist":   func(root string) { writeFile(t, filepath.Join(root, ChecklistFile), "") },
		"invalid checklist": func(root string) {
			writeFile(t, filepath.Join(root, ChecklistFile), "not frontmatter\n")
		},
		"no checklist items": func(root string) {
			writeFile(t, filepath.Join(root, ChecklistFile), "---\nitems: []\n---\n")
		},
		"agent done true": func(root string) {
			writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(validChecklist, "    command:", "    done: true\n    command:", 1))
		},
		"agent done false": func(root string) {
			writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(validChecklist, "    command:", "    done: false\n    command:", 1))
		},
		"missing recommendation": func(root string) { removeFile(t, filepath.Join(root, RecommendationFile)) },
		"empty recommendation":   func(root string) { writeFile(t, filepath.Join(root, RecommendationFile), "") },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			caseWorkdir := t.TempDir()
			caseRoot := filepath.Join(caseWorkdir, "ephemeral", "projects", "demo")
			writeFile(t, filepath.Join(caseRoot, BriefFile), "brief\n")
			writeFile(t, filepath.Join(caseRoot, ChecklistFile), validChecklist)
			writeFile(t, filepath.Join(caseRoot, RecommendationFile), recommendationText("SIMPLE", "Small.", "Execute the plan yourself."))
			mutate(caseRoot)
			if _, err := ValidatePlanArtifacts(caseWorkdir, "demo"); err == nil {
				t.Fatal("invalid artifacts passed validation")
			}
		})
	}
}

func TestRecommendation(t *testing.T) {
	tests := []struct {
		size string
		next string
	}{
		{"SIMPLE", "Execute the plan yourself."},
		{"MEDIUM", "tractor workflow run medium --project demo"},
		{"LARGE", "tractor workflow run large --project demo"},
	}
	for _, test := range tests {
		t.Run(test.size, func(t *testing.T) {
			workdir := t.TempDir()
			root := filepath.Join(workdir, "ephemeral", "projects", "demo")
			writeFile(t, filepath.Join(root, BriefFile), "brief\n")
			writeFile(t, filepath.Join(root, ChecklistFile), validChecklist)
			path := filepath.Join(root, RecommendationFile)
			writeFile(t, path, recommendationText(test.size, "Because the scope fits.", test.next))

			parsed, err := ParseRecommendation(path)
			if err != nil {
				t.Fatalf("ParseRecommendation: %v", err)
			}
			if parsed.Size != test.size || parsed.Next != test.next || parsed.Rationale == "" {
				t.Fatalf("parsed recommendation = %#v", parsed)
			}
			if _, err := ValidatePlanArtifacts(workdir, "demo"); err != nil {
				t.Fatalf("ValidatePlanArtifacts: %v", err)
			}
		})
	}

	invalid := []string{
		recommendationText("TINY", "No such size.", "Execute the plan yourself."),
		recommendationText("MEDIUM", "Wrong next.", "Execute the plan yourself."),
		"# Recommendation\nSize: SIMPLE\nRationale:\nNext: Execute the plan yourself.\n",
		"Size: SIMPLE\nRationale: Missing heading.\nNext: Execute the plan yourself.\n",
	}
	for index, contents := range invalid {
		t.Run("invalid-"+string(rune('a'+index)), func(t *testing.T) {
			workdir := t.TempDir()
			root := filepath.Join(workdir, "ephemeral", "projects", "demo")
			writeFile(t, filepath.Join(root, BriefFile), "brief\n")
			writeFile(t, filepath.Join(root, ChecklistFile), validChecklist)
			writeFile(t, filepath.Join(root, RecommendationFile), contents)
			if _, err := ValidatePlanArtifacts(workdir, "demo"); err == nil {
				t.Fatal("invalid recommendation passed validation")
			}
		})
	}
}

const validChecklist = `---
items:
  - name: implement behavior
    check: The behavior is observable through the public surface.
    command: go test ./...
---

# Plan
`

func recommendationText(size, rationale, next string) string {
	return "# Recommendation\nSize: " + size + "\nRationale: " + rationale + "\nNext: " + next + "\n"
}

func mustNode[T graph.Node](t *testing.T, pipeline *graph.Graph, id string) T {
	t.Helper()
	node, ok := pipeline.NodeByID(id)
	if !ok {
		t.Fatalf("missing node %q", id)
	}
	typed, ok := node.(T)
	if !ok {
		t.Fatalf("node %q has type %T", id, node)
	}
	return typed
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func removeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
