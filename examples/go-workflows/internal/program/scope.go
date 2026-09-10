package program

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Scope freezes the parent's effective values into a named child scope. Child
// writes, sibling writes, and later parent writes remain isolated. Immutable
// value files are shared; the child publishes its own index when next used.
// Files are immutable by API convention. Scope routes are not access controls:
// agent tools can read neighboring layers and otherwise use the filesystem.
//
// Return to the parent by using the original ctx again. There is no pop,
// automatic promotion, or cleanup: earlier snapshots must remain readable.
func Scope(ctx context.Context, name string) (context.Context, error) {
	parent, ok := ctx.Value(contextKey{}).(*contextStore)
	if !ok {
		return nil, errors.New("no context store is attached")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("scope name must not be blank")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parent.mu.Lock()
	defer parent.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Each child has its own physical layer inside its parent directory.
	// Its index refers back to inherited immutable files without copying them.
	dir, err := os.MkdirTemp(parent.dir, "scope-")
	if err != nil {
		return nil, err
	}
	path := append(slices.Clone(parent.scope), name)
	if len(contextRoute(filepath.Join(dir, "view-000000-0000000000", "index.json"), path...)) > parent.limits.PromptBytes {
		_ = os.Remove(dir)
		return nil, errors.New("scope prompt budget cannot fit its name and index entrypoint")
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(dir)
		return nil, err
	}
	child := &contextStore{
		dir: dir, limits: parent.limits, scope: path,
		values: maps.Clone(parent.values), revision: parent.revision,
	}
	return context.WithValue(ctx, contextKey{}, child), nil
}
