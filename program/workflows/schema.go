package workflows

func objectSchema(description string, required []string, properties map[string]any) map[string]any {
	return map[string]any{"$schema": "https://json-schema.org/draft/2020-12/schema", "type": "object", "description": description, "additionalProperties": false, "required": required, "properties": properties}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
