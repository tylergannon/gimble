package editor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tylergannon/gimble/lint"
)

const unreachableYAML = `name: t
start: a
nodes:
  - id: a
    type: command
    command: "true"
    edges:
      success: success
  - id: b
    type: command
    command: "true"
    edges:
      success: success
`

func newTestStore(t *testing.T, source string) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	pipeline := filepath.Join(dir, filepath.Base(source))
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pipeline, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := Open(pipeline, lint.New(lint.Options{}))
	if err != nil {
		t.Fatal(err)
	}
	store.pollInterval = 10 * time.Millisecond
	return store, pipeline
}

func load(t *testing.T, store *Store) Document {
	t.Helper()
	document, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestLoadReturnsFileAndDiagnostics(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	source, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	document := load(t, store)
	if document.Path != pipeline || document.YAML != string(source) {
		t.Fatalf("document = %+v", document)
	}
	if document.ParseError != "" {
		t.Fatalf("parse error = %q", document.ParseError)
	}
	if document.Diagnostics == nil || document.Layout == nil || document.Models == nil {
		t.Fatalf("nil slices cross the wire as null: %+v", document)
	}
	if document.HasLayout {
		t.Fatal("a pipeline without a sidecar reports a layout")
	}
	if len(document.Models) == 0 {
		t.Fatal("no model resolutions")
	}
	if document.Version == "" || document.Version != store.Version() {
		t.Fatalf("version = %q, store says %q", document.Version, store.Version())
	}
}

func TestLoadReportsLintDiagnosticsAndParseErrors(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	if err := os.WriteFile(pipeline, []byte(unreachableYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	document := load(t, store)
	if document.ParseError != "" {
		t.Fatalf("parse error = %q", document.ParseError)
	}
	var found bool
	for _, d := range document.Diagnostics {
		if d.NodeID == "b" {
			found = true
			if d.Edge == nil {
				t.Fatalf("edge is nil, which crosses the wire as null: %+v", d)
			}
		}
	}
	if !found {
		t.Fatalf("no diagnostic names the unreachable node: %+v", document.Diagnostics)
	}

	if err := os.WriteFile(pipeline, []byte("name: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	document = load(t, store)
	if document.ParseError == "" {
		t.Fatal("a file that does not parse reports no parse error")
	}
	if len(document.Diagnostics) != 0 {
		t.Fatalf("diagnostics on an unparsable file: %+v", document.Diagnostics)
	}
}

func TestSaveWithStaleVersionWritesNothing(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	before, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.Save(SaveRequest{YAML: "name: stale\n", Version: "not-the-version"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Saved {
		t.Fatal("a stale save was accepted")
	}
	if result.Document.YAML != string(before) {
		t.Fatal("the refusal did not carry what is on disk")
	}
	after, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("a stale save changed the file")
	}
}

func TestSaveWritesFileAndSidecar(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	current := load(t, store)
	updated := "name: moved\nstart: a\nnodes:\n  - id: a\n    type: command\n    command: \"true\"\n    edges:\n      success: success\n"
	result, err := store.Save(SaveRequest{
		YAML:        updated,
		Layout:      []Placement{{ID: "b", X: 10, Y: 20, W: 240, H: 120}, {ID: "a", X: 1, Y: 2}},
		WriteLayout: true,
		Version:     current.Version,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Saved {
		t.Fatal("save refused")
	}
	onDisk, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != updated {
		t.Fatalf("pipeline on disk:\n%s", onDisk)
	}

	// The sidecar is keyed by node id, so a person can edit it by hand, and a
	// placement without a size leaves the size out.
	raw, err := os.ReadFile(SidecarPath(pipeline))
	if err != nil {
		t.Fatal(err)
	}
	var sidecar map[string]map[string]float64
	if err := json.Unmarshal(raw, &sidecar); err != nil {
		t.Fatalf("sidecar: %v\n%s", err, raw)
	}
	if sidecar["b"]["w"] != 240 || sidecar["b"]["h"] != 120 || sidecar["a"]["x"] != 1 {
		t.Fatalf("sidecar = %s", raw)
	}
	if _, has := sidecar["a"]["w"]; has {
		t.Fatalf("an unsized placement wrote a width: %s", raw)
	}

	if !result.Document.HasLayout || result.Document.Version == current.Version {
		t.Fatalf("document after save = %+v", result.Document)
	}
	if len(result.Document.Layout) != 2 || result.Document.Layout[0].ID != "a" || result.Document.Layout[1].W != 240 {
		t.Fatalf("layout after save = %+v", result.Document.Layout)
	}
	if result.Document.Layout[0].W != 0 {
		t.Fatalf("an unsized placement reads back a width: %+v", result.Document.Layout[0])
	}

	// A save that does not touch the layout leaves the sidecar alone.
	result, err = store.Save(SaveRequest{YAML: updated + "# touched\n", Version: result.Document.Version})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Saved {
		t.Fatal("second save refused")
	}
	again, err := os.ReadFile(SidecarPath(pipeline))
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(raw) {
		t.Fatal("a save without a layout rewrote the sidecar")
	}
}

func TestSavePreservesModelVersionAndReportsEffectiveSelection(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	current := load(t, store)
	updated := `name: versioned
defaults:
  model:
    name: fable
    version: "5.1"
    effort: high
start: implement
nodes:
  - id: implement
    type: agent
    prompt: Implement the request.
    edges:
      - to: success
`
	result, err := store.Save(SaveRequest{YAML: updated, Version: current.Version})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Saved {
		t.Fatal("save refused")
	}
	onDisk, err := os.ReadFile(pipeline)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != updated {
		t.Fatalf("saved workflow changed:\n%s", onDisk)
	}
	if len(result.Document.Models) != 1 {
		t.Fatalf("models = %+v", result.Document.Models)
	}
	model := result.Document.Models[0]
	if model.NodeID != "implement" || model.Role != "agent" || model.AuthoredName != "fable" || model.AuthoredVersion != "5.1" || model.NativeModel != "claude-fable-5-1" || model.EffectiveEffort != "high" || model.Source != "pipeline defaults.model" {
		t.Fatalf("model = %+v", model)
	}
}

func TestWatchAnnouncesTheCurrentVersionThenEveryChange(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	changes := make(chan Change, 8)
	done := make(chan error, 1)
	go func() {
		done <- store.Watch(ctx, func(c Change) error {
			changes <- c
			return nil
		})
	}()

	first := <-changes
	if first.Version != store.Version() {
		t.Fatalf("first announcement = %q, want the version on disk %q", first.Version, store.Version())
	}

	if err := os.WriteFile(pipeline, []byte(unreachableYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case change := <-changes:
		if change.Version == first.Version || change.Version != store.Version() {
			t.Fatalf("announced %q after a write; disk is %q", change.Version, store.Version())
		}
	case <-ctx.Done():
		t.Fatal("no change announced after the file was written")
	}

	// The sidecar is part of the version too.
	if err := os.WriteFile(SidecarPath(pipeline), []byte(`{"a":{"x":1,"y":2}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case change := <-changes:
		if change.Version != store.Version() {
			t.Fatalf("announced %q after a sidecar write; disk is %q", change.Version, store.Version())
		}
	case <-ctx.Done():
		t.Fatal("no change announced after the sidecar was written")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("watch returned %v", err)
	}
}

func TestSidecarWithoutSizesReadsBackUnsized(t *testing.T) {
	store, pipeline := newTestStore(t, "../../examples/loops/bake-off.yaml")
	if err := os.WriteFile(SidecarPath(pipeline), []byte(`{"z":{"x":5,"y":6},"a":{"x":1,"y":2,"w":3,"h":4}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	document := load(t, store)
	if !document.HasLayout || len(document.Layout) != 2 {
		t.Fatalf("layout = %+v", document.Layout)
	}
	if document.Layout[0] != (Placement{ID: "a", X: 1, Y: 2, W: 3, H: 4}) || document.Layout[1] != (Placement{ID: "z", X: 5, Y: 6}) {
		t.Fatalf("layout = %+v", document.Layout)
	}

	// A sidecar that is not an object is read as absent, so the page can
	// still open the pipeline.
	if err := os.WriteFile(SidecarPath(pipeline), []byte(`[1, 2]`), 0o644); err != nil {
		t.Fatal(err)
	}
	document = load(t, store)
	if len(document.Layout) != 0 {
		t.Fatalf("layout from a broken sidecar = %+v", document.Layout)
	}
}
