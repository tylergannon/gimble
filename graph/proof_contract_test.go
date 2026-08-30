package graph

import (
	"strings"
	"testing"
)

const completeProofContractDocument = `{
  "start":"primary_assertion",
  "proof_contract":{
    "primary_outcome":"Representative input produces an observable result.",
    "primary_cases":[{
      "id":"qualifying_input",
      "node":"primary_assertion",
      "representative_input":"A known qualifying fixture.",
      "expected_output":"The expected value is visible.",
      "correctness_oracle":"tool_exit_zero",
      "evidence_mode":"operating_layer"
    }],
    "boundary_cases":[{
      "id":"empty_boundary",
      "node":"empty_assertion",
      "representative_input":"An empty fixture.",
      "expected_output":"The observed value remains empty.",
      "correctness_oracle":"tool_exit_zero",
      "evidence_mode":"operating_layer"
    }],
    "unknowns":[],
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
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(document)); err == nil {
				t.Fatal("ambiguous proof contract admitted")
			}
		})
	}
}
