package main

import (
	"fmt"
	"strings"
)

func proposalPrompt(input Input) string {
	return "Write your own independent proposal on this topic:\n" + input.Topic
}

func critiquePrompt(input Input, own Proposal, peers []Proposal) string {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Topic: %s\nYour proposal, for context only: %s\n", input.Topic, own.Text)
	prompt.WriteString("Critique only these two peers' proposals: their strengths, weaknesses or errors, and what all three of you agree on.\n")
	for _, peer := range peers {
		fmt.Fprintf(&prompt, "Proposal by %s:\n%s\n", peer.Author, peer.Text)
	}
	return prompt.String()
}
