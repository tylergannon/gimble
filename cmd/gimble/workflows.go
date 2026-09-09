package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tylergannon/gimble/internal/workflows"
)

func newWorkflowsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "workflows",
		Short: "List the workflows that ship in this binary",
		Long: "List the workflows that ship in this binary.\n\n" +
			"Run one by name with `gimble run <name>`, read one with\n" +
			"`gimble workflows show <name>`, or redirect that into a file to\n" +
			"start your own.\n\n" +
			"The third column is what a workflow wants before it will do\n" +
			"anything: `needs --goal` means it works on whatever you name and\n" +
			"refuses to start without one; `needs <path>` means it reads that\n" +
			"file in the workspace and fails on the first node if it is\n" +
			"missing.",
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			writer := tabwriter.NewWriter(command.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, workflow := range workflows.List() {
				line := fmt.Sprintf("%s\t%s", workflow.Name, workflow.When)
				switch {
				case workflow.NeedsGoal:
					line += "\tneeds --goal"
				case workflow.Needs != "":
					line += "\tneeds " + workflow.Needs
				}
				if _, err := fmt.Fprintln(writer, line); err != nil {
					return err
				}
			}
			return writer.Flush()
		},
	}
	command.AddCommand(newWorkflowsShowCommand())
	return command
}

func newWorkflowsShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Print a built-in workflow's pipeline",
		Long: "Print a built-in workflow's pipeline to stdout.\n\n" +
			"Redirect it into a file to adapt it: the copy is an ordinary\n" +
			"pipeline with no tie back to the binary.",
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			raw, err := workflows.Read(args[0])
			if err != nil {
				return fmt.Errorf("%w; run `gimble workflows` to list them", err)
			}
			_, err = command.OutOrStdout().Write(raw)
			return err
		},
	}
}

// builtinPipelineHint names the built-in workflows for an error message about
// a pipeline source that resolved to neither a file nor a workflow.
func builtinPipelineHint() string {
	return strings.Join(workflows.Names(), ", ")
}
