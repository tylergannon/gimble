package gimble

import (
	"context"
	"encoding/json"
	"time"
)

// HarnessAdapter is the agent-specific code for one coding-agent harness
// (Codex, Claude Code). It is untyped: a raw JSON Schema goes in and raw
// JSON comes out; Generate validates and decodes.
type HarnessAdapter interface {
	// CreateSession starts a native session and returns its id.
	CreateSession(ctx context.Context, model, workdir string) (string, error)

	// RunTurn runs one turn and blocks until it ends. With a schema it
	// returns the structured result; without one, the final message encoded
	// as a JSON string. It passes every event to onEvent as it arrives.
	// Cancelling ctx interrupts the native turn and returns ctx.Err().
	RunTurn(ctx context.Context, sessionID, prompt string, schema json.RawMessage, onEvent func(Event)) (json.RawMessage, error)

	// Steer sends a message into the session's running turn. With no turn
	// running it does nothing and returns nil.
	Steer(ctx context.Context, sessionID, message string) error

	// Fork returns a new native session with the conversation so far.
	Fork(ctx context.Context, sessionID string) (string, error)
}

// Event is one thing the harness did during a turn. Kind is "user" (the
// prompt, and each steer as it lands), "assistant", "thinking",
// "tool_call", "tool_result", or "usage". Text carries the message for the
// first three; CallID pairs a tool call with its result; Tool names the
// tool; Data carries tool arguments, tool output, or token usage as JSON.
type Event struct {
	Seq         uint64          `json:"seq,omitempty"`
	Time        time.Time       `json:"time,omitempty"`
	Kind        string          `json:"kind"`
	Scope       string          `json:"scope,omitempty"`
	Session     string          `json:"session,omitempty"`
	Turn        string          `json:"turn,omitempty"`
	Name        string          `json:"name,omitempty"`
	Task        string          `json:"task,omitempty"`
	Key         string          `json:"key,omitempty"`
	Value       json.RawMessage `json:"value,omitempty"`
	Error       string          `json:"error,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	OutputType  string          `json:"output_type,omitempty"`
	Tokens      any             `json:"tokens,omitempty"`
	Duration    time.Duration   `json:"duration,omitempty"`
	Interrupted bool            `json:"interrupted,omitempty"`
	ExitCode    int             `json:"exit_code,omitempty"`
	Command     string          `json:"command,omitempty"`
	Decision    string          `json:"decision,omitempty"`
	Adapter     string          `json:"adapter,omitempty"`
	Model       string          `json:"model,omitempty"`
	Workdir     string          `json:"workdir,omitempty"`
	Parent      string          `json:"parent,omitempty"`
	Target      string          `json:"target,omitempty"`
	Source      string          `json:"source,omitempty"`
	Message     string          `json:"message,omitempty"`
	Landed      *bool           `json:"landed,omitempty"`
	Reviewer    string          `json:"reviewer,omitempty"`
	Worker      string          `json:"worker,omitempty"`
	Instruction string          `json:"instruction,omitempty"`
	Interval    time.Duration   `json:"interval,omitempty"`
	Delta       string          `json:"delta,omitempty"`
	Text        string          `json:"text,omitempty"`
	CallID      string          `json:"call_id,omitempty"`
	Tool        string          `json:"tool,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}
