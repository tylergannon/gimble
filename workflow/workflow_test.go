package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tylergannon/tractor/checklist"
	"github.com/tylergannon/tractor/engine"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
	"github.com/tylergannon/tractor/lint"
)

func TestBuiltInPlan(t *testing.T) {
	listed := List()
	if len(listed) != 2 || listed[1].Name != PlanName || listed[1].Description == "" {
		t.Fatalf("List = %#v", listed)
	}

	params := Parameters{
		Project:    "quote-test",
		Workdir:    "/tmp/work dir",
		Executable: "/tmp/tractor's binary",
		Plan:       PlanParameters{Seed: "seed's brief.md"},
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
		"Size: SIMPLE|MEDIUM|LARGE", "SIMPLE: zero or one flat implementation sprint",
		"MEDIUM: more than one flat implementation sprint", "LARGE: more than one chapter item",
		"items: []", "unique doc path", "unique checklist path", "contains no traversal",
		"tractor workflow run medium", "tractor workflow run large",
	} {
		if !strings.Contains(prompt, required) {
			t.Errorf("planner prompt does not contain %q", required)
		}
	}

	validator := mustNode[*graph.ToolNode](t, pipeline, "validate")
	if validator.OnSuccess != graph.Success || !validator.OnError.Present || validator.OnError.Value != "planner" {
		t.Fatalf("validator routes = success %q, error %#v", validator.OnSuccess, validator.OnError)
	}
	if strings.Contains(validator.ToolCommand, params.Plan.Seed) {
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

func TestBuiltInMedium(t *testing.T) {
	listed := List()
	if len(listed) != 2 || listed[0].Name != MediumName || listed[1].Name != PlanName {
		t.Fatalf("List = %#v", listed)
	}
	for _, definition := range listed {
		if definition.Description == "" {
			t.Fatalf("workflow %q has no description", definition.Name)
		}
	}

	params := Parameters{
		Project:    "medium-demo",
		Workdir:    "/tmp/work dir",
		Executable: "/tmp/tractor's binary",
	}
	pipeline, err := Build(MediumName, params)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if pipeline.Name != MediumName || pipeline.Start != "sprints" || len(pipeline.Nodes) != 2 {
		t.Fatalf("graph identity/shape = %#v", pipeline)
	}

	sprints := mustNode[*graph.LoopNode](t, pipeline, "sprints")
	wantRoot := filepath.Join(params.Workdir, "ephemeral", "projects", params.Project)
	if !sprints.Checklist.Present || sprints.Checklist.Value != filepath.Join(wantRoot, ChecklistFile) {
		t.Fatalf("sprints checklist = %#v", sprints.Checklist)
	}
	if sprints.Body != "implement" || sprints.OnDone != graph.Success || !sprints.MaxVisits.Present || sprints.MaxVisits.Value != 40 {
		t.Fatalf("sprints lifecycle = %#v", sprints)
	}

	implement := mustNode[*graph.CodergenNode](t, pipeline, "implement")
	if !reflect.DeepEqual(implement.Edges, []graph.Edge{{To: "sprints"}}) || implement.MaxVisits.Present {
		t.Fatalf("implement routing/budget = edges %#v, max visits %#v", implement.Edges, implement.MaxVisits)
	}
	if implement.Fidelity.Value != "full" || implement.ThreadID.Value != MediumName {
		t.Fatalf("implement context = fidelity %q, thread %q", implement.Fidelity.Value, implement.ThreadID.Value)
	}
	prompt := implement.Prompt.Value
	for _, required := range []string{
		filepath.Join(wantRoot, BriefFile), filepath.Join(wantRoot, ChecklistFile), filepath.Join(wantRoot, "interview"),
		"current <iterate> frame", "item document", "Complete the whole current sprint", "Run the sprint's validator command yourself",
		"never add, remove, or edit a done field", "Commit the completed sprint", "repeated engine validation failure",
		"TRACTOR_INTERVIEW_DIR=", "ask <question-file>", "Do not start a child Tractor run", "do not add a failure-escalation mechanism",
	} {
		if !strings.Contains(prompt, required) {
			t.Errorf("implement prompt does not contain %q", required)
		}
	}
	if diagnostics, err := lint.ValidateOrError(*pipeline); err != nil {
		t.Fatalf("embedded graph lint: %v (%#v)", err, diagnostics)
	}

	unsafe := params
	unsafe.Project = "../escape"
	if _, err := Build(MediumName, unsafe); err == nil {
		t.Fatal("Build accepted unsafe project")
	}

	originalDefinition := mediumDefinition
	mediumDefinition = []byte("not a pipeline")
	t.Cleanup(func() { mediumDefinition = originalDefinition })
	if _, err := Build(MediumName, params); err == nil {
		t.Fatal("Build accepted an invalid embedded definition")
	}
}

func TestMediumRunsPlanningChecklist(t *testing.T) {
	workdir := t.TempDir()
	logsRoot := filepath.Join(t.TempDir(), "run")
	projectRoot := filepath.Join(workdir, "ephemeral", "projects", "demo")
	checklistPath := filepath.Join(projectRoot, ChecklistFile)
	const body = "\n\n# Planned work\n\nKeep this body byte-for-byte.\n"
	writeFile(t, filepath.Join(projectRoot, BriefFile), "# Brief\n\nExecute both planned sprints.\n")
	writeFile(t, checklistPath, `---
items:
  - name: First sprint
    check: The first command ran.
    command: "printf 'first\\n' >> executed.log"
  - name: Second sprint
    check: The second command ran.
    command: "printf 'second\\n' >> executed.log"
---`+body)

	pipeline, err := Build(MediumName, Parameters{
		Project: "demo", Workdir: workdir, Executable: "/tmp/tractor",
	})
	if err != nil {
		t.Fatal(err)
	}
	selected := make([]string, 0, 2)
	registry := engine.NewRegistry()
	registry.Register("codergen", engine.HandlerFunc(func(_ graph.Node, _ []graph.Edge, scope engine.ExecutionScope, _ *graph.Graph) (harness.Outcome, *harness.Error) {
		for _, name := range []string{"First sprint", "Second sprint"} {
			if strings.Contains(scope.Frame, "name: "+name) {
				selected = append(selected, name)
				return harness.Outcome{Notes: "implemented " + name}, nil
			}
		}
		t.Fatalf("codergen frame did not select a planned sprint: %s", scope.Frame)
		return harness.Outcome{}, nil
	}))
	runner, err := engine.NewRunner(*pipeline, registry, engine.RunnerConfig{
		LogsRoot: logsRoot,
		Workdir:  workdir,
		Validate: func(candidate graph.Graph) error {
			_, err := lint.ValidateOrError(candidate)
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != engine.RunCompleted || result.FailureReason != "" {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(selected, []string{"First sprint", "Second sprint"}) {
		t.Fatalf("selected items = %#v", selected)
	}
	if got := readFile(t, filepath.Join(workdir, "executed.log")); got != "first\nsecond\n" {
		t.Fatalf("executed commands = %q", got)
	}

	list, err := checklist.Load(checklistPath)
	if err != nil {
		t.Fatal(err)
	}
	if list.Body != body {
		t.Fatalf("markdown body = %q, want %q", list.Body, body)
	}
	for _, item := range list.Items {
		if !item.Done {
			t.Fatalf("item %q is not done", item.Name)
		}
	}
	if got := strings.Count(readFile(t, checklistPath), "done: true"); got != 2 {
		t.Fatalf("done fields = %d, want 2", got)
	}

	recordPaths, err := filepath.Glob(filepath.Join(logsRoot, "stages", "*-sprints", "validation.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(recordPaths) != 2 {
		t.Fatalf("validation records = %v", recordPaths)
	}
	for _, path := range recordPaths {
		var record struct {
			Item     string `json:"item"`
			Command  string `json:"command"`
			ExitCode int    `json:"exit_code"`
			Passed   bool   `json:"passed"`
		}
		if err := json.Unmarshal([]byte(readFile(t, path)), &record); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if !record.Passed || record.ExitCode != 0 || record.Item == "" || record.Command == "" {
			t.Fatalf("validation record %s = %#v", path, record)
		}
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

func TestPlanExecutionShapes(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, size := range []string{"SIMPLE", "MEDIUM", "LARGE"} {
			t.Run(size, func(t *testing.T) {
				workdir, _ := writeValidPlan(t, size)
				if _, err := ValidatePlanArtifacts(workdir, "demo"); err != nil {
					t.Fatalf("ValidatePlanArtifacts: %v", err)
				}
			})
		}
		t.Run("SIMPLE empty", func(t *testing.T) {
			workdir, root := writeValidPlan(t, "SIMPLE")
			writeFile(t, filepath.Join(root, ChecklistFile), emptyChecklist)
			if _, err := ValidatePlanArtifacts(workdir, "demo"); err != nil {
				t.Fatalf("ValidatePlanArtifacts: %v", err)
			}
		})
	})

	tests := map[string]struct {
		size   string
		mutate func(t *testing.T, workdir, root string)
	}{
		"SIMPLE too many items": {
			size: "SIMPLE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), mediumChecklist)
			},
		},
		"SIMPLE nested item": {
			size: "SIMPLE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(validChecklist, "    command:", "    checklist: ephemeral/projects/demo/sprints.md\n    command:", 1))
			},
		},
		"MEDIUM too few items": {
			size: "MEDIUM",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), validChecklist)
			},
		},
		"MEDIUM nested item": {
			size: "MEDIUM",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(mediumChecklist, "    command:", "    checklist: ephemeral/projects/demo/sprints.md\n    command:", 1))
			},
		},
		"LARGE too few chapters": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, largeChapterTwo, "", 1))
			},
		},
		"LARGE missing doc path": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "    doc: ephemeral/projects/demo/chapters/one.md\n", "", 1))
			},
		},
		"LARGE missing checklist path": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "    checklist: ephemeral/projects/demo/chapters/one-sprints.md\n", "", 1))
			},
		},
		"LARGE missing doc file": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				removeFile(t, filepath.Join(root, "chapters", "two.md"))
			},
		},
		"LARGE missing checklist file": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				removeFile(t, filepath.Join(root, "chapters", "two-sprints.md"))
			},
		},
		"LARGE traversal path": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "ephemeral/projects/demo/chapters/one.md", "ephemeral/projects/demo/../outside.md", 1))
			},
		},
		"LARGE absolute path": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "ephemeral/projects/demo/chapters/one-sprints.md", filepath.Join(root, "chapters", "one-sprints.md"), 1))
			},
		},
		"LARGE symlink escape": {
			size: "LARGE",
			mutate: func(t *testing.T, workdir, root string) {
				outside := filepath.Join(workdir, "outside.md")
				writeFile(t, outside, "# Outside\n")
				link := filepath.Join(root, "chapters", "escape.md")
				if err := os.Symlink(outside, link); err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "ephemeral/projects/demo/chapters/one.md", "ephemeral/projects/demo/chapters/escape.md", 1))
			},
		},
		"LARGE duplicate doc paths": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "ephemeral/projects/demo/chapters/two.md", "ephemeral/projects/demo/chapters/one.md", 1))
			},
		},
		"LARGE duplicate checklist paths": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "ephemeral/projects/demo/chapters/two-sprints.md", "ephemeral/projects/demo/chapters/one-sprints.md", 1))
			},
		},
		"LARGE empty chapter doc": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, "chapters", "one.md"), "")
			},
		},
		"LARGE non-empty sprint ledger": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, "chapters", "one-sprints.md"), validChecklist)
			},
		},
		"LARGE top-level done": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, ChecklistFile), strings.Replace(largeChecklist, "    doc:", "    done: false\n    doc:", 1))
			},
		},
		"LARGE nested done": {
			size: "LARGE",
			mutate: func(t *testing.T, _, root string) {
				writeFile(t, filepath.Join(root, "chapters", "one-sprints.md"), strings.Replace(validChecklist, "    command:", "    done: true\n    command:", 1))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			workdir, root := writeValidPlan(t, test.size)
			test.mutate(t, workdir, root)
			if _, err := ValidatePlanArtifacts(workdir, "demo"); err == nil {
				t.Fatal("invalid execution shape passed validation")
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
			workdir, root := writeValidPlan(t, test.size)
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

const emptyChecklist = `---
items: []
---
`

const mediumChecklist = `---
items:
  - name: implement first behavior
    check: The first behavior is observable.
    command: go test ./...
  - name: implement second behavior
    check: The second behavior is observable.
    command: go test ./...
---
`

const largeChapterTwo = `  - name: chapter two
    check: The second chapter is complete.
    doc: ephemeral/projects/demo/chapters/two.md
    checklist: ephemeral/projects/demo/chapters/two-sprints.md
`

const largeChecklist = `---
items:
  - name: chapter one
    check: The first chapter is complete.
    doc: ephemeral/projects/demo/chapters/one.md
    checklist: ephemeral/projects/demo/chapters/one-sprints.md
` + largeChapterTwo + `---
`

func writeValidPlan(t *testing.T, size string) (string, string) {
	t.Helper()
	workdir := t.TempDir()
	root := filepath.Join(workdir, "ephemeral", "projects", "demo")
	writeFile(t, filepath.Join(root, BriefFile), "# Brief\n")

	var checklistContents, next string
	switch size {
	case "SIMPLE":
		checklistContents = validChecklist
		next = "Execute the plan yourself."
	case "MEDIUM":
		checklistContents = mediumChecklist
		next = "tractor workflow run medium --project demo"
	case "LARGE":
		checklistContents = largeChecklist
		next = "tractor workflow run large --project demo"
		for _, chapter := range []string{"one", "two"} {
			writeFile(t, filepath.Join(root, "chapters", chapter+".md"), "# Chapter "+chapter+"\n")
			writeFile(t, filepath.Join(root, "chapters", chapter+"-sprints.md"), emptyChecklist)
		}
	default:
		t.Fatalf("unsupported test size %q", size)
	}
	writeFile(t, filepath.Join(root, ChecklistFile), checklistContents)
	writeFile(t, filepath.Join(root, RecommendationFile), recommendationText(size, "Fits the execution shape.", next))
	return workdir, root
}

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

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func removeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
