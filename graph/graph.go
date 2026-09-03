// Package graph defines and parses Tractor pipeline documents.
package graph

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

//go:generate go tool gen-jsonschema gen --pretty --validate --formats=both
//go:generate go run ./internal/schemafix

// Graph is a complete pipeline definition.
type Graph struct {
	// Name is the pipeline's display name.
	Name string `json:"name,omitzero"`

	// Goal is the pipeline objective exposed to prompt expansion.
	Goal string `json:"goal,omitzero"`

	// Defaults contains file-level defaults for node fields.
	Defaults Defaults `json:"defaults,omitzero"`

	// Start names the walk node where execution begins.
	Start string `json:"start"`

	// Nodes is the graph. Each node carries its outgoing edges.
	Nodes []Node `json:"nodes"`
}

// Defaults contains the six fields that may be inherited by nodes.
type Defaults struct {
	MaxRetries      jsonschema.Optional[int]      `json:"max_retries,omitzero"`
	Fidelity        jsonschema.Optional[string]   `json:"fidelity,omitzero"`
	Timeout         jsonschema.Optional[Duration] `json:"timeout,omitzero"`
	LLMModel        jsonschema.Optional[string]   `json:"llm_model,omitzero"`
	LLMProvider     jsonschema.Optional[string]   `json:"llm_provider,omitzero"`
	ReasoningEffort jsonschema.Optional[string]   `json:"reasoning_effort,omitzero"`
}

// Duration is an integer followed by ms, s, m, h, or d.
type Duration string

// Parse converts d to a time.Duration. Days are fixed 24-hour periods.
func (d Duration) Parse() (time.Duration, error) {
	s := string(d)
	var number string
	switch {
	case strings.HasSuffix(s, "ms"):
		number = strings.TrimSuffix(s, "ms")
	case len(s) > 0 && strings.ContainsRune("smhd", rune(s[len(s)-1])):
		number = s[:len(s)-1]
	default:
		return 0, fmt.Errorf("parse duration %q: expected integer followed by ms, s, m, h, or d", s)
	}
	if number == "" {
		return 0, fmt.Errorf("parse duration %q: expected integer followed by ms, s, m, h, or d", s)
	}
	for _, digit := range number {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("parse duration %q: expected integer followed by ms, s, m, h, or d", s)
		}
	}
	if strings.HasSuffix(s, "d") {
		days, err := strconv.ParseInt(number, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse duration %q: %w", s, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", s, err)
	}
	return parsed, nil
}

// Edge is an outgoing transition owned by its origin node.
type Edge struct {
	To        string `json:"to"`
	Condition string `json:"condition,omitzero"`
}

// Node is a pipeline stage.
type Node interface {
	isNode()
	Base() *NodeBase
	NodeType() string
}

// NodeBase contains fields admitted on every node type.
type NodeBase struct {
	ID    string                      `json:"id"`
	Label jsonschema.Optional[string] `json:"label,omitzero"`
}

// DisplayLabel returns the configured label or the node ID.
func (n *NodeBase) DisplayLabel() string {
	if n.Label.Present {
		return n.Label.Value
	}
	return n.ID
}

// AgentNode runs an LLM task.
type AgentNode struct {
	NodeBase
	Edges     []Edge                   `json:"edges,omitzero"`
	MaxVisits jsonschema.Optional[int] `json:"max_visits,omitzero"`
	LLMNodeFields
	synthesized bool
}

func (*AgentNode) isNode()           {}
func (n *AgentNode) Base() *NodeBase { return &n.NodeBase }
func (*AgentNode) NodeType() string  { return "agent" }

// IsSynthesized reports whether the node was resolved from a structured
// parallel branch.
func (n *AgentNode) IsSynthesized() bool { return n.synthesized }

// FanInNode evaluates parallel branch evidence with an LLM turn.
type FanInNode struct {
	NodeBase
	Edges     []Edge                   `json:"edges,omitzero"`
	MaxVisits jsonschema.Optional[int] `json:"max_visits,omitzero"`
	LLMNodeFields
}

func (*FanInNode) isNode()           {}
func (n *FanInNode) Base() *NodeBase { return &n.NodeBase }
func (*FanInNode) NodeType() string  { return "fan_in" }

// LLMNodeFields are shared by agent and fan-in nodes.
type LLMNodeFields struct {
	Prompt          jsonschema.Optional[string]   `json:"prompt,omitzero"`
	MaxRetries      jsonschema.Optional[int]      `json:"max_retries,omitzero"`
	Fidelity        jsonschema.Optional[string]   `json:"fidelity,omitzero"`
	ThreadID        jsonschema.Optional[string]   `json:"thread_id,omitzero"`
	Timeout         jsonschema.Optional[Duration] `json:"timeout,omitzero"`
	LLMModel        jsonschema.Optional[string]   `json:"llm_model,omitzero"`
	LLMProvider     jsonschema.Optional[string]   `json:"llm_provider,omitzero"`
	ReasoningEffort jsonschema.Optional[string]   `json:"reasoning_effort,omitzero"`
}

// AgentOverride selectively replaces fields inherited from a fan-out node's
// agent configuration. Every field is optional by design.
type AgentOverride struct {
	Label           jsonschema.Optional[string]   `json:"label,omitzero"`
	Prompt          jsonschema.Optional[string]   `json:"prompt,omitzero"`
	MaxRetries      jsonschema.Optional[int]      `json:"max_retries,omitzero"`
	MaxVisits       jsonschema.Optional[int]      `json:"max_visits,omitzero"`
	Fidelity        jsonschema.Optional[string]   `json:"fidelity,omitzero"`
	ThreadID        jsonschema.Optional[string]   `json:"thread_id,omitzero"`
	Timeout         jsonschema.Optional[Duration] `json:"timeout,omitzero"`
	LLMModel        jsonschema.Optional[string]   `json:"llm_model,omitzero"`
	LLMProvider     jsonschema.Optional[string]   `json:"llm_provider,omitzero"`
	ReasoningEffort jsonschema.Optional[string]   `json:"reasoning_effort,omitzero"`
}

// PromptValue returns a non-empty prompt or falls back to label.
func (n *LLMNodeFields) PromptValue(label string) string {
	if n.Prompt.Present && n.Prompt.Value != "" {
		return n.Prompt.Value
	}
	return label
}

// CommandEdges are the engine-known routes out of a command node.
type CommandEdges struct {
	Success string                      `json:"success"`
	Error   jsonschema.Optional[string] `json:"error,omitzero"`
}

// CommandNode executes one shell command.
type CommandNode struct {
	NodeBase
	Command   string                        `json:"command"`
	Edges     CommandEdges                  `json:"edges"`
	Timeout   jsonschema.Optional[Duration] `json:"timeout,omitzero"`
	MaxVisits jsonschema.Optional[int]      `json:"max_visits,omitzero"`
}

func (*CommandNode) isNode()           {}
func (n *CommandNode) Base() *NodeBase { return &n.NodeBase }
func (*CommandNode) NodeType() string  { return "command" }

// WorkspacePolicy controls whether parallel branches receive separate Git
// worktrees or run together in the caller's workspace.
type WorkspacePolicy string

const (
	WorkspaceIsolated WorkspacePolicy = "isolated"
	WorkspaceShared   WorkspacePolicy = "shared"
)

// FanOutBranch is either a legacy branch-root reference or a synthesized
// agent branch with declared output artifacts.
type FanOutBranch struct {
	ID        string                             `json:"id"`
	Artifacts []string                           `json:"artifacts"`
	Agent     jsonschema.Optional[AgentOverride] `json:"agent,omitzero"`
	legacy    bool
}

// LegacyFanOutBranch constructs an existing-style branch-root reference.
func LegacyFanOutBranch(id string) FanOutBranch {
	return FanOutBranch{ID: id, legacy: true}
}

// LegacyFanOutBranches constructs existing-style branch-root references.
func LegacyFanOutBranches(ids ...string) []FanOutBranch {
	branches := make([]FanOutBranch, len(ids))
	for index, id := range ids {
		branches[index] = LegacyFanOutBranch(id)
	}
	return branches
}

// IsLegacy reports whether the branch was authored as a string reference.
func (b FanOutBranch) IsLegacy() bool { return b.legacy }

// FanOutNode concurrently walks each outgoing branch.
type FanOutNode struct {
	NodeBase
	Branches    []FanOutBranch                       `json:"branches"`
	BranchEdges []Edge                               `json:"branch_edges,omitzero"`
	MaxParallel jsonschema.Optional[int]             `json:"max_parallel,omitzero"`
	MaxVisits   jsonschema.Optional[int]             `json:"max_visits,omitzero"`
	Workspace   jsonschema.Optional[WorkspacePolicy] `json:"workspace,omitzero"`
	LLMNodeFields
}

func (*FanOutNode) isNode()           {}
func (n *FanOutNode) Base() *NodeBase { return &n.NodeBase }
func (*FanOutNode) NodeType() string  { return "fan_out" }

// SupervisorNode observes declared nodes and coaches them outside the walk.
type SupervisorNode struct {
	NodeBase
	Prompt          string                        `json:"prompt"`
	Supervises      []string                      `json:"supervises"`
	Interval        jsonschema.Optional[Duration] `json:"interval,omitzero"`
	Timeout         jsonschema.Optional[Duration] `json:"timeout,omitzero"`
	LLMModel        jsonschema.Optional[string]   `json:"llm_model,omitzero"`
	LLMProvider     jsonschema.Optional[string]   `json:"llm_provider,omitzero"`
	ReasoningEffort jsonschema.Optional[string]   `json:"reasoning_effort,omitzero"`
}

func (*SupervisorNode) isNode()           {}
func (n *SupervisorNode) Base() *NodeBase { return &n.NodeBase }
func (*SupervisorNode) NodeType() string  { return "supervisor" }

// LoopNode iterates a checklist file: on every arrival it validates the
// previous lap's item, then an evaluator decides whether the definition of
// done is met or another open item should be dispatched.
type LoopNode struct {
	NodeBase
	// Checklist is the checklist path, relative to the workdir. Optional only
	// when this loop node lies inside another loop's body, in which case it
	// iterates the enclosing item's checklist field.
	Checklist jsonschema.Optional[string] `json:"checklist,omitzero"`
	// Edges are the engine-known routes for another lap and for leaving the loop.
	Edges LoopEdges `json:"edges"`
	// MaxVisits bounds arrivals at the loop node, laps plus one.
	MaxVisits jsonschema.Optional[int] `json:"max_visits,omitzero"`
	// Timeout bounds one item's validation command, the infer judge turn, and
	// the evaluator turn.
	Timeout jsonschema.Optional[Duration] `json:"timeout,omitzero"`
	// LLMModel selects the model used by the infer judge.
	LLMModel jsonschema.Optional[string] `json:"llm_model,omitzero"`
	// LLMProvider selects the provider used by the infer judge.
	LLMProvider jsonschema.Optional[string] `json:"llm_provider,omitzero"`
	// ReasoningEffort sets the reasoning effort of the infer judge.
	ReasoningEffort jsonschema.Optional[string] `json:"reasoning_effort,omitzero"`
	// EvaluatorLLMModel selects the model used by the loop evaluator.
	EvaluatorLLMModel jsonschema.Optional[string] `json:"evaluator_llm_model,omitzero"`
	// EvaluatorLLMProvider selects the provider used by the loop evaluator.
	EvaluatorLLMProvider jsonschema.Optional[string] `json:"evaluator_llm_provider,omitzero"`
	// EvaluatorReasoningEffort sets the loop evaluator's reasoning effort.
	EvaluatorReasoningEffort jsonschema.Optional[string] `json:"evaluator_reasoning_effort,omitzero"`
}

// LoopEdges are the engine-known routes out of a loop node.
type LoopEdges struct {
	Loop string `json:"loop"`
	Exit string `json:"exit"`
}

func (*LoopNode) isNode()           {}
func (n *LoopNode) Base() *NodeBase { return &n.NodeBase }
func (*LoopNode) NodeType() string  { return "loop" }

// MaxParallelValue returns the explicit maximum or the system default.
func (n *FanOutNode) MaxParallelValue() int {
	if n.MaxParallel.Present {
		return n.MaxParallel.Value
	}
	return 4
}

// WorkspacePolicyValue returns the explicit workspace policy or the
// compatibility-preserving isolated default.
func (n *FanOutNode) WorkspacePolicyValue() WorkspacePolicy {
	if n.Workspace.Present {
		return n.Workspace.Value
	}
	return WorkspaceIsolated
}

// BranchIDs returns branch roots in authored order.
func (n *FanOutNode) BranchIDs() []string {
	ids := make([]string, len(n.Branches))
	for index, branch := range n.Branches {
		ids[index] = branch.ID
	}
	return ids
}

// Branch returns branch metadata by ID.
func (n *FanOutNode) Branch(id string) (FanOutBranch, bool) {
	for _, branch := range n.Branches {
		if branch.ID == id {
			return branch, true
		}
	}
	return FanOutBranch{}, false
}

// IntervalValue returns the explicit patrol interval or the system default.
func (n *SupervisorNode) IntervalValue() Duration {
	if n.Interval.Present {
		return n.Interval.Value
	}
	return "60s"
}

const (
	// Success is the terminal pseudo-target that completes a run.
	Success = "success"
	// Failure is the terminal pseudo-target that deliberately fails a run.
	Failure = "failure"
)

// IsPseudoTarget reports whether target names a terminal pseudo-target.
func IsPseudoTarget(target string) bool { return target == Success || target == Failure }

// RoutingTargets returns the authored routing targets for a walk node.
func RoutingTargets(node Node) []string {
	switch node := node.(type) {
	case *AgentNode:
		return edgeTargets(node.Edges)
	case *FanInNode:
		return edgeTargets(node.Edges)
	case *CommandNode:
		targets := []string{node.Edges.Success}
		if node.Edges.Error.Present {
			targets = append(targets, node.Edges.Error.Value)
		}
		return targets
	case *FanOutNode:
		return node.BranchIDs()
	case *LoopNode:
		return []string{node.Edges.Loop, node.Edges.Exit}
	default:
		return nil
	}
}

// ChoiceEdges returns the authored choice edges for a chooser node.
func ChoiceEdges(node Node) []Edge {
	switch node := node.(type) {
	case *AgentNode:
		return node.Edges
	case *FanInNode:
		return node.Edges
	default:
		return nil
	}
}

// MaxVisits returns a node's optional visit budget.
func MaxVisits(node Node) jsonschema.Optional[int] {
	switch node := node.(type) {
	case *AgentNode:
		return node.MaxVisits
	case *FanInNode:
		return node.MaxVisits
	case *CommandNode:
		return node.MaxVisits
	case *FanOutNode:
		return node.MaxVisits
	case *LoopNode:
		return node.MaxVisits
	default:
		return jsonschema.Optional[int]{}
	}
}

func edgeTargets(edges []Edge) []string {
	targets := make([]string, len(edges))
	for i, edge := range edges {
		targets[i] = edge.To
	}
	return targets
}

// ThreadKey returns the explicit thread ID or the node ID.
func (n *LLMNodeFields) ThreadKey(nodeID string) string {
	if n.ThreadID.Present {
		return n.ThreadID.Value
	}
	return nodeID
}

// FidelityValue returns the resolved fidelity or its system default.
func (n *LLMNodeFields) FidelityValue() string {
	if n.Fidelity.Present {
		return n.Fidelity.Value
	}
	return "compacted"
}
