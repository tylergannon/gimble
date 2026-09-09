package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tylergannon/gimble/engine"
	"github.com/tylergannon/gimble/graph"
	"github.com/tylergannon/gimble/harness"
	"github.com/tylergannon/gimble/harness/agy"
	"github.com/tylergannon/gimble/harness/claude"
	"github.com/tylergannon/gimble/harness/codex"
	"github.com/tylergannon/gimble/internal/modelalias"
	"github.com/tylergannon/gimble/internal/workflows"
	"github.com/tylergannon/gimble/lint"
)

const (
	defaultModel           = "gpt-5.6-sol"
	defaultReasoningEffort = "high"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:               "gimble",
		Short:             "Run Gimble pipelines",
		SilenceErrors:     true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.AddCommand(newAskCommand(), newAnswerCommand(), newValidateCommand(), newInspectModelsCommand(), newRunCommand(), newRunPromptCommand(), newEditCommand(), newWorkflowsCommand(), newPrintSchemaCommand(), newMCPCommand(), newMCPRunnerCommand(), newPluginCommand())
	return root
}

func newInspectModelsCommand() *cobra.Command {
	var inlineJSON string
	var inlineYAML string
	command := &cobra.Command{
		Use:   "inspect-models [pipeline]",
		Short: "Print every effective model selection without executing",
		RunE: func(command *cobra.Command, args []string) error {
			pipeline, _, err := loadPipeline(args, inlineJSON, command.Flags().Changed("json"), inlineYAML, command.Flags().Changed("yaml"))
			if err != nil {
				return err
			}
			if err := validateAndReport(command, cliValidator(), *pipeline); err != nil {
				return err
			}
			resolved, err := engine.ResolveGraphModels(*pipeline, systemModelSelection())
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(command.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(resolved)
		},
	}
	command.Flags().StringVar(&inlineJSON, "json", "", "pipeline JSON")
	command.Flags().StringVar(&inlineYAML, "yaml", "", "pipeline YAML")
	return command
}

func newValidateCommand() *cobra.Command {
	var inlineJSON string
	var inlineYAML string
	command := &cobra.Command{
		Use:   "validate [pipeline]",
		Short: "Parse and validate a pipeline",
		RunE: func(command *cobra.Command, args []string) error {
			pipeline, source, err := loadPipeline(
				args,
				inlineJSON, command.Flags().Changed("json"),
				inlineYAML, command.Flags().Changed("yaml"),
			)
			if err != nil {
				return err
			}
			if err := validateAndReport(command, cliValidator(), *pipeline); err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.OutOrStdout(), "valid %s\n", source.Display)
			return err
		},
	}
	command.Flags().StringVar(&inlineJSON, "json", "", "pipeline JSON")
	command.Flags().StringVar(&inlineYAML, "yaml", "", "pipeline YAML")
	return command
}

func newRunCommand() *cobra.Command {
	var inlineJSON string
	var inlineYAML string
	var workdir string
	var logsRoot string
	var resume bool
	var goal string
	var wake string
	var wakeInterval time.Duration
	command := &cobra.Command{
		Use:   "run [pipeline]",
		Short: "Run a pipeline",
		RunE: func(command *cobra.Command, args []string) error {
			pipeline, source, err := loadPipeline(
				args,
				inlineJSON, command.Flags().Changed("json"),
				inlineYAML, command.Flags().Changed("yaml"),
			)
			if err != nil {
				return err
			}
			if err := applyGoal(pipeline, source, goal, command.Flags().Changed("goal")); err != nil {
				return err
			}
			wakeConfig, err := wakeConfiguration(wake, wakeInterval)
			if err != nil {
				return err
			}
			return runPipeline(command, *pipeline, source.Provenance, workdir, logsRoot, resume, wakeConfig)
		},
	}
	command.Flags().StringVar(&inlineJSON, "json", "", "pipeline JSON")
	command.Flags().StringVar(&inlineYAML, "yaml", "", "pipeline YAML")
	command.Flags().StringVar(&workdir, "workdir", ".", "pipeline workspace")
	command.Flags().StringVar(&logsRoot, "logs", "", "run log directory")
	command.Flags().BoolVar(&resume, "resume", false, "resume from the logs checkpoint")
	command.Flags().StringVar(&goal, "goal", "", "replace the pipeline's goal; every $goal in a prompt expands to this")
	command.Flags().StringVar(&wake, "wake", string(engine.WakeAuto), "wake the agent session that started the run: auto (background sessions only), on (any session), off")
	command.Flags().DurationVar(&wakeInterval, "wake-interval", engine.DefaultWakeInterval, "how often an armed run checks whether it has news to deliver")
	return command
}

// wakeConfiguration turns the --wake flags into engine configuration.
func wakeConfiguration(mode string, interval time.Duration) (engine.WakeConfig, error) {
	switch engine.WakeMode(strings.ToLower(strings.TrimSpace(mode))) {
	case engine.WakeAuto:
		return engine.WakeConfig{Mode: engine.WakeAuto, Interval: interval}, nil
	case engine.WakeOn:
		return engine.WakeConfig{Mode: engine.WakeOn, Interval: interval}, nil
	case engine.WakeOff:
		return engine.WakeConfig{Mode: engine.WakeOff}, nil
	default:
		return engine.WakeConfig{}, fmt.Errorf("unknown --wake value %q: use auto, on, or off", mode)
	}
}

func newPrintSchemaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "print-schema",
		Short: "Print the pipeline JSON Schema",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := command.OutOrStdout().Write(graph.Graph{}.Schema())
			return err
		},
	}
}

type pipelineSource struct {
	Display    string
	Provenance string
	Builtin    bool
}

func loadPipeline(args []string, inlineJSON string, jsonSet bool, inlineYAML string, yamlSet bool) (*graph.Graph, pipelineSource, error) {
	if len(args) > 1 {
		return nil, pipelineSource{}, fmt.Errorf("accepts exactly one pipeline source")
	}
	if (jsonSet && yamlSet) || (len(args) == 1 && (jsonSet || yamlSet)) {
		return nil, pipelineSource{}, fmt.Errorf("pipeline file, --json, and --yaml are mutually exclusive")
	}
	if !jsonSet && !yamlSet && len(args) == 0 {
		return nil, pipelineSource{}, fmt.Errorf("pipeline source is required: provide a file, --json, or --yaml")
	}
	if jsonSet {
		pipeline, err := graph.Parse([]byte(inlineJSON))
		return pipeline, pipelineSource{Display: "--json", Provenance: "inline"}, err
	}
	if yamlSet {
		pipeline, err := graph.ParseYAML([]byte(inlineYAML))
		return pipeline, pipelineSource{Display: "--yaml", Provenance: "inline"}, err
	}
	return loadPipelineSource(args[0])
}

// loadPipelineSource reads a pipeline from a file, or from the built-in
// catalogue when no such file exists. A file on disk always wins, so a local
// pipeline is never shadowed by a workflow that ships in the binary.
func loadPipelineSource(source string) (*graph.Graph, pipelineSource, error) {
	raw, err := os.ReadFile(source)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, pipelineSource{}, fmt.Errorf("read pipeline %q: %w", source, err)
		}
		builtin, builtinErr := workflows.Read(source)
		if builtinErr != nil {
			return nil, pipelineSource{}, fmt.Errorf(
				"pipeline %q is neither a file nor a built-in workflow; built-in workflows are %s",
				source, builtinPipelineHint(),
			)
		}
		pipeline, parseErr := graph.ParseYAML(builtin)
		return pipeline, pipelineSource{Display: source, Provenance: "builtin:" + source, Builtin: true}, parseErr
	}
	absolute, absoluteErr := filepath.Abs(source)
	if absoluteErr != nil {
		return nil, pipelineSource{}, fmt.Errorf("resolve pipeline path %q: %w", source, absoluteErr)
	}
	var pipeline *graph.Graph
	if extension := strings.ToLower(filepath.Ext(source)); extension == ".yaml" || extension == ".yml" {
		pipeline, err = graph.ParseYAML(raw)
	} else {
		pipeline, err = graph.Parse(raw)
	}
	return pipeline, pipelineSource{Display: source, Provenance: "file:" + absolute}, err
}

// applyGoal replaces the pipeline's goal with the one supplied at the command
// line. A built-in workflow that exists to work on something the operator
// names refuses to start without one, rather than running against the generic
// goal its file carries.
func applyGoal(pipeline *graph.Graph, source pipelineSource, goal string, goalSet bool) error {
	if goalSet {
		if strings.TrimSpace(goal) == "" {
			return fmt.Errorf("--goal requires text")
		}
		pipeline.Goal = goal
		return nil
	}
	if workflow, ok := workflows.Lookup(source.Display); source.Builtin && ok && workflow.NeedsGoal {
		return fmt.Errorf(
			"the %s workflow needs a goal: pass --goal with what it should work on.\n%s",
			workflow.Name, workflow.GoalHint,
		)
	}
	return nil
}

func runPipeline(command *cobra.Command, pipeline graph.Graph, pipelineSource, workdir, logsRoot string, resume bool, wake engine.WakeConfig) error {
	if strings.TrimSpace(logsRoot) == "" {
		return fmt.Errorf("--logs is required")
	}
	workdir, err := absoluteDirectory(workdir)
	if err != nil {
		return err
	}
	logsRoot, err = filepath.Abs(logsRoot)
	if err != nil {
		return fmt.Errorf("resolve logs directory: %w", err)
	}
	validator := cliValidator()
	if err := validateAndReport(command, validator, pipeline); err != nil {
		return err
	}

	var bindings map[string]harness.ThreadBinding
	if resume {
		checkpoint, err := engine.LoadCheckpoint(logsRoot)
		if err != nil {
			return err
		}
		bindings = checkpoint.Sessions
	}
	codexAdapter := codex.New()
	codexAdapter.SetStderr(command.ErrOrStderr())
	defer codexAdapter.Close()
	claudeAdapter := claude.New()
	claudeAdapter.SetStderr(command.ErrOrStderr())
	defer claudeAdapter.Close()
	agyAdapter := agy.New()
	agyAdapter.SetStderr(command.ErrOrStderr())
	defer agyAdapter.Close()
	backend, backendErr := harness.NewHarnessBackend(
		logsRoot,
		map[string]harness.HarnessAdapter{"agy": agyAdapter, "codex": codexAdapter, "claude": claudeAdapter},
		harness.DefaultProviderRoutes(),
		bindings,
	)
	if backendErr != nil {
		return backendErr
	}

	registry := engine.NewRegistry()
	agentConfig := engine.AgentConfig{
		Backend:                backend,
		DefaultModel:           defaultModel,
		DefaultReasoningEffort: defaultReasoningEffort,
	}
	registry.Register("agent", engine.NewAgentHandler(agentConfig))
	registry.Register("fan_in", engine.NewFanInHandler(agentConfig))
	runnerConfig := engine.RunnerConfig{
		LogsRoot:               logsRoot,
		Workdir:                workdir,
		PipelineSource:         pipelineSource,
		DefaultModel:           defaultModel,
		DefaultReasoningEffort: defaultReasoningEffort,
		Validate: func(candidate graph.Graph) error {
			_, err := validator.ValidateOrError(candidate)
			return err
		},
		Backend: backend,
		Wake:    wake,
	}
	var runner *engine.Runner
	if resume {
		runner, err = engine.ResumeRunner(pipeline, registry, runnerConfig)
	} else {
		runner, err = engine.NewRunner(pipeline, registry, runnerConfig)
	}
	if err != nil {
		return err
	}

	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			runner.Stop()
		case <-done:
		}
	}()
	result, err := runner.Run()
	if err != nil {
		return err
	}
	if result.Status != engine.RunCompleted {
		return fmt.Errorf("pipeline failed: %s", result.FailureReason)
	}
	_, err = fmt.Fprintln(command.OutOrStdout(), result.Status)
	return err
}

func absoluteDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve workdir: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("inspect workdir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workdir %q is not a directory", absolute)
	}
	return absolute, nil
}

func cliValidator() *lint.Validator {
	return lint.New(lint.Options{ResolveHarness: resolveHarness})
}

func resolveHarness(provider, model string) (string, error) {
	if model == "" {
		model = defaultModel
	}
	var err error
	provider, model, err = modelalias.ResolveSelection(provider, model)
	if err != nil {
		return "", err
	}
	if provider == "" {
		provider = engine.DetectProvider(model)
	}
	harnessName := harness.DefaultProviderRoutes()[provider]
	if harnessName == "" {
		return "", fmt.Errorf("no harness route for provider %q", provider)
	}
	return harnessName, nil
}

func validateAndReport(command *cobra.Command, validator *lint.Validator, pipeline graph.Graph) error {
	if _, err := engine.ResolveGraphModels(pipeline, systemModelSelection()); err != nil {
		return err
	}
	diagnostics := validator.Validate(pipeline)
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == lint.SeverityError {
			continue
		}
		if _, err := fmt.Fprintf(command.ErrOrStderr(), "%s %s: %s\n", diagnostic.Severity, diagnostic.Rule, diagnostic.Message); err != nil {
			return err
		}
	}
	if lint.HasErrors(diagnostics) {
		return &lint.ValidationError{Diagnostics: diagnostics}
	}
	return nil
}

func systemModelSelection() engine.SystemModelSelection {
	return engine.SystemModelSelection{Name: defaultModel, Effort: defaultReasoningEffort}
}
