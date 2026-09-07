// Package editor is the file side of the graph editor: it reads, lints,
// writes, and watches one pipeline file and its layout sidecar. The page that
// draws the graph is a SvelteKit app served by skgo; the remote functions it
// calls live beside its routes in web/editor/src/routes and are thin wrappers
// over the Store here.
package editor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tylergannon/tractor/engine"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/lint"
)

const defaultPollInterval = 500 * time.Millisecond

// Document is what the page reads: the pipeline file, its layout sidecar, a
// version that changes whenever either file does, and what lint makes of it.
type Document struct {
	// Path is the absolute path of the pipeline file.
	Path string `json:"path"`
	// YAML is the file's text, exactly as it is on disk.
	YAML string `json:"yaml"`
	// Layout is every node placement the sidecar holds.
	Layout []Placement `json:"layout"`
	// HasLayout reports whether a layout sidecar exists beside the pipeline.
	HasLayout bool `json:"has_layout"`
	// Version identifies this state of the two files. A save must quote it.
	Version string `json:"version"`
	// Diagnostics are the lint findings, the same ones `tractor validate` prints.
	Diagnostics []Diagnostic `json:"diagnostics"`
	// ParseError is set when the file does not parse; Diagnostics is empty then.
	ParseError string `json:"parse_error"`
	// Models says which model each node resolved to.
	Models []ModelResolution `json:"models"`
}

// Placement is where one node sits on the canvas.
type Placement struct {
	// ID is the node's id in the pipeline.
	ID string `json:"id"`
	// X is the left edge, in canvas units.
	X float64 `json:"x"`
	// Y is the top edge, in canvas units.
	Y float64 `json:"y"`
	// W is the width; zero means the page decides.
	W float64 `json:"w"`
	// H is the height; zero means the page decides.
	H float64 `json:"h"`
}

// Diagnostic is one lint finding.
type Diagnostic struct {
	// Rule names the lint rule that fired.
	Rule string `json:"rule"`
	// Severity is error, warning, or info.
	Severity string `json:"severity"`
	// Message says what is wrong.
	Message string `json:"message"`
	// NodeID is the node the finding belongs to, or empty.
	NodeID string `json:"node_id"`
	// Edge is the authored edge the finding belongs to, as from and to, or empty.
	Edge []string `json:"edge"`
	// Fix suggests what to change, or is empty.
	Fix string `json:"fix"`
}

// ModelResolution says which model a node resolved to and why.
type ModelResolution struct {
	NodeID          string `json:"node_id"`
	Role            string `json:"role"`
	Source          string `json:"source"`
	AuthoredName    string `json:"authored_name"`
	AuthoredVersion string `json:"authored_version"`
	NativeModel     string `json:"native_model"`
	EffectiveEffort string `json:"effective_effort"`
	Provider        string `json:"provider"`
	Harness         string `json:"harness"`
}

// SaveRequest is what the page sends to write the file.
type SaveRequest struct {
	// YAML is the whole file to write.
	YAML string `json:"yaml"`
	// Layout is the placements to write to the sidecar when WriteLayout is set.
	Layout []Placement `json:"layout"`
	// WriteLayout says whether to write the sidecar at all. A page that has
	// not moved a node leaves a pipeline without one alone.
	WriteLayout bool `json:"write_layout"`
	// Version is the Document.Version the page last saw. A save against any
	// other version is refused and the current document is returned instead.
	Version string `json:"version"`
}

// SaveResult is the answer to a save.
type SaveResult struct {
	// Saved is false when the file changed on disk since the page last read
	// it; nothing was written and Document is what is on disk now.
	Saved bool `json:"saved"`
	// Document is the state after the save, or the state that refused it.
	Document Document `json:"document"`
}

// Change announces a new version of the files on disk.
type Change struct {
	// Version is Document.Version for the state now on disk.
	Version string `json:"version"`
}

// SidecarPath returns the layout sidecar beside a pipeline file:
// examples/x.yaml -> examples/x.layout.json.
func SidecarPath(pipeline string) string {
	return strings.TrimSuffix(pipeline, filepath.Ext(pipeline)) + ".layout.json"
}

// Store reads, lints, writes, and watches one pipeline file.
type Store struct {
	path         string
	sidecar      string
	validator    *lint.Validator
	pollInterval time.Duration

	mu sync.Mutex // serializes disk reads and writes of the pipeline and sidecar
}

// Open builds a store for the pipeline at path. validator may be nil, in which
// case only parse errors are reported.
func Open(pipeline string, validator *lint.Validator) (*Store, error) {
	absolute, err := filepath.Abs(pipeline)
	if err != nil {
		return nil, fmt.Errorf("resolve pipeline path: %w", err)
	}
	return &Store{
		path:         absolute,
		sidecar:      SidecarPath(absolute),
		validator:    validator,
		pollInterval: defaultPollInterval,
	}, nil
}

// Path returns the absolute pipeline path the store edits.
func (s *Store) Path() string { return s.path }

// Load reads the pipeline and sidecar, lints, and assembles the document.
func (s *Store) Load() (Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

// Save writes the pipeline, and the sidecar when asked, provided the files
// still have the version the request quotes.
func (s *Store) Save(request SaveRequest) (SaveResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.loadLocked()
	if err != nil {
		return SaveResult{}, err
	}
	if request.Version != current.Version {
		return SaveResult{Saved: false, Document: current}, nil
	}
	if err := writeFileAtomic(s.path, []byte(request.YAML)); err != nil {
		return SaveResult{}, err
	}
	if request.WriteLayout {
		encoded, err := encodeSidecar(request.Layout)
		if err != nil {
			return SaveResult{}, err
		}
		if err := writeFileAtomic(s.sidecar, encoded); err != nil {
			return SaveResult{}, err
		}
	}
	updated, err := s.loadLocked()
	if err != nil {
		return SaveResult{}, err
	}
	return SaveResult{Saved: true, Document: updated}, nil
}

// Watch yields the current version at once, then every change to either file
// until ctx ends or yield returns an error.
func (s *Store) Watch(ctx context.Context, yield func(Change) error) error {
	last := s.Version()
	if err := yield(Change{Version: last}); err != nil {
		return err
	}
	poll := time.NewTicker(s.pollInterval)
	defer poll.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-poll.C:
			version := s.Version()
			if version == last {
				continue
			}
			last = version
			if err := yield(Change{Version: version}); err != nil {
				return err
			}
		}
	}
}

// Version hashes the on-disk content without lint; missing files hash as empty.
func (s *Store) Version() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, _ := readOptional(s.path)
	sidecar, _ := readOptional(s.sidecar)
	return hashVersion(raw, sidecar)
}

func (s *Store) loadLocked() (Document, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return Document{}, fmt.Errorf("read pipeline: %w", err)
	}
	sidecar, err := readOptional(s.sidecar)
	if err != nil {
		return Document{}, fmt.Errorf("read layout: %w", err)
	}
	document := Document{
		Path:        s.path,
		YAML:        string(raw),
		Layout:      []Placement{},
		Version:     hashVersion(raw, sidecar),
		Diagnostics: []Diagnostic{},
		Models:      []ModelResolution{},
	}
	if sidecar != nil {
		document.HasLayout = true
		document.Layout = decodeSidecar(sidecar)
	}
	document.Diagnostics, document.ParseError, document.Models = s.inspect(raw)
	return document, nil
}

func (s *Store) inspect(raw []byte) ([]Diagnostic, string, []ModelResolution) {
	var (
		pipeline *graph.Graph
		err      error
	)
	if strings.EqualFold(filepath.Ext(s.path), ".json") {
		pipeline, err = graph.Parse(raw)
	} else {
		pipeline, err = graph.ParseYAML(raw)
	}
	if err != nil {
		return []Diagnostic{}, err.Error(), []ModelResolution{}
	}
	resolved, modelErr := engine.ResolveGraphModels(*pipeline, engine.SystemModelSelection{})
	if modelErr != nil {
		return []Diagnostic{}, modelErr.Error(), []ModelResolution{}
	}
	models := make([]ModelResolution, 0, len(resolved))
	for _, m := range resolved {
		models = append(models, ModelResolution(m))
	}
	if s.validator == nil {
		return []Diagnostic{}, "", models
	}
	found := s.validator.Validate(*pipeline)
	diagnostics := make([]Diagnostic, 0, len(found))
	for _, d := range found {
		diagnostics = append(diagnostics, wireDiagnostic(d))
	}
	return diagnostics, "", models
}

func wireDiagnostic(d lint.Diagnostic) Diagnostic {
	out := Diagnostic{
		Rule:     d.Rule,
		Severity: string(d.Severity),
		Message:  d.Message,
		NodeID:   d.NodeID,
		Edge:     []string{},
		Fix:      d.Fix,
	}
	if d.Edge != nil {
		out.Edge = []string{d.Edge[0], d.Edge[1]}
	}
	return out
}

// sidecarEntry is one node's record in <stem>.layout.json, which is keyed by
// node id so a person or an agent can edit it by hand.
type sidecarEntry struct {
	X float64  `json:"x"`
	Y float64  `json:"y"`
	W *float64 `json:"w,omitempty"`
	H *float64 `json:"h,omitempty"`
}

// decodeSidecar reads placements out of the sidecar; a sidecar that is not a
// JSON object yields none, as if it were absent.
func decodeSidecar(raw []byte) []Placement {
	var entries map[string]sidecarEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return []Placement{}
	}
	ids := make([]string, 0, len(entries))
	for id := range entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Placement, 0, len(entries))
	for _, id := range ids {
		e := entries[id]
		p := Placement{ID: id, X: e.X, Y: e.Y}
		if e.W != nil {
			p.W = *e.W
		}
		if e.H != nil {
			p.H = *e.H
		}
		out = append(out, p)
	}
	return out
}

func encodeSidecar(layout []Placement) ([]byte, error) {
	entries := make(map[string]sidecarEntry, len(layout))
	for _, p := range layout {
		if p.ID == "" {
			return nil, errors.New("layout: a placement needs a node id")
		}
		e := sidecarEntry{X: p.X, Y: p.Y}
		if p.W > 0 {
			e.W = &p.W
		}
		if p.H > 0 {
			e.H = &p.H
		}
		entries[p.ID] = e
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		return nil, fmt.Errorf("layout: %w", err)
	}
	return buf.Bytes(), nil
}

func hashVersion(yaml, sidecar []byte) string {
	digest := sha256.New()
	_, _ = fmt.Fprintf(digest, "%d:", len(yaml))
	digest.Write(yaml)
	_, _ = fmt.Fprintf(digest, "%d:", len(sidecar))
	digest.Write(sidecar)
	return hex.EncodeToString(digest.Sum(nil))
}

func readOptional(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return data, err
}

// writeFileAtomic writes data to a temp file in the same directory, then
// renames it over name, preserving the existing file mode.
func writeFileAtomic(name string, data []byte) error {
	mode := fs.FileMode(0o644)
	if info, err := os.Stat(name); err == nil {
		mode = info.Mode().Perm()
	}
	temp, err := os.CreateTemp(filepath.Dir(name), "."+filepath.Base(name)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempName := temp.Name()
	cleanup := func(err error) error {
		_ = temp.Close()
		_ = os.Remove(tempName)
		return err
	}
	if _, err := temp.Write(data); err != nil {
		return cleanup(fmt.Errorf("write %s: %w", name, err))
	}
	if err := temp.Sync(); err != nil {
		return cleanup(fmt.Errorf("sync %s: %w", name, err))
	}
	if err := temp.Chmod(mode); err != nil {
		return cleanup(fmt.Errorf("chmod %s: %w", name, err))
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempName)
		return fmt.Errorf("close %s: %w", name, err)
	}
	if err := os.Rename(tempName, name); err != nil {
		_ = os.Remove(tempName)
		return fmt.Errorf("rename into %s: %w", name, err)
	}
	return nil
}
