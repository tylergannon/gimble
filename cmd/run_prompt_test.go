package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestResolvePromptModelPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		options runPromptOptions
		caller  promptCaller
		model   string
		harness string
		wantErr bool
	}{
		{name: "Codex gets opposite provider", caller: promptCallerCodex, model: "claude-fable-5-1", harness: "claude"},
		{name: "Claude gets opposite provider", caller: promptCallerClaude, model: "gpt-5.6-sol", harness: "codex"},
		{name: "explicit flash effort", caller: promptCallerCodex, options: runPromptOptions{model: "flash", effort: "low"}, model: "gemini-3.8-flash-low", harness: "agy"},
		{name: "standalone requires model", caller: promptCallerNone, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolvePromptModel(test.options, test.caller)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Model != test.model || got.Harness != test.harness {
				t.Fatalf("selection = %#v", got)
			}
		})
	}
}

func TestDetectPromptCallerPrefersClaude(t *testing.T) {
	environment := map[string]string{"CLAUDE_CODE_SESSION_ID": "claude", "CODEX_THREAD_ID": "codex"}
	if got := detectPromptCaller(func(name string) string { return environment[name] }); got != promptCallerClaude {
		t.Fatalf("caller = %q", got)
	}
}

func TestRunPromptJSONValidatesExactSchema(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"],"additionalProperties":false}`)
	compiled, err := compilePromptSchema(schema)
	if err != nil {
		t.Fatal(err)
	}
	runPromptSchemaMu.Lock()
	defer runPromptSchemaMu.Unlock()
	runPromptSchema, runPromptValidator = schema, compiled
	if err := (runPromptJSON{}).ValidateJSON([]byte(`{"answer":"yes"}`)); err != nil {
		t.Fatal(err)
	}
	if err := (runPromptJSON{}).ValidateJSON([]byte(`{"answer":1}`)); err == nil {
		t.Fatal("invalid structured output was accepted")
	}
}

func TestPromptProjectDirRejectsNonemptyDirectory(t *testing.T) {
	dir := t.TempDir()
	if _, err := promptProjectDir(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/occupied", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := promptProjectDir(dir); err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("error = %v", err)
	}
}
