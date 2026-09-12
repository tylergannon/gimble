package observation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Registry is a project's live runs. The web runtime puts one in the
// context it serves from, and gimble.Run finds it there: there is no global
// map, and a run started without the web runtime simply does not find one.
type Registry struct {
	dir string

	mu   sync.RWMutex
	runs map[string]*Store
}

// NewRegistry returns the registry for a project directory. A finished run
// is served from the checkpoint in its run directory below it.
func NewRegistry(projectDir string) *Registry {
	return &Registry{dir: projectDir, runs: map[string]*Store{}}
}

type registryKey struct{}

// WithRegistry puts the registry in ctx.
func WithRegistry(ctx context.Context, registry *Registry) context.Context {
	return context.WithValue(ctx, registryKey{}, registry)
}

// FromContext returns the registry in ctx, or nil.
func FromContext(ctx context.Context) *Registry {
	if ctx == nil {
		return nil
	}
	registry, _ := ctx.Value(registryKey{}).(*Registry)
	return registry
}

func (r *Registry) add(s *Store) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[s.id] = s
}

// finish removes a run after its final observation has reached disk.
func (r *Registry) finish(id string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runs, id)
}

// Live returns the store of a run that is still going.
func (r *Registry) Live(id string) (*Store, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	store, ok := r.runs[id]
	return store, ok
}

// ErrNoRun is returned for a run id this project has never had.
var ErrNoRun = errors.New("observation: no such run")

// Snapshot returns a run's complete observation: from its live store while
// it is going, and from its checkpoint once it has finished. Nothing here
// replays a log, and nothing resumes an agent process.
func (r *Registry) Snapshot(id string) (RunSnapshot, error) {
	if r == nil {
		return RunSnapshot{}, ErrNoRun
	}
	if store, ok := r.Live(id); ok {
		return store.Snapshot(), nil
	}
	dir, err := r.runDir(id)
	if err != nil {
		return RunSnapshot{}, err
	}
	checkpoint, err := loadCheckpoint(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return RunSnapshot{}, ErrNoRun
		}
		return RunSnapshot{}, err
	}
	return checkpoint, nil
}

// runDir is the run's directory below the project. A run id that could name
// anything outside it is refused rather than cleaned into something else.
func (r *Registry) runDir(id string) (string, error) {
	if r.dir == "" || id == "" || id == "." || id == ".." ||
		strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return "", ErrNoRun
	}
	return filepath.Join(r.dir, "runs", id), nil
}
