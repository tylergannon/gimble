package modelalias

import (
	"strings"
	"testing"
)

func TestResolveModelAliasesVersionsAndEffort(t *testing.T) {
	tests := []struct {
		name      string
		selection Selection
		want      ResolvedSelection
	}{
		{name: "Flash default", selection: Selection{Name: "flash"}, want: ResolvedSelection{Name: "flash", Version: "3.8", Model: "gemini-3.8-flash-medium", Effort: "medium", Provider: "gemini", Harness: "agy"}},
		{name: "Flash explicit effort translates native ID", selection: Selection{Name: "flash", Effort: "high", EffortPresent: true}, want: ResolvedSelection{Name: "flash", Version: "3.8", Model: "gemini-3.8-flash-high", Effort: "high", Provider: "gemini", Harness: "agy"}},
		{name: "Flash older release pin", selection: Selection{Name: "flash", Version: "3.7", VersionPresent: true, Effort: "low", EffortPresent: true}, want: ResolvedSelection{Name: "flash", Version: "3.7", Model: "gemini-3.7-flash-low", Effort: "low", Provider: "gemini", Harness: "agy"}},
		{name: "Fable default release", selection: Selection{Name: "fable"}, want: ResolvedSelection{Name: "fable", Version: "5.1", Model: "claude-fable-5-1", Effort: "high", Provider: "anthropic", Harness: "claude"}},
		{name: "Fable 5.1 pin", selection: Selection{Name: "fable", Version: "5.1", VersionPresent: true, Effort: "low", EffortPresent: true}, want: ResolvedSelection{Name: "fable", Version: "5.1", Model: "claude-fable-5-1", Effort: "low", Provider: "anthropic", Harness: "claude"}},
		{name: "Fable 5 pin", selection: Selection{Name: "fable", Version: "5", VersionPresent: true}, want: ResolvedSelection{Name: "fable", Version: "5", Model: "claude-fable-5", Effort: "high", Provider: "anthropic", Harness: "claude"}},
		{name: "Version-bearing compatibility alias", selection: Selection{Name: "fable-5"}, want: ResolvedSelection{Name: "fable-5", Version: "5", Model: "claude-fable-5", Effort: "high", Provider: "anthropic", Harness: "claude"}},
		{name: "OpenAI native", selection: Selection{Name: "gpt-5.6-sol", Effort: "medium", EffortPresent: true}, want: ResolvedSelection{Name: "gpt-5.6-sol", Model: "gpt-5.6-sol", Effort: "medium", Provider: "openai", Harness: "codex"}},
		{name: "Gemini fixed native effort", selection: Selection{Name: "gemini-3.1-pro-low"}, want: ResolvedSelection{Name: "gemini-3.1-pro-low", Model: "gemini-3.1-pro-low", Effort: "low", Provider: "gemini", Harness: "agy"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveModel(test.selection)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("ResolveModel() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestResolveModelRejectsInvalidSelectionsWithoutFallback(t *testing.T) {
	tests := []struct {
		name      string
		selection Selection
		want      string
	}{
		{name: "blank name", selection: Selection{Name: " "}, want: "nonblank"},
		{name: "explicit blank version", selection: Selection{Name: "fable", VersionPresent: true}, want: "version must be a nonblank"},
		{name: "unsupported version", selection: Selection{Name: "fable", Version: "4", VersionPresent: true}, want: `does not support version "4"`},
		{name: "version with versioned alias", selection: Selection{Name: "fable-5", Version: "5", VersionPresent: true}, want: "already selects a version"},
		{name: "version with native", selection: Selection{Name: "gpt-5.6-sol", Version: "5.6", VersionPresent: true}, want: "already selects a version"},
		{name: "explicit blank effort", selection: Selection{Name: "fable", EffortPresent: true}, want: "unsupported model effort"},
		{name: "unknown effort", selection: Selection{Name: "fable", Effort: "max", EffortPresent: true}, want: "unsupported model effort"},
		{name: "fixed native effort conflict", selection: Selection{Name: "gemini-3.1-pro-low", Effort: "high", EffortPresent: true}, want: "fixes effort"},
		{name: "unknown provider", selection: Selection{Name: "mystery-model"}, want: "cannot determine provider"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ResolveModel(test.selection)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ResolveModel() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestAvailableReturnsCopy(t *testing.T) {
	models := Available()
	models[0].Model = "changed"
	resolved, _ := Resolve("flash")
	if resolved.Model != "gemini-3.8-flash-medium" {
		t.Fatalf("registry mutated through Available(): %#v", resolved)
	}
}
