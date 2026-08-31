package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
mode: delivery
start: silence_boundary
proof_contract:
  intended_architecture: Input crosses the operating boundary and produces a user-visible result.
  primary_outcome: Representative input produces the expected observable result.
  primary_cases:
    - id: qualifying_input
      actor: A user
      job: Submit qualifying input and observe the promised result.
      node: primary_assertion
      qualifying_input_criteria: Input independently established to qualify.
      expected_output: The expected value is visible at the operating boundary.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
      evidence_source: current_run
      independence: independent_execution
      status: unproven
      required_capabilities: []
      evidence_artifacts: []
  boundary_cases:
    - id: silence_boundary
      actor: A user
      job: Submit empty input without receiving invented output.
      node: silence_boundary
      qualifying_input_criteria: Input is empty.
      expected_output: The observed value remains empty.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
      evidence_source: current_run
      independence: independent_execution
      status: unproven
      required_capabilities: []
      evidence_artifacts: []
  required_capabilities: []
  unknowns: []
  scope_gaps: []
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
		_, boundaryExists := checkpoint.ProofEvidence["silence_boundary"]
		if checkpoint.NextNode != "silence_boundary" || !checkpoint.RetryVisit ||
			!boundaryExists || len(checkpoint.ProofEvidence) != 1 {
			t.Fatalf("checkpoint = %#v", checkpoint)
		}
	})

	t.Run("independent primary and boundary assertions complete", func(t *testing.T) {
		pipeline := parseProofPipeline(t, `
mode: delivery
start: primary_assertion
proof_contract:
  intended_architecture: Input crosses the operating boundary and produces a user-visible result.
  primary_outcome: Representative input produces the expected observable result.
  primary_cases:
    - id: qualifying_input
      actor: A user
      job: Submit qualifying input and observe the promised result.
      node: primary_assertion
      qualifying_input_criteria: Input independently established to qualify.
      expected_output: The expected value is visible at the operating boundary.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
      evidence_source: current_run
      independence: independent_execution
      status: unproven
      required_capabilities: []
      evidence_artifacts: []
  boundary_cases:
    - id: silence_boundary
      actor: A user
      job: Submit empty input without receiving invented output.
      node: silence_boundary
      qualifying_input_criteria: Input is empty.
      expected_output: The observed value remains empty.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
      evidence_source: current_run
      independence: independent_execution
      status: unproven
      required_capabilities: []
      evidence_artifacts: []
  required_capabilities: []
  unknowns: []
  scope_gaps: []
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
		_, primaryExists := checkpoint.ProofEvidence["qualifying_input"]
		_, boundaryExists := checkpoint.ProofEvidence["silence_boundary"]
		if !primaryExists || !boundaryExists || len(checkpoint.ProofEvidence) != 2 {
			t.Fatalf("proof evidence = %v", checkpoint.ProofEvidence)
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
	if len(checkpoint.ProofEvidence) != 0 {
		t.Fatalf("legacy checkpoint gained proof state: %v", checkpoint.ProofEvidence)
	}
}

func TestExplicitDiscoveryCompletesLearningWithoutClaimingDelivery(t *testing.T) {
	pipeline := parseProofPipeline(t, `
mode: discovery
start: explore
nodes:
  - id: explore
    type: tool
    tool_command: "true"
    on_success: success
`)
	workdir := t.TempDir()
	logsRoot := t.TempDir()
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
	if result.Status != RunLearningCompleted || result.FailureReason != "" {
		t.Fatalf("result = %#v", result)
	}
	events, err := readTimeline(filepath.Join(logsRoot, "timeline.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if got := events[len(events)-1]["type"]; got != "PipelineLearningCompleted" {
		t.Fatalf("last timeline event = %v", got)
	}
	checkpoint := mustCheckpoint(t, logsRoot)
	if checkpoint.NextNode != graph.Success {
		t.Fatalf("checkpoint = %#v", checkpoint)
	}

	resumed, err := ResumeRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: logsRoot,
		Workdir:  workdir,
		Validate: func(graph.Graph) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err = resumed.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunLearningCompleted {
		t.Fatalf("resumed result = %#v", result)
	}
}

func TestFailedProofAssertionRoutesToFixBeforeItCanPass(t *testing.T) {
	pipeline := parseProofPipeline(t, `
mode: delivery
start: primary_assertion
proof_contract:
  intended_architecture: A repair step establishes the capability before the assertion passes.
  primary_outcome: The repaired workspace exposes the required marker.
  primary_cases:
    - id: repaired_marker
      actor: A user
      job: Observe the promised marker after repair.
      node: primary_assertion
      qualifying_input_criteria: The workspace initially lacks ready.txt.
      expected_output: ready.txt exists after the repair step.
      correctness_oracle: tool_exit_zero
      evidence_mode: component
      evidence_source: current_run
      independence: independent_execution
      status: blocked
      required_capabilities: []
      evidence_artifacts: []
  boundary_cases: []
  required_capabilities: []
  unknowns: []
  scope_gaps: []
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
	_, proofExists := checkpoint.ProofEvidence["repaired_marker"]
	if !reflect.DeepEqual(checkpoint.CompletedNodes, []string{"primary_assertion", "fix", "primary_assertion"}) ||
		!proofExists || len(checkpoint.ProofEvidence) != 1 {
		t.Fatalf("checkpoint = %#v", checkpoint)
	}
}

func TestCompletedNodeRecordCannotSubstituteForProofExecution(t *testing.T) {
	pipeline := parseProofPipeline(t, `
mode: delivery
start: primary_assertion
proof_contract:
  intended_architecture: Input reaches a deterministic assertion at the product boundary.
  primary_outcome: Representative input produces the expected observable result.
  primary_cases:
    - id: qualifying_input
      actor: A user
      job: Submit qualifying input and observe the promised result.
      node: primary_assertion
      qualifying_input_criteria: Input independently established to qualify.
      expected_output: The expected value is visible.
      correctness_oracle: tool_exit_zero
      evidence_mode: operating_layer
      evidence_source: current_run
      independence: independent_execution
      status: unproven
      required_capabilities: []
      evidence_artifacts: []
  boundary_cases: []
  required_capabilities: []
  unknowns: []
  scope_gaps: []
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

func TestSelfAuthoredProofRecordCannotSubstituteForEngineEvidence(t *testing.T) {
	pipeline := singlePromiseGraph(graph.WorkflowModeDelivery)
	root := t.TempDir()
	runID := "self-authored-run"
	store, err := openRunStore(root, newEngineState())
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(root, "manifest.json"), runManifest{ID: runID}); err != nil {
		t.Fatal(err)
	}
	contractSHA256, err := proofCaseDigest(
		pipeline.ProofContract.Value.PrimaryCases[0], pipeline.Nodes[0].(*graph.ToolNode),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.saveCheckpoint(Checkpoint{
		CurrentNode:    "primary_assertion",
		NextNode:       graph.Success,
		CompletedNodes: []string{"primary_assertion"},
		NodeVisits:     map[string]int{"primary_assertion": 1},
		NodeAttempts:   map[string]int{"primary_assertion": 1},
		ProofEvidence: map[string]ProofEvidenceRecord{"primary": {
			RunID: runID, CaseID: "primary",
			NodeID: "primary_assertion", ExecutionRef: "stages/000001-primary_assertion",
			ContractSHA256: contractSHA256,
			Edge:           ProofEdgeRef{From: "primary_assertion", To: graph.Success},
		}},
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
	if result.Status != RunFailed || !strings.Contains(result.FailureReason, "invalid execution reference") {
		t.Fatalf("result = %#v", result)
	}
}

func TestDiscoveryLearningAndDeliveryScopeGapSemantics(t *testing.T) {
	t.Run("discovery prototype", func(t *testing.T) {
		pipeline := singlePromiseGraph(graph.WorkflowModeDiscovery)
		pipeline.ProofContract.Value.PrimaryCases[0].Status = graph.ProofStatusSimulated
		result, _ := runProofPipeline(t, pipeline, func(string) {})
		if result.Status != RunLearningCompleted || result.FailureReason != "" {
			t.Fatalf("result = %#v", result)
		}
	})

	t.Run("structured material gap", func(t *testing.T) {
		pipeline := singlePromiseGraph(graph.WorkflowModeDelivery)
		pipeline.ProofContract.Value.ScopeGaps = []graph.ProofScopeGap{{
			PromiseID: "primary", OriginalPromise: "A user observes the expected value.",
			CurrentProvenBehavior: "Only setup is available.", MissingCapability: "The primary operating behavior.",
			Impact: "The user's job remains unmet.", RecommendedNextMove: "Implement and rerun the primary proof.",
		}}
		result, _ := runProofPipeline(t, pipeline, func(string) {})
		if result.Status != RunFailed || !strings.Contains(result.FailureReason, `scope gap for promise "primary"`) {
			t.Fatalf("result = %#v", result)
		}
	})
}

func TestProofEvidenceCarriesSameRunArtifactAndEdgeProvenance(t *testing.T) {
	pipeline := singlePromiseGraph(graph.WorkflowModeDelivery)
	pipeline.Nodes[0].(*graph.ToolNode).ToolCommand = `observed=$(cat input.txt); test "$observed" = expected; printf '%s\n' "$observed" > output.txt`
	pipeline.ProofContract.Value.PrimaryCases[0].EvidenceArtifacts = []graph.ProofArtifactRequirement{
		{Path: "input.txt", Role: graph.ProofArtifactInput, ArchitectureEdge: "caller_to_product"},
		{Path: "output.txt", Role: graph.ProofArtifactOutput, ArchitectureEdge: "product_to_user"},
	}
	result, checkpoint := runProofPipeline(t, pipeline, func(workdir string) {
		writeFixture(t, workdir, "input.txt", "expected\n")
	})
	if result.Status != RunCompleted {
		t.Fatalf("result = %#v", result)
	}
	raw, err := json.Marshal(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	var persisted struct {
		ProofEvidence map[string]struct {
			RunID          string `json:"run_id"`
			NodeID         string `json:"node_id"`
			ExecutionRef   string `json:"execution_ref"`
			ContractSHA256 string `json:"contract_sha256"`
			OutcomeSHA256  string `json:"outcome_sha256"`
			ToolLogSHA256  string `json:"tool_log_sha256"`
			Edge           struct {
				From string `json:"from"`
				To   string `json:"to"`
			} `json:"edge"`
			ArtifactSHA256 []string `json:"artifact_sha256"`
		} `json:"proof_evidence"`
	}
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	evidence, exists := persisted.ProofEvidence["primary"]
	if !exists || evidence.RunID == "" || evidence.NodeID != "primary_assertion" ||
		evidence.ExecutionRef == "" || evidence.Edge.From != "primary_assertion" || evidence.Edge.To != graph.Success {
		t.Fatalf("proof evidence = %#v", evidence)
	}
	if evidence.ContractSHA256 == "" || evidence.OutcomeSHA256 == "" || evidence.ToolLogSHA256 == "" {
		t.Fatalf("proof evidence is missing a digest: %#v", evidence)
	}
	wantHash := sha256.Sum256([]byte("expected\n"))
	wantSHA := hex.EncodeToString(wantHash[:])
	if !reflect.DeepEqual(evidence.ArtifactSHA256, []string{wantSHA, wantSHA}) {
		t.Fatalf("artifact digests = %#v", evidence.ArtifactSHA256)
	}
	if strings.Contains(string(raw), `"status"`) || strings.Contains(string(raw), `"evidence_mode"`) ||
		strings.Contains(string(raw), `"source"`) || strings.Contains(string(raw), `"role"`) ||
		strings.Contains(string(raw), `"stored_path"`) || strings.Contains(string(raw), `"captured_at"`) ||
		strings.Contains(string(raw), `"architecture_edge"`) {
		t.Fatalf("checkpoint repeats contract-derived provenance: %s", raw)
	}
}

func TestChangedProofEvidenceInvalidatesTerminalResume(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*testing.T, *graph.Graph, string, ProofEvidenceRecord)
		wantReason string
	}{
		{"contract", func(_ *testing.T, pipeline *graph.Graph, _ string, _ ProofEvidenceRecord) {
			pipeline.ProofContract.Value.PrimaryCases[0].ExpectedOutput = "A substituted expected value."
		}, "does not match the current contract"},
		{"assertion", func(_ *testing.T, pipeline *graph.Graph, _ string, _ ProofEvidenceRecord) {
			pipeline.Nodes[0].(*graph.ToolNode).ToolCommand = "false"
		}, "does not match the current contract"},
		{"declared snapshot", func(t *testing.T, _ *graph.Graph, root string, e ProofEvidenceRecord) {
			changeProofArtifact(t, root, filepath.Join(filepath.FromSlash(e.ExecutionRef), "proof-evidence", "primary", "000-input.snapshot"))
		}, "missing or changed evidence artifact"},
		{"engine outcome", func(t *testing.T, _ *graph.Graph, root string, e ProofEvidenceRecord) {
			changeProofArtifact(t, root, filepath.Join(filepath.FromSlash(e.ExecutionRef), "outcome.json"))
		}, "missing or changed engine-owned outcome.json"},
		{"tool log", func(t *testing.T, _ *graph.Graph, root string, e ProofEvidenceRecord) {
			changeProofArtifact(t, root, filepath.Join(filepath.FromSlash(e.ExecutionRef), "tool.log"))
		}, "missing or changed engine-owned tool.log"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pipeline := singlePromiseGraph(graph.WorkflowModeDelivery)
			pipeline.ProofContract.Value.PrimaryCases[0].EvidenceArtifacts = []graph.ProofArtifactRequirement{
				{Path: "input.txt", Role: graph.ProofArtifactInput, ArchitectureEdge: "caller_to_product"},
			}
			workdir := t.TempDir()
			logsRoot := t.TempDir()
			writeFixture(t, workdir, "input.txt", "expected\n")
			runner, err := NewRunner(pipeline, NewRegistry(), RunnerConfig{
				LogsRoot: logsRoot,
				Workdir:  workdir,
				Validate: func(graph.Graph) error { return nil },
			})
			if err != nil {
				t.Fatal(err)
			}
			result, err := runner.Run()
			if err != nil || result.Status != RunCompleted {
				t.Fatalf("run result = %#v, err = %v", result, err)
			}
			checkpoint := mustCheckpoint(t, logsRoot)
			evidence := checkpoint.ProofEvidence["primary"]
			if len(evidence.ArtifactSHA256) == 0 {
				t.Fatal("proof evidence has no artifacts")
			}
			test.mutate(t, &pipeline, logsRoot, evidence)
			resumed, err := ResumeRunner(pipeline, NewRegistry(), RunnerConfig{
				LogsRoot: logsRoot,
				Workdir:  workdir,
				Validate: func(graph.Graph) error { return nil },
			})
			if err != nil {
				t.Fatal(err)
			}
			result, err = resumed.Run()
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != RunFailed || !strings.Contains(result.FailureReason, test.wantReason) {
				t.Fatalf("resume result = %#v", result)
			}
		})
	}
}

func changeProofArtifact(t *testing.T, root, relative string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, relative), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func singlePromiseGraph(mode graph.WorkflowMode) graph.Graph {
	return graph.Graph{
		Mode: optional(mode), Start: "primary_assertion",
		ProofContract: optional(graph.ProofContract{
			IntendedArchitecture: "Input crosses a deterministic assertion and produces a user-visible result.",
			PrimaryOutcome:       "A qualifying input produces the expected value.",
			PrimaryCases: []graph.ProofCase{{
				ID: "primary", Actor: "A user", Job: "Submit qualifying input and observe the expected value.",
				Node: "primary_assertion", QualifyingInputCriteria: "Input meets the declared product criteria.",
				ExpectedOutput: "The expected value is observable.", CorrectnessOracle: graph.ProofOracleToolExitZero,
				EvidenceMode: graph.ProofEvidenceOperatingLayer, EvidenceSource: graph.ProofEvidenceCurrentRun,
				Independence: graph.ProofIndependentExecution, Status: graph.ProofStatusUnproven,
				RequiredCapabilities: []string{}, EvidenceArtifacts: []graph.ProofArtifactRequirement{},
			}},
			BoundaryCases: []graph.ProofCase{}, RequiredCapabilities: []graph.ProofCapability{},
			Unknowns: []string{}, ScopeGaps: []graph.ProofScopeGap{},
			TerminalSuccess: graph.ProofTerminalSuccess{RequiredCases: []string{"primary"}},
		}),
		Nodes: []graph.Node{
			&graph.ToolNode{NodeBase: graph.NodeBase{ID: "primary_assertion"}, ToolCommand: "true", OnSuccess: graph.Success},
		},
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
