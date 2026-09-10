package main

import (
	"fmt"
	"strings"
)

func buildPrompt(input Input) string {
	return input.Goal + "\nLeave REPORT.md recording what you built and how you demonstrated it works."
}

func judgePrompt(input Input, candidates []Candidate) string {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "%s\nRun each candidate yourself; judge demonstrated behavior, not its report.\n", input.Goal)
	fmt.Fprintf(&prompt, "Acceptance after integration: %s\n", input.Acceptance.Command)
	for i, candidate := range candidates {
		fmt.Fprintf(&prompt, "Candidate %d (%s), workspace %s: %s\n", i, candidate.Builder, candidate.Workspace, candidate.Report)
	}
	prompt.WriteString("Return the winning candidate's zero-based index and your reason. Record the reason in WINNER.md in the main workspace. The caller integrates the winner next.")
	return prompt.String()
}
