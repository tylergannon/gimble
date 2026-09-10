package main

import (
	"fmt"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func implementPrompt(goal string, item program.Item) string {
	return fmt.Sprintf("Goal: %s. Implement sprint %q. Demonstrate it with %s.", goal, item.Name, item.Command)
}

func reviewPrompt(item program.Item) string {
	return fmt.Sprintf("Run the software for sprint %q. Report specific material defects.", item.Name)
}

func repairPrompt(item program.Item, review Review) string {
	return fmt.Sprintf("Repair sprint %q: %s", item.Name, review.Notes)
}

func evaluatePrompt(item program.Item, check program.Check) string {
	return fmt.Sprintf("Does this observation demonstrate sprint %q?\n%s", item.Name, check.Output)
}
