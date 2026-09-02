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
			"Use plan to produce a sized project, medium to execute one sprint checklist, or large to plan and execute nested chapter checklists. " +
			"The execution command printed as Next runs the whole plan in one foreground run.",
		Example: "  tractor workflow list\n" +
			"  tractor workflow run plan --project demo --seed seed.md\n" +
			"  tractor workflow run medium --project demo\n" +
			"  tractor workflow run large --project demo",
		Args: cobra.NoArgs,
	}
	command.AddCommand(newWorkflowListCommand(), newWorkflowRunCommand(run), newWorkflowValidatePlanCommand())
	return command
}

func newWorkflowListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available built-in workflows",
		Long:  "List every runnable built-in workflow in this Tractor binary. Run one with 'tractor workflow run <name>'.",
		Example: "  tractor workflow list\n" +
			"  tractor workflow run plan --project demo --seed seed.md",
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
		Long: "Run plan, medium, or large through Tractor's foreground runner. Plan requires --seed, writes brief.md, checklist.md, " +
			"and recommendation.md, then prints Size and Next. Medium runs one checklist loop. Large plans each chapter and runs its nested sprint loop. " +
			"Both execution workflows run the whole project without child runs; the loop engine validates and marks items. All three configure the project's " +
			"interview directory for blocking questions, print 'Logs: <absolute-path>' before starting, and allocate fresh logs under Tractor's state root unless --logs is set.",
		Example: "  tractor workflow run plan --project demo --seed seed.md\n" +
			"  tractor workflow run medium --project demo\n" +
			"  tractor workflow run large --project demo --workdir /path/to/repo --logs ./tractor-large-logs",
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runWorkflow(command, run, args[0], project, seed, workdir, logsRoot)
		},
	}
	command.Flags().StringVar(&project, "project", "", "safe project directory name (required)")
	command.Flags().StringVar(&seed, "seed", "", "plan seed file, relative to --workdir (required only for plan)")
	command.Flags().StringVar(&workdir, "workdir", ".", "workflow workspace")
	command.Flags().StringVar(&logsRoot, "logs", "", "fresh run log directory, relative to --workdir (default: Tractor state root)")
	return command
}

func runWorkflow(command *cobra.Command, run pipelineRunner, name, project, seed, workdir, logsRoot string) error {
	if !registeredWorkflow(name) {
		return fmt.Errorf("unknown built-in workflow %q", name)
	}
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("--project is required")
	}
	if err := workflowlib.ValidateProject(project); err != nil {
		return fmt.Errorf("invalid --project: %w", err)
	}

	absoluteWorkdir, err := absoluteDirectory(workdir)
	if err != nil {
		return fmt.Errorf("invalid --workdir: %w", err)
	}
	var absoluteSeed string
	if name == workflowlib.PlanName {
		if strings.TrimSpace(seed) == "" {
			return fmt.Errorf("--seed is required for plan")
		}
		absoluteSeed, err = readableSeed(absoluteWorkdir, seed)
		if err != nil {
			return err
		}
	} else if strings.TrimSpace(seed) != "" {
		return fmt.Errorf("--seed is only valid for plan")
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
		Project: project, Workdir: absoluteWorkdir, Executable: executable,
		Plan: workflowlib.PlanParameters{Seed: absoluteSeed},
	})
	if err != nil {
		return err
	}
	logsRoot, err = workflowLogsRoot(logsRoot, absoluteWorkdir, name, project)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(command.OutOrStdout(), "Logs: %s\n", logsRoot); err != nil {
		return fmt.Errorf("print workflow logs: %w", err)
	}
	interviewDir := filepath.Join(projectDir, "interview")
	if err := withEnvironment(interviewDirectoryEnv, interviewDir, func() error {
		return run(command, *pipeline, absoluteWorkdir, logsRoot, false)
	}); err != nil {
		return err
	}

	if name == workflowlib.PlanName {
		recommendation, err := workflowlib.ValidatePlanArtifacts(absoluteWorkdir, project)
		if err != nil {
			return fmt.Errorf("read completed plan handoff: %w", err)
		}
		return printWorkflowHandoff(command, projectDir, recommendation)
	}
	return printExecutionHandoff(command, projectDir, name, logsRoot)
}

func registeredWorkflow(name string) bool {
	for _, definition := range workflowlib.List() {
		if definition.Name == name {
			return true
		}
	}
	return false
}

func workflowLogsRoot(configured, workdir, name, project string) (string, error) {
	if strings.TrimSpace(configured) != "" {
		path := configured
		if !filepath.IsAbs(path) {
			path = filepath.Join(workdir, path)
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("resolve --logs: %w", err)
		}
		if err := requireFreshLogsRoot(absolute); err != nil {
			return "", err
		}
		return absolute, nil
	}

	stateRoot, err := tractorStateRoot()
	if err != nil {
		return "", err
	}
	runsRoot := filepath.Join(stateRoot, "workflow-runs")
	if err := os.MkdirAll(runsRoot, 0o700); err != nil {
		return "", fmt.Errorf("create workflow runs directory: %w", err)
	}
	logsRoot, err := os.MkdirTemp(runsRoot, name+"-"+project+"-")
	if err != nil {
		return "", fmt.Errorf("allocate workflow logs: %w", err)
	}
	return logsRoot, nil
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

func printExecutionHandoff(command *cobra.Command, projectDir, name, logsRoot string) error {
	if _, err := fmt.Fprintf(command.OutOrStdout(), "Project: %s\nWorkflow: %s\nCompleted logs: %s\n", projectDir, name, logsRoot); err != nil {
		return fmt.Errorf("print workflow completion: %w", err)
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
