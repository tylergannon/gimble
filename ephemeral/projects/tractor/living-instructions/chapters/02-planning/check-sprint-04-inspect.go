// Command check-sprint-04-inspect verifies the semantic content of the live
// planning proof's generated loop checklist.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tylergannon/tractor/checklist"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: check-sprint-04-inspect <checklist>")
		os.Exit(2)
	}

	list, err := checklist.Load(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "load generated checklist: %v\n", err)
		os.Exit(1)
	}
	if len(list.Items) != 1 {
		fmt.Fprintf(os.Stderr, "generated checklist has %d items, want exactly 1\n", len(list.Items))
		os.Exit(1)
	}

	item := list.Items[0]
	if item.Done || item.DonePresent {
		fmt.Fprintln(os.Stderr, "generated checklist item contains engine-owned done state")
		os.Exit(1)
	}
	for label, value := range map[string]string{
		"check":   item.Check,
		"command": item.Command,
	} {
		if !strings.Contains(value, "greet.sh") || !strings.Contains(value, "Hello, Tractor!") {
			fmt.Fprintf(os.Stderr, "generated checklist %s is not specific to the chosen greeting contract: %q\n", label, value)
			os.Exit(1)
		}
	}

	fmt.Println("generated checklist: one open, seed-specific item")
}
