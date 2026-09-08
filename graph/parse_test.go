package graph

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

const representative = `{
  "name":"all_fields",
  "goal":"exercise parsing",
  "defaults":{
    "max_retries":2,
    "fidelity":"full",
    "timeout":"15m",
    "model":{"name":"fable","effort":"high"}
  },
  "start":"code",
  "nodes":[
    {"id":"code","type":"agent","label":"Code","prompt":"work","model":{"name":"gpt-5.6-sol","effort":"low"},"thread_id":"shared","max_visits":3,"edges":[{"to":"command"}]},
    {"id":"command","type":"command","command":"go test ./...","edges":{"success":"fanout","error":"code"}},
    {"id":"fanout","type":"fan_out","max_parallel":3,"branches":["left","right"]},
    {"id":"left","type":"command","command":"true","edges":{"success":"join"}},
    {"id":"right","type":"command","command":"true","edges":{"success":"join"}},
    {"id":"join","type":"fan_in","prompt":"choose","edges":[{"to":"success"}]},
    {"id":"coach","type":"supervisor","prompt":"keep scope","supervises":["code","join"],"interval":"120s"}
  ]
}`

func TestParseRepresentativeGraphAndResolveDefaults(t *testing.T) {
	pipeline, err := Parse([]byte(representative))
	if err != nil {
		t.Fatal(err)
	}
	if pipeline.Start != "code" || pipeline.Name != "all_fields" || len(pipeline.Nodes) != 7 {
		t.Fatalf("graph = %#v", pipeline)
	}
	code := mustNode[*AgentNode](t, pipeline, "code")
	if code.DisplayLabel() != "Code" || code.Model.Value.Name != "gpt-5.6-sol" || code.ThreadKey(code.ID) != "shared" {
		t.Fatalf("agent = %#v", code)
	}
	if code.MaxRetries.Value != 2 || code.FidelityValue() != "full" || code.Timeout.Value != "15m" {
		t.Fatalf("agent defaults = %#v", code.LLMNodeFields)
	}
	command := mustNode[*CommandNode](t, pipeline, "command")
	if command.Edges.Success != "fanout" || !command.Edges.Error.Present || command.Edges.Error.Value != "code" || command.Timeout.Value != "15m" {
		t.Fatalf("command = %#v", command)
	}
	fanOut := mustNode[*FanOutNode](t, pipeline, "fanout")
	if fanOut.MaxParallelValue() != 3 || !reflect.DeepEqual(fanOut.BranchIDs(), []string{"left", "right"}) {
		t.Fatalf("fan_out = %#v", fanOut)
	}
	join := mustNode[*FanInNode](t, pipeline, "join")
	if join.Model.Present {
		t.Fatalf("parser flattened model defaults into fan-in: %#v", join.LLMNodeFields)
	}
	coach := mustNode[*SupervisorNode](t, pipeline, "coach")
	if coach.IntervalValue() != "120s" || coach.Timeout.Value != "15m" || coach.Model.Present {
		t.Fatalf("supervisor defaults = %#v", coach)
	}
}

func TestParseAcceptsEveryNodeShape(t *testing.T) {
	tests := []string{
		`{"id":"c","type":"agent","prompt":"p","max_retries":0,"fidelity":"none","thread_id":"t","timeout":"250ms","model":{"name":"fable","version":"5.1","effort":"low"},"edges":[{"to":"success"}]}`,
		`{"id":"p","type":"fan_out","branches":["c"],"max_parallel":4}`,
		`{"id":"f","type":"fan_in","prompt":"p","edges":[{"to":"success"}]}`,
		`{"id":"t","type":"command","command":"true","edges":{"success":"success","error":"failure"},"timeout":"2h"}`,
		`{"id":"s","type":"supervisor","prompt":"watch","supervises":["c"]}`,
		`{"id":"l","type":"loop","checklist":"list.md","edges":{"loop":"c","exit":"success"},"max_visits":3,"timeout":"1m","item_judge":{"model":{"name":"flash","effort":"medium"}},"goal_evaluator":{"model":{"name":"fable","version":"5","effort":"high"}}}`,
	}
	for _, node := range tests {
		document := `{"start":"c","nodes":[` + node + `]}`
		if _, err := Parse([]byte(document)); err != nil {
			t.Errorf("Parse(%s): %v", node, err)
		}
	}
}

func TestParseRejectsObsoleteAndMalformedModelSelections(t *testing.T) {
	tests := map[string]string{
		"string shorthand":         `{"start":"x","nodes":[{"id":"x","type":"agent","model":"fable","edges":[{"to":"success"}]}]}`,
		"missing name":             `{"start":"x","nodes":[{"id":"x","type":"agent","model":{"effort":"high"},"edges":[{"to":"success"}]}]}`,
		"numeric version":          `{"start":"x","nodes":[{"id":"x","type":"agent","model":{"name":"fable","version":5},"edges":[{"to":"success"}]}]}`,
		"authored provider":        `{"start":"x","nodes":[{"id":"x","type":"agent","model":{"name":"fable","provider":"anthropic"},"edges":[{"to":"success"}]}]}`,
		"command model":            `{"start":"x","nodes":[{"id":"x","type":"command","command":"true","model":{"name":"fable"},"edges":{"success":"success"}}]}`,
		"loop top-level model":     `{"start":"x","nodes":[{"id":"x","type":"loop","model":{"name":"fable"},"edges":{"loop":"x","exit":"success"}}]}`,
		"obsolete defaults":        `{"defaults":{"llm_model":"fable"},"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`,
		"obsolete provider":        `{"start":"x","nodes":[{"id":"x","type":"agent","llm_provider":"anthropic","edges":[{"to":"success"}]}]}`,
		"obsolete effort":          `{"start":"x","nodes":[{"id":"x","type":"agent","reasoning_effort":"high","edges":[{"to":"success"}]}]}`,
		"obsolete evaluator":       `{"start":"x","nodes":[{"id":"x","type":"loop","evaluator_llm_model":"fable","edges":{"loop":"x","exit":"success"}}]}`,
		"unknown role field":       `{"start":"x","nodes":[{"id":"x","type":"loop","item_judge":{"prompt":"no"},"edges":{"loop":"x","exit":"success"}}]}`,
		"branch authored provider": `{"start":"p","nodes":[{"id":"p","type":"fan_out","branches":[{"id":"b","artifacts":["x"],"agent":{"model":{"name":"fable","provider":"anthropic"}}}],"branch_edges":[{"to":"j"}]},{"id":"j","type":"fan_in","edges":[{"to":"success"}]}]}`,
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(document)); err == nil {
				t.Fatal("Parse accepted invalid model selection")
			}
		})
	}
}

func TestParseObsoleteModelDiagnosticsNameTheReplacement(t *testing.T) {
	tests := []struct {
		name     string
		document string
		want     string
	}{
		{"agent model", `{"start":"x","nodes":[{"id":"x","type":"agent","llm_model":"fable","edges":[{"to":"success"}]}]}`, `replace it with "model.name"`},
		{"loop judge effort", `{"start":"x","nodes":[{"id":"x","type":"loop","reasoning_effort":"medium","edges":{"loop":"x","exit":"success"}}]}`, `replace it with "item_judge.model.effort"`},
		{"loop evaluator", `{"start":"x","nodes":[{"id":"x","type":"loop","evaluator_llm_model":"fable","edges":{"loop":"x","exit":"success"}}]}`, `replace it with "goal_evaluator.model.name"`},
		{"provider", `{"start":"x","nodes":[{"id":"x","type":"agent","llm_provider":"anthropic","edges":[{"to":"success"}]}]}`, `authored provider was removed because model.name determines provider`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse([]byte(test.document))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestParsePreservesOptionalPresence(t *testing.T) {
	document := `{"start":"omitted","nodes":[
    {"id":"omitted","type":"agent"},
    {"id":"empty","type":"agent","label":"","prompt":""}
  ]}`
	pipeline, err := Parse([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	omitted := mustNode[*AgentNode](t, pipeline, "omitted")
	if omitted.Label.Present || omitted.Prompt.Present || omitted.DisplayLabel() != "omitted" {
		t.Fatalf("omitted = %#v", omitted)
	}
	empty := mustNode[*AgentNode](t, pipeline, "empty")
	if !empty.Label.Present || !empty.Prompt.Present || empty.DisplayLabel() != "" {
		t.Fatalf("empty = %#v", empty)
	}
}

func TestParseResolvesStructuredFanOutAgentBranches(t *testing.T) {
	document := `{
  "start":"fanout",
  "nodes":[
    {
      "id":"fanout",
      "type":"fan_out",
      "workspace":"shared",
      "prompt":"Build the parent artifact",
      "model":{"name":"gpt-5.6-sol","effort":"high"},
      "timeout":"3m",
      "branch_edges":[{"to":"join"}],
      "branches":[
        {"id":"openai_branch","artifacts":["openai.txt"],"agent":{"prompt":"Build OpenAI output"}},
        {"id":"anthropic_branch","artifacts":["anthropic.txt"],"agent":{"model":{"name":"claude-sonnet-5","effort":"medium"}}}
      ]
    },
    {"id":"join","type":"fan_in","edges":[{"to":"success"}]}
  ]
}`
	pipeline, err := Parse([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	fanOut := mustNode[*FanOutNode](t, pipeline, "fanout")
	if fanOut.WorkspacePolicyValue() != WorkspaceShared || !reflect.DeepEqual(fanOut.BranchIDs(), []string{"openai_branch", "anthropic_branch"}) {
		t.Fatalf("fan_out = %#v", fanOut)
	}
	openai := mustNode[*AgentNode](t, pipeline, "openai_branch")
	if openai.Prompt.Value != "Build OpenAI output" || openai.Model.Value.Name != "gpt-5.6-sol" || openai.Model.Value.Effort.Value != "high" || openai.Timeout.Value != "3m" {
		t.Fatalf("inherited branch = %#v", openai)
	}
	anthropic := mustNode[*AgentNode](t, pipeline, "anthropic_branch")
	if anthropic.Prompt.Value != "Build the parent artifact" || anthropic.Model.Value.Name != "claude-sonnet-5" || anthropic.Model.Value.Effort.Value != "medium" || !reflect.DeepEqual(anthropic.Edges, []Edge{{To: "join"}}) {
		t.Fatalf("overridden branch = %#v", anthropic)
	}
}

func TestParseRejectsInvalidStructuredFanOutBranches(t *testing.T) {
	tests := map[string]string{
		"mixed branches":       `{"start":"p","nodes":[{"id":"p","type":"fan_out","branch_edges":[{"to":"join"}],"branches":["legacy",{"id":"variant","artifacts":["out.txt"]}]},{"id":"legacy","type":"agent","edges":[{"to":"join"}]},{"id":"join","type":"fan_in","edges":[{"to":"success"}]}]}`,
		"missing artifacts":    `{"start":"p","nodes":[{"id":"p","type":"fan_out","branch_edges":[{"to":"join"}],"branches":[{"id":"variant"}]},{"id":"join","type":"fan_in","edges":[{"to":"success"}]}]}`,
		"empty artifacts":      `{"start":"p","nodes":[{"id":"p","type":"fan_out","branch_edges":[{"to":"join"}],"branches":[{"id":"variant","artifacts":[]}]},{"id":"join","type":"fan_in","edges":[{"to":"success"}]}]}`,
		"unsafe artifact":      `{"start":"p","nodes":[{"id":"p","type":"fan_out","branch_edges":[{"to":"join"}],"branches":[{"id":"variant","artifacts":["../out.txt"]}]},{"id":"join","type":"fan_in","edges":[{"to":"success"}]}]}`,
		"missing branch edges": `{"start":"p","nodes":[{"id":"p","type":"fan_out","branches":[{"id":"variant","artifacts":["out.txt"]}]},{"id":"join","type":"fan_in","edges":[{"to":"success"}]}]}`,
		"node collision":       `{"start":"p","nodes":[{"id":"p","type":"fan_out","branch_edges":[{"to":"join"}],"branches":[{"id":"join","artifacts":["out.txt"]}]},{"id":"join","type":"fan_in","edges":[{"to":"success"}]}]}`,
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(document)); err == nil {
				t.Fatal("invalid structured branch admitted")
			}
		})
	}
}

func TestParseRejectsStructuralViolations(t *testing.T) {
	tests := map[string]string{
		"missing start":             `{"nodes":[]}`,
		"unknown type":              `{"start":"x","nodes":[{"id":"x","type":"notify.slack"}]}`,
		"cross type field":          `{"start":"x","nodes":[{"id":"x","type":"command","command":"true","edges":{"success":"success"},"prompt":"no"}]}`,
		"missing command":           `{"start":"x","nodes":[{"id":"x","type":"command","edges":{"success":"success"}}]}`,
		"missing command edges":     `{"start":"x","nodes":[{"id":"x","type":"command","command":"true"}]}`,
		"missing command success":   `{"start":"x","nodes":[{"id":"x","type":"command","command":"true","edges":{}}]}`,
		"missing branches":          `{"start":"x","nodes":[{"id":"x","type":"fan_out"}]}`,
		"missing supervisor prompt": `{"start":"x","nodes":[{"id":"x","type":"supervisor","supervises":["x"]}]}`,
		"missing supervises":        `{"start":"x","nodes":[{"id":"x","type":"supervisor","prompt":"watch"}]}`,
		"missing loop edges":        `{"start":"x","nodes":[{"id":"x","type":"loop"}]}`,
		"missing loop route":        `{"start":"x","nodes":[{"id":"x","type":"loop","edges":{"loop":"x"}}]}`,
		"loop unknown field":        `{"start":"x","nodes":[{"id":"x","type":"loop","edges":{"loop":"x","exit":"success"},"prompt":"no"}]}`,
		"loop role effort":          `{"start":"x","nodes":[{"id":"x","type":"loop","edges":{"loop":"x","exit":"success"},"goal_evaluator":{"model":{"name":"fable","effort":"extreme"}}}]}`,
		"old agent type":            `{"start":"x","nodes":[{"id":"x","type":"codergen"}]}`,
		"old command type":          `{"start":"x","nodes":[{"id":"x","type":"tool","command":"true","edges":{"success":"success"}}]}`,
		"old fan_out type":          `{"start":"x","nodes":[{"id":"x","type":"parallel","branches":[]}]}`,
		"old fan_in type":           `{"start":"x","nodes":[{"id":"x","type":"parallel.fan_in","edges":[{"to":"success"}]}]}`,
		"old command fields":        `{"start":"x","nodes":[{"id":"x","type":"command","tool_command":"true","on_success":"success"}]}`,
		"old loop fields":           `{"start":"x","nodes":[{"id":"x","type":"loop","body":"x","on_done":"success"}]}`,
		"old fan_out edges":         `{"start":"x","nodes":[{"id":"x","type":"fan_out","branches":[{"id":"b","artifacts":["b.txt"]}],"edges":[{"to":"success"}]}]}`,
		"unknown top field":         `{"start":"x","nodes":[],"extra":1}`,
		"unknown defaults":          `{"start":"x","defaults":{"max_visits":1},"nodes":[]}`,
		"null":                      `{"start":"x","name":null,"nodes":[]}`,
		"wrong type":                `{"start":1,"nodes":[]}`,
		"trailing value":            `{"start":"x","nodes":[]} {}`,
		"comment":                   `{"start":"x","nodes":[]} // no`,
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(document)); err == nil {
				t.Fatal("invalid pipeline admitted")
			}
		})
	}
}

func TestParseAcceptsZeroOrOneNamedProcfileService(t *testing.T) {
	without := `{"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`
	if _, err := Parse([]byte(without)); err != nil {
		t.Fatalf("workflow without services: %v", err)
	}
	with := `{"system_file":"dev/Procfile.dev","services":["web"],"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`
	pipeline, err := Parse([]byte(with))
	if err != nil {
		t.Fatal(err)
	}
	if pipeline.SystemFile.Value != "dev/Procfile.dev" || !reflect.DeepEqual(pipeline.Services, []string{"web"}) {
		t.Fatalf("service declaration = %#v, %#v", pipeline.SystemFile, pipeline.Services)
	}
}

func TestParseRejectsBroaderServiceConfigurations(t *testing.T) {
	tests := map[string]string{
		"multiple services": `{"system_file":"Procfile","services":["web","worker"],"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`,
		"compose":           `{"system_file":"docker-compose.yaml","services":["web"],"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`,
		"file alone":        `{"system_file":"Procfile","start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`,
		"service alone":     `{"services":["web"],"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`,
		"outside workdir":   `{"system_file":"../Procfile","services":["web"],"start":"x","nodes":[{"id":"x","type":"agent","edges":[{"to":"success"}]}]}`,
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(document)); err == nil {
				t.Fatal("configuration was accepted")
			}
		})
	}
}

func TestParseRejectsDuplicateMembersAndNodeIDs(t *testing.T) {
	for name, document := range map[string]string{
		"top":  `{"start":"x","start":"y","nodes":[]}`,
		"node": `{"start":"a","nodes":[{"id":"a","id":"b","type":"agent"}]}`,
		"edge": `{"start":"a","nodes":[{"id":"a","type":"agent","edges":[{"to":"success","to":"failure"}]}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte(document))
			if err == nil || !strings.Contains(err.Error(), "duplicate object member") {
				t.Fatalf("error = %v", err)
			}
		})
	}
	duplicate := `{"start":"same","nodes":[{"id":"same","type":"agent"},{"id":"same","type":"agent"}]}`
	if _, err := Parse([]byte(duplicate)); err == nil || !strings.Contains(err.Error(), "duplicate node ID") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestParseRejectsInvalidAndReservedNodeIDs(t *testing.T) {
	for _, id := range []string{"", "1bad", "bad-name", "bad name", Success, Failure} {
		document := `{"start":"x","nodes":[{"id":` + quoted(id) + `,"type":"agent"}]}`
		if _, err := Parse([]byte(document)); err == nil {
			t.Errorf("invalid ID %q admitted", id)
		}
	}
}

func TestParseYAMLUsesSameContract(t *testing.T) {
	document := `
name: yaml
defaults:
  timeout: 2m
start: work
nodes:
  # Keep commands readable.
  - id: work
    type: command
    label: Run checks
    command: |
      printf '%s\n' first
      printf '%s\n' second
    edges:
      success: success
  - id: coach
    type: supervisor
    prompt: Watch the command
    supervises: [work]
`
	pipeline, err := ParseYAML([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	command := mustNode[*CommandNode](t, pipeline, "work")
	if command.Command != "printf '%s\\n' first\nprintf '%s\\n' second\n" || command.Timeout.Value != "2m" {
		t.Fatalf("command = %#v", command)
	}
	if mustNode[*SupervisorNode](t, pipeline, "coach").IntervalValue() != "60s" {
		t.Fatal("supervisor interval default missing")
	}
	for name, invalid := range map[string]string{
		"duplicate": "start: a\nstart: b\nnodes: []\n",
		"unknown":   "start: a\nunknown: true\nnodes: []\n",
		"null":      "start: a\nname: null\nnodes: []\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseYAML([]byte(invalid)); err == nil {
				t.Fatal("invalid YAML admitted")
			}
		})
	}
}

func TestParseLoopNode(t *testing.T) {
	document := `
defaults:
  timeout: 15m
  model:
    name: gpt-5.6-sol
    effort: low
start: items
nodes:
  - id: items
    type: loop
    checklist: ephemeral/checklist.md
    edges:
      loop: implement
      exit: success
    max_visits: 40
  - id: bare
    type: loop
    edges:
      loop: implement
      exit: items
    timeout: 1m
    item_judge:
      model:
        name: flash
        effort: high
    goal_evaluator:
      model:
        name: fable
        version: "5"
        effort: medium
  - id: implement
    type: agent
    prompt: Implement the current item.
    edges:
      - to: items
`
	pipeline, err := ParseYAML([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	items := mustNode[*LoopNode](t, pipeline, "items")
	if items.NodeType() != "loop" || items.Checklist.Value != "ephemeral/checklist.md" || items.Edges.Loop != "implement" || items.Edges.Exit != Success || items.MaxVisits.Value != 40 {
		t.Fatalf("loop = %#v", items)
	}
	if items.Timeout.Value != "15m" || items.ItemJudge.Present || items.GoalEvaluator.Present {
		t.Fatalf("loop defaults = %#v", items)
	}
	bare := mustNode[*LoopNode](t, pipeline, "bare")
	if bare.Checklist.Present || bare.MaxVisits.Present || bare.Timeout.Value != "1m" || bare.ItemJudge.Value.Model.Value.Name != "flash" || bare.ItemJudge.Value.Model.Value.Effort.Value != "high" || bare.GoalEvaluator.Value.Model.Value.Name != "fable" || bare.GoalEvaluator.Value.Model.Value.Version.Value != "5" || bare.GoalEvaluator.Value.Model.Value.Effort.Value != "medium" {
		t.Fatalf("explicit loop fields = %#v", bare)
	}
	if !reflect.DeepEqual(RoutingTargets(items), []string{"implement", Success}) || !reflect.DeepEqual(RoutingTargets(bare), []string{"implement", "items"}) {
		t.Fatalf("routing targets = %v, %v", RoutingTargets(items), RoutingTargets(bare))
	}
	if ChoiceEdges(items) != nil || !MaxVisits(items).Present || MaxVisits(items).Value != 40 || MaxVisits(bare).Present {
		t.Fatalf("loop accessors = %v, %v", ChoiceEdges(items), MaxVisits(items))
	}
}

func TestDurationSyntaxAndParsing(t *testing.T) {
	for value, want := range map[string]time.Duration{
		"250ms": 250 * time.Millisecond,
		"900s":  900 * time.Second,
		"15m":   15 * time.Minute,
		"2h":    2 * time.Hour,
		"1d":    24 * time.Hour,
	} {
		document := `{"start":"command","nodes":[{"id":"command","type":"command","command":"true","edges":{"success":"success"},"timeout":` + quoted(value) + `}]}`
		if _, err := Parse([]byte(document)); err != nil {
			t.Errorf("duration %q rejected: %v", value, err)
		}
		got, err := Duration(value).Parse()
		if err != nil || got != want {
			t.Errorf("Parse(%q) = %v, %v; want %v", value, got, err, want)
		}
	}
	for _, value := range []string{"1.5s", "-1s", "1", "1 second", "1h30m", ""} {
		document := `{"start":"command","nodes":[{"id":"command","type":"command","command":"true","edges":{"success":"success"},"timeout":` + quoted(value) + `}]}`
		if _, err := Parse([]byte(document)); err == nil {
			t.Errorf("duration %q admitted", value)
		}
	}
}

func TestGraphSchemaIsCommittedAndClosed(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal((Graph{}).Schema(), &root); err != nil {
		t.Fatal(err)
	}
	if root["additionalProperties"] != false || !reflect.DeepEqual(root["required"], []any{"start", "nodes"}) {
		t.Fatalf("top-level schema = %#v", root)
	}
	properties := root["properties"].(map[string]any)
	options := properties["nodes"].(map[string]any)["items"].(map[string]any)["anyOf"].([]any)
	if len(options) != 6 {
		t.Fatalf("node union has %d cases", len(options))
	}
	for _, raw := range options {
		if raw.(map[string]any)["additionalProperties"] != false {
			t.Fatal("node schema is not closed")
		}
	}
	agent := options[0].(map[string]any)
	if !reflect.DeepEqual(agent["required"], []any{"type", "id"}) {
		t.Fatalf("canonical agent required fields = %#v", agent["required"])
	}
	fanOut := options[1].(map[string]any)
	branchOptions := fanOut["properties"].(map[string]any)["branches"].(map[string]any)["items"].(map[string]any)["anyOf"].([]any)
	structured := branchOptions[1].(map[string]any)
	override := structured["properties"].(map[string]any)["agent"].(map[string]any)
	if required, exists := override["required"]; exists && len(required.([]any)) != 0 {
		t.Fatalf("agent override required fields = %#v", required)
	}
}

func mustNode[T Node](t *testing.T, graph *Graph, id string) T {
	t.Helper()
	node, exists := graph.NodeByID(id)
	if !exists {
		t.Fatalf("node %q missing", id)
	}
	typed, ok := node.(T)
	if !ok {
		t.Fatalf("node %q has type %T", id, node)
	}
	return typed
}

func quoted(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}
