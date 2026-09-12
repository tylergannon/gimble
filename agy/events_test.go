package agy

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/tylergannon/gimble"
)

func TestProjectorTranslatesResponseAndToolSteps(t *testing.T) {
	var events []gimble.AgentEvent
	p := newProjector("conversation-1", "gemini-test-low", func(event gimble.AgentEvent) error {
		events = append(events, event)
		return nil
	})
	p.setConversation("conversation-1")
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "conversation-1", StepIndex: 1, StepType: "agent_response", State: "ACTIVE", TextDelta: "hel",
	}}))
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "conversation-1", StepIndex: 1, StepType: "agent_response", State: "DONE", TextDelta: "lo",
		Usage: map[string]any{"input_tokens": float64(10), "output_tokens": float64(4), "thinking_tokens": float64(1), "cache_read_tokens": float64(2)},
	}}))
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "conversation-1", StepIndex: 2, StepType: "tool", State: "ACTIVE",
		ToolInfo: &toolInfo{Name: "run_command", Parameters: map[string]any{"CommandLine": "pwd"}},
	}}))
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "conversation-1", StepIndex: 2, StepType: "tool", State: "DONE",
		ToolInfo: &toolInfo{Name: "run_command", Parameters: map[string]any{"CommandLine": "pwd"}, Output: "/work\n"},
	}}))

	want := []string{
		"session.step.started", "session.text.started", "session.text.delta", "session.text.delta",
		"session.text.ended", "session.step.streamed", "session.step.ended",
		"session.step.started", "session.tool.input.started", "session.tool.input.ended",
		"session.tool.called", "session.tool.success", "session.step.streamed", "session.step.ended",
	}
	if got := eventTypes(events); !slices.Equal(got, want) {
		t.Fatalf("event types = %v, want %v", got, want)
	}
	var ended map[string]any
	if err := json.Unmarshal(events[6].Data, &ended); err != nil {
		t.Fatal(err)
	}
	tokens := ended["tokens"].(map[string]any)
	if tokens["output"] != float64(3) || tokens["reasoning"] != float64(1) {
		t.Fatalf("normalized tokens = %#v", tokens)
	}
	if !json.Valid(events[6].NativeRef) {
		t.Fatal("invalid native ref")
	}
	var ref map[string]any
	if err := json.Unmarshal(events[6].NativeRef, &ref); err != nil {
		t.Fatal(err)
	}
	accounting := ref["accounting"].(map[string]any)
	if accounting["tokensAvailable"] != false || accounting["costAvailable"] != false {
		t.Fatalf("accounting = %#v", accounting)
	}
}

func TestProjectorPropagatesCallbackAndProtocolErrors(t *testing.T) {
	boom := errors.New("observer stopped")
	p := newProjector("conversation-1", "model", func(gimble.AgentEvent) error { return boom })
	err := p.envelope(envelope{StepUpdate: &stepUpdate{ConversationID: "conversation-1", StepIndex: 1, StepType: "agent_response", State: "DONE"}})
	if !errors.Is(err, boom) {
		t.Fatalf("callback error = %v", err)
	}

	p = newProjector("conversation-1", "model", func(gimble.AgentEvent) error { return nil })
	p.setConversation("conversation-1")
	err = p.envelope(envelope{StepUpdate: &stepUpdate{ConversationID: "other", StepIndex: 1, StepType: "agent_response", State: "DONE"}})
	if err == nil {
		t.Fatal("conversation mismatch was accepted")
	}
}

func TestProjectorPreservesNativeConversationIdentity(t *testing.T) {
	var event gimble.AgentEvent
	p := newProjector("adapter-session", "model", func(value gimble.AgentEvent) error {
		event = value
		return nil
	})
	p.setConversation("native-conversation")
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "native-conversation", StepIndex: 3, StepType: "agent_response", State: "DONE",
	}}))
	var data, ref map[string]any
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(event.NativeRef, &ref); err != nil {
		t.Fatal(err)
	}
	if data["sessionID"] != "native-conversation" || ref["sessionID"] != "native-conversation" {
		t.Fatalf("data=%#v ref=%#v", data, ref)
	}
	if ref["messageID"] != "native-conversation/step.3" {
		t.Fatalf("messageID = %v", ref["messageID"])
	}
}

func TestProjectorSettlesStructuredFinishToolFromResult(t *testing.T) {
	var events []gimble.AgentEvent
	p := newProjector("adapter-session", "model", func(value gimble.AgentEvent) error {
		events = append(events, value)
		return nil
	})
	p.setConversation("native-conversation")
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "native-conversation", StepIndex: 4, StepType: "tool", State: "ACTIVE",
		ToolInfo: &toolInfo{Name: "finish", Parameters: map[string]any{"answer": "yes"}},
	}}))
	mustProject(t, p.envelope(envelope{Result: &result{
		ConversationID: "native-conversation", Status: "SUCCESS", StructuredOutput: map[string]any{"answer": "yes"},
	}}))
	want := []string{
		"session.step.started", "session.tool.input.started", "session.tool.input.ended", "session.tool.called",
		"session.tool.success", "session.step.streamed", "session.step.ended",
	}
	if got := eventTypes(events); !slices.Equal(got, want) {
		t.Fatalf("event types = %v, want %v", got, want)
	}
}

func TestProjectorRejectsUnsettledNonFinishStepAtResult(t *testing.T) {
	p := newProjector("adapter-session", "model", func(gimble.AgentEvent) error { return nil })
	p.setConversation("native-conversation")
	mustProject(t, p.envelope(envelope{StepUpdate: &stepUpdate{
		ConversationID: "native-conversation", StepIndex: 2, StepType: "tool", State: "ACTIVE",
		ToolInfo: &toolInfo{Name: "run_command", Parameters: map[string]any{"CommandLine": "sleep 1"}},
	}}))
	if err := p.envelope(envelope{Result: &result{Status: "SUCCESS"}}); err == nil {
		t.Fatal("result accepted an unsettled non-finish tool")
	}
}

func TestScanStreamRejectsMalformedNDJSON(t *testing.T) {
	output := make(chan streamItem)
	go scanStream(strings.NewReader("not-json\n"), output)
	item := <-output
	if item.err == nil {
		t.Fatal("malformed stream was accepted")
	}
}

func eventTypes(events []gimble.AgentEvent) []string {
	out := make([]string, len(events))
	for i, event := range events {
		out[i] = event.Type
	}
	return out
}

func mustProject(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
