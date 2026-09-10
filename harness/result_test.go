package harness

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestValidateResultUsesExactRuntimeSchema(t *testing.T) {
	property := "decision_" + t.Name()
	schema := json.RawMessage(fmt.Sprintf(`{
		"type":"object",
		"properties":{
			%q:{"type":"string","enum":["continue","stop"]},
			"details":{"$ref":"#/$defs/details"}
		},
		"required":[%q,"details"],
		"additionalProperties":false,
		"$defs":{"details":{
			"type":"object",
			"properties":{"order":{"type":"array","prefixItems":[{"const":"first"},{"const":"second"}],"items":false}},
			"required":["order"],
			"additionalProperties":false
		}}
	}`, property, property))
	result := []byte(fmt.Sprintf(`{%q:"continue","details":{"order":["first","second"]}}`, property))

	got, err := ValidateResult(schema, result)
	if err != nil {
		t.Fatalf("ValidateResult() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded[property] != "continue" {
		t.Fatalf("ValidateResult()[%q] = %v, want continue", property, decoded[property])
	}

	withExtraProperty := []byte(fmt.Sprintf(`{%q:"continue","details":{"order":["first","second"]},"extra":true}`, property))
	if _, err := ValidateResult(schema, withExtraProperty); err == nil {
		t.Fatal("expected error for extra property")
	}

	wrongNestedOrder := []byte(fmt.Sprintf(`{%q:"continue","details":{"order":["second","first"]}}`, property))
	if _, err := ValidateResult(schema, wrongNestedOrder); err == nil {
		t.Fatal("expected error for wrong nested order")
	}
}

func TestResultValidatorRejectsInvalidResults(t *testing.T) {
	schema := json.RawMessage(`{
		"type":"object",
		"properties":{"next":{"type":"string","enum":["done"]},"notes":{"type":"string"}},
		"required":["next","notes"],
		"additionalProperties":false
	}`)
	validator, err := NewResultValidator(schema)
	if err != nil {
		t.Fatalf("NewResultValidator() error = %v", err)
	}

	tests := map[string][]byte{
		"malformed":      []byte(`{"next":`),
		"trailing value": []byte(`{"next":"done","notes":"ok"} {}`),
		"non-object":     []byte(`[]`),
		"wrong enum":     []byte(`{"next":"retry","notes":"no"}`),
		"missing field":  []byte(`{"next":"done"}`),
		"extra field":    []byte(`{"next":"done","notes":"ok","other":1}`),
	}
	for name, result := range tests {
		t.Run(name, func(t *testing.T) {
			if _, gotErr := validator.Validate(result); gotErr == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestNewResultValidatorRejectsInvalidObjectSchemas(t *testing.T) {
	tests := map[string]json.RawMessage{
		"empty":           nil,
		"malformed":       json.RawMessage(`{"type":`),
		"trailing value":  json.RawMessage(`{"type":"object"} {}`),
		"non-object JSON": json.RawMessage(`[]`),
		"non-object type": json.RawMessage(`{"type":"array"}`),
		"invalid schema":  json.RawMessage(`{"type":"object","properties":[]}`),
	}
	for name, schema := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewResultValidator(schema); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestTextResultEncodesAJSONString(t *testing.T) {
	got, err := TextResult("hello\n")
	if err != nil {
		t.Fatal(err)
	}
	var text string
	if err := json.Unmarshal(got, &text); err != nil || text != "hello\n" {
		t.Fatalf("TextResult() = %s, err = %v", got, err)
	}
}
