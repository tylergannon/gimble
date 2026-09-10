package main

import "github.com/tylergannon/gimble/examples/go-workflows/internal/program"

func exampleInput() Input {
	return Input{
		Goal: "Deliver a quote CLI with free shipping from a 50.00 subtotal and demonstrate the boundary.",
		Chapters: []Chapter{
			{Name: "Quote command", Command: "check quote chapter", Sprints: []Sprint{
				{Name: "Parse amounts", Command: "check amount parsing"},
				{Name: "Compute shipping", Command: "check shipping calculation"},
			}},
			{Name: "Behavioral proof", Command: "check proof chapter", Sprints: []Sprint{
				{Name: "Boundary evidence", Command: "check price boundaries"},
			}},
		},
	}
}

type reviewAssignment struct {
	Name  string
	Role  program.Role
	Focus string
}

func reviewAssignments() []reviewAssignment {
	return []reviewAssignment{
		{Name: "evidence", Role: tester, Focus: "Independently demonstrate the exact price boundary."},
		{Name: "scope-check", Role: engMgr, Focus: "Check that the change serves only the declared scope."},
	}
}
