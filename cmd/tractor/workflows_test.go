package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tylergannon/tractor/internal/workflows"
)

func TestWorkflowsListsEveryBuiltinWithItsSituation(t *testing.T) {
	stdout, _, err := executeCommand("workflows")
	if err != nil {
		t.Fatal(err)
	}
	for _, workflow := range workflows.List() {
		if !strings.Contains(stdout, workflow.Name) {
			t.Errorf("listing omits %q:\n%s", workflow.Name, stdout)
		}
		if !strings.Contains(stdout, workflow.When) {
			t.Errorf("listing omits the situation for %q:\n%s", workflow.Name, stdout)
		}
	}
}

func TestWorkflowsShowPrintsPipelineSourceVerbatim(t *testing.T) {
	stdout, _, err := executeCommand("workflows", "show", "sprint-execute")
	if err != nil {
		t.Fatal(err)
	}
	want, err := workflows.Read("sprint-execute")
	if err != nil {
		t.Fatal(err)
	}
	if stdout != string(want) {
		t.Error("show did not print the pipeline source verbatim")
	}
}

func TestWorkflowsShowNamesTheListingWhenTheNameIsUnknown(t *testing.T) {
	_, _, err := executeCommand("workflows", "show", "no-such-workflow")
	if err == nil {
		t.Fatal("expected an error for an unknown workflow")
	}
	if !strings.Contains(err.Error(), "tractor workflows") {
		t.Errorf("error does not point at the listing: %v", err)
	}
}

func TestValidateAcceptsABuiltinWorkflowName(t *testing.T) {
	for _, workflow := range workflows.List() {
		t.Run(workflow.Name, func(t *testing.T) {
			stdout, _, err := executeCommand("validate", workflow.Name)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout, "valid "+workflow.Name) {
				t.Errorf("unexpected output: %q", stdout)
			}
		})
	}
}

// A pipeline file always wins, so a workflow shipped in the binary can never
// silently replace one on disk.
func TestAPipelineFileShadowsABuiltinOfTheSameName(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "sprint-execute")
	if err := os.WriteFile(path, []byte(linearPipeline), 0o600); err != nil {
		t.Fatal(err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })

	stdout, _, err := executeCommand("validate", "sprint-execute")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "valid sprint-execute") {
		t.Errorf("unexpected output: %q", stdout)
	}
	models, _, err := executeCommand("inspect-models", "sprint-execute")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(models, "claude-opus-5") {
		t.Error("the built-in workflow shadowed the file on disk")
	}
}

func TestUnknownPipelineSourceNamesTheBuiltins(t *testing.T) {
	_, _, err := executeCommand("validate", "no-such-pipeline")
	if err == nil {
		t.Fatal("expected an error for an unknown pipeline source")
	}
	for _, name := range workflows.Names() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error omits built-in %q: %v", name, err)
		}
	}
}

func TestWorkflowsListingMarksTheOnesThatNeedAGoal(t *testing.T) {
	stdout, _, err := executeCommand("workflows")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		name, _, _ := strings.Cut(line, " ")
		workflow, ok := workflows.Lookup(name)
		if !ok {
			t.Fatalf("listing names an unknown workflow: %q", line)
		}
		if got := strings.Contains(line, "needs --goal"); got != workflow.NeedsGoal {
			t.Errorf("%s: marked needs --goal = %v, want %v", name, got, workflow.NeedsGoal)
		}
	}
}

func TestRunRefusesAGoalHungryWorkflowWithoutOne(t *testing.T) {
	for _, workflow := range workflows.List() {
		if !workflow.NeedsGoal {
			continue
		}
		t.Run(workflow.Name, func(t *testing.T) {
			_, _, err := executeCommand("run", workflow.Name, "--workdir", t.TempDir(), "--logs", t.TempDir())
			if err == nil {
				t.Fatal("expected a refusal")
			}
			if !strings.Contains(err.Error(), "--goal") {
				t.Errorf("refusal does not name the flag: %v", err)
			}
			if !strings.Contains(err.Error(), workflow.GoalHint) {
				t.Errorf("refusal does not show what a goal looks like: %v", err)
			}
		})
	}
}

func TestApplyGoalReplacesTheGoalEveryPromptExpands(t *testing.T) {
	pipeline, source, err := loadPipelineSource("sprint-plan")
	if err != nil {
		t.Fatal(err)
	}
	if err := applyGoal(pipeline, source, "Ship the CSV exporter", true); err != nil {
		t.Fatal(err)
	}
	if pipeline.Goal != "Ship the CSV exporter" {
		t.Errorf("goal = %q", pipeline.Goal)
	}
}

func TestApplyGoalRejectsBlankText(t *testing.T) {
	pipeline, source, err := loadPipelineSource("sprint-plan")
	if err != nil {
		t.Fatal(err)
	}
	if err := applyGoal(pipeline, source, "   ", true); err == nil {
		t.Fatal("expected an error for a blank goal")
	}
}

// A workflow that reads its work from the workspace starts without a goal.
func TestApplyGoalLeavesLedgerDrivenWorkflowsAlone(t *testing.T) {
	pipeline, source, err := loadPipelineSource("sprint-execute")
	if err != nil {
		t.Fatal(err)
	}
	before := pipeline.Goal
	if err := applyGoal(pipeline, source, "", false); err != nil {
		t.Fatal(err)
	}
	if pipeline.Goal != before {
		t.Error("goal changed without --goal")
	}
}
