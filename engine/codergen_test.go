package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

func TestAgentHandlerBuildsExactTurnSchemaAndArtifacts(t *testing.T) {
	stageDir := t.TempDir()
	backend := &captureBackend{outcome: harness.Outcome{Next: "review", Notes: "approved\nwith details"}}
	handler := NewAgentHandler(AgentConfig{
		Backend:                backend,
		DefaultModel:           "system-model",
		DefaultProvider:        "system-provider",
		DefaultReasoningEffort: "low",
	})
	node := &graph.AgentNode{
		NodeBase: graph.NodeBase{ID: "plan", Label: optional("Plan")},
		LLMNodeFields: graph.LLMNodeFields{
			Prompt:          optional("Do $goal, then $goal"),
			LLMModel:        optional("claude-opus-4-6"),
			LLMProvider:     optional("anthropic"),
			ReasoningEffort: optional("high"),
			Fidelity:        optional("none"),
			ThreadID:        optional("ignored-thread"),
			Timeout:         optional(graph.Duration("3s")),
		},
	}
	pipeline := &graph.Graph{
		Goal: "unused graph goal",
		Defaults: graph.Defaults{
			LLMModel:        optional("file-model"),
			LLMProvider:     optional("file-provider"),
			ReasoningEffort: optional("medium"),
			Fidelity:        optional("full"),
			Timeout:         optional(graph.Duration("2s")),
		},
		Nodes: []graph.Node{
			node,
			&graph.AgentNode{NodeBase: graph.NodeBase{ID: "implement", Label: optional("Implement")}},
			&graph.AgentNode{NodeBase: graph.NodeBase{ID: "review", Label: optional("Review")}},
		},
	}
	offered := []graph.Edge{
		{To: "implement", Condition: "Needs work"},
		{To: "review", Condition: "Looks good"},
	}

	outcome, runErr := handler.Execute(node, offered, ExecutionScope{
		Workdir:  "/workspace",
		StageDir: stageDir,
		RunLog:   filepath.Join(stageDir, "events.jsonl"),
		Goal:     "ship it",
		Stop:     NewStopSignal(),
	}, pipeline)
	if runErr != nil {
		t.Fatal(runErr)
	}
	if !reflect.DeepEqual(outcome, backend.outcome) {
		t.Fatalf("outcome = %#v", outcome)
	}
	wantSchema := `{"type":"object","properties":{"next":{"type":"string","enum":["implement","review"],"description":"Choose the next stage. Needs work: implement; Looks good: review"},"notes":{"type":"string","description":"Your account of this stage."}},"required":["next","notes"],"additionalProperties":false}`
	wantTurn := harness.AgentTurn{
		NodeID:          "plan",
		Parts:           []harness.ContentPart{{Type: harness.ContentPartText, Text: "Do ship it, then ship it"}},
		OutputSchema:    json.RawMessage(wantSchema),
		Model:           "claude-opus-4-6",
		Provider:        "anthropic",
		ReasoningEffort: "high",
		Fidelity:        harness.FidelityNone,
		ThreadKey:       "",
		Workdir:         "/workspace",
		RunLog:          filepath.Join(stageDir, "events.jsonl"),
		Timeout:         3 * time.Second,
	}
	if !reflect.DeepEqual(backend.turns, []harness.AgentTurn{wantTurn}) {
		t.Fatalf("turn = %#v\nwant %#v", backend.turns, wantTurn)
	}
	assertTextFile(t, filepath.Join(stageDir, "prompt.md"), "Do ship it, then ship it")
	assertTextFile(t, filepath.Join(stageDir, "response.md"), "---\nnext: review\n---\napproved\nwith details")
}

func TestAgentHandlerResolutionPrecedenceAndProviderAutodetection(t *testing.T) {
	tests := []struct {
		name          string
		nodeFields    graph.LLMNodeFields
		defaults      graph.Defaults
		config        AgentConfig
		wantModel     string
		wantProvider  string
		wantReasoning string
		wantFidelity  harness.FidelityMode
		wantThread    string
		wantTimeout   time.Duration
	}{
		{
			name: "node beats file and system",
			nodeFields: graph.LLMNodeFields{
				LLMModel: optional("gpt-5.3-codex"), LLMProvider: optional("openai"),
				ReasoningEffort: optional("low"), Fidelity: optional("full"),
				ThreadID: optional("shared"), Timeout: optional(graph.Duration("9s")),
			},
			defaults: graph.Defaults{
				LLMModel: optional("file-model"), LLMProvider: optional("file-provider"),
				ReasoningEffort: optional("medium"), Fidelity: optional("none"),
				Timeout: optional(graph.Duration("8s")),
			},
			config:        AgentConfig{DefaultModel: "system-model", DefaultProvider: "system-provider", DefaultReasoningEffort: "high"},
			wantModel:     "gpt-5.3-codex",
			wantProvider:  "openai",
			wantReasoning: "low",
			wantFidelity:  harness.FidelityFull,
			wantThread:    "shared",
			wantTimeout:   9 * time.Second,
		},
		{
			name:       "file beats system",
			nodeFields: graph.LLMNodeFields{ThreadID: optional("file-thread")},
			defaults: graph.Defaults{
				LLMModel: optional("file-model"), LLMProvider: optional("file-provider"),
				ReasoningEffort: optional("medium"), Fidelity: optional("compacted"),
				Timeout: optional(graph.Duration("7s")),
			},
			config:        AgentConfig{DefaultModel: "system-model", DefaultProvider: "system-provider", DefaultReasoningEffort: "high"},
			wantModel:     "file-model",
			wantProvider:  "file-provider",
			wantReasoning: "medium",
			wantFidelity:  harness.FidelityCompacted,
			wantThread:    "file-thread",
			wantTimeout:   7 * time.Second,
		},
		{
			name:          "system defaults",
			config:        AgentConfig{DefaultModel: "system-model", DefaultProvider: "system-provider", DefaultReasoningEffort: "medium"},
			wantModel:     "system-model",
			wantProvider:  "system-provider",
			wantReasoning: "medium",
			wantFidelity:  harness.FidelityCompacted,
			wantThread:    "work",
		},
		{
			name:          "system model auto-detects provider",
			config:        AgentConfig{DefaultModel: "claude-sonnet-4-5"},
			wantModel:     "claude-sonnet-4-5",
			wantProvider:  "anthropic",
			wantReasoning: "high",
			wantFidelity:  harness.FidelityCompacted,
			wantThread:    "work",
		},
		{
			name:          "unversioned Fable resolves to 5.1",
			config:        AgentConfig{DefaultModel: "fable"},
			wantModel:     "claude-fable-5-1",
			wantProvider:  "anthropic",
			wantReasoning: "high",
			wantFidelity:  harness.FidelityCompacted,
			wantThread:    "work",
		},
		{
			name:          "none fidelity has no thread key",
			nodeFields:    graph.LLMNodeFields{Fidelity: optional("none"), ThreadID: optional("ignored")},
			config:        AgentConfig{DefaultModel: "gemini-2.5-pro"},
			wantModel:     "gemini-2.5-pro",
			wantProvider:  "gemini",
			wantReasoning: "high",
			wantFidelity:  harness.FidelityNone,
			wantThread:    "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := &captureBackend{outcome: harness.Outcome{Notes: "ok"}}
			test.config.Backend = backend
			node := &graph.AgentNode{NodeBase: graph.NodeBase{ID: "work", Label: optional("Work")}, LLMNodeFields: test.nodeFields}
			pipeline := &graph.Graph{Defaults: test.defaults, Nodes: []graph.Node{node, exitNode("done")}}
			stageDir := t.TempDir()
			_, runErr := NewAgentHandler(test.config).Execute(node, []graph.Edge{{To: "done"}}, ExecutionScope{
				Workdir: "/workspace", StageDir: stageDir, RunLog: filepath.Join(stageDir, "events.jsonl"), Stop: NewStopSignal(),
			}, pipeline)
			if runErr != nil {
				t.Fatal(runErr)
			}
			turn := backend.turns[0]
			if turn.Model != test.wantModel || turn.Provider != test.wantProvider || turn.ReasoningEffort != test.wantReasoning ||
				turn.Fidelity != test.wantFidelity || turn.ThreadKey != test.wantThread || turn.Timeout != test.wantTimeout {
				t.Fatalf("resolved turn = %#v", turn)
			}
		})
	}
}

func TestAgentHandlerRejectsProviderConflictWithModelAlias(t *testing.T) {
	node := &graph.AgentNode{
		NodeBase: graph.NodeBase{ID: "work"},
		LLMNodeFields: graph.LLMNodeFields{
			LLMProvider: optional("openai"),
			LLMModel:    optional("fable"),
		},
	}
	pipeline := &graph.Graph{Nodes: []graph.Node{node, exitNode("done")}}
	stageDir := t.TempDir()
	_, runErr := NewAgentHandler(AgentConfig{}).Execute(node, []graph.Edge{{To: "done"}}, ExecutionScope{
		Workdir: "/workspace", StageDir: stageDir, RunLog: filepath.Join(stageDir, "events.jsonl"), Stop: NewStopSignal(),
	}, pipeline)
	if runErr == nil || !strings.Contains(runErr.Message, `provider "openai" conflicts with model alias "fable"`) {
		t.Fatalf("conflict error = %#v", runErr)
	}
}

func TestChoiceSchemaOmitsNextForZeroOrOneSuccessor(t *testing.T) {
	want := `{"type":"object","properties":{"notes":{"type":"string","description":"Your account of this stage."}},"required":["notes"],"additionalProperties":false}`
	for _, offered := range [][]graph.Edge{nil, {{To: "done"}}} {
		schema, err := choiceSchema(offered, &graph.Graph{})
		if err != nil {
			t.Fatal(err)
		}
		if string(schema) != want {
			t.Fatalf("schema = %s", schema)
		}
	}
}

func TestChoiceSchemaRouteDescriptionFallsBackToLabelThenID(t *testing.T) {
	pipeline := &graph.Graph{Nodes: []graph.Node{
		&graph.SupervisorNode{NodeBase: graph.NodeBase{ID: "retry", Label: optional("Try again")}, Prompt: "__test_exit__"},
		exitNode("done"),
	}}
	schema, err := choiceSchema([]graph.Edge{{To: "retry"}, {To: "done"}}, pipeline)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"type":"object","properties":{"next":{"type":"string","enum":["retry","done"],"description":"Choose the next stage. Try again: retry; done: done"},"notes":{"type":"string","description":"Your account of this stage."}},"required":["next","notes"],"additionalProperties":false}`
	if string(schema) != want {
		t.Fatalf("schema = %s", schema)
	}
}

func TestAgentHandlerSimulationRoutingAndPromptFallback(t *testing.T) {
	tests := []struct {
		name     string
		offered  []graph.Edge
		wantNext string
	}{
		{name: "single has no choice", offered: []graph.Edge{{To: "left"}}},
		{name: "multiple chooses first", offered: []graph.Edge{{To: "left"}, {To: "right"}}, wantNext: "left"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stageDir := t.TempDir()
			node := &graph.AgentNode{
				NodeBase:      graph.NodeBase{ID: "plan", Label: optional("Plan $goal")},
				LLMNodeFields: graph.LLMNodeFields{Prompt: optional("")},
			}
			pipeline := &graph.Graph{Nodes: []graph.Node{node, exitNode("left"), exitNode("right")}}
			outcome, runErr := NewAgentHandler(AgentConfig{DefaultModel: "gpt-5.3-codex"}).Execute(
				node, test.offered, ExecutionScope{Workdir: "/workspace", StageDir: stageDir, Goal: "release", Stop: NewStopSignal()}, pipeline,
			)
			if runErr != nil {
				t.Fatal(runErr)
			}
			want := harness.Outcome{Next: test.wantNext, Notes: "[Simulated] Stage completed: plan"}
			if !reflect.DeepEqual(outcome, want) {
				t.Fatalf("outcome = %#v", outcome)
			}
			assertTextFile(t, filepath.Join(stageDir, "prompt.md"), "Plan release")
			response := "---\n"
			if test.wantNext != "" {
				response += "next: " + test.wantNext + "\n"
			}
			response += "---\n[Simulated] Stage completed: plan"
			assertTextFile(t, filepath.Join(stageDir, "response.md"), response)
		})
	}
}

func TestAgentHandlerPassesBackendErrorUnchanged(t *testing.T) {
	stageDir := t.TempDir()
	wantError := &harness.Error{Category: harness.ErrorRetryable, Message: "try later"}
	backend := &captureBackend{runErr: wantError}
	node := &graph.AgentNode{NodeBase: graph.NodeBase{ID: "work", Label: optional("Work")}}
	_, runErr := NewAgentHandler(AgentConfig{Backend: backend, DefaultModel: "gpt-5.3-codex"}).Execute(
		node, []graph.Edge{{To: "done"}}, ExecutionScope{Workdir: "/workspace", StageDir: stageDir, RunLog: filepath.Join(stageDir, "events.jsonl"), Stop: NewStopSignal()},
		&graph.Graph{Nodes: []graph.Node{node, exitNode("done")}},
	)
	if runErr != wantError {
		t.Fatalf("error = %#v, want original pointer %#v", runErr, wantError)
	}
	assertTextFile(t, filepath.Join(stageDir, "prompt.md"), "Work")
	if _, err := os.Stat(filepath.Join(stageDir, "response.md")); !os.IsNotExist(err) {
		t.Fatalf("response exists after backend error: %v", err)
	}
}

func TestAgentHandlerRejectsInvalidResolvedTurnInSimulation(t *testing.T) {
	tests := []struct {
		name       string
		config     AgentConfig
		fields     graph.LLMNodeFields
		wantReason string
	}{
		{name: "missing provider", wantReason: "provider must not be empty"},
		{name: "missing model", config: AgentConfig{DefaultProvider: "openai"}, wantReason: "model must not be empty"},
		{name: "unsupported fidelity", config: AgentConfig{DefaultModel: "gpt-5.3-codex"}, fields: graph.LLMNodeFields{Fidelity: optional("unknown")}, wantReason: `unsupported fidelity "unknown"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stageDir := t.TempDir()
			node := &graph.AgentNode{NodeBase: graph.NodeBase{ID: "work", Label: optional("Work")}, LLMNodeFields: test.fields}
			_, runErr := NewAgentHandler(test.config).Execute(
				node, []graph.Edge{{To: "done"}}, ExecutionScope{Workdir: "/workspace", StageDir: stageDir, Stop: NewStopSignal()},
				&graph.Graph{Nodes: []graph.Node{node, exitNode("done")}},
			)
			if runErr == nil || runErr.Category != harness.ErrorTerminal || runErr.Message != test.wantReason {
				t.Fatalf("error = %#v", runErr)
			}
			assertTextFile(t, filepath.Join(stageDir, "prompt.md"), "Work")
			if _, err := os.Stat(filepath.Join(stageDir, "response.md")); !os.IsNotExist(err) {
				t.Fatalf("response exists after invalid turn: %v", err)
			}
		})
	}
}

func TestDetectProvider(t *testing.T) {
	tests := map[string]string{
		"gpt-5.6-sol":        "openai",
		"o3":                 "openai",
		"claude-opus-4-6":    "anthropic",
		"gemini-2.5-pro":     "gemini",
		"unrecognized-model": "",
	}
	for model, want := range tests {
		if got := DetectProvider(model); got != want {
			t.Errorf("DetectProvider(%q) = %q, want %q", model, got, want)
		}
	}
}

func optional[T any](value T) jsonschema.Optional[T] {
	return jsonschema.Optional[T]{Present: true, Value: value}
}

func assertTextFile(t *testing.T, path, want string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != want {
		t.Fatalf("%s = %q, want %q", path, raw, want)
	}
}

type captureBackend struct {
	turns   []harness.AgentTurn
	outcome harness.Outcome
	runErr  *harness.Error
}

func (b *captureBackend) Run(turn harness.AgentTurn) (harness.Outcome, *harness.Error) {
	b.turns = append(b.turns, turn)
	return b.outcome, b.runErr
}

func (*captureBackend) RunSupervisor(harness.SupervisorTurn) (harness.Verdict, *harness.Error) {
	return harness.Verdict{}, nil
}

func (*captureBackend) Steer([]harness.ContentPart) harness.SteerStatus {
	return harness.SteerNotActive
}

func (*captureBackend) InterruptAll() {}

func (*captureBackend) Bindings() map[string]harness.ThreadBinding { return nil }

func (*captureBackend) SetBindingOpened(harness.BindingOpened) {}
