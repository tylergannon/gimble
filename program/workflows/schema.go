package workflows

import (
	"encoding/json"
	"fmt"
)

// Schema returns the narrow input schema for a shipped program. It is kept
// beside the program rather than derived through a general graph schema.
func Schema(name string) (json.RawMessage, error) {
	var schema map[string]any
	switch name {
	case "sprint-execute":
		schema = objectSchema("Sprint ledger execution", []string{"goal", "checklist"}, map[string]any{
			"goal":            stringSchema("The outcome that frames each sprint."),
			"checklist":       stringSchema("Sprint checklist path, relative to workdir."),
			"implement_model": stringSchema("Model for implementation turns."),
			"review_model":    stringSchema("Model for independent review turns."),
			"evaluate_model":  stringSchema("Model for checklist evaluation turns."),
			"max_iterations":  positiveIntSchema("Maximum loop arrivals."),
		})
	case "chapter-loop":
		schema = objectSchema("Chapter ledger with nested sprint ledgers", []string{"goal", "checklist"}, map[string]any{
			"goal":                   stringSchema("The outcome that frames the chapters."),
			"checklist":              stringSchema("Chapter checklist path, relative to workdir."),
			"implement_model":        stringSchema("Model for implementation turns."),
			"review_model":           stringSchema("Model for independent review turns."),
			"evaluate_model":         stringSchema("Model for checklist evaluation turns."),
			"chapter_max_iterations": positiveIntSchema("Maximum chapter loop arrivals."),
			"sprint_max_iterations":  positiveIntSchema("Maximum sprint loop arrivals per chapter."),
		})
	case "delivery-loop":
		schema = objectSchema("Specification planning and delivery", []string{"goal"}, map[string]any{
			"goal":           stringSchema("Specification path or outcome to deliver."),
			"checklist":      stringSchema("Plan checklist output path; defaults to plan.md."),
			"plan_model":     stringSchema("Model for planning."),
			"critique_model": stringSchema("Model for plan critique."),
			"coding_model":   stringSchema("Model for coding turns."),
			"review_model":   stringSchema("Model for independent review turns."),
			"evaluate_model": stringSchema("Model for checklist evaluation turns."),
			"max_iterations": positiveIntSchema("Maximum work loop arrivals."),
		})
	case "sprint-plan":
		return nil, fmt.Errorf("program %q is deferred: its fan-out is not implemented in the Go builtins", name)
	default:
		return nil, fmt.Errorf("unknown program %q", name)
	}
	encoded, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func objectSchema(description string, required []string, properties map[string]any) map[string]any {
	return map[string]any{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "object", "description": description, "additionalProperties": false, "required": required, "properties": properties}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func positiveIntSchema(description string) map[string]any {
	return map[string]any{"type": "integer", "minimum": 1, "description": description}
}
