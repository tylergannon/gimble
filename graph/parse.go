package graph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"go.yaml.in/yaml/v4"
)

// Parse validates and decodes one pipeline document, then applies file-level
// defaults to the node types that admit each field.
func Parse(data []byte) (*Graph, error) {
	if err := preflight(data); err != nil {
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}
	var authored map[string]any
	if err := json.Unmarshal(data, &authored); err == nil {
		if err := rejectObsoleteModelKeys(authored); err != nil {
			return nil, fmt.Errorf("parse pipeline: %w", err)
		}
	}
	if err := (Graph{}).ValidateJSON(data); err != nil {
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}

	var graph Graph
	if err := json.Unmarshal(data, &graph); err != nil {
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}
	if err := finishGraph(&graph); err != nil {
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}
	return &graph, nil
}

func finishGraph(graph *Graph) error {
	if err := graph.expandFanOutBranches(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(graph.Nodes))
	for _, node := range graph.Nodes {
		id := node.Base().ID
		if IsPseudoTarget(id) {
			return fmt.Errorf("node ID %q is reserved for a terminal pseudo-target", id)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate node ID %q", id)
		}
		seen[id] = struct{}{}
	}
	graph.applyDefaults()
	return nil
}

// UnmarshalJSON accepts both string branch roots and structured agent branches
// without weakening the generated object schema.
func (n *FanOutNode) UnmarshalJSON(data []byte) error {
	type alias FanOutNode
	var payload struct {
		*alias
		Branches []json.RawMessage `json:"branches"`
	}
	payload.alias = (*alias)(n)
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	branches := make([]FanOutBranch, len(payload.Branches))
	for index, raw := range payload.Branches {
		var id string
		if err := json.Unmarshal(raw, &id); err == nil {
			branches[index] = LegacyFanOutBranch(id)
			continue
		}
		if err := json.Unmarshal(raw, &branches[index]); err != nil {
			return fmt.Errorf("field branches[%d]: %w", index, err)
		}
	}
	n.Branches = branches
	return nil
}

func (g *Graph) expandFanOutBranches() error {
	authored := make(map[string]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		authored[node.Base().ID] = struct{}{}
	}
	var synthesized []Node
	for _, node := range g.Nodes {
		fanOut, ok := node.(*FanOutNode)
		if !ok || len(fanOut.Branches) == 0 {
			continue
		}
		legacy := fanOut.Branches[0].IsLegacy()
		for _, branch := range fanOut.Branches {
			if branch.IsLegacy() != legacy {
				return fmt.Errorf("fan_out node %q cannot mix string and object branches", fanOut.ID)
			}
		}
		if legacy {
			continue
		}
		if len(fanOut.BranchEdges) == 0 {
			return fmt.Errorf("fan_out node %q with agent branches must declare branch_edges", fanOut.ID)
		}
		for _, branch := range fanOut.Branches {
			if _, exists := authored[branch.ID]; exists {
				return fmt.Errorf("fan_out node %q branch ID %q collides with a declared node", fanOut.ID, branch.ID)
			}
			if err := validateArtifactPaths(fanOut.ID, branch); err != nil {
				return err
			}
			resolved := resolveFanOutAgent(fanOut, branch)
			synthesized = append(synthesized, resolved)
			authored[branch.ID] = struct{}{}
		}
	}
	g.Nodes = append(g.Nodes, synthesized...)
	return nil
}

func resolveFanOutAgent(parent *FanOutNode, branch FanOutBranch) *AgentNode {
	resolved := &AgentNode{
		ID: branch.ID, Label: parent.Label,
		Edges:         append([]Edge(nil), parent.BranchEdges...),
		LLMNodeFields: parent.LLMNodeFields,
		synthesized:   true,
	}
	if parent.Model.Present {
		resolved.synthesizedModelSource = fmt.Sprintf("node %s fan_out template", parent.ID)
	}
	if !branch.Agent.Present {
		return resolved
	}
	override := branch.Agent.Value
	overrideOptional(&resolved.Label, override.Label)
	overrideOptional(&resolved.Prompt, override.Prompt)
	overrideOptional(&resolved.MaxRetries, override.MaxRetries)
	overrideOptional(&resolved.MaxVisits, override.MaxVisits)
	overrideOptional(&resolved.Fidelity, override.Fidelity)
	overrideOptional(&resolved.ThreadID, override.ThreadID)
	overrideOptional(&resolved.Timeout, override.Timeout)
	overrideOptional(&resolved.Model, override.Model)
	if override.Model.Present {
		resolved.synthesizedModelSource = fmt.Sprintf("node %s branch %s", parent.ID, branch.ID)
	}
	return resolved
}

func overrideOptional[T any](destination *jsonschema.Optional[T], source jsonschema.Optional[T]) {
	if source.Present {
		*destination = source
	}
}

func validateArtifactPaths(fanOutID string, branch FanOutBranch) error {
	seen := make(map[string]struct{}, len(branch.Artifacts))
	for _, artifact := range branch.Artifacts {
		cleaned := filepath.Clean(artifact)
		if strings.TrimSpace(artifact) == "" || filepath.IsAbs(artifact) || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			return fmt.Errorf("fan_out node %q branch %q has invalid artifact path %q", fanOutID, branch.ID, artifact)
		}
		if _, duplicate := seen[cleaned]; duplicate {
			return fmt.Errorf("fan_out node %q branch %q repeats artifact path %q", fanOutID, branch.ID, artifact)
		}
		seen[cleaned] = struct{}{}
	}
	return nil
}

// ParseYAML validates and decodes one YAML pipeline using the generated
// JSON Schema contract.
func ParseYAML(data []byte) (*Graph, error) {
	var authored map[string]any
	if err := yaml.Load(data, &authored, yaml.WithV4Defaults()); err == nil {
		if err := rejectObsoleteModelKeys(authored); err != nil {
			return nil, fmt.Errorf("parse pipeline YAML: %w", err)
		}
	}
	if err := (Graph{}).ValidateYAML(data); err != nil {
		return nil, fmt.Errorf("parse pipeline YAML: %w", err)
	}

	var graph Graph
	if err := yaml.Load(data, &graph, yaml.WithV4Defaults()); err != nil {
		return nil, fmt.Errorf("parse pipeline YAML: %w", err)
	}
	if err := finishGraph(&graph); err != nil {
		return nil, fmt.Errorf("parse pipeline YAML: %w", err)
	}
	return &graph, nil
}

func rejectObsoleteModelKeys(document map[string]any) error {
	if defaults, ok := document["defaults"].(map[string]any); ok {
		if err := obsoleteAt(defaults, "defaults", false); err != nil {
			return err
		}
	}
	nodes, _ := document["nodes"].([]any)
	for index, raw := range nodes {
		node, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		typeName, _ := node["type"].(string)
		location := fmt.Sprintf("nodes[%d]", index)
		if id, ok := node["id"].(string); ok && id != "" {
			location = fmt.Sprintf("node %q", id)
		}
		if err := obsoleteAt(node, location, typeName == "loop"); err != nil {
			return err
		}
		if typeName != "fan_out" {
			continue
		}
		branches, _ := node["branches"].([]any)
		for branchIndex, branchRaw := range branches {
			branch, _ := branchRaw.(map[string]any)
			agent, _ := branch["agent"].(map[string]any)
			if agent == nil {
				continue
			}
			branchLocation := fmt.Sprintf("%s branch[%d].agent", location, branchIndex)
			if id, ok := branch["id"].(string); ok && id != "" {
				branchLocation = fmt.Sprintf("%s branch %q agent", location, id)
			}
			if err := obsoleteAt(agent, branchLocation, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func obsoleteAt(object map[string]any, location string, loop bool) error {
	type replacement struct {
		old string
		new string
	}
	replacements := []replacement{{"llm_model", "model.name"}, {"reasoning_effort", "model.effort"}}
	if loop {
		replacements = []replacement{
			{"llm_model", "item_judge.model.name"},
			{"reasoning_effort", "item_judge.model.effort"},
			{"evaluator_llm_model", "goal_evaluator.model.name"},
			{"evaluator_reasoning_effort", "goal_evaluator.model.effort"},
		}
	}
	for _, replacement := range replacements {
		if _, exists := object[replacement.old]; exists {
			return fmt.Errorf("%s uses obsolete %q; replace it with %q", location, replacement.old, replacement.new)
		}
	}
	providerKeys := []string{"llm_provider"}
	if loop {
		providerKeys = append(providerKeys, "evaluator_llm_provider")
	}
	for _, key := range providerKeys {
		if _, exists := object[key]; exists {
			return fmt.Errorf("%s uses obsolete %q; authored provider was removed because model.name determines provider", location, key)
		}
	}
	return nil
}

// NodeByID returns the node with id, if present.
func (g *Graph) NodeByID(id string) (Node, bool) {
	for _, node := range g.Nodes {
		if node.Base().ID == id {
			return node, true
		}
	}
	return nil, false
}

func (g *Graph) applyDefaults() {
	for _, node := range g.Nodes {
		switch current := node.(type) {
		case *AgentNode:
			inheritLLM(&current.LLMNodeFields, g.Defaults)
		case *FanInNode:
			inheritLLM(&current.LLMNodeFields, g.Defaults)
		case *CommandNode:
			inherit(&current.Timeout, g.Defaults.Timeout)
		case *SupervisorNode:
			inherit(&current.Timeout, g.Defaults.Timeout)
		case *LoopNode:
			inherit(&current.Timeout, g.Defaults.Timeout)
		}
	}
}

func inheritLLM(fields *LLMNodeFields, defaults Defaults) {
	inherit(&fields.MaxRetries, defaults.MaxRetries)
	inherit(&fields.Fidelity, defaults.Fidelity)
	inherit(&fields.Timeout, defaults.Timeout)
}

func inherit[T any](destination *jsonschema.Optional[T], source jsonschema.Optional[T]) {
	if !destination.Present && source.Present {
		*destination = source
	}
}

func preflight(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("more than one JSON value")
		}
		return err
	}
	return nil
}

func scanValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("null is not allowed")
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		members := map[string]struct{}{}
		for decoder.More() {
			nameToken, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := nameToken.(string)
			if !ok {
				return errors.New("object member name is not a string")
			}
			if _, duplicate := members[name]; duplicate {
				return fmt.Errorf("duplicate object member %q", name)
			}
			members[name] = struct{}{}
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	_, err = decoder.Token()
	return err
}
