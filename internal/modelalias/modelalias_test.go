package modelalias

import (
	"reflect"
	"strings"
	"testing"
)

func TestAvailableModels(t *testing.T) {
	want := []Model{
		{Alias: "flash", Provider: "gemini", Model: "gemini-3.8-flash-medium"},
		{Alias: "fable", Provider: "anthropic", Model: "claude-fable-5-1"},
		{Alias: "fable-5.1", Provider: "anthropic", Model: "claude-fable-5-1"},
		{Alias: "fable-5", Provider: "anthropic", Model: "claude-fable-5"},
	}
	if got := Available(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Available() = %#v, want %#v", got, want)
	}
}

func TestResolveSelection(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		model        string
		wantProvider string
		wantModel    string
		wantErr      string
	}{
		{name: "Flash resolves to Gemini 3.8 medium", model: "flash", wantProvider: "gemini", wantModel: "gemini-3.8-flash-medium"},
		{name: "explicit Gemini provider accepts Flash", provider: "gemini", model: "flash", wantProvider: "gemini", wantModel: "gemini-3.8-flash-medium"},
		{name: "unversioned Fable defaults to 5.1", model: "fable", wantProvider: "anthropic", wantModel: "claude-fable-5-1"},
		{name: "explicit Fable 5.1", provider: "anthropic", model: "fable-5.1", wantProvider: "anthropic", wantModel: "claude-fable-5-1"},
		{name: "explicit Fable 5 remains available", model: "fable-5", wantProvider: "anthropic", wantModel: "claude-fable-5"},
		{name: "raw model passes through", provider: "anthropic", model: "claude-experimental", wantProvider: "anthropic", wantModel: "claude-experimental"},
		{name: "conflicting provider", provider: "openai", model: "flash", wantErr: `provider "openai" conflicts with model alias "flash"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider, model, err := ResolveSelection(test.provider, test.model)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("ResolveSelection() error = %v, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil || provider != test.wantProvider || model != test.wantModel {
				t.Fatalf("ResolveSelection() = %q, %q, %v; want %q, %q", provider, model, err, test.wantProvider, test.wantModel)
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
