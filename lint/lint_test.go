package lint_test

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/tylergannon/gimble/graph"
	"github.com/tylergannon/gimble/lint"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
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
		{"fan_out_fan_in", lint.SeverityError, func() graph.Graph {
			g := validParallel()
			coder(g, "left").Edges = []graph.Edge{edge(graph.Success)}
			return g
		}, lint.Options{}},
		{"branch_disjoint", lint.SeverityError, overlappingParallel, lint.Options{}},
		{"no_nested_fan_out", lint.SeverityError, nestedParallel, lint.Options{}},
		{"fan_in_single_fan_out", lint.SeverityError, func() graph.Graph {
			g := validLinear()
			g.Nodes = append(g.Nodes, fanIn("unowned", edge(graph.Success)))
			return g
		}, lint.Options{}},
		{"fan_out_thread_disjoint", lint.SeverityError, func() graph.Graph {
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
		{"loop_body_entry", lint.SeverityError, externalLoopBodyEntry, lint.Options{}},
		{"loop_body_returns", lint.SeverityError, func() graph.Graph {
			g := validLoop()
			coder(g, "lap").Edges = []graph.Edge{edge(graph.Success)}
			return g
		}, lint.Options{}},
		{"loop_body_exit", lint.SeverityError, loopBodyExitsToSuccess, lint.Options{}},
		{"loop_checklist_required", lint.SeverityError, func() graph.Graph {
			g := validLoop()
			loopNode(g, "items").Checklist = jsonschema.Optional[string]{}
			return g
		}, lint.Options{}},
		{"loop_in_fan_out", lint.SeverityError, loopInsideParallel, lint.Options{}},
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
	if len(tests) != 32 {
		t.Fatalf("covered %d built-in rules, want 32", len(tests))
	}
}

func TestValidGraphsAndSupervisorExemption(t *testing.T) {
	for name, g := range map[string]graph.Graph{
		"linear":   validLinear(),
		"parallel": validParallel(),
		"loop":     validLoop(),
		"nested":   validNestedLoop(),
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

func TestLoopRulesAreSilentOnValidLoops(t *testing.T) {
	loopRules := []string{"loop_body_entry", "loop_body_returns", "loop_body_exit", "loop_checklist_required", "loop_in_fan_out"}
	for name, g := range map[string]graph.Graph{"loop": validLoop(), "nested": validNestedLoop()} {
		t.Run(name, func(t *testing.T) {
			diagnostics := lint.Validate(g)
			for _, rule := range loopRules {
				assertNoRule(t, diagnostics, rule)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", diagnostics)
			}
		})
	}
}

func TestLoopRuleDetails(t *testing.T) {
	g := validLoop()
	loopNode(g, "items").Edges.Exit = "lap"
	finding, ok := findDiagnostic(lint.Validate(g), "loop_body_entry")
	if !ok || finding.Edge == nil || *finding.Edge != (lint.EdgeRef{"items", "lap"}) {
		t.Fatalf("edges.exit into body = %#v", finding)
	}

	g = externalLoopBodyEntry()
	finding, ok = findDiagnostic(lint.Validate(g), "loop_body_entry")
	if !ok || finding.Edge == nil || *finding.Edge != (lint.EdgeRef{"pre", "lap"}) {
		t.Fatalf("external body entry = %#v", finding)
	}

	g = validLoop()
	loopNode(g, "items").Edges.Loop = "missing"
	diagnostics := lint.Validate(g)
	if _, ok := findDiagnostic(diagnostics, "edge_target_exists"); !ok {
		t.Fatal("missing edge_target_exists")
	}
	assertNoRule(t, diagnostics, "loop_body_returns")

	g = validLoop()
	loopNode(g, "items").MaxVisits = set(0)
	if _, ok := findDiagnostic(lint.Validate(g), "max_visits_positive"); !ok {
		t.Fatal("missing max_visits_positive on loop node")
	}
	assertNoRule(t, lint.Validate(g), "edge_condition_missing")
}

func TestLoopBodyMustNameANode(t *testing.T) {
	for _, target := range []string{graph.Success, graph.Failure} {
		g := validLoop()
		loopNode(g, "items").Edges.Loop = target
		finding, ok := findDiagnostic(lint.Validate(g), "loop_body_returns")
		if !ok || finding.NodeID != "items" || finding.Message != `edges.loop must name a node, not "`+target+`"` {
			t.Fatalf("body %s = %#v", target, finding)
		}
	}
}

func TestLoopBodyExit(t *testing.T) {
	g := loopBodyExitsToSuccess()
	finding, ok := findDiagnostic(lint.Validate(g), "loop_body_exit")
	if !ok || finding.Edge == nil || *finding.Edge != (lint.EdgeRef{"lap", graph.Success}) {
		t.Fatalf("body exit = %#v", finding)
	}
	if want := `loop body node "lap" may not route to success; route back to loop "items" and let edges.exit end the run`; finding.Message != want {
		t.Fatalf("message = %q, want %q", finding.Message, want)
	}

	// A nested loop whose edges.exit is success would end the run with outer
	// items open: it is a body node of the outer loop.
	g = validNestedLoop()
	loopNode(g, "inner").Edges.Exit = graph.Success
	finding, ok = findDiagnostic(lint.Validate(g), "loop_body_exit")
	if !ok || finding.Edge == nil || *finding.Edge != (lint.EdgeRef{"inner", graph.Success}) {
		t.Fatalf("nested edges.exit = %#v", finding)
	}

	// failure stays an escape hatch.
	g = validLoop()
	coder(g, "lap").Edges = []graph.Edge{condition("items", "item done"), condition(graph.Failure, "give up")}
	diagnostics := lint.Validate(g)
	assertNoRule(t, diagnostics, "loop_body_exit")
	if lint.HasErrors(diagnostics) {
		t.Fatalf("unexpected errors: %#v", diagnostics)
	}
}

func TestStartMustNotNameLoopBodyNode(t *testing.T) {
	g := validLoop()
	g.Start = "lap"
	finding, ok := findDiagnostic(lint.Validate(g), "loop_body_entry")
	if !ok || finding.NodeID != "lap" || finding.Message != `start must not name loop body node "lap"; start at the loop or before it` {
		t.Fatalf("start at body = %#v", finding)
	}

	g = validNestedLoop()
	g.Start = "inner"
	finding, ok = findDiagnostic(lint.Validate(g), "loop_body_entry")
	if !ok || finding.NodeID != "inner" || finding.Message != `start must not name loop body node "inner"; start at the loop or before it` {
		t.Fatalf("start at nested loop = %#v", finding)
	}

	g = validNestedLoop()
	pre := codergen("pre", edge("outer"))
	g.Start = "pre"
	g.Nodes = append(g.Nodes, pre)
	assertNoRule(t, lint.Validate(g), "loop_body_entry")
}

func TestOutermostLoop(t *testing.T) {
	g := validNestedLoop()
	for node, want := range map[string]string{"plan": "outer", "inner": "outer", "implement": "outer"} {
		if got, ok := lint.OutermostLoop(g, node); !ok || got != want {
			t.Fatalf("OutermostLoop(%s) = %q, %v; want %q", node, got, ok, want)
		}
	}
	for _, node := range []string{"outer", "missing", graph.Success} {
		if got, ok := lint.OutermostLoop(g, node); ok {
			t.Fatalf("OutermostLoop(%s) = %q, want none", node, got)
		}
	}
	if got, ok := lint.OutermostLoop(validLoop(), "lap"); !ok || got != "items" {
		t.Fatalf("OutermostLoop(lap) = %q, %v", got, ok)
	}
	if _, ok := lint.OutermostLoop(validLinear(), "work"); ok {
		t.Fatal("OutermostLoop found a loop in a graph without one")
	}
}

func TestLoopBodyTopologyLookups(t *testing.T) {
	g := validNestedLoop()
	if got, ok := lint.LoopBodyNodes(g, "outer"); !ok || !reflect.DeepEqual(got, []string{"plan", "inner", "implement"}) {
		t.Fatalf("LoopBodyNodes(outer) = %v, %v", got, ok)
	}
	if got, ok := lint.LoopBodyNodes(g, "inner"); !ok || !reflect.DeepEqual(got, []string{"implement"}) {
		t.Fatalf("LoopBodyNodes(inner) = %v, %v", got, ok)
	}
	if got, ok := lint.LoopBodyNodes(g, "missing"); ok || got != nil {
		t.Fatalf("LoopBodyNodes(missing) = %v, %v", got, ok)
	}
	for node, want := range map[string][]string{
		"plan":      {"outer"},
		"inner":     {"outer"},
		"implement": {"outer", "inner"},
		"outer":     {},
	} {
		if got := lint.EnclosingLoops(g, node); !reflect.DeepEqual(got, want) {
			t.Fatalf("EnclosingLoops(%s) = %v, want %v", node, got, want)
		}
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
		&graph.CommandNode{ID: "check", Command: "true", Edges: graph.CommandEdges{Success: graph.Success, Error: set(graph.Success)}},
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
	parallel(g).Branches[0] = graph.LegacyFanOutBranch("missing")
	diagnostics := lint.Validate(g)
	if _, ok := findDiagnostic(diagnostics, "edge_target_exists"); !ok {
		t.Fatal("missing edge_target_exists")
	}
	for _, rule := range []string{"fan_out_fan_in", "branch_disjoint", "no_nested_fan_out", "fan_out_thread_disjoint", "thread_branch_boundary", "fan_in_entry", "branch_entry"} {
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
	assertNoRule(t, lint.Validate(g), "fan_out_thread_disjoint")

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
		&graph.FanOutNode{ID: "parallel", Branches: graph.LegacyFanOutBranches("left", "right")},
		codergen("left", edge("join")),
		codergen("right", edge("join")),
		fanIn("join", edge(graph.Success)),
	}}
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
	g.Nodes = append(g.Nodes, &graph.FanOutNode{ID: "nested", Branches: graph.LegacyFanOutBranches("join")})
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

func validLoop() graph.Graph {
	return graph.Graph{Start: "items", Nodes: []graph.Node{
		loop("items", "checklist.md", "lap", graph.Success),
		codergen("lap", edge("items")),
	}}
}

func validNestedLoop() graph.Graph {
	return graph.Graph{Start: "outer", Nodes: []graph.Node{
		loop("outer", "checklist.md", "plan", graph.Success),
		codergen("plan", edge("inner")),
		loop("inner", "", "implement", "outer"),
		codergen("implement", edge("inner")),
	}}
}

func externalLoopBodyEntry() graph.Graph {
	g := validLoop()
	pre := codergen("pre", condition("items", "start the loop"), condition("lap", "skip the loop"))
	g.Start = "pre"
	g.Nodes = append(g.Nodes, pre)
	return g
}

func loopBodyExitsToSuccess() graph.Graph {
	g := validLoop()
	coder(g, "lap").Edges = []graph.Edge{condition("items", "item done"), condition(graph.Success, "everything done")}
	return g
}

func loopInsideParallel() graph.Graph {
	g := validParallel()
	coder(g, "left").Edges = []graph.Edge{edge("items")}
	g.Nodes = append(g.Nodes, loop("items", "checklist.md", "lap", "join"), codergen("lap", edge("items")))
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
	first.Model = set(graph.ModelSelection{Name: "gpt-5.6-sol"})
	second := codergen("second", edge(graph.Success))
	second.ThreadID = set("shared")
	second.Model = set(graph.ModelSelection{Name: "claude-sonnet-5"})
	g.Nodes = slices.Insert(g.Nodes, 1, graph.Node(second))
	return g
}

func codergen(id string, edges ...graph.Edge) *graph.AgentNode {
	return &graph.AgentNode{ID: id, Edges: edges, Prompt: set("work")}
}

func fanIn(id string, edges ...graph.Edge) *graph.FanInNode {
	return &graph.FanInNode{ID: id, Edges: edges, Prompt: set("evaluate")}
}

func loop(id, checklist, body, onDone string) *graph.LoopNode {
	node := &graph.LoopNode{ID: id, Edges: graph.LoopEdges{Loop: body, Exit: onDone}}
	if checklist != "" {
		node.Checklist = set(checklist)
	}
	return node
}

func loopNode(g graph.Graph, id string) *graph.LoopNode {
	for _, node := range g.Nodes {
		if node.Base().ID == id {
			return node.(*graph.LoopNode)
		}
	}
	panic("missing loop " + id)
}

func supervisor(id string, supervises ...string) *graph.SupervisorNode {
	return &graph.SupervisorNode{ID: id, Prompt: "watch", Supervises: supervises}
}

func coder(g graph.Graph, id string) *graph.AgentNode {
	for _, node := range g.Nodes {
		if node.Base().ID == id {
			return node.(*graph.AgentNode)
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

func parallel(g graph.Graph) *graph.FanOutNode { return g.Nodes[0].(*graph.FanOutNode) }

func llm(g graph.Graph, id string) *graph.LLMNodeFields {
	for _, node := range g.Nodes {
		if node.Base().ID != id {
			continue
		}
		switch node := node.(type) {
		case *graph.AgentNode:
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
