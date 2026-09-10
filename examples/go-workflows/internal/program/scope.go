package program

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Scope creates a child with its own declared values. It retains its parent;
// each SnapshotContext sees current ancestor values, while earlier snapshots
// remain immutable. A child cannot edit a Key declared by its parent.
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
	parent.tree.mu.Lock()
	defer parent.tree.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Each child has its own physical layer inside its parent directory.
	// Its index refers back to inherited immutable files without copying them.
	dir, err := os.MkdirTemp(parent.dir, scopePrefix(name))
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
		tree: parent.tree, parent: parent, id: parent.tree.nextID,
		keys: make(map[string]Key), values: make(map[string]contextValue),
	}
	parent.tree.nextID++
	return context.WithValue(ctx, contextKey{}, child), nil
}

func scopePrefix(name string) string {
	var prefix strings.Builder
	for _, char := range strings.ToLower(name) {
		if prefix.Len() == 32 {
			break
		}
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '_' {
			prefix.WriteRune(char)
		} else {
			prefix.WriteByte('-')
		}
	}
	result := strings.Trim(prefix.String(), "-")
	if result == "" {
		result = "scope"
	}
	return result + "-"
}
