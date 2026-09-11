package gimble

import (
	"time"

	"github.com/tylergannon/polytype"
)

// JSONText is JSON source text carried opaquely by an event. Complete events
// contain one encoded value; delta events may contain only the next fragment.
// Text preserves provider-specific payloads while the event remains a closed
// Polytype shape.
type JSONText string

// AgentEvent is one typed piece of a harness transcript. The interface is
// sealed: adapters emit the concrete variants Gimble defines here.
type AgentEvent interface{ agentEvent() }

// UserMessage is one complete prompt or landed steer sent to an agent.
type UserMessage struct {
	// Text is the complete message sent to the agent.
	Text string `json:"text"`
}

func (UserMessage) agentEvent() {}

// AssistantMessage is one complete prose message from an agent.
type AssistantMessage struct {
	// ID identifies this message within its turn.
	ID string `json:"id"`
	// Text is the complete assistant message.
	Text string `json:"text"`
}

func (AssistantMessage) agentEvent() {}

// AssistantMessageDelta is the next fragment of an assistant message.
type AssistantMessageDelta struct {
	// ID identifies the assistant message this fragment extends.
	ID string `json:"id"`
	// Text is the next fragment of the assistant message.
	Text string `json:"text"`
}

func (AssistantMessageDelta) agentEvent() {}

// Thinking is one complete reasoning item exposed by a harness.
type Thinking struct {
	// ID identifies this thinking item within its turn.
	ID string `json:"id"`
	// Text is the complete thinking item.
	Text string `json:"text"`
}

func (Thinking) agentEvent() {}

// ThinkingDelta is the next fragment of a reasoning item.
type ThinkingDelta struct {
	// ID identifies the thinking item this fragment extends.
	ID string `json:"id"`
	// Text is the next fragment of the thinking item.
	Text string `json:"text"`
}

func (ThinkingDelta) agentEvent() {}

// ToolCall is one complete tool invocation.
type ToolCall struct {
	// CallID pairs the call with its input fragments and result.
	CallID string `json:"call_id"`
	// Tool is the harness's tool name.
	Tool string `json:"tool"`
	// Input is the complete tool input encoded as JSON text.
	Input JSONText `json:"input"`
}

func (ToolCall) agentEvent() {}

// ToolInputDelta is the next fragment of a tool call's input.
type ToolInputDelta struct {
	// CallID identifies the tool call whose input this fragment extends.
	CallID string `json:"call_id"`
	// Input is the next fragment of the tool input's JSON source text.
	Input JSONText `json:"input"`
}

func (ToolInputDelta) agentEvent() {}

// ToolResult is one complete result of a tool call.
type ToolResult struct {
	// CallID identifies the completed tool call.
	CallID string `json:"call_id"`
	// Output is the complete tool result encoded as JSON text.
	Output JSONText `json:"output"`
}

func (ToolResult) agentEvent() {}

// ToolResultDelta is the next fragment of a tool result.
type ToolResultDelta struct {
	// CallID identifies the tool result this fragment extends.
	CallID string `json:"call_id"`
	// Output is the next fragment of the tool result's JSON source text.
	Output JSONText `json:"output"`
}

func (ToolResultDelta) agentEvent() {}

// Usage is a harness's token-usage payload for one model call.
type Usage struct {
	// Data is the harness's token-usage payload encoded as JSON text.
	Data JSONText `json:"data"`
}

func (Usage) agentEvent() {}

// HarnessError is an error reported by the harness while a turn is running.
type HarnessError struct {
	// Error is the error reported by the harness.
	Error string `json:"error"`
}

func (HarnessError) agentEvent() {}

// Retry reports that the harness will attempt part of a turn again.
type Retry struct {
	// Attempt is the one-based attempt that will run next.
	Attempt int `json:"attempt"`
	// Error is why the preceding attempt did not complete.
	Error string `json:"error"`
}

func (Retry) agentEvent() {}

// ApprovalRequest is a harness operation waiting for approval.
type ApprovalRequest struct {
	// ID identifies this request within its turn.
	ID string `json:"id"`
	// Tool names the operation awaiting approval.
	Tool string `json:"tool"`
	// Input is the approval payload encoded as JSON text.
	Input JSONText `json:"input"`
}

func (ApprovalRequest) agentEvent() {}

// NestedTranscript is a subagent transcript produced inside a tool call.
type NestedTranscript struct {
	// CallID identifies the tool call that started the nested agent.
	CallID string `json:"call_id"`
	// Transcript is the nested harness transcript encoded as JSON text.
	Transcript JSONText `json:"transcript"`
}

func (NestedTranscript) agentEvent() {}

// LifecycleEvent is one typed change to a run's lifecycle. The interface is
// sealed: Gimble produces the concrete variants defined here.
type LifecycleEvent interface{ lifecycleEvent() }

// RunStarted records the beginning of a workflow run.
type RunStarted struct {
	Name string `json:"name"`
}

func (RunStarted) lifecycleEvent() {}

// RunEnded records the result of a workflow run.
type RunEnded struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

func (RunEnded) lifecycleEvent() {}

// RunCancelled records a workflow run stopped by context cancellation.
type RunCancelled struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Error  string `json:"error"`
}

func (RunCancelled) lifecycleEvent() {}

// ScopeBegan records entry into one scope instance.
type ScopeBegan struct {
	Name string `json:"name"`
	Task string `json:"task"`
}

func (ScopeBegan) lifecycleEvent() {}

// ScopeEnded records exit from one scope instance.
type ScopeEnded struct {
	Error string `json:"error"`
}

func (ScopeEnded) lifecycleEvent() {}

// LoopCommand records one command run by a loop.
type LoopCommand struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
}

func (LoopCommand) lifecycleEvent() {}

// PlannerDecision records the task selected for a loop's next lap.
type PlannerDecision struct {
	Decision string `json:"decision"`
}

func (PlannerDecision) lifecycleEvent() {}

// ValueSet records one value written into a scope.
type ValueSet struct {
	Key   string   `json:"key"`
	Value JSONText `json:"value"`
}

func (ValueSet) lifecycleEvent() {}

// SessionCreated records a new agent session or fork.
type SessionCreated struct {
	Name    string `json:"name"`
	Adapter string `json:"adapter"`
	Model   string `json:"model"`
	Workdir string `json:"workdir"`
	Parent  string `json:"parent"`
}

func (SessionCreated) lifecycleEvent() {}

// SessionClosed records that a scope closed one of its sessions.
type SessionClosed struct{}

func (SessionClosed) lifecycleEvent() {}

// TurnStarted records the beginning of one agent turn.
type TurnStarted struct {
	Prompt     string `json:"prompt"`
	OutputType string `json:"output_type"`
}

func (TurnStarted) lifecycleEvent() {}

// TurnEnded records the result and accounting for one agent turn.
type TurnEnded struct {
	Result      JSONText      `json:"result"`
	Error       string        `json:"error"`
	Tokens      []JSONText    `json:"tokens"`
	Duration    time.Duration `json:"duration"`
	Interrupted bool          `json:"interrupted"`
}

func (TurnEnded) lifecycleEvent() {}

// SuperviseAttached records a reviewer attached to a worker turn.
type SuperviseAttached struct {
	Reviewer    string        `json:"reviewer"`
	Worker      string        `json:"worker"`
	Instruction string        `json:"instruction"`
	Interval    time.Duration `json:"interval"`
}

func (SuperviseAttached) lifecycleEvent() {}

// Steer records a message sent to a running session or dropped after it ended.
type Steer struct {
	Target  string `json:"target"`
	Source  string `json:"source"`
	Message string `json:"message"`
	Landed  bool   `json:"landed"`
}

func (Steer) lifecycleEvent() {}

// Interrupt records a request to stop a running session.
type Interrupt struct {
	Target string `json:"target"`
	Source string `json:"source"`
}

func (Interrupt) lifecycleEvent() {}

// Complete marks the durable end of a run log. RecordingError reports an
// earlier failure in another log owned by the run; an absent Complete means
// the run log itself did not finish durably.
type Complete struct {
	RecordingError string `json:"recording_error"`
}

func (Complete) lifecycleEvent() {}

// LifecycleRecord places one lifecycle event in a run or project log.
type LifecycleRecord struct {
	Seq     uint64                    `json:"seq"`
	Time    time.Time                 `json:"time"`
	Scope   string                    `json:"scope"`
	Session polytype.Optional[string] `json:"session,omitzero"`
	Turn    polytype.Optional[string] `json:"turn,omitzero"`
	Event   LifecycleEvent            `json:"event"`
}

// AgentRecord places one harness event in a session transcript.
type AgentRecord struct {
	Seq     uint64     `json:"seq"`
	Time    time.Time  `json:"time"`
	Scope   string     `json:"scope"`
	Session string     `json:"session"`
	Turn    string     `json:"turn"`
	Event   AgentEvent `json:"event"`
}
