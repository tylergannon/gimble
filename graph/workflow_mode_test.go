package graph

import (
	"strings"
	"testing"
)

const explicitModeDocument = `{
  "mode":"delivery",
  "start":"check",
  "nodes":[{"id":"check","type":"tool","tool_command":"true","on_success":"success"}]
}`

func TestParseAcceptsExplicitWorkflowModes(t *testing.T) {
	for _, mode := range []string{"delivery", "discovery"} {
		document := strings.Replace(explicitModeDocument, `"mode":"delivery"`, `"mode":"`+mode+`"`, 1)
		if _, err := Parse([]byte(document)); err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
	}
}

func TestParseRejectsUnknownWorkflowMode(t *testing.T) {
	document := strings.Replace(explicitModeDocument, `"mode":"delivery"`, `"mode":"experiment"`, 1)
	if _, err := Parse([]byte(document)); err == nil {
		t.Fatal("unknown workflow mode admitted")
	}
}
