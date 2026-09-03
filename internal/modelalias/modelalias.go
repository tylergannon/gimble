// Package modelalias resolves Tractor-owned model aliases to provider-native
// model IDs. Provider-native IDs remain valid without registration here.
package modelalias

import "fmt"

// Model is one supported Tractor model alias.
type Model struct {
	Alias    string
	Provider string
	Model    string
}

var available = []Model{
	{Alias: "flash", Provider: "gemini", Model: "gemini-3.8-flash-medium"},
	{Alias: "fable", Provider: "anthropic", Model: "claude-fable-5-1"},
	{Alias: "fable-5.1", Provider: "anthropic", Model: "claude-fable-5-1"},
	{Alias: "fable-5", Provider: "anthropic", Model: "claude-fable-5"},
}

// Available returns the maintained aliases in display order.
func Available() []Model {
	return append([]Model(nil), available...)
}

// Resolve returns the provider and concrete model for a known alias. Unknown
// values are left to callers because they may be provider-native model IDs.
func Resolve(alias string) (Model, bool) {
	for _, candidate := range available {
		if candidate.Alias == alias {
			return candidate, true
		}
	}
	return Model{}, false
}

// ResolveSelection applies a maintained alias and rejects an explicitly
// conflicting provider. Unknown model names pass through unchanged.
func ResolveSelection(provider, model string) (string, string, error) {
	resolved, known := Resolve(model)
	if !known {
		return provider, model, nil
	}
	if provider != "" && provider != resolved.Provider {
		return "", "", fmt.Errorf("provider %q conflicts with model alias %q (provider %q)", provider, model, resolved.Provider)
	}
	return resolved.Provider, resolved.Model, nil
}
