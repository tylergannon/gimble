// Package program contains executable stubs for exploring Go workflow authoring.
// It does not call agents, execute commands, create worktrees, or integrate files.
package program

import (
	"context"
	"fmt"
	"io"
	"path"
	"sync"
)

type Role string

type Call struct {
	Name      string
	Role      Role
	Prompt    string
	Workspace string
}

type Check struct {
	Passed bool
	Output string
}

// Runtime supplies one synchronous effect at a time. Callbacks are demo fixtures
// and must be safe for concurrent calls; orchestration belongs to the caller.
type Runtime struct {
	Agent      func(context.Context, Call) (any, error)
	RunCommand func(ctx context.Context, workspace, command string) (Check, error)
	Output     io.Writer
	mu         sync.Mutex // Protects trace output only, never an agent invocation.
}

// Codergen stands in for the recovered POC's typed native-agent operation.
func Codergen[T any](ctx context.Context, rt *Runtime, name string, role Role, prompt string) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	if rt.Agent == nil {
		return zero, fmt.Errorf("%s: agent stub is missing", name)
	}
	rt.trace("agent %s role=%s workspace=%s", name, role, Workspace(ctx))
	value, err := rt.Agent(ctx, Call{name, role, prompt, Workspace(ctx)})
	if err != nil {
		return zero, err
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	result, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("%s: stub returned %T, want %T", name, value, zero)
	}
	return result, nil
}

// Command returns a scripted observation. A failed check is data; an inability
// to execute the check is an error. No shell command is actually executed.
func (rt *Runtime) Command(ctx context.Context, command string) (Check, error) {
	if err := ctx.Err(); err != nil {
		return Check{}, err
	}
	if rt.RunCommand == nil {
		return Check{}, fmt.Errorf("command stub is missing")
	}
	check, err := rt.RunCommand(ctx, Workspace(ctx), command)
	if err != nil {
		return Check{}, err
	}
	if err := ctx.Err(); err != nil {
		return Check{}, err
	}
	rt.trace("command %q workspace=%s passed=%t", command, Workspace(ctx), check.Passed)
	return check, nil
}

type workspaceKey struct{}

func Workspace(ctx context.Context) string {
	if workspace, ok := ctx.Value(workspaceKey{}).(string); ok {
		return workspace
	}
	return "main"
}

// Worktree gives a branch an inherited context with its own symbolic workspace.
// These names demonstrate ownership; there is no filesystem isolation here.
func (rt *Runtime) Worktree(ctx context.Context, baseline, name string) (context.Context, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	workspace := path.Join(baseline, name)
	rt.trace("worktree %s", workspace)
	return context.WithValue(ctx, workspaceKey{}, workspace), nil
}

func (rt *Runtime) Integrate(ctx context.Context, candidate string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	rt.trace("integrate %s into %s", candidate, Workspace(ctx))
	return nil
}

func (rt *Runtime) trace(format string, args ...any) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Output != nil {
		_, _ = fmt.Fprintf(rt.Output, "[stub] "+format+"\n", args...)
	}
}
