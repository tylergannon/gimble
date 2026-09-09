package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tylergannon/gimble/harness"
	"github.com/tylergannon/gimble/harness/agy"
	"github.com/tylergannon/gimble/harness/claude"
	"github.com/tylergannon/gimble/harness/codex"
	"github.com/tylergannon/gimble/internal/modelalias"
	"github.com/tylergannon/gimble/internal/runlog"
)

type promptCaller string

const (
	promptCallerNone   promptCaller = ""
	promptCallerClaude promptCaller = "claude"
	promptCallerCodex  promptCaller = "codex"
)

type runPromptOptions struct {
	model        string
	modelVersion string
	effort       string
	outputSchema string
	workdir      string
	logsRoot     string
	sessionID    string
	timeout      time.Duration
}

func newRunPromptCommand() *cobra.Command {
	var options runPromptOptions
	command := &cobra.Command{
		Use:   "run-prompt [flags] PROMPT",
		Short: "Run one coding-agent prompt",
		Long:  "Run one coding-agent prompt with real tool access. Output is plain text unless --output-schema supplies an exact JSON Schema.",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runPrompt(command, options, detectPromptCaller(os.Getenv), args[0])
		},
	}
	command.Flags().StringVar(&options.model, "model", "", "model name or alias; omitted inside Codex or Claude selects the opposite provider")
	command.Flags().StringVar(&options.modelVersion, "model-version", "", "exact model-family version")
	command.Flags().StringVar(&options.effort, "effort", "", "reasoning effort: low, medium, high, xhigh, or max")
	command.Flags().StringVar(&options.outputSchema, "output-schema", "", "exact JSON Schema for structured output; omit for plain text")
	command.Flags().StringVar(&options.workdir, "workdir", ".", "agent workspace")
	command.Flags().StringVar(&options.logsRoot, "logs", "", "run log directory; default is a new temporary directory")
	command.Flags().StringVar(&options.sessionID, "session", "", "existing native session ID")
	command.Flags().DurationVar(&options.timeout, "timeout", 20*time.Minute, "turn timeout")
	return command
}

func detectPromptCaller(getenv func(string) string) promptCaller {
	if strings.TrimSpace(getenv("CLAUDE_CODE_SESSION_ID")) != "" {
		return promptCallerClaude
	}
	if strings.TrimSpace(getenv("CODEX_THREAD_ID")) != "" {
		return promptCallerCodex
	}
	return promptCallerNone
}

func resolvePromptModel(options runPromptOptions, caller promptCaller) (modelalias.ResolvedSelection, error) {
	name := strings.TrimSpace(options.model)
	if name == "" {
		switch caller {
		case promptCallerCodex:
			name = "fable"
		case promptCallerClaude:
			name = "gpt"
		default:
			return modelalias.ResolvedSelection{}, fmt.Errorf("--model is required outside Codex or Claude")
		}
	}
	selection := modelalias.Selection{Name: name}
	if options.modelVersion != "" {
		selection.Version = options.modelVersion
		selection.VersionPresent = true
	}
	if options.effort != "" {
		selection.Effort = options.effort
		selection.EffortPresent = true
	}
	return modelalias.ResolveModel(selection)
}

func runPrompt(command *cobra.Command, options runPromptOptions, caller promptCaller, prompt string) error {
	selection, err := resolvePromptModel(options, caller)
	if err != nil {
		return err
	}
	workdir, err := absoluteDirectory(options.workdir)
	if err != nil {
		return err
	}
	logsRoot := options.logsRoot
	if logsRoot == "" {
		logsRoot, err = os.MkdirTemp("", "gimble-run-prompt-")
	} else {
		logsRoot, err = filepath.Abs(logsRoot)
	}
	if err != nil {
		return fmt.Errorf("prepare logs directory: %w", err)
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
	bindings := map[string]harness.ThreadBinding(nil)
	if options.sessionID != "" {
		bindings = map[string]harness.ThreadBinding{
			"run-prompt": {Harness: selection.Harness, SessionID: options.sessionID, Workdir: workdir},
		}
	}
	backend, backendErr := harness.NewHarnessBackend(logsRoot,
		map[string]harness.HarnessAdapter{"agy": agyAdapter, "claude": claudeAdapter, "codex": codexAdapter},
		harness.DefaultProviderRoutes(), bindings)
	if backendErr != nil {
		return backendErr
	}
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			backend.InterruptAll()
		case <-done:
		}
	}()
	allocator, err := runlog.New(logsRoot)
	if err != nil {
		return err
	}
	segment, err := allocator.Allocate("run-prompt")
	if err != nil {
		return err
	}
	common := harness.TextTurn{
		NodeID: "run-prompt", Role: "agent", Parts: []harness.ContentPart{{Type: harness.ContentPartText, Text: prompt}},
		Model: selection.Model, Provider: selection.Provider, ReasoningEffort: selection.Effort,
		Fidelity: harness.FidelityFull, ThreadKey: "run-prompt", Workdir: workdir,
		RunLog: segment.Path, Timeout: options.timeout,
	}
	if options.outputSchema == "" {
		text, runErr := backend.RunText(common)
		if runErr != nil {
			return runErr
		}
		if _, err := fmt.Fprint(command.OutOrStdout(), text); err != nil {
			return err
		}
		if !strings.HasSuffix(text, "\n") {
			if _, err := fmt.Fprintln(command.OutOrStdout()); err != nil {
				return err
			}
		}
	} else {
		turn := harness.AgentTurn{
			NodeID: common.NodeID, Role: common.Role, Parts: common.Parts,
			OutputSchema: json.RawMessage(options.outputSchema), Model: common.Model,
			Provider: common.Provider, ReasoningEffort: common.ReasoningEffort,
			Fidelity: common.Fidelity, ThreadKey: common.ThreadKey, Workdir: common.Workdir,
			RunLog: common.RunLog, Timeout: common.Timeout,
		}
		result, runErr := backend.RunResult(turn)
		if runErr != nil {
			return runErr
		}
		if err := json.NewEncoder(command.OutOrStdout()).Encode(result); err != nil {
			return err
		}
	}
	binding := backend.Bindings()["run-prompt"]
	_, err = fmt.Fprintf(command.ErrOrStderr(), "Session: %s\nLogs: %s\n", binding.SessionID, logsRoot)
	return err
}
