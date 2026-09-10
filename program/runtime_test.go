package program

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tylergannon/gimble/checklist"
	"github.com/tylergannon/gimble/harness"
)

func TestCommandReturnsNonzeroExitAndCombinedOutputAsData(t *testing.T) {
	runtime, err := NewRuntime(Config{Workdir: t.TempDir(), RunDir: t.TempDir(), Backend: &programBackend{}})
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(runtime.RunDir, "agents")); err != nil || !info.IsDir() {
		t.Fatalf("native agent log directory not prepared: %v", err)
	}
	result, err := runtime.Command(context.Background(), `printf 'out'; printf 'err' >&2; exit 7`)
	if err != nil {
		t.Fatalf("Command infrastructure error: %v", err)
	}
	if result.ExitCode != 7 || result.Output != "outerr" {
		t.Fatalf("result = %#v", result)
	}
	contents, err := os.ReadFile(filepath.Join(runtime.RunDir, "operations.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `"type":"operation_start"`) || !strings.Contains(string(contents), `"exit_code":7`) {
		t.Fatalf("operation log = %s", contents)
	}
}

func TestCommandReturnsInfrastructureFailureSeparately(t *testing.T) {
	runtime, err := NewRuntime(Config{Workdir: t.TempDir(), RunDir: t.TempDir(), Backend: &programBackend{}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = runtime.Command(ctx, "echo never")
	if !errors.Is(err, context.Canceled) && (err == nil || !strings.Contains(err.Error(), "context canceled")) {
		t.Fatalf("error = %v", err)
	}
}

func TestCodergenDecodesValidatedResultIntoRequestedType(t *testing.T) {
	backend := &programBackend{result: harness.Result{"passed": true, "notes": "observed"}}
	runtime, err := NewRuntime(Config{Workdir: t.TempDir(), RunDir: t.TempDir(), Backend: backend})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Codergen[Validation](context.Background(), runtime, CodergenRequest{
		Prompt: "judge", JSONSchema: validationSchema, Model: "gpt-5.6-sol",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != (Validation{Passed: true, Notes: "observed"}) {
		t.Fatalf("result = %#v", got)
	}
	if backend.turn.Model != "gpt-5.6-sol" || backend.turn.Provider != "openai" || backend.turn.Workdir != runtime.Workdir {
		t.Fatalf("turn = %#v", backend.turn)
	}
}

func TestValidateRunsInferenceAfterCommandFailure(t *testing.T) {
	workdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workdir, "evidence.txt"), []byte("actual evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	backend := &programBackend{result: harness.Result{"passed": true, "notes": "evidence itself looks right"}}
	runtime, err := NewRuntime(Config{Workdir: workdir, RunDir: t.TempDir(), Backend: backend})
	if err != nil {
		t.Fatal(err)
	}
	got, err := runtime.Validate(context.Background(), checklist.Item{Name: "proof", Check: "observable",
		Command: "echo command-ran; exit 4", Infer: &checklist.Infer{Files: []string{"evidence.txt"}, Prompt: "inspect it"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Passed || backend.turn.NodeID != "validate-proof" {
		t.Fatalf("validation = %#v, inference turn = %#v", got, backend.turn)
	}
	if !strings.Contains(backend.turn.Parts[0].Text, "command exit: 4") || !strings.Contains(backend.turn.Parts[0].Text, "evidence.txt") {
		t.Fatalf("judge prompt = %q", backend.turn.Parts[0].Text)
	}
}

type programBackend struct {
	result harness.Result
	runErr *harness.Error
	turn   harness.AgentTurn
}

func (b *programBackend) RunResult(turn harness.AgentTurn) (harness.Result, *harness.Error) {
	b.turn = turn
	if err := harness.ValidateAgentTurn(turn); err != nil {
		return nil, err
	}
	return b.result, b.runErr
}
func (b *programBackend) Run(harness.AgentTurn) (harness.Outcome, *harness.Error) {
	return harness.Outcome{}, nil
}
func (b *programBackend) RunSupervisor(harness.SupervisorTurn) (harness.Verdict, *harness.Error) {
	return harness.Verdict{}, nil
}
func (b *programBackend) Steer([]harness.ContentPart) harness.SteerStatus {
	return harness.SteerNotActive
}
func (b *programBackend) InterruptAll()                              {}
func (b *programBackend) Bindings() map[string]harness.ThreadBinding { return nil }
func (b *programBackend) SetBindingOpened(harness.BindingOpened)     {}
