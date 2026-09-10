package program

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func viewContext(t *testing.T) context.Context {
	t.Helper()
	ctx, err := NewContext(t.Context(), t.TempDir(), ContextLimits{ValueBytes: 256, PromptBytes: 2000})
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func readViewValue(t *testing.T, snapshot ContextSnapshot, key string) (string, string) {
	t.Helper()
	path := filepath.Join(snapshot.View, "values", key+".json")
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw), target
}

func TestViewMaterializesCompleteNativeDirectory(t *testing.T) {
	ctx := viewContext(t)
	for key, value := range map[string]any{"goal": "deliver", "index": "ordinary key", "research": strings.Repeat("x", 500)} {
		if err := SetContext(ctx, key, value); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := scopedSnapshot(t, ctx)
	entries, err := os.ReadDir(filepath.Join(snapshot.View, "values"))
	if err != nil || len(entries) != 3 {
		t.Fatalf("incomplete view: %#v, %v", entries, err)
	}
	for _, entry := range readContextIndex(t, snapshot).Entries {
		_, target := readViewValue(t, snapshot, entry.Key)
		if target != entry.Path {
			t.Fatalf("view %s points to %s, index specifies %s", entry.Key, target, entry.Path)
		}
	}
	if value, _ := readViewValue(t, snapshot, "index"); value != `"ordinary key"` {
		t.Fatalf("index key collided with routing index: %s", value)
	}
	indexLink := filepath.Join(snapshot.View, "index.json")
	if target, err := os.Readlink(indexLink); err != nil || target != snapshot.Index {
		t.Fatalf("view index target: %s, %v", target, err)
	}
	if !strings.Contains(snapshot.Prompt, indexLink) || strings.Contains(snapshot.Prompt, snapshot.Index) {
		t.Fatalf("prompt should use one pointer into its view: %s", snapshot.Prompt)
	}
	if _, err := os.ReadFile(indexLink); err != nil {
		t.Fatalf("view index is not readable at publication: %v", err)
	}
}

func TestViewReplacementPreservesEarlierLinks(t *testing.T) {
	ctx := viewContext(t)
	if err := SetContext(ctx, "focus", "before"); err != nil {
		t.Fatal(err)
	}
	before := scopedSnapshot(t, ctx)
	if err := SetContext(ctx, "focus", "after"); err != nil {
		t.Fatal(err)
	}
	after := scopedSnapshot(t, ctx)
	oldValue, oldTarget := readViewValue(t, before, "focus")
	newValue, newTarget := readViewValue(t, after, "focus")
	if oldValue != `"before"` || newValue != `"after"` || oldTarget == newTarget || before.View == after.View {
		t.Fatalf("replacement changed an earlier view: old=%s new=%s oldTarget=%s newTarget=%s", oldValue, newValue, oldTarget, newTarget)
	}
	if again := scopedSnapshot(t, ctx); again.View != after.View {
		t.Fatal("unchanged snapshot created another view")
	}
}

func TestViewRejectsCaseCollidingKeysWithoutInvalidation(t *testing.T) {
	ctx := viewContext(t)
	if err := SetContext(ctx, "goal", "original"); err != nil {
		t.Fatal(err)
	}
	before := scopedSnapshot(t, ctx)
	if err := SetContext(ctx, "Goal", "collision"); err == nil {
		t.Fatal("accepted keys that collide on case-insensitive filesystems")
	}
	if after := scopedSnapshot(t, ctx); !reflect.DeepEqual(before, after) {
		t.Fatalf("rejected setter invalidated the published snapshot: before=%#v after=%#v", before, after)
	}
	files, err := filepath.Glob(filepath.Join(filepath.Dir(before.Index), "value-*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("rejected setter wrote another value file: %v, %v", files, err)
	}
	if err := SetContext(ctx, "research", "new data"); err != nil {
		t.Fatal(err)
	}
	after := scopedSnapshot(t, ctx)
	if after.Revision != before.Revision+1 {
		t.Fatalf("rejected setter advanced revision: before=%d after=%d", before.Revision, after.Revision)
	}
	if value, _ := readViewValue(t, after, "goal"); value != `"original"` {
		t.Fatalf("original data changed: %s", value)
	}
	if value, _ := readViewValue(t, after, "research"); value != `"new data"` {
		t.Fatalf("subsequent write failed to publish: %s", value)
	}
	if err := SetContext(ctx, "goal", "replacement"); err != nil {
		t.Fatalf("same-key replacement rejected: %v", err)
	}
	if value, _ := readViewValue(t, scopedSnapshot(t, ctx), "goal"); value != `"replacement"` {
		t.Fatalf("same-key replacement failed: %s", value)
	}
}

func TestViewNestedScopeReusesInheritedValueFiles(t *testing.T) {
	parent := viewContext(t)
	if err := SetContext(parent, "goal", "shared"); err != nil {
		t.Fatal(err)
	}
	root := scopedSnapshot(t, parent)
	child, err := Scope(parent, "chapter")
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(child, "focus", "child only"); err != nil {
		t.Fatal(err)
	}
	nested := scopedSnapshot(t, child)
	_, parentTarget := readViewValue(t, root, "goal")
	_, childTarget := readViewValue(t, nested, "goal")
	if parentTarget != childTarget {
		t.Fatal("child copied inherited JSON instead of linking the original")
	}
	if _, err := os.Stat(filepath.Join(root.View, "values", "focus.json")); !os.IsNotExist(err) {
		t.Fatalf("child key leaked into parent view: %v", err)
	}
	if filepath.Dir(filepath.Dir(nested.View)) != filepath.Dir(root.View) {
		t.Fatal("child view is outside its nested scope layer")
	}
}

func TestViewPublicationFailureRemovesUnpublishedFiles(t *testing.T) {
	ctx := viewContext(t)
	if err := SetContext(ctx, "goal", "data to preserve"); err != nil {
		t.Fatal(err)
	}
	store := ctx.Value(contextKey{}).(*contextStore)
	// Force projection to fail after the index and complete view were built.
	store.limits.PromptBytes = 1
	if _, err := SnapshotContext(ctx); err == nil {
		t.Fatal("published a view that cannot fit its context route")
	}
	entries, err := os.ReadDir(store.dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.HasPrefix(entries[0].Name(), "value-") || store.snapshot != nil {
		t.Fatalf("failed publication left index/view files or a cached snapshot: %#v", entries)
	}
	store.limits.PromptBytes = 2000
	if value, _ := readViewValue(t, scopedSnapshot(t, ctx), "goal"); value != `"data to preserve"` {
		t.Fatalf("publication failure lost pending data: %s", value)
	}
}
