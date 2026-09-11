package gimble_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tylergannon/gimble"
)

func TestPublishedEventTypesAreUsableOutsideGimble(t *testing.T) {
	lifecycle := []gimble.LifecycleEvent{
		gimble.RunStarted{},
		gimble.RunEnded{},
		gimble.RunCancelled{},
		gimble.ScopeBegan{},
		gimble.ScopeEnded{},
		gimble.LoopCommand{},
		gimble.PlannerDecision{},
		gimble.ValueSet{},
		gimble.SessionCreated{},
		gimble.SessionClosed{},
		gimble.TurnStarted{},
		gimble.TurnEnded{Tokens: []gimble.JSONText{}},
		gimble.SuperviseAttached{},
		gimble.Steer{},
		gimble.Interrupt{},
		gimble.Complete{},
	}
	agent := []gimble.AgentEvent{
		gimble.UserMessage{},
		gimble.AssistantMessage{},
		gimble.AssistantMessageDelta{},
		gimble.Thinking{},
		gimble.ThinkingDelta{},
		gimble.ToolCall{},
		gimble.ToolInputDelta{},
		gimble.ToolResult{},
		gimble.ToolResultDelta{},
		gimble.Usage{},
		gimble.HarnessError{},
		gimble.Retry{},
		gimble.ApprovalRequest{},
		gimble.NestedTranscript{},
	}

	lifecycleRaw, err := json.Marshal(gimble.LifecycleRecord{Seq: 1, Time: time.Now().UTC(), Event: lifecycle[0]})
	if err != nil {
		t.Fatal(err)
	}
	var lifecycleRecord gimble.LifecycleRecord
	if err := json.Unmarshal(lifecycleRaw, &lifecycleRecord); err != nil {
		t.Fatal(err)
	}
	if _, ok := lifecycleRecord.Event.(gimble.RunStarted); !ok {
		t.Fatalf("decoded lifecycle event = %T", lifecycleRecord.Event)
	}

	agentRaw, err := json.Marshal(gimble.AgentRecord{Seq: 1, Time: time.Now().UTC(), Session: "coder.1", Turn: "coder.1/turn.1", Event: agent[0]})
	if err != nil {
		t.Fatal(err)
	}
	var agentRecord gimble.AgentRecord
	if err := json.Unmarshal(agentRaw, &agentRecord); err != nil {
		t.Fatal(err)
	}
	if _, ok := agentRecord.Event.(gimble.UserMessage); !ok {
		t.Fatalf("decoded agent event = %T", agentRecord.Event)
	}
}
