package program

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// ContextLimits uses encoded JSON bytes as a prototype proxy for token budgets.
type ContextLimits struct {
	ValueBytes  int
	PromptBytes int
}

type ContextSnapshot struct {
	Revision int      `json:"revision"`
	Scope    []string `json:"scope,omitempty"`
	Prompt   string   `json:"prompt"`
	Index    string   `json:"index"`
	View     string   `json:"view"`
	Inline   []string `json:"inline"`
	External []string `json:"external"`
}

type contextKey struct{}

type contextValue struct {
	raw  []byte
	path string
}

// Key identifies a declared value and the one scope allowed to edit it.
// Its display name does not identify a value in any other scope.
type Key struct {
	owner *contextStore
	name  string
}

type contextTree struct {
	mu       sync.Mutex
	revision int
	nextID   int
}

type contextStore struct {
	tree     *contextTree
	parent   *contextStore
	id       int
	dir      string
	limits   ContextLimits
	keys     map[string]Key
	values   map[string]contextValue
	scope    []string
	snapshot *ContextSnapshot
}

// NewContext attaches a filesystem-backed context to ctx. Each attachment owns
// a fresh subdirectory; old snapshots and their value files are never replaced.
func NewContext(ctx context.Context, dir string, limits ContextLimits) (context.Context, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(dir) == "" || limits.ValueBytes <= 0 || limits.PromptBytes <= 0 {
		return nil, errors.New("context directory and positive byte limits are required")
	}
	parent, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, err
	}
	root, err := os.MkdirTemp(parent, "context-")
	if err != nil {
		return nil, err
	}
	if len(contextRoute(filepath.Join(root, "view-000000-0000000000", "index.json"))) > limits.PromptBytes {
		_ = os.Remove(root)
		return nil, errors.New("context prompt budget cannot fit its index entrypoint")
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(root)
		return nil, err
	}
	store := &contextStore{
		dir: root, limits: limits, values: make(map[string]contextValue),
		keys: make(map[string]Key), tree: &contextTree{nextID: 1},
	}
	return context.WithValue(ctx, contextKey{}, store), nil
}

// DeclareContext binds a name to the current scope before it has a value.
// Repeating a declaration in that scope returns the same Key. A declaration in
// another scope creates a distinct binding, even if its display name is equal.
func DeclareContext(ctx context.Context, name string) (Key, error) {
	store, ok := ctx.Value(contextKey{}).(*contextStore)
	if !ok {
		return Key{}, errors.New("no context store is attached")
	}
	if name == "" || len(name) > 128 || strings.Trim(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-") != "" {
		return Key{}, errors.New("context key must use 1-128 ASCII letters, digits, underscores, dots, or hyphens")
	}
	if err := ctx.Err(); err != nil {
		return Key{}, err
	}
	store.tree.mu.Lock()
	defer store.tree.mu.Unlock()
	normalized := strings.ToLower(name)
	if key, exists := store.keys[normalized]; exists {
		if key.name != name {
			return Key{}, fmt.Errorf("context key %q conflicts with existing key %q on case-insensitive filesystems", name, key.name)
		}
		return key, nil
	}
	key := Key{owner: store, name: name}
	store.keys[normalized] = key
	return key, nil
}

// SetContext preserves the complete JSON value immediately. Only the scope
// that declared key may set it. Ordinary derived Go contexts share that scope.
// Indexing is deferred until SnapshotContext, which also reads current parents.
func SetContext(ctx context.Context, key Key, value any) error {
	store, ok := ctx.Value(contextKey{}).(*contextStore)
	if !ok {
		return errors.New("no context store is attached")
	}
	if key.owner == nil {
		return errors.New("context key has not been declared")
	}
	if key.owner != store {
		return fmt.Errorf("context key %q cannot be written in scope %q: owned by scope %q", key.name, store.scope, key.owner.scope)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode context %q: %w", key.name, err)
	}
	store.tree.mu.Lock()
	defer store.tree.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := writeContextFile(store.dir, fmt.Sprintf("value-%06d-*.json", store.tree.revision+1), raw)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(file)
		return err
	}
	store.values[key.name] = contextValue{raw: raw, path: file}
	store.tree.revision++
	return nil
}

// effectiveValues resolves the current ancestor chain under the tree lock.
// Qualifying every name on any collision avoids hiding either binding and
// prevents a declared name from colliding with a generated qualified name.
func (store *contextStore) effectiveValues() map[string]contextValue {
	names := make(map[string]bool)
	qualify := false
	for scope := store; scope != nil; scope = scope.parent {
		for name := range scope.values {
			normalized := strings.ToLower(name)
			qualify = qualify || names[normalized]
			names[normalized] = true
		}
	}
	values := make(map[string]contextValue)
	for scope := store; scope != nil; scope = scope.parent {
		for name, value := range scope.values {
			if qualify {
				name = fmt.Sprintf("scope-%d.%s", scope.id, name)
			}
			values[name] = value
		}
	}
	return values
}

// SnapshotContext is the indexing barrier before an agent call. Pending setters
// and current ancestor values are coalesced into one coherent revision. Earlier
// snapshots stay fixed. The deterministic JSON index is a
// routing placeholder, not semantic indexing or an LLM-generated description.
func SnapshotContext(ctx context.Context) (ContextSnapshot, error) {
	store, ok := ctx.Value(contextKey{}).(*contextStore)
	if !ok {
		return ContextSnapshot{}, nil
	}
	if err := ctx.Err(); err != nil {
		return ContextSnapshot{}, err
	}
	store.tree.mu.Lock()
	defer store.tree.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return ContextSnapshot{}, err
	}
	if store.snapshot != nil && store.snapshot.Revision == store.tree.revision {
		return cloneContextSnapshot(*store.snapshot), nil
	}
	type entry struct {
		Key         string `json:"key"`
		Path        string `json:"path"`
		Bytes       int    `json:"bytes"`
		Description string `json:"description"`
	}
	values := store.effectiveValues()
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	entries := make([]entry, 0, len(keys))
	for _, key := range keys {
		value := values[key]
		entries = append(entries, entry{key, value.path, len(value.raw), fmt.Sprintf("JSON value (%d bytes); placeholder description", len(value.raw))})
	}
	index, err := json.MarshalIndent(struct {
		Revision int      `json:"revision"`
		Scope    []string `json:"scope,omitempty"`
		Kind     string   `json:"kind"`
		Entries  []entry  `json:"entries"`
	}{store.tree.revision, store.scope, "deterministic routing placeholder", entries}, "", "  ")
	if err != nil {
		return ContextSnapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return ContextSnapshot{}, err
	}
	indexPath, err := writeContextFile(store.dir, fmt.Sprintf("index-%06d-*.json", store.tree.revision), index)
	if err != nil {
		return ContextSnapshot{}, err
	}
	view, err := store.materializeView(ctx, indexPath, keys, values)
	if err != nil {
		_ = os.Remove(indexPath)
		return ContextSnapshot{}, err
	}
	snapshot, err := store.project(keys, values, indexPath, view)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		_ = os.RemoveAll(view)
		_ = os.Remove(indexPath)
		return ContextSnapshot{}, err
	}
	store.snapshot = &snapshot
	return cloneContextSnapshot(snapshot), nil
}

func (store *contextStore) project(keys []string, values map[string]contextValue, index, view string) (ContextSnapshot, error) {
	root := contextRoute(filepath.Join(view, "index.json"), store.scope...)
	if len(root) > store.limits.PromptBytes {
		return ContextSnapshot{}, errors.New("context prompt budget cannot fit its index entrypoint")
	}
	external := make(map[string]bool)
	var candidates []string
	for _, key := range keys {
		if len(values[key].raw) > store.limits.ValueBytes {
			external[key] = true
		} else {
			candidates = append(candidates, key)
		}
	}
	// Largest values leave first; alphabetical keys break ties deterministically.
	slices.SortFunc(candidates, func(a, b string) int {
		if difference := len(values[b].raw) - len(values[a].raw); difference != 0 {
			return difference
		}
		return strings.Compare(a, b)
	})
	render := func() string {
		var prompt strings.Builder
		prompt.WriteString(root)
		for _, key := range keys {
			value := values[key]
			if external[key] {
				fmt.Fprintf(&prompt, "%s: see index (%d JSON bytes)\n", key, len(value.raw))
			} else {
				fmt.Fprintf(&prompt, "%s = %s\n", key, value.raw)
			}
		}
		return prompt.String()
	}
	prompt := render()
	for _, key := range candidates {
		if len(prompt) <= store.limits.PromptBytes {
			break
		}
		external[key] = true
		prompt = render()
	}
	if len(prompt) > store.limits.PromptBytes {
		// Routing entries themselves can exceed the budget. Their complete set
		// remains in the versioned index; one entrypoint is sufficient here.
		prompt = root
	}
	snapshot := ContextSnapshot{Revision: store.tree.revision, Scope: slices.Clone(store.scope), Prompt: prompt, Index: index, View: view, Inline: []string{}, External: []string{}}
	for _, key := range keys {
		if external[key] {
			snapshot.External = append(snapshot.External, key)
		} else {
			snapshot.Inline = append(snapshot.Inline, key)
		}
	}
	return snapshot, nil
}

func contextRoute(index string, scope ...string) string {
	root := "Context JSON data. Read this index; sibling values/ contains every key:\n" + index + "\n"
	if len(scope) == 0 {
		return root
	}
	// JSON quoting keeps arbitrary scope names from introducing prompt lines.
	name, _ := json.Marshal(scope)
	return "Scope: " + string(name) + "\n" + root
}

func cloneContextSnapshot(snapshot ContextSnapshot) ContextSnapshot {
	snapshot.Scope = slices.Clone(snapshot.Scope)
	snapshot.Inline = slices.Clone(snapshot.Inline)
	snapshot.External = slices.Clone(snapshot.External)
	return snapshot
}

func writeContextFile(dir, pattern string, raw []byte) (string, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	_, writeErr := file.Write(raw)
	err = errors.Join(writeErr, file.Close())
	if err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}
