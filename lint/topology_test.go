package lint_test

import (
	"strings"
	"testing"

	"github.com/tylergannon/gimble/graph"
	"github.com/tylergannon/gimble/lint"
)

func TestFanOutForFanInUsesConvergenceAnalysis(t *testing.T) {
	owner, err := lint.FanOutForFanIn(validParallel(), "join")
	if err != nil {
		t.Fatal(err)
	}
	if owner.ID != "parallel" {
		t.Fatalf("owner = %q", owner.ID)
	}
}

func TestFanOutForFanInRejectsMissingOrAmbiguousOwner(t *testing.T) {
	if _, err := lint.FanOutForFanIn(validParallel(), "missing"); err == nil || !strings.Contains(err.Error(), "found 0") {
		t.Fatalf("missing owner error = %v", err)
	}

	ambiguous := graph.Graph{Start: "first", Nodes: []graph.Node{
		&graph.FanOutNode{ID: "first", Branches: graph.LegacyFanOutBranches("left")},
		&graph.FanOutNode{ID: "second", Branches: graph.LegacyFanOutBranches("right")},
		codergen("left", edge("join")),
		codergen("right", edge("join")),
		fanIn("join"),
	}}
	if _, err := lint.FanOutForFanIn(ambiguous, "join"); err == nil || !strings.Contains(err.Error(), "found 2") {
		t.Fatalf("ambiguous owner error = %v", err)
	}
}
