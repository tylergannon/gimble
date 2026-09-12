package gimble

import (
	"context"
	"encoding/json"
)

// HarnessAdapter is the agent-specific code for one coding-agent harness
// (for example, Codex or Claude Code). It is untyped: a raw JSON Schema goes
// in, raw JSON comes out; Generate validates and decodes.
type HarnessAdapter interface {
	// CreateSession starts a native session and returns its id.
	CreateSession(ctx context.Context, model, workdir string) (string, error)

	// RunTurn runs one turn and blocks until it ends. With a schema it
	// returns the structured result; without one, the final message encoded
	// as a JSON string. It passes every event to onEvent as it arrives.
	// Cancelling ctx interrupts the native turn and returns ctx.Err().
	RunTurn(ctx context.Context, sessionID, prompt string, schema json.RawMessage, onEvent func(AgentEvent) error) (json.RawMessage, error)

	// Steer sends a message into the session's running turn. With no turn
	// running it does nothing and returns nil.
	Steer(ctx context.Context, sessionID, message string) error

	// Fork returns a new native session with the conversation so far.
	Fork(ctx context.Context, sessionID string) (string, error)
}
