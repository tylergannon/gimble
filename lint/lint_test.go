package lint_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/lint"
)

func TestEveryBuiltInRule(t *testing.T) {
	tests := []struct {
		rule     string
		severity lint.Severity
		graph    func() graph.Graph
		options  lint.Options
	}{
		{"start_target", lint.SeverityError, func() graph.Graph { g := validLinear(); g.Start = "missing"; return g }, lint.Options{}},
		{"terminal_reachable", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			coder(g, "work").Edges = []graph.Edge{edge(graph.Failure)}
			return g
		}, lint.Options{}},
		{"reachability", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			g.Nodes = append(g.Nodes, codergen("orphan", edge(graph.Success)))
			return g
		}, lint.Options{}},
		{"edge_target_exists", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			coder(g, "work").Edges = []graph.Edge{edge("missing")}
			return g
		}, lint.Options{}},
		{"edge_target_unique", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			coder(g, "work").Edges = []graph.Edge{condition("success", "a"), condition("success", "b")}
			return g
		}, lint.Options{}},
		{"dead_end", lint.SeverityError, func() graph.Graph { g := validLinear(); coder(g, "work").Edges = nil; return g }, lint.Options{}},
		{"parallel_fan_in", lint.SeverityError, func() graph.Graph {
			g := validParallel()
			coder(g, "left").Edges = []graph.Edge{edge(graph.Success)}
			return g
		}, lint.Options{}},
		{"branch_disjoint", lint.SeverityError, overlappingParallel, lint.Options{}},
		{"no_nested_parallel", lint.SeverityError, nestedParallel, lint.Options{}},
		{"fan_in_single_parallel", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			g.Nodes = append(g.Nodes, fanIn("unowned", edge(graph.Success)))
			return g
		}, lint.Options{}},
		{"parallel_thread_disjoint", lint.SeverityError, func() graph.Graph {
			g := validParallel()
			llm(g, "left").ThreadID = set("shared")
			llm(g, "right").ThreadID = set("shared")
			return g
		}, lint.Options{}},
		{"thread_branch_boundary", lint.SeverityError, func() graph.Graph {
			g := validParallel()
			llm(g, "left").ThreadID = set("shared")
			llm(g, "join").ThreadID = set("shared")
			return g
		}, lint.Options{}},
		{"fan_in_entry", lint.SeverityError, directFanInParallel, lint.Options{}},
		{"branch_entry", lint.SeverityError, externalBranchEntry, lint.Options{}},
		{"max_visits_positive", lint.SeverityError, func() graph.Graph { g := validLinear(); coder(g, "work").MaxVisits = set(0); return g }, lint.Options{}},
		{"max_parallel_positive", lint.SeverityError, func() graph.Graph { g := validParallel(); parallel(g).MaxParallel = set(0); return g }, lint.Options{}},
		{"max_retries_nonnegative", lint.SeverityError, func() graph.Graph { g := validLinear(); coder(g, "work").MaxRetries = set(-1); return g }, lint.Options{}},
		{"edge_condition_missing", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			g.Nodes = append(g.Nodes, codergen("other", edge(graph.Success)))
			coder(g, "work").Edges = []graph.Edge{edge("other"), condition(graph.Success, "done")}
			return g
		}, lint.Options{}},
		{"supervises_valid", lint.SeverityError, func() graph.Graph { g := validLinear(); g.Nodes = append(g.Nodes, supervisor("coach")); return g }, lint.Options{}},
		{"supervisor_not_targeted", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			g.Nodes = append(g.Nodes, supervisor("coach", "work"))
			coder(g, "work").Edges = []graph.Edge{edge("coach")}
			return g
		}, lint.Options{}},
		{"supervisor_cycle", lint.SeverityError, supervisedCycle, lint.Options{}},
		{"fidelity_valid", lint.SeverityError, func() graph.Graph { g := validLinear(); coder(g, "work").Fidelity = set("summary"); return g }, lint.Options{}},
		{"thread_id_collision", lint.SeverityError, func() graph.Graph { g := validLinear(); coder(g, "work").ThreadID = set("work"); return g }, lint.Options{}},
		{"thread_harness_consistent", lint.SeverityError, sharedThreadLinear, lint.Options{ResolveHarness: func(provider, _ string) (string, error) { return provider, nil }}},
		{"workflow_mode", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			g.Mode = set(graph.WorkflowModeDelivery)
			return g
		}, lint.Options{}},
		{"proof_contract", lint.SeverityError, func() graph.Graph {
			g := validProofGraph()
			g.ProofContract.Value.PrimaryCases = nil
			return g
		}, lint.Options{}},
		{"fan_in_max_visits", lint.SeverityWarning, func() graph.Graph { g := validParallel(); fan(g, "join").MaxVisits = set(2); return g }, lint.Options{}},
		{"branch_root_max_visits", lint.SeverityWarning, func() graph.Graph { g := validParallel(); coder(g, "left").MaxVisits = set(2); return g }, lint.Options{}},
		{"prompt_on_llm_nodes", lint.SeverityWarning, func() graph.Graph {
			g := validLinear()
			coder(g, "work").Prompt = jsonschema.Optional[string]{}
			return g
		}, lint.Options{}},
	}

	for _, test := range tests {
		t.Run(test.rule, func(t *testing.T) {
			finding, ok := findDiagnostic(lint.New(test.options).Validate(test.graph()), test.rule)
			if !ok {
				t.Fatalf("missing %s", test.rule)
			}
			if finding.Severity != test.severity {
				t.Fatalf("severity = %q, want %q", finding.Severity, test.severity)
			}
		})
	}
	if len(tests) != 29 {
		t.Fatalf("covered %d built-in rules, want 29", len(tests))
	}
}

func TestValidGraphsAndSupervisorExemption(t *testing.T) {
	for name, g := range map[string]graph.Graph{
		"linear":   validLinear(),
		"parallel": validParallel(),
		"proof":    validProofGraph(),
		"supervised": func() graph.Graph {
			g := validLinear()
			g.Nodes = append(g.Nodes, supervisor("coach", "work"))
			return g
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if diagnostics := lint.Validate(g); lint.HasErrors(diagnostics) {
				t.Fatalf("unexpected errors: %#v", diagnostics)
			}
		})
	}
}

func TestProofContractValidationDetails(t *testing.T) {
	tests := map[string]func(*graph.ProofContract){
		"blank primary outcome": func(contract *graph.ProofContract) { contract.PrimaryOutcome = "  " },
		"blank expected output": func(contract *graph.ProofContract) { contract.PrimaryCases[0].ExpectedOutput = "\t" },
		"ambiguous oracle": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].CorrectnessOracle = "review_record"
		},
		"unknown promise status": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].Status = "done"
		},
		"missing capability": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].RequiredCapabilities = []string{"missing"}
		},
		"unsafe evidence artifact": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].EvidenceArtifacts[0].Path = "../outside.txt"
		},
		"unknown artifact role": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].EvidenceArtifacts[0].Role = "diagram"
		},
		"missing artifact edge": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].EvidenceArtifacts[0].ArchitectureEdge = " "
		},
		"unknown evidence mode": func(contract *graph.ProofContract) {
			contract.PrimaryCases[0].EvidenceMode = "mock"
		},
		"case node is not a tool": func(contract *graph.ProofContract) { contract.PrimaryCases[0].Node = "worker" },
		"same node proves two cases": func(contract *graph.ProofContract) {
			contract.BoundaryCases[0].Node = contract.PrimaryCases[0].Node
		},
		"required boundary omitted": func(contract *graph.ProofContract) {
			contract.TerminalSuccess.RequiredCases = []string{"qualifying_input"}
		},
		"unknown required case": func(contract *graph.ProofContract) {
			contract.TerminalSuccess.RequiredCases = append(contract.TerminalSuccess.RequiredCases, "invented")
		},
		"blank unknown": func(contract *graph.ProofContract) { contract.Unknowns = []string{" "} },
		"scope gap names unknown promise": func(contract *graph.ProofContract) {
			contract.ScopeGaps = []graph.ProofScopeGap{{
				PromiseID: "missing", OriginalPromise: "A user-visible promise.",
				CurrentProvenBehavior: "Only a boundary works.", MissingCapability: "Primary behavior.",
				Impact: "The user job is not met.", RecommendedNextMove: "Implement the primary behavior.",
			}}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			g := validProofGraph()
			g.Nodes = append(g.Nodes, codergen("worker", edge(graph.Success)))
			mutate(&g.ProofContract.Value)
			if _, ok := findDiagnostic(lint.Validate(g), "proof_contract"); !ok {
				t.Fatal("missing proof_contract diagnostic")
			}
		})
	}
}

func TestExplicitWorkflowModeValidation(t *testing.T) {
	t.Run("delivery requires a proof contract", func(t *testing.T) {
		g := validLinear()
		g.Mode = set(graph.WorkflowModeDelivery)
		diagnostic, ok := findDiagnostic(lint.Validate(g), "workflow_mode")
		if !ok || !strings.Contains(diagnostic.Message, "delivery mode requires proof_contract") {
			t.Fatalf("diagnostics = %#v", lint.Validate(g))
		}
	})

	t.Run("discovery may remain contract-light", func(t *testing.T) {
		g := validLinear()
		g.Mode = set(graph.WorkflowModeDiscovery)
		if diagnostics := lint.Validate(g); lint.HasErrors(diagnostics) {
			t.Fatalf("diagnostics = %#v", diagnostics)
		}
	})

	t.Run("legacy omission remains compatible", func(t *testing.T) {
		if diagnostics := lint.Validate(validLinear()); lint.HasErrors(diagnostics) {
			t.Fatalf("diagnostics = %#v", diagnostics)
		}
	})
}

func TestProofContractRejectsAdvisoryOracleRoute(t *testing.T) {
	g := validProofGraph()
	tool := g.Nodes[0].(*graph.ToolNode)
	tool.OnError = set(tool.OnSuccess)
	if _, ok := findDiagnostic(lint.Validate(g), "proof_contract"); !ok {
		t.Fatal("proof tool with indistinguishable success and error routes was admitted")
	}
}

func TestProofContractRejectsParallelBranchOracle(t *testing.T) {
	g := graph.Graph{
		Mode: set(graph.WorkflowModeDelivery), Start: "fanout",
		ProofContract: set(graph.ProofContract{
			IntendedArchitecture: "A branch produces a result before fan-in.",
			PrimaryOutcome:       "A parallel branch produces the expected result.",
			PrimaryCases:         []graph.ProofCase{proofCase("branch_assertion", "left")},
			BoundaryCases:        []graph.ProofCase{}, Unknowns: []string{},
			TerminalSuccess: graph.ProofTerminalSuccess{RequiredCases: []string{"branch_assertion"}},
		}),
		Nodes: []graph.Node{
			&graph.ParallelNode{NodeBase: graph.NodeBase{ID: "fanout"}, Branches: graph.LegacyParallelBranches("left", "right")},
			&graph.ToolNode{NodeBase: graph.NodeBase{ID: "left"}, ToolCommand: "true", OnSuccess: "join"},
			codergen("right", edge("join")),
			fanIn("join", edge(graph.Success)),
		},
	}
	if _, ok := findDiagnostic(lint.Validate(g), "proof_contract"); !ok {
		t.Fatal("parallel branch proof oracle was admitted")
	}
}

func TestDiscoveryContractMayDeclareManualEvidenceButCannotPretendItIsIndependent(t *testing.T) {
	g := validProofGraph()
	g.Mode = set(graph.WorkflowModeDiscovery)
	manual := &g.ProofContract.Value.PrimaryCases[0]
	manual.Node = ""
	manual.CorrectnessOracle = graph.ProofOracleHumanAttestation
	manual.EvidenceSource = graph.ProofEvidenceManual
	manual.Independence = graph.ProofHumanAttestation
	manual.Status = graph.ProofStatusSimulated
	if finding, ok := findDiagnostic(lint.Validate(g), "proof_contract"); ok {
		t.Fatalf("discovery declaration rejected: %#v", finding)
	}

	g.Mode = set(graph.WorkflowModeDelivery)
	if _, ok := findDiagnostic(lint.Validate(g), "proof_contract"); !ok {
		t.Fatal("manual evidence was admitted for delivery success")
	}
}

func TestSupervisorValidationDetails(t *testing.T) {
	for name, mutate := range map[string]func(*graph.SupervisorNode){
		"duplicate":     func(node *graph.SupervisorNode) { node.Supervises = []string{"work", "work"} },
		"missing":       func(node *graph.SupervisorNode) { node.Supervises = []string{"missing"} },
		"self":          func(node *graph.SupervisorNode) { node.Supervises = []string{"coach"} },
		"zero interval": func(node *graph.SupervisorNode) { node.Interval = set(graph.Duration("0s")) },
	} {
		t.Run(name, func(t *testing.T) {
			g := validLinear()
			coach := supervisor("coach", "work")
			mutate(coach)
			g.Nodes = append(g.Nodes, coach)
			if _, ok := findDiagnostic(lint.Validate(g), "supervises_valid"); !ok {
				t.Fatal("missing supervises_valid")
			}
		})
	}

	g := validLinear()
	g.Start = "coach"
	g.Nodes = append(g.Nodes, supervisor("coach", "work"))
	if _, ok := findDiagnostic(lint.Validate(g), "start_target"); !ok {
		t.Fatal("supervisor admitted as start")
	}
}

func TestPseudoTargetsAndMechanicalToolRoutes(t *testing.T) {
	g := graph.Graph{Start: "check", Nodes: []graph.Node{
		&graph.ToolNode{NodeBase: graph.NodeBase{ID: "check"}, ToolCommand: "true", OnSuccess: graph.Success, OnError: set(graph.Success)},
	}}
	diagnostics := lint.Validate(g)
	assertNoRule(t, diagnostics, "edge_target_exists")
	assertNoRule(t, diagnostics, "edge_target_unique")
	if lint.HasErrors(diagnostics) {
		t.Fatalf("unexpected errors: %#v", diagnostics)
	}
}

func TestParallelConvergenceAllowsCycleWithRouteToFanIn(t *testing.T) {
	g := validParallel()
	coder(g, "left").Edges = []graph.Edge{
		condition("left", "continue another iteration"),
		condition("join", "branch work is complete"),
	}
	if diagnostics := lint.Validate(g); lint.HasErrors(diagnostics) {
		t.Fatalf("unexpected errors: %#v", diagnostics)
	}
}

func TestMissingBranchTargetDoesNotCascade(t *testing.T) {
	g := validParallel()
	parallel(g).Branches[0] = graph.LegacyParallelBranch("missing")
	diagnostics := lint.Validate(g)
	if _, ok := findDiagnostic(diagnostics, "edge_target_exists"); !ok {
		t.Fatal("missing edge_target_exists")
	}
	for _, rule := range []string{"parallel_fan_in", "branch_disjoint", "no_nested_parallel", "parallel_thread_disjoint", "thread_branch_boundary", "fan_in_entry", "branch_entry"} {
		assertNoRule(t, diagnostics, rule)
	}
}

func TestThreadRulesUseResolvedReusableSessions(t *testing.T) {
	g := validParallel()
	for _, id := range []string{"left", "right"} {
		fields := llm(g, id)
		fields.ThreadID = set("shared")
		fields.Fidelity = set("none")
	}
	assertNoRule(t, lint.Validate(g), "parallel_thread_disjoint")

	g = sharedThreadLinear()
	validator := lint.New(lint.Options{ResolveHarness: func(_, _ string) (string, error) { return "shared", nil }})
	assertNoRule(t, validator.Validate(g), "thread_harness_consistent")

	validator = lint.New(lint.Options{ResolveHarness: func(_, _ string) (string, error) { return "", errors.New("unroutable") }})
	if _, ok := findDiagnostic(validator.Validate(g), "thread_harness_consistent"); !ok {
		t.Fatal("resolver failure not reported")
	}
}

func TestDiagnosticsAreDeterministicAndCarryRouteIdentity(t *testing.T) {
	g := validLinear()
	g.Nodes = append(g.Nodes, codergen("other", edge(graph.Success)))
	coder(g, "work").Edges = []graph.Edge{edge("other"), edge(graph.Success)}
	first, second := lint.Validate(g), lint.Validate(g)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("diagnostics differ:\n%#v\n%#v", first, second)
	}
	finding, ok := findDiagnostic(first, "edge_condition_missing")
	if !ok || finding.Edge == nil || *finding.Edge != (lint.EdgeRef{"work", "other"}) {
		t.Fatalf("edge diagnostic = %#v", finding)
	}
}

func TestValidateOrErrorAndExtraRule(t *testing.T) {
	validator := lint.New(lint.Options{})
	diagnostics, err := validator.ValidateOrError(validLinear(), namedRule{})
	if err != nil {
		t.Fatal(err)
	}
	if finding, ok := findDiagnostic(diagnostics, "custom_note"); !ok || finding.Severity != lint.SeverityInfo {
		t.Fatalf("custom diagnostic = %#v", diagnostics)
	}
	invalid := validLinear()
	coder(invalid, "work").Edges = nil
	diagnostics, err = validator.ValidateOrError(invalid)
	var validationError *lint.ValidationError
	if !errors.As(err, &validationError) || len(validationError.Diagnostics) != len(diagnostics) {
		t.Fatalf("error = %#v; diagnostics = %#v", err, diagnostics)
	}
}

type namedRule struct{}

func (namedRule) Name() string { return "custom_note" }
func (namedRule) Apply(graph.Graph) []lint.Diagnostic {
	return []lint.Diagnostic{{Severity: lint.SeverityInfo, Message: "custom rule ran"}}
}

func validLinear() graph.Graph {
	return graph.Graph{Start: "work", Nodes: []graph.Node{codergen("work", edge(graph.Success))}}
}

func validParallel() graph.Graph {
	return graph.Graph{Start: "parallel", Nodes: []graph.Node{
		&graph.ParallelNode{NodeBase: graph.NodeBase{ID: "parallel"}, Branches: graph.LegacyParallelBranches("left", "right")},
		codergen("left", edge("join")),
		codergen("right", edge("join")),
		fanIn("join", edge(graph.Success)),
	}}
}

func validProofGraph() graph.Graph {
	return graph.Graph{
		Mode: set(graph.WorkflowModeDelivery), Start: "primary_assertion",
		ProofContract: set(graph.ProofContract{
			IntendedArchitecture: "Input crosses the operating boundary and produces a user-visible result.",
			PrimaryOutcome:       "Representative input produces an observable result.",
			PrimaryCases:         []graph.ProofCase{proofCase("qualifying_input", "primary_assertion")},
			BoundaryCases:        []graph.ProofCase{proofCase("empty_boundary", "empty_assertion")},
			RequiredCapabilities: []graph.ProofCapability{{ID: "operating_runtime", Description: "The operating surface is available."}},
			Unknowns:             []string{},
			ScopeGaps:            []graph.ProofScopeGap{},
			TerminalSuccess: graph.ProofTerminalSuccess{
				RequiredCases: []string{"qualifying_input", "empty_boundary"},
			},
		}),
		Nodes: []graph.Node{
			&graph.ToolNode{NodeBase: graph.NodeBase{ID: "primary_assertion"}, ToolCommand: "true", OnSuccess: "empty_assertion"},
			&graph.ToolNode{NodeBase: graph.NodeBase{ID: "empty_assertion"}, ToolCommand: "true", OnSuccess: graph.Success},
		},
	}
}

func proofCase(id, node string) graph.ProofCase {
	return graph.ProofCase{
		ID: id, Actor: "A user", Job: "Submit qualifying input and observe the promised result.", Node: node,
		QualifyingInputCriteria: "Input independently established to meet the case criteria.",
		ExpectedOutput:          "The expected value is visible at the declared evidence surface.",
		CorrectnessOracle:       graph.ProofOracleToolExitZero,
		EvidenceMode:            graph.ProofEvidenceOperatingLayer,
		EvidenceSource:          graph.ProofEvidenceCurrentRun,
		Independence:            graph.ProofIndependentExecution,
		Status:                  graph.ProofStatusUnproven,
		RequiredCapabilities:    []string{"operating_runtime"},
		EvidenceArtifacts: []graph.ProofArtifactRequirement{{
			Path: id + ".txt", Role: graph.ProofArtifactInput, ArchitectureEdge: "caller_to_product",
		}},
	}
}

func overlappingParallel() graph.Graph {
	g := validParallel()
	g.Nodes = append(g.Nodes, codergen("shared", edge("join")))
	coder(g, "left").Edges = []graph.Edge{edge("shared")}
	coder(g, "right").Edges = []graph.Edge{edge("shared")}
	return g
}

func nestedParallel() graph.Graph {
	g := validParallel()
	g.Nodes = append(g.Nodes, &graph.ParallelNode{NodeBase: graph.NodeBase{ID: "nested"}, Branches: graph.LegacyParallelBranches("join")})
	coder(g, "left").Edges = []graph.Edge{edge("nested")}
	return g
}

func directFanInParallel() graph.Graph {
	g := validParallel()
	pre := codergen("pre", condition("parallel", "fan out"), condition("join", "skip branches"))
	g.Start = "pre"
	g.Nodes = append(g.Nodes, pre)
	return g
}

func externalBranchEntry() graph.Graph {
	g := validParallel()
	pre := codergen("pre", condition("parallel", "fan out"), condition("left", "enter branch"))
	g.Start = "pre"
	g.Nodes = append(g.Nodes, pre)
	return g
}

func supervisedCycle() graph.Graph {
	g := validLinear()
	g.Nodes = append(g.Nodes, supervisor("manager", "director"), supervisor("director", "manager"))
	return g
}

func sharedThreadLinear() graph.Graph {
	g := validLinear()
	first := coder(g, "work")
	first.Edges = []graph.Edge{edge("second")}
	first.ThreadID = set("shared")
	first.LLMProvider = set("openai")
	second := codergen("second", edge(graph.Success))
	second.ThreadID = set("shared")
	second.LLMProvider = set("anthropic")
	g.Nodes = slices.Insert(g.Nodes, 1, graph.Node(second))
	return g
}

func codergen(id string, edges ...graph.Edge) *graph.CodergenNode {
	return &graph.CodergenNode{NodeBase: graph.NodeBase{ID: id}, Edges: edges, LLMNodeFields: graph.LLMNodeFields{Prompt: set("work")}}
}

func fanIn(id string, edges ...graph.Edge) *graph.FanInNode {
	return &graph.FanInNode{NodeBase: graph.NodeBase{ID: id}, Edges: edges, LLMNodeFields: graph.LLMNodeFields{Prompt: set("evaluate")}}
}

func supervisor(id string, supervises ...string) *graph.SupervisorNode {
	return &graph.SupervisorNode{NodeBase: graph.NodeBase{ID: id}, Prompt: "watch", Supervises: supervises}
}

func coder(g graph.Graph, id string) *graph.CodergenNode {
	for _, node := range g.Nodes {
		if node.Base().ID == id {
			return node.(*graph.CodergenNode)
		}
	}
	panic("missing codergen " + id)
}

func fan(g graph.Graph, id string) *graph.FanInNode {
	for _, node := range g.Nodes {
		if node.Base().ID == id {
			return node.(*graph.FanInNode)
		}
	}
	panic("missing fan-in " + id)
}

func parallel(g graph.Graph) *graph.ParallelNode { return g.Nodes[0].(*graph.ParallelNode) }

func llm(g graph.Graph, id string) *graph.LLMNodeFields {
	for _, node := range g.Nodes {
		if node.Base().ID != id {
			continue
		}
		switch node := node.(type) {
		case *graph.CodergenNode:
			return &node.LLMNodeFields
		case *graph.FanInNode:
			return &node.LLMNodeFields
		}
	}
	panic("missing LLM node " + id)
}

func edge(to string) graph.Edge            { return graph.Edge{To: to} }
func condition(to, text string) graph.Edge { return graph.Edge{To: to, Condition: text} }

func set[T any](value T) jsonschema.Optional[T] {
	return jsonschema.Optional[T]{Present: true, Value: value}
}

func findDiagnostic(diagnostics []lint.Diagnostic, rule string) (lint.Diagnostic, bool) {
	for _, diagnostic := range diagnostics {
		if diagnostic.Rule == rule {
			return diagnostic, true
		}
	}
	return lint.Diagnostic{}, false
}

func assertNoRule(t *testing.T, diagnostics []lint.Diagnostic, rule string) {
	t.Helper()
	if diagnostic, ok := findDiagnostic(diagnostics, rule); ok {
		t.Fatalf("unexpected %s: %#v", rule, diagnostic)
	}
}
