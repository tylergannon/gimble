package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tylergannon/tractor/engine"
	"github.com/tylergannon/tractor/graph"
)

const linearPipeline = `{"name":"linear","start":"done","nodes":[{"id":"done","type":"command","command":"true","edges":{"success":"success"}}]}`

const linearYAML = `name: linear
start: done
nodes:
  - id: done
    type: command
    command: "true"
    edges:
      success: success
`

func TestRootExposesOnlyRequestedCommands(t *testing.T) {
	stdout, _, err := executeCommand("--help")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout, "completion") {
		t.Fatalf("help exposes unrequested completion command:\n%s", stdout)
	}
	for _, command := range []string{"answer", "ask", "inspect-models", "mcp", "print-schema", "run", "validate"} {
		if !strings.Contains(stdout, command) {
			t.Fatalf("help omits %q:\n%s", command, stdout)
		}
	}
}

func TestInspectModelsReportsEffectiveSelectionsForEveryRole(t *testing.T) {
	pipeline := `defaults:
  model:
    name: fable
    version: "5"
    effort: low
start: items
nodes:
  - id: items
    type: loop
    checklist: work.md
    item_judge:
      model:
        name: flash
        version: "3.7"
        effort: high
    edges:
      loop: implement
      exit: success
  - id: implement
    type: agent
    prompt: Implement the current item.
    edges:
      - to: items
`
	stdout, _, err := executeCommand("inspect-models", "--yaml", pipeline)
	if err != nil {
		t.Fatal(err)
	}
	var resolutions []engine.ModelResolution
	if err := json.Unmarshal([]byte(stdout), &resolutions); err != nil {
		t.Fatalf("decode inspection: %v\n%s", err, stdout)
	}
	byRole := make(map[string]engine.ModelResolution, len(resolutions))
	for _, resolution := range resolutions {
		byRole[resolution.NodeID+"/"+resolution.Role] = resolution
	}
	if got := byRole["implement/agent"]; got.NativeModel != "claude-fable-5" || got.AuthoredVersion != "5" || got.EffectiveEffort != "low" || got.Source != "pipeline defaults.model" {
		t.Fatalf("agent resolution = %+v", got)
	}
	if got := byRole["items/item_judge"]; got.NativeModel != "gemini-3.7-flash-high" || got.AuthoredVersion != "3.7" || got.Provider != "gemini" || got.Harness != "agy" {
		t.Fatalf("judge resolution = %+v", got)
	}
	if got := byRole["items/goal_evaluator"]; got.NativeModel != "claude-fable-5" || got.Source != "pipeline defaults.model" {
		t.Fatalf("evaluator resolution = %+v", got)
	}
}

func TestValidateRequiresExactlyOneExplicitPipelineSource(t *testing.T) {
	file := filepath.Join(t.TempDir(), "pipeline.json")
	if err := os.WriteFile(file, []byte(linearPipeline), 0o644); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		args       []string
		wantError  string
		wantOutput string
	}{
		{name: "missing", args: []string{"validate"}, wantError: "pipeline source is required"},
		{name: "two files", args: []string{"validate", file, file}, wantError: "exactly one pipeline source"},
		{name: "file and json", args: []string{"validate", file, "--json", linearPipeline}, wantError: "mutually exclusive"},
		{name: "file and yaml", args: []string{"validate", file, "--yaml", linearYAML}, wantError: "mutually exclusive"},
		{name: "json and yaml", args: []string{"validate", "--json", linearPipeline, "--yaml", linearYAML}, wantError: "mutually exclusive"},
		{name: "explicit empty json", args: []string{"validate", "--json", ""}, wantError: "parse pipeline"},
		{name: "explicit empty yaml", args: []string{"validate", "--yaml", ""}, wantError: "parse pipeline YAML"},
		{name: "file", args: []string{"validate", file}, wantOutput: "valid " + file + "\n"},
		{name: "json", args: []string{"validate", "--json", linearPipeline}, wantOutput: "valid --json\n"},
		{name: "yaml", args: []string{"validate", "--yaml", linearYAML}, wantOutput: "valid --yaml\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, _, err := executeCommand(test.args...)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want containing %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if stdout != test.wantOutput {
				t.Fatalf("stdout = %q, want %q", stdout, test.wantOutput)
			}
		})
	}
}

func TestValidateDetectsYAMLFileExtensions(t *testing.T) {
	for _, extension := range []string{".yaml", ".yml", ".YAML"} {
		file := filepath.Join(t.TempDir(), "pipeline"+extension)
		if err := os.WriteFile(file, []byte(linearYAML), 0o644); err != nil {
			t.Fatal(err)
		}
		stdout, _, err := executeCommand("validate", file)
		if err != nil {
			t.Fatal(err)
		}
		if stdout != "valid "+file+"\n" {
			t.Fatalf("stdout = %q", stdout)
		}
	}
}

func TestValidateReturnsLintFailure(t *testing.T) {
	invalid := `{"start":"work","nodes":[{"id":"work","type":"agent"}]}`
	_, _, err := executeCommand("validate", "--json", invalid)
	if err == nil || !strings.Contains(err.Error(), "pipeline validation failed") {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateSurfacesWarningsAndPreservesSuccessOutput(t *testing.T) {
	pipeline := `{"start":"work","nodes":[` +
		`{"id":"work","type":"agent","edges":[{"to":"success"}]}]}`
	stdout, stderr, err := executeCommand("validate", "--json", pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "valid --json\n" {
		t.Fatalf("stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "warning prompt_on_llm_nodes:") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestPrintSchemaEmitsGraphSchemaExactly(t *testing.T) {
	stdout, _, err := executeCommand("print-schema")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(stdout), graph.Graph{}.Schema()) {
		t.Fatal("print-schema output differs from graph.Graph schema")
	}
	if _, _, err := executeCommand("print-schema", "extra"); err == nil {
		t.Fatal("print-schema accepted an argument")
	}
}

func TestRunFreshThenResumeWithRealBackendWiring(t *testing.T) {
	workdir := t.TempDir()
	logsRoot := filepath.Join(t.TempDir(), "run")
	stdout, _, err := executeCommand("run", "--json", linearPipeline, "--workdir", workdir, "--logs", logsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "COMPLETED\n" {
		t.Fatalf("fresh stdout = %q", stdout)
	}
	checkpoint, err := engine.LoadCheckpoint(logsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.CurrentNode != "done" || checkpoint.NextNode != graph.Success {
		t.Fatalf("checkpoint = %#v", checkpoint)
	}

	stdout, _, err = executeCommand("run", "--resume", "--json", linearPipeline, "--workdir", workdir, "--logs", logsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "COMPLETED\n" {
		t.Fatalf("resume stdout = %q", stdout)
	}
}

func TestRunReturnsFailedRunAsCommandError(t *testing.T) {
	pipeline := `{"start":"check","nodes":[` +
		`{"id":"check","type":"command","command":"exit 7","edges":{"success":"success"}}]}`
	stdout, _, err := executeCommand(
		"run", "--json", pipeline, "--workdir", t.TempDir(), "--logs", filepath.Join(t.TempDir(), "run"),
	)
	if err == nil || !strings.Contains(err.Error(), "pipeline failed: exit 7 exited 7") {
		t.Fatalf("error = %v", err)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestRunSurfacesWarningsBeforeExecution(t *testing.T) {
	pipeline := `{"start":"fanout","nodes":[` +
		`{"id":"fanout","type":"fan_out","branches":["left","right"]},` +
		`{"id":"left","type":"command","command":"true","max_visits":1,"edges":{"success":"join"}},` +
		`{"id":"right","type":"command","command":"true","edges":{"success":"join"}},` +
		`{"id":"join","type":"fan_in","prompt":"evaluate","edges":[{"to":"success"}]}]}`
	_, stderr, err := executeCommand(
		"run", "--json", pipeline, "--workdir", t.TempDir(), "--logs", filepath.Join(t.TempDir(), "run"),
	)
	if err == nil || !strings.Contains(err.Error(), "pipeline failed:") {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(stderr, "warning branch_root_max_visits:") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestRunRequiresLogsAndResumeCheckpoint(t *testing.T) {
	workdir := t.TempDir()
	if _, _, err := executeCommand("run", "--json", linearPipeline, "--workdir", workdir); err == nil || err.Error() != "--logs is required" {
		t.Fatalf("missing logs error = %v", err)
	}
	if _, _, err := executeCommand(
		"run", "--resume", "--json", linearPipeline, "--workdir", workdir, "--logs", filepath.Join(t.TempDir(), "missing"),
	); err == nil || !strings.Contains(err.Error(), "read checkpoint") {
		t.Fatalf("missing checkpoint error = %v", err)
	}
}

func TestResolveHarnessUsesExecutionProviderDetectionAndSystemModel(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		want     string
		wantErr  string
	}{
		{model: "claude-opus-4-6", want: "claude"},
		{model: "fable", want: "claude"},
		{provider: "anthropic", model: "fable-5.1", want: "claude"},
		{model: "gpt-5.6-sol", want: "codex"},
		{want: "codex"},
		{model: "gemini-2.5-pro", want: "agy"},
		{provider: "openai", model: "fable", wantErr: "conflicts with model"},
	}
	for _, test := range tests {
		got, err := resolveHarness(test.provider, test.model)
		if test.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Errorf("resolveHarness(%q, %q) error = %v", test.provider, test.model, err)
			}
			continue
		}
		if err != nil || got != test.want {
			t.Errorf("resolveHarness(%q, %q) = %q, %v; want %q", test.provider, test.model, got, err, test.want)
		}
	}
	if defaultModel != "gpt-5.6-sol" || defaultReasoningEffort != "high" {
		t.Fatalf("system defaults = model %q reasoning %q", defaultModel, defaultReasoningEffort)
	}
}

func executeCommand(args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := newRootCommand()
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs(args)
	err := command.Execute()
	return stdout.String(), stderr.String(), err
}
