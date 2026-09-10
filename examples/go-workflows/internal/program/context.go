package program

import "context"

type ContextLimits struct {
	ValueBytes  int
	PromptBytes int
}
type ContextSnapshot struct {
	Revision int
	Scope    []string
	Prompt   string
	Index    string
	View     string
	Inline   []string
	External []string
}
type contextKey struct{}
type Key struct{ name string }

// Context is a no-op stub; intended ownership is per scope with per-call ancestor inheritance.
func NewContext(ctx context.Context, _ string, _ ContextLimits) (context.Context, error) {
	return context.WithValue(ctx, contextKey{}, ContextSnapshot{}), nil
}
func DeclareContext(ctx context.Context, n string) (Key, error)    { return Key{name: n}, nil }
func SetContext(context.Context, Key, any) error                   { return nil }
func SnapshotContext(ctx context.Context) (ContextSnapshot, error) { return ContextSnapshot{}, nil }
