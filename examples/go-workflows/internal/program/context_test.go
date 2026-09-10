package program

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// putContext keeps routine fixtures short; ownership tests declare handles explicitly.
func putContext(t *testing.T, ctx context.Context, name string, value any) Key {
	t.Helper()
	key, err := DeclareContext(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, key, value); err != nil {
		t.Fatal(err)
	}
	return key
}

type contextIndex struct {
	Revision int `json:"revision"`
	Entries  []struct {
		Key   string `json:"key"`
		Path  string `json:"path"`
		Bytes int    `json:"bytes"`
	} `json:"entries"`
}

func readContextIndex(t *testing.T, snapshot ContextSnapshot) contextIndex {
	t.Helper()
	raw, err := os.ReadFile(snapshot.Index)
	if err != nil {
		t.Fatal(err)
	}
	var index contextIndex
	if err := json.Unmarshal(raw, &index); err != nil {
		t.Fatal(err)
	}
	if index.Revision != snapshot.Revision {
		t.Fatalf("index revision %d != snapshot %d", index.Revision, snapshot.Revision)
	}
	return index
}

func TestContextIndividualSpillPreservesCompleteJSON(t *testing.T) {
	ctx, err := NewContext(context.Background(), t.TempDir(), ContextLimits{ValueBytes: 64, PromptBytes: 4096})
	if err != nil {
		t.Fatal(err)
	}
	large := map[string]any{"note": strings.Repeat("é\n", 50), "tags": []string{"first", "last"}}
	for key, value := range map[string]any{"research": large, "goal": "Ship it"} {
		putContext(t, ctx, key, value)
	}
	store := ctx.Value(contextKey{}).(*contextStore)
	indexes, err := filepath.Glob(filepath.Join(store.dir, "index-*.json"))
	if err != nil || len(indexes) != 0 {
		t.Fatalf("Set eagerly indexed: %v, %v", indexes, err)
	}
	snapshot, err := SnapshotContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Revision != 2 || !reflect.DeepEqual(snapshot.Inline, []string{"goal"}) || !reflect.DeepEqual(snapshot.External, []string{"research"}) {
		t.Fatalf("snapshot: %#v", snapshot)
	}
	if !strings.Contains(snapshot.Prompt, `goal = "Ship it"`) || strings.Contains(snapshot.Prompt, "é") {
		t.Fatalf("unexpected projection: %s", snapshot.Prompt)
	}
	index := readContextIndex(t, snapshot)
	if len(index.Entries) != 2 {
		t.Fatalf("index omitted an entry: %#v", index)
	}
	for _, entry := range index.Entries {
		if entry.Key != "research" {
			continue
		}
		got, err := os.ReadFile(entry.Path)
		want, marshalErr := json.Marshal(large)
		if err != nil || marshalErr != nil || string(got) != string(want) || len(got) != entry.Bytes {
			t.Fatalf("lossy value: %q, read=%v marshal=%v", got, err, marshalErr)
		}
	}
}

func TestContextAggregateSpillAndCompactRouting(t *testing.T) {
	ctx, err := NewContext(context.Background(), t.TempDir(), ContextLimits{ValueBytes: 1024, PromptBytes: 800})
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range []struct {
		key  string
		size int
	}{{"large", 500}, {"medium", 400}, {"small", 20}} {
		putContext(t, ctx, pair.key, strings.Repeat("x", pair.size))
	}
	snapshot, err := SnapshotContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Prompt) > 800 || !slices.Contains(snapshot.External, "large") {
		t.Fatalf("aggregate budget not applied: %#v", snapshot)
	}
	// Enough independently short keys force routing through just the index.
	for i := range 40 {
		putContext(t, ctx, strings.Repeat("a", i+1), "tiny")
	}
	snapshot, err = SnapshotContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Prompt != contextRoute(filepath.Join(snapshot.View, "index.json")) || len(snapshot.Inline) != 0 || len(snapshot.External) != 43 {
		t.Fatalf("routing was not collapsed coherently: %#v", snapshot)
	}
	if len(readContextIndex(t, snapshot).Entries) != 43 {
		t.Fatal("collapsed projection lost index entries")
	}
}

func TestContextReplacementKeepsOldSnapshotCoherent(t *testing.T) {
	ctx, err := NewContext(context.Background(), t.TempDir(), ContextLimits{ValueBytes: 64, PromptBytes: 4096})
	if err != nil {
		t.Fatal(err)
	}
	goal := putContext(t, ctx, "goal", "before")
	old, err := SnapshotContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	oldIndex := readContextIndex(t, old)
	oldBytes, err := os.ReadFile(old.Index)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, goal, "after"); err != nil {
		t.Fatal(err)
	}
	current, err := SnapshotContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	newIndex := readContextIndex(t, current)
	stillOld, err := os.ReadFile(old.Index)
	value, valueErr := os.ReadFile(oldIndex.Entries[0].Path)
	if err != nil || valueErr != nil || string(stillOld) != string(oldBytes) || string(value) != `"before"` {
		t.Fatalf("old snapshot changed: index=%v value=%q error=%v", err, value, valueErr)
	}
	if current.Revision != 2 || current.Index == old.Index || newIndex.Entries[0].Path == oldIndex.Entries[0].Path {
		t.Fatalf("replacement reused a published revision: %#v", current)
	}
	current.Inline[0] = "caller mutation"
	again, err := SnapshotContext(ctx)
	if err != nil || again.Inline[0] != "goal" || again.Index != current.Index {
		t.Fatalf("cached snapshot exposed mutable state: %#v, %v", again, err)
	}
}

func TestContextRejectsMissingInvalidAndUnencodableInputs(t *testing.T) {
	if snapshot, err := SnapshotContext(context.Background()); err != nil || snapshot.Index != "" {
		t.Fatalf("unattached context: %#v, %v", snapshot, err)
	}
	if err := SetContext(context.Background(), Key{}, "work"); err == nil {
		t.Fatal("accepted Set without a store")
	}
	for _, limits := range []ContextLimits{{}, {ValueBytes: 1, PromptBytes: 1}} {
		if _, err := NewContext(context.Background(), t.TempDir(), limits); err == nil {
			t.Fatalf("accepted unusable limits: %#v", limits)
		}
	}
	ctx, err := NewContext(context.Background(), t.TempDir(), ContextLimits{ValueBytes: 64, PromptBytes: 4096})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "../../outside", "heading\n# injected"} {
		if _, err := DeclareContext(ctx, key); err == nil {
			t.Fatalf("accepted unsafe key %q", key)
		}
	}
	unsupported, err := DeclareContext(ctx, "unsupported")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, unsupported, make(chan int)); err == nil {
		t.Fatal("accepted non-JSON data")
	}
	if ctx.Value(contextKey{}).(*contextStore).tree.revision != 0 {
		t.Fatal("failed setters advanced the revision")
	}
}

func TestContextCancellationAndIndexFailureStopSnapshot(t *testing.T) {
	ctx, err := NewContext(context.Background(), t.TempDir(), ContextLimits{ValueBytes: 64, PromptBytes: 4096})
	if err != nil {
		t.Fatal(err)
	}
	goal := putContext(t, ctx, "goal", "work")
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := SnapshotContext(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("snapshot ignored cancellation: %v", err)
	}
	if err := SetContext(canceled, goal, "replacement"); !errors.Is(err, context.Canceled) {
		t.Fatalf("setter ignored cancellation: %v", err)
	}
	store := ctx.Value(contextKey{}).(*contextStore)
	if err := os.Rename(store.dir, store.dir+"-unavailable"); err != nil {
		t.Fatal(err)
	}
	if _, err := SnapshotContext(ctx); err == nil {
		t.Fatal("published snapshot without a writable index")
	}
}
