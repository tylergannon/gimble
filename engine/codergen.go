package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

// AgentConfig supplies the backend and implementation-level model defaults.
type AgentConfig struct {
	Backend                harness.AgentBackend
	DefaultModel           string
	DefaultProvider        string
	DefaultReasoningEffort string
}

// AgentHandler executes agent nodes through an AgentBackend.
type AgentHandler struct {
	config AgentConfig
}

// NewAgentHandler constructs an agent handler. A nil backend enables simulation.
func NewAgentHandler(config AgentConfig) *AgentHandler {
	return &AgentHandler{config: config}
}

// Execute renders and executes one agent turn.
func (h *AgentHandler) Execute(node graph.Node, offered []graph.Edge, scope ExecutionScope, pipeline *graph.Graph) (harness.Outcome, *harness.Error) {
	agent, ok := node.(*graph.AgentNode)
	if !ok {
		return harness.Outcome{}, terminalError(fmt.Sprintf("agent handler cannot execute node type %s", node.NodeType()))
	}

	prompt := agent.PromptValue(agent.DisplayLabel())
	prompt = expandPrompt(prompt, scope.Goal)
	prompt = prependFrame(scope.Frame, prompt)
	role := RoleAgent
	if agent.IsSynthesized() {
		role = "branch_agent"
	}
	return h.executeTurn(role, agent, &agent.LLMNodeFields, offered, scope, pipeline, prompt)
}

// prependFrame places the rendered loop frame stack ahead of a node's own
// prompt, separated by a blank line. Empty frames leave the prompt alone.
func prependFrame(frame, prompt string) string {
	if frame == "" {
		return prompt
	}
	return frame + "\n\n" + prompt
}

func (h *AgentHandler) executeTurn(role string, node graph.Node, fields *graph.LLMNodeFields, offered []graph.Edge, scope ExecutionScope, pipeline *graph.Graph, prompt string) (harness.Outcome, *harness.Error) {
	return h.executeTurnAt(role, node, fields, offered, scope, pipeline, prompt,
		filepath.Join(scope.StageDir, "prompt.md"), filepath.Join(scope.StageDir, "response.md"))
}

// executeTurnAt runs a turn whose prompt and response artifacts have
// caller-selected paths. Composite handlers use this to keep multiple turns
// within one stage from overwriting one another.
func (h *AgentHandler) executeTurnAt(role string, node graph.Node, fields *graph.LLMNodeFields, offered []graph.Edge, scope ExecutionScope, pipeline *graph.Graph, prompt, promptPath, responsePath string) (harness.Outcome, *harness.Error) {
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return harness.Outcome{}, terminalError(fmt.Sprintf("write prompt: %v", err))
	}

	turn, err := h.turn(role, node, fields, offered, scope, pipeline, prompt)
	if err != nil {
		return harness.Outcome{}, err
	}
	validationTurn := turn
	if h.config.Backend == nil && validationTurn.RunLog == "" {
		validationTurn.RunLog = filepath.Join(scope.StageDir, "simulation.jsonl")
	}
	if validationErr := harness.ValidateAgentTurn(validationTurn); validationErr != nil {
		return harness.Outcome{}, validationErr
	}
	if agent, ok := node.(*graph.AgentNode); ok && agent.IsSynthesized() {
		if err := writeJSON(filepath.Join(scope.StageDir, "resolved.json"), resolvedAgentRecord{
			Type:            "agent",
			ID:              node.Base().ID,
			Prompt:          prompt,
			Provider:        validationTurn.Provider,
			Model:           validationTurn.Model,
			ReasoningEffort: validationTurn.ReasoningEffort,
			Fidelity:        validationTurn.Fidelity,
			ThreadID:        validationTurn.ThreadKey,
			Timeout:         validationTurn.Timeout.String(),
			Workdir:         validationTurn.Workdir,
			RunLog:          validationTurn.RunLog,
		}); err != nil {
			return harness.Outcome{}, terminalError(fmt.Sprintf("write resolved agent configuration: %v", err))
		}
	}
	var outcome harness.Outcome
	if h.config.Backend == nil {
		outcome = harness.Outcome{Notes: "[Simulated] Stage completed: " + node.Base().ID}
		if len(offered) > 1 {
			outcome.Next = offered[0].To
		}
	} else {
		var runErr *harness.Error
		outcome, runErr = h.config.Backend.Run(turn)
		if runErr != nil {
			return harness.Outcome{}, runErr
		}
	}

	if err := writeResponse(responsePath, outcome); err != nil {
		return harness.Outcome{}, terminalError(fmt.Sprintf("write response: %v", err))
	}
	return outcome, nil
}

type resolvedAgentRecord struct {
	Type            string               `json:"type"`
	ID              string               `json:"id"`
	Prompt          string               `json:"prompt"`
	Provider        string               `json:"llm_provider"`
	Model           string               `json:"llm_model"`
	ReasoningEffort string               `json:"reasoning_effort"`
	Fidelity        harness.FidelityMode `json:"fidelity"`
	ThreadID        string               `json:"thread_id,omitempty"`
	Timeout         string               `json:"timeout"`
	Workdir         string               `json:"workdir"`
	RunLog          string               `json:"run_log"`
}

func (h *AgentHandler) turn(role string, node graph.Node, fields *graph.LLMNodeFields, offered []graph.Edge, scope ExecutionScope, pipeline *graph.Graph, prompt string) (harness.AgentTurn, *harness.Error) {
	resolution, selectionErr := resolveNodeModel(node.Base().ID, role, fields.Model, pipeline.Defaults.Model, SystemModelSelection{Name: h.config.DefaultModel, Effort: h.config.DefaultReasoningEffort})
	if selectionErr != nil {
		return harness.AgentTurn{}, terminalError(selectionErr.Error())
	}
	fidelity := resolveString(fields.Fidelity, pipeline.Defaults.Fidelity, string(harness.FidelityCompacted))
	threadKey := ""
	if fidelity != string(harness.FidelityNone) {
		threadKey = fields.ThreadKey(node.Base().ID)
	}
	timeout, timeoutErr := resolveTimeout(fields.Timeout, pipeline.Defaults.Timeout)
	if timeoutErr != nil {
		return harness.AgentTurn{}, terminalError(timeoutErr.Error())
	}
	schema, schemaErr := choiceSchema(offered, pipeline)
	if schemaErr != nil {
		return harness.AgentTurn{}, terminalError(schemaErr.Error())
	}
	return harness.AgentTurn{
		NodeID:          node.Base().ID,
		Role:            role,
		Parts:           []harness.ContentPart{{Type: harness.ContentPartText, Text: prompt}},
		OutputSchema:    schema,
		Model:           resolution.NativeModel,
		Provider:        resolution.Provider,
		ReasoningEffort: resolution.EffectiveEffort,
		Fidelity:        harness.FidelityMode(fidelity),
		ThreadKey:       threadKey,
		Workdir:         scope.Workdir,
		RunLog:          scope.RunLog,
		Timeout:         timeout,
	}, nil
}

func resolveString(nodeValue, fileValue jsonschema.Optional[string], systemValue string) string {
	if nodeValue.Present {
		return nodeValue.Value
	}
	if fileValue.Present {
		return fileValue.Value
	}
	return systemValue
}

func resolveTimeout(nodeValue, fileValue jsonschema.Optional[graph.Duration]) (time.Duration, error) {
	if nodeValue.Present {
		return nodeValue.Value.Parse()
	}
	if fileValue.Present {
		return fileValue.Value.Parse()
	}
	return 0, nil
}

// DetectProvider returns the provider implied by a known provider-native model name.
func DetectProvider(model string) string {
	lower := strings.ToLower(model)
	switch {
	case strings.HasPrefix(lower, "claude"):
		return "anthropic"
	case strings.HasPrefix(lower, "gpt-"), strings.HasPrefix(lower, "o1"), strings.HasPrefix(lower, "o3"), strings.HasPrefix(lower, "o4"), strings.Contains(lower, "codex"):
		return "openai"
	case strings.HasPrefix(lower, "gemini"):
		return "gemini"
	default:
		return ""
	}
}

type outputSchema struct {
	Type                 string           `json:"type"`
	Properties           schemaProperties `json:"properties"`
	Required             []string         `json:"required"`
	AdditionalProperties bool             `json:"additionalProperties"`
}

type schemaProperties struct {
	Next  *schemaProperty `json:"next,omitempty"`
	Notes schemaProperty  `json:"notes"`
}

type schemaProperty struct {
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description"`
}

func choiceSchema(offered []graph.Edge, pipeline *graph.Graph) (json.RawMessage, error) {
	properties := schemaProperties{
		Notes: schemaProperty{Type: "string", Description: "Your account of this stage."},
	}
	required := []string{"notes"}
	if len(offered) > 1 {
		targets := make([]string, len(offered))
		for index, edge := range offered {
			targets[index] = edge.To
		}
		properties.Next = &schemaProperty{
			Type:        "string",
			Enum:        targets,
			Description: describeRoutes(offered, pipeline),
		}
		required = []string{"next", "notes"}
	}
	encoded, err := json.Marshal(outputSchema{
		Type:                 "object",
		Properties:           properties,
		Required:             required,
		AdditionalProperties: false,
	})
	if err != nil {
		return nil, fmt.Errorf("encode choice schema: %w", err)
	}
	return encoded, nil
}

func describeRoutes(offered []graph.Edge, pipeline *graph.Graph) string {
	clauses := make([]string, len(offered))
	for index, edge := range offered {
		description := strings.TrimSpace(edge.Condition)
		if description == "" {
			if target, ok := pipeline.NodeByID(edge.To); ok {
				description = strings.TrimSpace(target.Base().DisplayLabel())
			}
			if description == "" {
				description = edge.To
			}
		}
		clauses[index] = description + ": " + edge.To
	}
	return "Choose the next stage. " + strings.Join(clauses, "; ")
}

func writeResponse(path string, outcome harness.Outcome) error {
	var response strings.Builder
	response.WriteString("---\n")
	if outcome.Next != "" {
		response.WriteString("next: ")
		response.WriteString(outcome.Next)
		response.WriteByte('\n')
	}
	response.WriteString("---\n")
	response.WriteString(outcome.Notes)
	return os.WriteFile(path, []byte(response.String()), 0o644)
}
