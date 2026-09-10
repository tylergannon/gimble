package program

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tylergannon/gimble/harness"
	"github.com/tylergannon/gimble/internal/modelalias"
)

// CodergenRequest is one typed structured-output turn.
type CodergenRequest struct {
	Name       string
	Prompt     string
	JSONSchema map[string]any
	Model      string
	Effort     string
}

type resultBackend interface {
	RunResult(harness.AgentTurn) (harness.Result, *harness.Error)
}

// Codergen runs a native agent against the request's exact JSON Schema and
// decodes its validated object into T.
func Codergen[T any](ctx context.Context, runtime *Runtime, request CodergenRequest) (T, error) {
	var zero T
	if runtime == nil {
		return zero, errors.New("runtime is nil")
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return zero, errors.New("prompt must not be empty")
	}
	schema, err := json.Marshal(request.JSONSchema)
	if err != nil {
		return zero, fmt.Errorf("encode JSON schema: %w", err)
	}
	model := request.Model
	if model == "" {
		model = runtime.DefaultModel
	}
	selection := modelalias.Selection{Name: model}
	effort := request.Effort
	if effort == "" {
		effort = runtime.ReasoningEffort
	}
	if effort != "" {
		selection.Effort, selection.EffortPresent = effort, true
	}
	resolved, err := modelalias.ResolveModel(selection)
	if err != nil {
		return zero, err
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = "codergen"
	}
	id, started, err := runtime.startOperation("codergen", "", map[string]any{"name": name, "model": resolved.Model})
	if err != nil {
		return zero, err
	}
	backend, ok := runtime.Backend.(resultBackend)
	if !ok {
		return zero, errors.New("runtime backend does not support structured results")
	}
	logPath := filepath.Join(runtime.RunDir, "agents", id+".jsonl")
	// The harness appends to an allocated segment; it deliberately does not
	// create one itself. IDs are unique within this runtime's new run directory.
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return zero, fmt.Errorf("allocate agent log: %w", err)
	}
	if err := logFile.Close(); err != nil {
		return zero, err
	}
	timeout := runtime.Timeout
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			runtime.Backend.InterruptAll()
		case <-done:
		}
	}()
	result, runErr := backend.RunResult(harness.AgentTurn{NodeID: name, Role: "agent",
		Parts: []harness.ContentPart{{Type: harness.ContentPartText, Text: request.Prompt}}, OutputSchema: schema,
		Model: resolved.Model, Provider: resolved.Provider, ReasoningEffort: resolved.Effort,
		Fidelity: harness.FidelityNone, Workdir: runtime.Workdir, RunLog: logPath, Timeout: timeout})
	close(done)
	if ctx.Err() != nil {
		_ = runtime.endOperation(id, "codergen", started, map[string]any{"error": ctx.Err().Error(), "run_log": logPath})
		return zero, ctx.Err()
	}
	if runErr != nil {
		_ = runtime.endOperation(id, "codergen", started, map[string]any{"error": runErr.Error(), "run_log": logPath})
		return zero, runErr
	}
	raw, err := json.Marshal(result)
	if err == nil {
		err = json.Unmarshal(raw, &zero)
	}
	endErr := runtime.endOperation(id, "codergen", started, map[string]any{"error": errorString(err), "run_log": logPath})
	if err != nil {
		return zero, fmt.Errorf("decode structured result: %w", err)
	}
	return zero, endErr
}
