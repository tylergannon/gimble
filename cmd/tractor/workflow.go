package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tylergannon/tractor/graph"
	workflowlib "github.com/tylergannon/tractor/workflow"
)

const interviewDirectoryEnv = "TRACTOR_INTERVIEW_DIR"

type pipelineRunner func(*cobra.Command, graph.Graph, string, string, bool) error

func newWorkflowCommand(run pipelineRunner) *cobra.Command {
	command := &cobra.Command{
		Use:   "workflow",
		Short: "List and run Tractor's built-in workflows",
		Long: "List and run workflows embedded in Tractor. Start with 'tractor workflow list'. " +
			"Use the plan workflow when you need to turn a seed into an interviewed, sized plan.",
		Example: "  tractor workflow list\n" +
			"  tractor workflow run plan --project demo --seed seed.md --logs ./tractor-plan-logs",
		Args: cobra.NoArgs,
	}
	command.AddCommand(newWorkflowListCommand(), newWorkflowRunCommand(run), newWorkflowValidatePlanCommand())
	return command
}

func newWorkflowListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available built-in workflows",
		Long:  "List the built-in workflows available in this Tractor binary. Start with plan when you need to turn a seed into an interviewed, sized plan.",
		Example: "  tractor workflow list\n" +
			"  tractor workflow run plan --project demo --seed seed.md --logs ./tractor-plan-logs",
		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			for _, definition := range workflowlib.List() {
				if _, err := fmt.Fprintf(command.OutOrStdout(), "%s\t%s\n", definition.Name, definition.Description); err != nil {
					return fmt.Errorf("print built-in workflows: %w", err)
				}
			}
			return nil
		},
	}
}

func newWorkflowRunCommand(run pipelineRunner) *cobra.Command {
	var project string
	var seed string
	var workdir string
	var logsRoot string
	command := &cobra.Command{
		Use:   "run <name>",
		Short: "Run a built-in workflow",
		Long: "Run a workflow embedded in Tractor. The plan workflow reads --seed relative to --workdir, " +
			"interviews the caller through numbered QuestionAsked events, and writes brief.md, checklist.md, " +
			"and recommendation.md under ephemeral/projects/<project>/.",
		Example: "  tractor workflow run plan --project demo --seed seed.md --logs ./tractor-plan-logs\n" +
			"  tractor workflow run plan --project demo --seed notes/seed.md --workdir /path/to/repo --logs /tmp/demo-plan",
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runWorkflow(command, run, args[0], project, seed, workdir, logsRoot)
		},
	}
	command.Flags().StringVar(&project, "project", "", "safe project directory name (required)")
	command.Flags().StringVar(&seed, "seed", "", "seed file, relative to --workdir (required)")
	command.Flags().StringVar(&workdir, "workdir", ".", "workflow workspace")
	command.Flags().StringVar(&logsRoot, "logs", "", "run log directory (required)")
	return command
}

func runWorkflow(command *cobra.Command, run pipelineRunner, name, project, seed, workdir, logsRoot string) error {
	if name != workflowlib.PlanName {
		return fmt.Errorf("unknown built-in workflow %q", name)
	}
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("--project is required")
	}
	if err := workflowlib.ValidateProject(project); err != nil {
		return fmt.Errorf("invalid --project: %w", err)
	}
	if strings.TrimSpace(seed) == "" {
		return fmt.Errorf("--seed is required")
	}
	if strings.TrimSpace(logsRoot) == "" {
		return fmt.Errorf("--logs is required")
	}

	absoluteWorkdir, err := absoluteDirectory(workdir)
	if err != nil {
		return fmt.Errorf("invalid --workdir: %w", err)
	}
	absoluteSeed, err := readableSeed(absoluteWorkdir, seed)
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve Tractor executable: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return fmt.Errorf("resolve Tractor executable: %w", err)
	}
	projectDir, err := workflowlib.ProjectDir(absoluteWorkdir, project)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("create project directory %q: %w", projectDir, err)
	}

	pipeline, err := workflowlib.Build(name, workflowlib.Parameters{
		Project: project, Seed: absoluteSeed, Workdir: absoluteWorkdir, Executable: executable,
	})
	if err != nil {
		return err
	}
	interviewDir := filepath.Join(projectDir, "interview")
	if err := withEnvironment(interviewDirectoryEnv, interviewDir, func() error {
		return run(command, *pipeline, absoluteWorkdir, logsRoot, false)
	}); err != nil {
		return err
	}

	recommendation, err := workflowlib.ValidatePlanArtifacts(absoluteWorkdir, project)
	if err != nil {
		return fmt.Errorf("read completed plan handoff: %w", err)
	}
	return printWorkflowHandoff(command, projectDir, recommendation)
}

func readableSeed(workdir, seed string) (string, error) {
	path := seed
	if !filepath.IsAbs(path) {
		path = filepath.Join(workdir, path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve --seed %q: %w", seed, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("inspect --seed %q: %w", seed, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("--seed %q is not a regular file", seed)
	}
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read --seed %q: %w", seed, err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close --seed %q: %w", seed, err)
	}
	return path, nil
}

func withEnvironment(name, value string, action func() error) error {
	previous, present := os.LookupEnv(name)
	if err := os.Setenv(name, value); err != nil {
		return fmt.Errorf("set %s: %w", name, err)
	}
	defer func() {
		if present {
			_ = os.Setenv(name, previous)
		} else {
			_ = os.Unsetenv(name)
		}
	}()
	return action()
}

func printWorkflowHandoff(command *cobra.Command, projectDir string, recommendation *workflowlib.Recommendation) error {
	for _, artifact := range []struct {
		label string
		name  string
	}{
		{label: "Brief", name: workflowlib.BriefFile},
		{label: "Checklist", name: workflowlib.ChecklistFile},
		{label: "Recommendation", name: workflowlib.RecommendationFile},
	} {
		if _, err := fmt.Fprintf(command.OutOrStdout(), "%s: %s\n", artifact.label, filepath.Join(projectDir, artifact.name)); err != nil {
			return fmt.Errorf("print workflow handoff: %w", err)
		}
	}
	if _, err := fmt.Fprintf(command.OutOrStdout(), "Size: %s\nNext: %s\n", recommendation.Size, recommendation.Next); err != nil {
		return fmt.Errorf("print workflow handoff: %w", err)
	}
	return nil
}

func newWorkflowValidatePlanCommand() *cobra.Command {
	var project string
	var workdir string
	command := &cobra.Command{
		Use:    "validate-plan",
		Short:  "Validate artifacts written by the built-in plan workflow",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(project) == "" {
				return fmt.Errorf("--project is required")
			}
			absoluteWorkdir, err := absoluteDirectory(workdir)
			if err != nil {
				return fmt.Errorf("invalid --workdir: %w", err)
			}
			_, err = workflowlib.ValidatePlanArtifacts(absoluteWorkdir, project)
			return err
		},
	}
	command.Flags().StringVar(&project, "project", "", "safe project directory name (required)")
	command.Flags().StringVar(&workdir, "workdir", ".", "workflow workspace")
	return command
}
