package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/tylergannon/tractor/graph"
	workflowlib "github.com/tylergannon/tractor/workflow"
)

func TestWorkflowList(t *testing.T) {
	stdout, _, err := executeCommand("workflow", "list")
	if err != nil {
		t.Fatal(err)
	}
	want := "plan\tInterview the caller and write a planning brief, checklist, and size recommendation.\n"
	if stdout != want {
		t.Fatalf("workflow list stdout = %q, want %q", stdout, want)
	}
	if strings.Contains(stdout, workflowlib.MediumName) || strings.Contains(stdout, workflowlib.LargeName) {
		t.Fatalf("workflow list claims unavailable workflows:\n%s", stdout)
	}

	help, _, err := executeCommand("workflow", "--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(help, "tractor workflow list") || !strings.Contains(help, "run") {
		t.Fatalf("workflow help does not lead callers from list to run:\n%s", help)
	}
}

func TestWorkflowRun(t *testing.T) {
	workdir := t.TempDir()
	seed := filepath.Join("seeds", "idea.md")
	writeWorkflowFile(t, filepath.Join(workdir, seed), "Build a small command.\n")
	t.Setenv(interviewDirectoryEnv, "outer-interview")

	called := 0
	runner := func(command *cobra.Command, pipeline graph.Graph, gotWorkdir, logs string, resume bool) error {
		called++
		if gotWorkdir != workdir {
			t.Errorf("runner workdir = %q, want %q", gotWorkdir, workdir)
		}
		if logs != "relative-logs" || resume {
			t.Errorf("runner logs/resume = %q/%t", logs, resume)
		}
		wantInterview := filepath.Join(workdir, "ephemeral", "projects", "demo", "interview")
		if got := os.Getenv(interviewDirectoryEnv); got != wantInterview {
			t.Errorf("%s = %q, want %q", interviewDirectoryEnv, got, wantInterview)
		}
		if pipeline.Name != workflowlib.PlanName || pipeline.Start != "planner" {
			t.Fatalf("materialized graph identity = %#v", pipeline)
		}
		plannerNode, ok := pipeline.NodeByID("planner")
		if !ok {
			t.Fatal("materialized graph has no planner")
		}
		planner, ok := plannerNode.(*graph.CodergenNode)
		if !ok {
			t.Fatalf("planner type = %T", plannerNode)
		}
		for _, value := range []string{filepath.Join(workdir, seed), workdir, "ephemeral/projects/demo"} {
			if !strings.Contains(planner.Prompt.Value, value) {
				t.Errorf("planner prompt does not contain %q", value)
			}
		}
		validatorNode, ok := pipeline.NodeByID("validate")
		if !ok || !strings.Contains(validatorNode.(*graph.ToolNode).ToolCommand, "workflow validate-plan") {
			t.Fatalf("validator node = %#v", validatorNode)
		}
		writePlanHandoff(t, gotWorkdir, "demo", "SIMPLE", "Execute the plan yourself.")
		return nil
	}

	stdout, _, err := executeWorkflowCommand(runner,
		"workflow", "run", "plan", "--project", "demo", "--seed", seed,
		"--workdir", workdir, "--logs", "relative-logs",
	)
	if err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Fatalf("runner calls = %d, want 1", called)
	}
	if got := os.Getenv(interviewDirectoryEnv); got != "outer-interview" {
		t.Fatalf("restored %s = %q", interviewDirectoryEnv, got)
	}
	if !strings.Contains(stdout, filepath.Join(workdir, "ephemeral", "projects", "demo", workflowlib.BriefFile)) {
		t.Fatalf("handoff omits project output:\n%s", stdout)
	}
}

func TestWorkflowRejects(t *testing.T) {
	workdir := t.TempDir()
	writeWorkflowFile(t, filepath.Join(workdir, "seed.md"), "seed\n")
	missingWorkdir := filepath.Join(workdir, "missing-workdir")
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "unknown workflow", args: []string{"workflow", "run", "medium", "--project", "demo", "--seed", "seed.md", "--workdir", workdir, "--logs", "logs"}, wantErr: `unknown built-in workflow "medium"`},
		{name: "missing project", args: []string{"workflow", "run", "plan", "--seed", "seed.md", "--workdir", workdir, "--logs", "logs"}, wantErr: "--project is required"},
		{name: "unsafe project", args: []string{"workflow", "run", "plan", "--project", "../escape", "--seed", "seed.md", "--workdir", workdir, "--logs", "logs"}, wantErr: "invalid --project"},
		{name: "missing seed flag", args: []string{"workflow", "run", "plan", "--project", "demo", "--workdir", workdir, "--logs", "logs"}, wantErr: "--seed is required"},
		{name: "missing seed file", args: []string{"workflow", "run", "plan", "--project", "demo", "--seed", "missing.md", "--workdir", workdir, "--logs", "logs"}, wantErr: "inspect --seed"},
		{name: "seed is directory", args: []string{"workflow", "run", "plan", "--project", "demo", "--seed", ".", "--workdir", workdir, "--logs", "logs"}, wantErr: "--seed"},
		{name: "missing logs", args: []string{"workflow", "run", "plan", "--project", "demo", "--seed", "seed.md", "--workdir", workdir}, wantErr: "--logs is required"},
		{name: "missing workdir", args: []string{"workflow", "run", "plan", "--project", "demo", "--seed", "seed.md", "--workdir", missingWorkdir, "--logs", "logs"}, wantErr: "invalid --workdir"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			called := false
			_, _, err := executeWorkflowCommand(func(*cobra.Command, graph.Graph, string, string, bool) error {
				called = true
				return nil
			}, test.args...)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, test.wantErr)
			}
			if called {
				t.Fatal("invalid input started the pipeline runner")
			}
		})
	}
}

func TestWorkflowHandoff(t *testing.T) {
	workdir := t.TempDir()
	writeWorkflowFile(t, filepath.Join(workdir, "seed.md"), "seed\n")
	runner := func(command *cobra.Command, _ graph.Graph, gotWorkdir, _ string, _ bool) error {
		writePlanHandoff(t, gotWorkdir, "handoff", "SIMPLE", "Execute the plan yourself.")
		_, err := fmt.Fprintln(command.OutOrStdout(), "COMPLETED")
		return err
	}
	stdout, _, err := executeWorkflowCommand(runner,
		"workflow", "run", "plan", "--project", "handoff", "--seed", "seed.md",
		"--workdir", workdir, "--logs", filepath.Join(t.TempDir(), "logs"),
	)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(workdir, "ephemeral", "projects", "handoff")
	want := "COMPLETED\n" +
		"Brief: " + filepath.Join(root, workflowlib.BriefFile) + "\n" +
		"Checklist: " + filepath.Join(root, workflowlib.ChecklistFile) + "\n" +
		"Recommendation: " + filepath.Join(root, workflowlib.RecommendationFile) + "\n" +
		"Size: SIMPLE\n" +
		"Next: Execute the plan yourself.\n"
	if stdout != want {
		t.Fatalf("handoff stdout = %q, want %q", stdout, want)
	}

	validatorOut, _, err := executeCommand("workflow", "validate-plan", "--project", "handoff", "--workdir", workdir)
	if err != nil {
		t.Fatalf("hidden validator bridge: %v", err)
	}
	if validatorOut != "" {
		t.Fatalf("hidden validator stdout = %q", validatorOut)
	}
}

func executeWorkflowCommand(run pipelineRunner, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := &cobra.Command{Use: "tractor", SilenceErrors: true, SilenceUsage: true}
	command.AddCommand(newWorkflowCommand(run))
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs(args)
	err := command.Execute()
	return stdout.String(), stderr.String(), err
}

func writePlanHandoff(t *testing.T, workdir, project, size, next string) {
	t.Helper()
	root := filepath.Join(workdir, "ephemeral", "projects", project)
	writeWorkflowFile(t, filepath.Join(root, workflowlib.BriefFile), "# Brief\n\nA bounded plan.\n")
	writeWorkflowFile(t, filepath.Join(root, workflowlib.ChecklistFile), "---\nitems:\n  - name: implement\n    check: The requested behavior works.\n    command: go test ./...\n---\n")
	writeWorkflowFile(t, filepath.Join(root, workflowlib.RecommendationFile), "# Recommendation\nSize: "+size+"\nRationale: The scope fits.\nNext: "+next+"\n")
}

func writeWorkflowFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
