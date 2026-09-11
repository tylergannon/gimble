package gimble

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestAgentEventUnionRoundTripsEveryVariant(t *testing.T) {
	t.Parallel()
	when := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		kind  string
		event AgentEvent
	}{
		{"user_message", UserMessage{Text: "hello"}},
		{"assistant_message", AssistantMessage{ID: "message-1", Text: "hello"}},
		{"assistant_message_delta", AssistantMessageDelta{ID: "message-1", Text: "hel"}},
		{"thinking", Thinking{ID: "thought-1", Text: "considering"}},
		{"thinking_delta", ThinkingDelta{ID: "thought-1", Text: "cons"}},
		{"tool_call", ToolCall{CallID: "call-1", Tool: "shell", Input: JSONText(`{"command":"go test ./..."}`)}},
		{"tool_input_delta", ToolInputDelta{CallID: "call-1", Input: JSONText(`{"command":"go`)}},
		{"tool_result", ToolResult{CallID: "call-1", Output: JSONText(`{"exit_code":0}`)}},
		{"tool_result_delta", ToolResultDelta{CallID: "call-1", Output: JSONText(`{"exit`)}},
		{"usage", Usage{Data: JSONText(`{"input_tokens":12,"output_tokens":5}`)}},
		{"harness_error", HarnessError{Error: "connection closed"}},
		{"retry", Retry{Attempt: 2, Error: "rate limited"}},
		{"approval_request", ApprovalRequest{ID: "approval-1", Tool: "shell", Input: JSONText(`{"command":"deploy"}`)}},
		{"nested_transcript", NestedTranscript{CallID: "call-2", Transcript: JSONText(`[{"kind":"user_message","text":"inspect"}]`)}},
	}

	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			record := AgentRecord{Seq: 1, Time: when, Scope: "lap.1", Session: "coder.1", Turn: "coder.1/turn.1", Event: test.event}
			raw, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(raw, []byte(`"kind":"`+test.kind+`"`)) {
				t.Fatalf("record lacks discriminator %q: %s", test.kind, raw)
			}
			if err := record.ValidateJSON(raw); err != nil {
				t.Fatalf("generated schema rejected %s: %v", raw, err)
			}
			var decoded AgentRecord
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			if got := agentKind(decoded.Event); got != test.kind {
				t.Fatalf("decoded kind = %q, want %q", got, test.kind)
			}
			roundTrip, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(roundTrip, raw) {
				t.Fatalf("round trip changed record:\n%s\n%s", raw, roundTrip)
			}
		})
	}
}

func TestLifecycleEventUnionRoundTripsEveryVariant(t *testing.T) {
	t.Parallel()
	when := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		kind  string
		event LifecycleEvent
	}{
		{"run_started", RunStarted{Name: "sprint"}},
		{"run_ended", RunEnded{Name: "sprint", Error: ""}},
		{"run_cancelled", RunCancelled{Name: "sprint", Source: "operator", Error: "context canceled"}},
		{"scope_began", ScopeBegan{Name: "lap", Task: "implement"}},
		{"scope_ended", ScopeEnded{Error: ""}},
		{"loop_command", LoopCommand{Command: "go test ./...", ExitCode: 0}},
		{"planner_decision", PlannerDecision{Decision: "implement events"}},
		{"value_set", ValueSet{Key: "goal", Value: JSONText(`"ship"`)}},
		{"session_created", SessionCreated{Name: "coder", Adapter: "codex", Model: "gpt", Workdir: "/work", Parent: "researcher.1"}},
		{"session_closed", SessionClosed{}},
		{"turn_started", TurnStarted{Prompt: "build", OutputType: "gimble.Text"}},
		{"turn_ended", TurnEnded{Result: JSONText(`"done"`), Tokens: []JSONText{JSONText(`{"input":1}`)}, Duration: time.Second}},
		{"supervise_attached", SuperviseAttached{Reviewer: "reviewer.1", Worker: "worker.1/turn.1", Instruction: "watch", Interval: time.Minute}},
		{"steer", Steer{Target: "worker.1", Source: "reviewer.1", Message: "fix it", Landed: true}},
		{"interrupt", Interrupt{Target: "worker.1", Source: "operator"}},
		{"complete", Complete{}},
	}

	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			record := LifecycleRecord{Seq: 1, Time: when, Scope: "lap.1", Session: optionalString("coder.1"), Turn: optionalString("coder.1/turn.1"), Event: test.event}
			raw, err := json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(raw, []byte(`"kind":"`+test.kind+`"`)) {
				t.Fatalf("record lacks discriminator %q: %s", test.kind, raw)
			}
			if err := record.ValidateJSON(raw); err != nil {
				t.Fatalf("generated schema rejected %s: %v", raw, err)
			}
			var decoded LifecycleRecord
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatal(err)
			}
			if got := lifecycleKind(decoded.Event); got != test.kind {
				t.Fatalf("decoded kind = %q, want %q", got, test.kind)
			}
			roundTrip, err := json.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(roundTrip, raw) {
				t.Fatalf("round trip changed record:\n%s\n%s", raw, roundTrip)
			}
		})
	}
}

func TestEventUnionsRejectUnknownVariants(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"seq":1,"time":"2026-09-11T12:00:00Z","scope":"","session":"coder.1","turn":"coder.1/turn.1","event":{"kind":"unknown"}}`)
	if err := (AgentRecord{}).ValidateJSON(raw); err == nil {
		t.Fatal("generated schema accepted an unknown agent event")
	}
	var record AgentRecord
	if err := json.Unmarshal(raw, &record); err == nil {
		t.Fatal("generated codec accepted an unknown agent event")
	}
}
