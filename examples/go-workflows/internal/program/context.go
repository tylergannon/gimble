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
	Prompt   string   `json:"prompt"`
	Index    string   `json:"index"`
	Inline   []string `json:"inline"`
	External []string `json:"external"`
}

type contextKey struct{}

type contextValue struct {
	raw  []byte
	path string
}

type contextStore struct {
	mu       sync.Mutex
	dir      string
	limits   ContextLimits
	values   map[string]contextValue
	revision int
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
	if len(contextRoute(filepath.Join(root, "index-000000-0000000000.json"))) > limits.PromptBytes {
		_ = os.Remove(root)
		return nil, errors.New("context prompt budget cannot fit its index entrypoint")
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(root)
		return nil, err
	}
	store := &contextStore{dir: root, limits: limits, values: make(map[string]contextValue)}
	return context.WithValue(ctx, contextKey{}, store), nil
}

// SetContext preserves the complete JSON value immediately. It invalidates the
// projection but does no indexing: ordinary Go can continue until an agent asks
// for a snapshot. Child contexts inherit this store; it is not a branch fork.
func SetContext(ctx context.Context, key string, value any) error {
	store, ok := ctx.Value(contextKey{}).(*contextStore)
	if !ok {
		return errors.New("no context store is attached")
	}
	if key == "" || len(key) > 128 || strings.Trim(key, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-") != "" {
		return errors.New("context key must use 1-128 ASCII letters, digits, underscores, dots, or hyphens")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode context %q: %w", key, err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := writeContextFile(store.dir, fmt.Sprintf("value-%06d-*.json", store.revision+1), raw)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		_ = os.Remove(file)
		return err
	}
	store.values[key] = contextValue{raw: raw, path: file}
	store.revision++
	store.snapshot = nil
	return nil
}

// SnapshotContext is the indexing barrier before an agent call. Pending setters
// are coalesced into one coherent revision. The deterministic JSON index is a
// routing placeholder, not semantic indexing or an LLM-generated description.
func SnapshotContext(ctx context.Context) (ContextSnapshot, error) {
	store, ok := ctx.Value(contextKey{}).(*contextStore)
	if !ok {
		return ContextSnapshot{}, nil
	}
	if err := ctx.Err(); err != nil {
		return ContextSnapshot{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return ContextSnapshot{}, err
	}
	if store.snapshot != nil {
		return cloneContextSnapshot(*store.snapshot), nil
	}
	type entry struct {
		Key         string `json:"key"`
		Path        string `json:"path"`
		Bytes       int    `json:"bytes"`
		Description string `json:"description"`
	}
	keys := make([]string, 0, len(store.values))
	for key := range store.values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	entries := make([]entry, 0, len(keys))
	for _, key := range keys {
		value := store.values[key]
		entries = append(entries, entry{key, value.path, len(value.raw), fmt.Sprintf("JSON value (%d bytes); placeholder description", len(value.raw))})
	}
	index, err := json.MarshalIndent(struct {
		Revision int     `json:"revision"`
		Kind     string  `json:"kind"`
		Entries  []entry `json:"entries"`
	}{store.revision, "deterministic routing placeholder", entries}, "", "  ")
	if err != nil {
		return ContextSnapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return ContextSnapshot{}, err
	}
	indexPath, err := writeContextFile(store.dir, fmt.Sprintf("index-%06d-*.json", store.revision), index)
	if err != nil {
		return ContextSnapshot{}, err
	}
	snapshot, err := store.project(keys, indexPath)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		_ = os.Remove(indexPath)
		return ContextSnapshot{}, err
	}
	store.snapshot = &snapshot
	return cloneContextSnapshot(snapshot), nil
}

func (store *contextStore) project(keys []string, index string) (ContextSnapshot, error) {
	root := contextRoute(index)
	if len(root) > store.limits.PromptBytes {
		return ContextSnapshot{}, errors.New("context prompt budget cannot fit its index entrypoint")
	}
	external := make(map[string]bool)
	var candidates []string
	for _, key := range keys {
		if len(store.values[key].raw) > store.limits.ValueBytes {
			external[key] = true
		} else {
			candidates = append(candidates, key)
		}
	}
	// Largest values leave first; alphabetical keys break ties deterministically.
	slices.SortFunc(candidates, func(a, b string) int {
		if difference := len(store.values[b].raw) - len(store.values[a].raw); difference != 0 {
			return difference
		}
		return strings.Compare(a, b)
	})
	render := func() string {
		var prompt strings.Builder
		prompt.WriteString(root)
		for _, key := range keys {
			value := store.values[key]
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
	snapshot := ContextSnapshot{Revision: store.revision, Prompt: prompt, Index: index, Inline: []string{}, External: []string{}}
	for _, key := range keys {
		if external[key] {
			snapshot.External = append(snapshot.External, key)
		} else {
			snapshot.Inline = append(snapshot.Inline, key)
		}
	}
	return snapshot, nil
}

func contextRoute(index string) string {
	return "Context JSON data. Read the index for all keys and exact value files:\n" + index + "\n"
}

func cloneContextSnapshot(snapshot ContextSnapshot) ContextSnapshot {
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
