package engine

import (
	"fmt"
	"sort"

	"github.com/tylergannon/gimble/graph"
	"github.com/tylergannon/gimble/internal/modelalias"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
)

const (
	RoleAgent          = "agent"
	RoleSupervisor     = "supervisor"
	RoleFanIn          = "fan_in"
	RoleBranchTemplate = "branch_agent_template"
	RoleItemJudge      = "item_judge"
	RoleGoalEvaluator  = "goal_evaluator"
)

// ModelResolution is one effective model selection exposed by inspection and
// consumed unchanged by execution.
type ModelResolution struct {
	NodeID          string `json:"node_id"`
	Role            string `json:"role"`
	Source          string `json:"source"`
	AuthoredName    string `json:"authored_name"`
	AuthoredVersion string `json:"authored_version,omitempty"`
	NativeModel     string `json:"native_model"`
	EffectiveEffort string `json:"effective_effort"`
	Provider        string `json:"provider"`
	Harness         string `json:"harness"`
}

// SystemModelSelection is the final whole-object fallback used by a runner.
type SystemModelSelection struct {
	Name   string
	Effort string
}

// ResolveGraphModels resolves every model-capable authored location, including
// hidden loop roles, supervisors, fan-out templates, and synthesized branches.
func ResolveGraphModels(pipeline graph.Graph, system SystemModelSelection) ([]ModelResolution, error) {
	system = normalizeSystemModelSelection(system)

	if pipeline.Defaults.Model.Present {
		if _, err := resolveSelected("", "pipeline_default", "pipeline defaults.model", pipeline.Defaults.Model, jsonschema.Optional[graph.ModelSelection]{}, system, nil); err != nil {
			return nil, err
		}
	}
	var resolutions []ModelResolution
	for _, node := range pipeline.Nodes {
		var resolved ModelResolution
		var err error
		switch current := node.(type) {
		case *graph.AgentNode:
			role := RoleAgent
			source := ""
			if current.IsSynthesized() {
				role = "branch_agent"
				source = current.SynthesizedModelSource()
			}
			resolved, err = resolveSelected(current.ID, role, source, current.Model, pipeline.Defaults.Model, system, nil)
		case *graph.FanInNode:
			resolved, err = resolveSelected(current.ID, RoleFanIn, "", current.Model, pipeline.Defaults.Model, system, nil)
		case *graph.SupervisorNode:
			resolved, err = resolveSelected(current.ID, RoleSupervisor, "", current.Model, pipeline.Defaults.Model, system, nil)
		case *graph.FanOutNode:
			// The fan-out does not execute a turn. Its selection remains inspectable
			// as the template for each synthesized branch agent.
			resolved, err = resolveSelected(current.ID, RoleBranchTemplate, "", current.Model, pipeline.Defaults.Model, system, nil)
		case *graph.LoopNode:
			judge := optionalRoleModel(current.ItemJudge)
			independent := graph.ModelSelection{Name: "flash", Effort: optionalString("medium")}
			resolved, err = resolveSelected(current.ID, RoleItemJudge, "", judge, jsonschema.Optional[graph.ModelSelection]{}, system, &independent)
			if err == nil {
				resolutions = append(resolutions, resolved)
				resolved, err = resolveSelected(current.ID, RoleGoalEvaluator, "", optionalRoleModel(current.GoalEvaluator), pipeline.Defaults.Model, system, nil)
			}
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		resolutions = append(resolutions, resolved)
	}
	sort.SliceStable(resolutions, func(i, j int) bool {
		if resolutions[i].NodeID == resolutions[j].NodeID {
			return resolutions[i].Role < resolutions[j].Role
		}
		return resolutions[i].NodeID < resolutions[j].NodeID
	})
	return resolutions, nil
}

func normalizeSystemModelSelection(system SystemModelSelection) SystemModelSelection {
	if system.Name == "" {
		system.Name = "gpt-5.6-sol"
	}
	if system.Effort == "" {
		system.Effort = "high"
	}
	return system
}

func resolveNodeModel(nodeID, role string, selected, pipelineDefault jsonschema.Optional[graph.ModelSelection], system SystemModelSelection) (ModelResolution, error) {
	return resolveSelected(nodeID, role, "", selected, pipelineDefault, system, nil)
}

func resolveSelected(nodeID, role, selectedSource string, selected, pipelineDefault jsonschema.Optional[graph.ModelSelection], system SystemModelSelection, independent *graph.ModelSelection) (ModelResolution, error) {
	system = normalizeSystemModelSelection(system)
	choice := graph.ModelSelection{}
	source := selectedSource
	switch {
	case selected.Present:
		choice = selected.Value
		if source == "" {
			source = fmt.Sprintf("node %s %s", nodeID, role)
		}
	case independent != nil:
		choice = *independent
		source = "independent item_judge default"
	case pipelineDefault.Present:
		choice = pipelineDefault.Value
		source = "pipeline defaults.model"
	default:
		choice = graph.ModelSelection{Name: system.Name}
		if system.Effort != "" {
			choice.Effort = optionalString(system.Effort)
		}
		source = "system default"
	}
	selection := modelalias.Selection{Name: choice.Name}
	if choice.Version.Present {
		selection.Version = choice.Version.Value
		selection.VersionPresent = true
	}
	if choice.Effort.Present {
		selection.Effort = choice.Effort.Value
		selection.EffortPresent = true
	}
	effective, err := modelalias.ResolveModel(selection)
	if err != nil {
		return ModelResolution{}, fmt.Errorf("model selection at %s: %w", source, err)
	}
	return ModelResolution{
		NodeID: nodeID, Role: role, Source: source,
		AuthoredName: choice.Name, AuthoredVersion: selection.Version,
		NativeModel: effective.Model, EffectiveEffort: effective.Effort,
		Provider: effective.Provider, Harness: effective.Harness,
	}, nil
}

func optionalRoleModel(role jsonschema.Optional[graph.LoopRole]) jsonschema.Optional[graph.ModelSelection] {
	if role.Present {
		return role.Value.Model
	}
	return jsonschema.Optional[graph.ModelSelection]{}
}

func optionalString(value string) jsonschema.Optional[string] {
	return jsonschema.Optional[string]{Present: true, Value: value}
}
