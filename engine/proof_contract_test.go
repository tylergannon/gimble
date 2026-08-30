package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tylergannon/tractor/graph"
)

func TestProofContractRequiresPrimaryAndBoundaryToolExecutions(t *testing.T) {
	t.Run("boundary-only evidence cannot complete", func(t *testing.T) {
		pipeline := parseProofPipeline(t, `
start: silence_boundary
proof_contract:
  primary_outcome: Representative input produces the expected observable result.
  primary_cases:
    - id: qualifying_input
      node: primary_assertion
      representative_input: A known qualifying fixture.
      expected_output: The expected value is visible at the operating boundary.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
  boundary_cases:
    - id: silence_boundary
      node: silence_boundary
      representative_input: An empty fixture.
      expected_output: The observed value remains empty.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
  unknowns: []
  terminal_success:
    required_cases: [qualifying_input, silence_boundary]
nodes:
  - id: silence_boundary
    type: tool
    tool_command: test ! -s silence.txt
    on_success: success
  - id: primary_assertion
    type: tool
    tool_command: test "$(cat qualifying.txt)" = expected-words
    on_success: success
`)
		result, checkpoint := runProofPipeline(t, pipeline, func(workdir string) {
			writeFixture(t, workdir, "silence.txt", "")
			writeFixture(t, workdir, "qualifying.txt", "expected-words\n")
		})
		if result.Status != RunFailed || !strings.Contains(result.FailureReason, `required proof case "qualifying_input" has not passed`) {
			t.Fatalf("result = %#v", result)
		}
		if checkpoint.NextNode != "silence_boundary" || !checkpoint.RetryVisit ||
			!reflect.DeepEqual(checkpoint.PassedProofCases, map[string]bool{"silence_boundary": true}) {
			t.Fatalf("checkpoint = %#v", checkpoint)
		}
	})

	t.Run("independent primary and boundary assertions complete", func(t *testing.T) {
		pipeline := parseProofPipeline(t, `
start: primary_assertion
proof_contract:
  primary_outcome: Representative input produces the expected observable result.
  primary_cases:
    - id: qualifying_input
      node: primary_assertion
      representative_input: A known qualifying fixture.
      expected_output: The expected value is visible at the operating boundary.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
  boundary_cases:
    - id: silence_boundary
      node: silence_boundary
      representative_input: An empty fixture.
      expected_output: The observed value remains empty.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
  unknowns: []
  terminal_success:
    required_cases: [qualifying_input, silence_boundary]
nodes:
  - id: primary_assertion
    type: tool
    tool_command: test "$(cat qualifying.txt)" = expected-words
    on_success: silence_boundary
  - id: silence_boundary
    type: tool
    tool_command: test ! -s silence.txt
    on_success: success
`)
		result, checkpoint := runProofPipeline(t, pipeline, func(workdir string) {
			writeFixture(t, workdir, "qualifying.txt", "expected-words\n")
			writeFixture(t, workdir, "silence.txt", "")
		})
		if result.Status != RunCompleted || result.FailureReason != "" {
			t.Fatalf("result = %#v", result)
		}
		if !reflect.DeepEqual(checkpoint.PassedProofCases, map[string]bool{
			"qualifying_input": true,
			"silence_boundary": true,
		}) {
			t.Fatalf("passed proof cases = %v", checkpoint.PassedProofCases)
		}
	})
}

func TestRunnerWithoutProofContractRemainsCompatible(t *testing.T) {
	pipeline := parseProofPipeline(t, `
start: existing_check
nodes:
  - id: existing_check
    type: tool
    tool_command: test -f existing.txt
    on_success: success
`)
	result, checkpoint := runProofPipeline(t, pipeline, func(workdir string) {
		writeFixture(t, workdir, "existing.txt", "existing workflow\n")
	})
	if result.Status != RunCompleted || result.FailureReason != "" {
		t.Fatalf("result = %#v", result)
	}
	if len(checkpoint.PassedProofCases) != 0 {
		t.Fatalf("legacy checkpoint gained proof state: %v", checkpoint.PassedProofCases)
	}
}

func TestFailedProofAssertionRoutesToFixBeforeItCanPass(t *testing.T) {
	pipeline := parseProofPipeline(t, `
start: primary_assertion
proof_contract:
  primary_outcome: The repaired workspace exposes the required marker.
  primary_cases:
    - id: repaired_marker
      node: primary_assertion
      representative_input: A workspace without ready.txt.
      expected_output: ready.txt exists after the repair step.
      correctness_oracle: tool_exit_zero
      evidence_mode: component
  boundary_cases: []
  unknowns: []
  terminal_success:
    required_cases: [repaired_marker]
nodes:
  - id: primary_assertion
    type: tool
    tool_command: test -f ready.txt
    on_success: success
    on_error: fix
  - id: fix
    type: tool
    tool_command: touch ready.txt
    on_success: primary_assertion
`)
	result, checkpoint := runProofPipeline(t, pipeline, func(string) {})
	if result.Status != RunCompleted {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(checkpoint.CompletedNodes, []string{"primary_assertion", "fix", "primary_assertion"}) ||
		!reflect.DeepEqual(checkpoint.PassedProofCases, map[string]bool{"repaired_marker": true}) {
		t.Fatalf("checkpoint = %#v", checkpoint)
	}
}

func TestCompletedNodeRecordCannotSubstituteForProofExecution(t *testing.T) {
	pipeline := parseProofPipeline(t, `
start: primary_assertion
proof_contract:
  primary_outcome: Representative input produces the expected observable result.
  primary_cases:
    - id: qualifying_input
      node: primary_assertion
      representative_input: A known qualifying fixture.
      expected_output: The expected value is visible.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
  boundary_cases: []
  unknowns: []
  terminal_success:
    required_cases: [qualifying_input]
nodes:
  - id: primary_assertion
    type: tool
    tool_command: "true"
    on_success: success
`)
	root := t.TempDir()
	store, err := openRunStore(root, newEngineState())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.saveCheckpoint(Checkpoint{
		CurrentNode:    "primary_assertion",
		NextNode:       graph.Success,
		CompletedNodes: []string{"primary_assertion"},
		NodeVisits:     map[string]int{"primary_assertion": 1},
		NodeAttempts:   map[string]int{"primary_assertion": 1},
	}); err != nil {
		t.Fatal(err)
	}
	runner, err := ResumeRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: root,
		Workdir:  t.TempDir(),
		Validate: func(graph.Graph) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunFailed || !strings.Contains(result.FailureReason, `required proof case "qualifying_input" has not passed`) {
		t.Fatalf("result = %#v", result)
	}
}

func parseProofPipeline(t *testing.T, document string) graph.Graph {
	t.Helper()
	pipeline, err := graph.ParseYAML([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	return *pipeline
}

func runProofPipeline(t *testing.T, pipeline graph.Graph, setup func(string)) (RunResult, Checkpoint) {
	t.Helper()
	workdir := t.TempDir()
	logsRoot := t.TempDir()
	setup(workdir)
	runner, err := NewRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: logsRoot,
		Workdir:  workdir,
		Validate: func(graph.Graph) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	return result, mustCheckpoint(t, logsRoot)
}

func writeFixture(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
