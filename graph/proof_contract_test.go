package graph

import (
	"strings"
	"testing"
)

const completeProofContractDocument = `{
  "mode":"delivery",
  "start":"primary_assertion",
  "proof_contract":{
    "intended_architecture":"Input crosses the operating boundary and produces a user-visible result.",
    "primary_outcome":"Representative input produces an observable result.",
    "primary_cases":[{
      "id":"qualifying_input",
      "actor":"A user",
      "job":"Submit qualifying input and observe the intended result.",
      "node":"primary_assertion",
      "qualifying_input_criteria":"Input independently established to meet the product's qualifying criteria.",
      "expected_output":"The expected value is visible.",
      "correctness_oracle":"tool_exit_zero",
      "evidence_mode":"operating_layer",
      "evidence_source":"current_run",
      "independence":"independent_execution",
      "status":"unproven",
      "required_capabilities":["operating_runtime"],
      "evidence_artifacts":[
        {"path":"evidence/qualifying-input.txt","role":"input","architecture_edge":"client_to_product"},
        {"path":"evidence/observable-output.txt","role":"observation","architecture_edge":"product_to_user"}
      ]
    }],
    "boundary_cases":[{
      "id":"empty_boundary",
      "actor":"A user",
      "job":"Submit empty input without receiving invented output.",
      "node":"empty_assertion",
      "qualifying_input_criteria":"Input is empty.",
      "expected_output":"The observed value remains empty.",
      "correctness_oracle":"tool_exit_zero",
      "evidence_mode":"operating_layer",
      "evidence_source":"current_run",
      "independence":"independent_execution",
      "status":"unproven",
      "required_capabilities":["operating_runtime"],
      "evidence_artifacts":[]
    }],
    "required_capabilities":[{
      "id":"operating_runtime",
      "description":"The operating product surface is available to the assertion command."
    }],
    "unknowns":[],
    "scope_gaps":[],
    "terminal_success":{"required_cases":["qualifying_input","empty_boundary"]}
  },
  "nodes":[
    {"id":"primary_assertion","type":"tool","tool_command":"true","on_success":"empty_assertion"},
    {"id":"empty_assertion","type":"tool","tool_command":"true","on_success":"success"}
  ]
}`

func TestParseAcceptsCompleteProofContract(t *testing.T) {
	if _, err := Parse([]byte(completeProofContractDocument)); err != nil {
		t.Fatal(err)
	}
}

func TestParseRejectsStructurallyAmbiguousProofContract(t *testing.T) {
	tests := map[string]string{
		"missing expected output": strings.Replace(completeProofContractDocument,
			`"expected_output":"The expected value is visible.",`, "", 1),
		"empty expected output": strings.Replace(completeProofContractDocument,
			`"expected_output":"The expected value is visible."`, `"expected_output":""`, 1),
		"missing oracle": strings.Replace(completeProofContractDocument,
			`"correctness_oracle":"tool_exit_zero",`, "", 1),
		"unknown oracle": strings.Replace(completeProofContractDocument,
			`"correctness_oracle":"tool_exit_zero"`, `"correctness_oracle":"review_record"`, 1),
		"unknown evidence mode": strings.Replace(completeProofContractDocument,
			`"evidence_mode":"operating_layer"`, `"evidence_mode":"mock"`, 1),
		"unknown promise status": strings.Replace(completeProofContractDocument,
			`"status":"unproven"`, `"status":"done"`, 1),
		"unknown evidence source": strings.Replace(completeProofContractDocument,
			`"evidence_source":"current_run"`, `"evidence_source":"report"`, 1),
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(document)); err == nil {
				t.Fatal("ambiguous proof contract admitted")
			}
		})
	}
}
