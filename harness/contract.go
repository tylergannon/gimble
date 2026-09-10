// Package harness defines the provider-neutral contract shared by coding
// agent harness adapters and their callers.
package harness

import (
	"context"
	"encoding/json"
)

// ContentPart is one ordered part of a user message.
type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

const ContentPartText = "text"

// RunTurnInput is the fully resolved input to one harness turn. OutputSchema
// is optional; when present the adapter validates the agent's result against
// it exactly.
type RunTurnInput struct {
	SessionID       string
	Model           string
	ReasoningEffort string
	OutputSchema    json.RawMessage
	Workdir         string
	Parts           []ContentPart
}

// Event is one complete, provider-neutral logical item emitted during a turn.
// The event type set is open, so the representation preserves unknown fields.
type Event map[string]any

// OnEvent receives complete events in turn order.
type OnEvent func(Event)

const (
	EventUser       = "user"
	EventAssistant  = "assistant"
	EventThinking   = "thinking"
	EventToolCall   = "tool_call"
	EventToolResult = "tool_result"
	EventUsage      = "usage"
)

// HarnessAdapter translates the neutral contract into one native coding-agent
// harness. It is a primitive: sessions keyed by ID, one live turn per
// session, no routing, no policy.
//
// RunTurn blocks until the turn ends. With an OutputSchema it returns the
// validated JSON object; without one it returns the assistant's final text
// encoded as a JSON string. Cancelling ctx interrupts the native turn and
// RunTurn returns ctx.Err().
type HarnessAdapter interface {
	CreateSession(model, workdir string) (string, error)
	RunTurn(ctx context.Context, input RunTurnInput, onEvent OnEvent) (json.RawMessage, error)
	Steer(sessionID string, parts []ContentPart)
	Interrupt(sessionID string)
	Compact(sessionID, workdir string) error
}
