package program

import (
	"context"
	"fmt"
)

type Role string
type Call struct {
	Name      string
	Role      Role
	Prompt    string
	Workspace string
	Context   ContextSnapshot
}
type Check struct {
	Passed bool
	Output string
}
type Runtime struct {
	Agent      func(context.Context, Call) (any, error)
	RunCommand func(context.Context, string, string) (Check, error)
}

func Codergen[T any](ctx context.Context, rt *Runtime, name string, role Role, prompt string) (T, error) {
	var z T
	if rt == nil || rt.Agent == nil {
		return z, fmt.Errorf("%s: agent stub is missing", name)
	}
	snap, _ := SnapshotContext(ctx)
	v, e := rt.Agent(ctx, Call{Name: name, Role: role, Prompt: prompt, Workspace: Workspace(ctx), Context: snap})
	if e != nil {
		return z, e
	}
	r, ok := v.(T)
	if !ok {
		return z, fmt.Errorf("%s: stub returned %T, want %T", name, v, z)
	}
	return r, nil
}
func (rt *Runtime) Command(ctx context.Context, c string) (Check, error) {
	if rt == nil || rt.RunCommand == nil {
		return Check{Passed: true, Output: "Canned check; no command executed."}, nil
	}
	return rt.RunCommand(ctx, Workspace(ctx), c)
}

type workspaceKey struct{}

func Workspace(ctx context.Context) string {
	if s, ok := ctx.Value(workspaceKey{}).(string); ok {
		return s
	}
	return "main"
}

// Worktree is symbolic; Integrate is a no-op.
func (rt *Runtime) Worktree(ctx context.Context, b, n string) (context.Context, error) {
	return context.WithValue(ctx, workspaceKey{}, b+"/"+n), nil
}
func (rt *Runtime) Integrate(context.Context, string) error { return nil }
