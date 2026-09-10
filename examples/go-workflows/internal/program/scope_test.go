package program

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"
)

func scopedSnapshot(t *testing.T, ctx context.Context) ContextSnapshot {
	t.Helper()
	snapshot, err := SnapshotContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func scopedValue(t *testing.T, snapshot ContextSnapshot, key string) (string, string) {
	t.Helper()
	for _, entry := range readContextIndex(t, snapshot).Entries {
		if entry.Key == key {
			raw, err := os.ReadFile(entry.Path)
			if err != nil {
				t.Fatal(err)
			}
			return string(raw), entry.Path
		}
	}
	t.Fatalf("missing context key %q in %s", key, snapshot.Index)
	return "", ""
}

func TestScopeNestedValuesRestoreParent(t *testing.T) {
	type callerKey struct{}
	base := context.WithValue(t.Context(), callerKey{}, "caller state")
	root, err := NewContext(base, t.TempDir(), ContextLimits{ValueBytes: 256, PromptBytes: 2000})
	if err != nil {
		t.Fatal(err)
	}
	putContext(t, root, "goal", "deliver")
	original := scopedSnapshot(t, root)
	chapter, err := Scope(root, "chapter")
	if err != nil {
		t.Fatal(err)
	}
	putContext(t, chapter, "chapter_focus", "chapter work")
	sprint, err := Scope(chapter, "sprint")
	if err != nil {
		t.Fatal(err)
	}
	putContext(t, sprint, "sprint_focus", "sprint work")
	current := scopedSnapshot(t, sprint)
	if !reflect.DeepEqual(current.Scope, []string{"chapter", "sprint"}) || !strings.HasPrefix(current.Prompt, "Scope: [\"chapter\",\"sprint\"]\n") {
		t.Fatalf("scope ancestry is missing: %#v", current)
	}
	if sprint.Value(callerKey{}) != "caller state" {
		t.Fatal("scope discarded another Go context value")
	}
	if value, _ := scopedValue(t, current, "sprint_focus"); value != `"sprint work"` {
		t.Fatalf("sprint focus: %s", value)
	}
	parent := scopedSnapshot(t, chapter)
	childDir, parentDir := filepath.Dir(current.Index), filepath.Dir(parent.Index)
	if filepath.Dir(childDir) != parentDir {
		t.Fatalf("child storage %s is not inside parent layer %s", childDir, parentDir)
	}
	if value, _ := scopedValue(t, parent, "chapter_focus"); value != `"chapter work"` {
		t.Fatalf("child write leaked to parent: %s", value)
	}
	restored := scopedSnapshot(t, root)
	if entries := readContextIndex(t, restored).Entries; len(entries) != 1 || entries[0].Key != "goal" {
		t.Fatalf("child bindings leaked into root: %#v", entries)
	}
	_, rootPath := scopedValue(t, original, "goal")
	_, inheritedPath := scopedValue(t, current, "goal")
	if rootPath != inheritedPath || current.Index == parent.Index || parent.Index == original.Index {
		t.Fatal("scopes must share immutable values but publish their own indexes")
	}
	raw, err := os.ReadFile(current.Index)
	if err != nil {
		t.Fatal(err)
	}
	var index struct{ Scope []string }
	if err := json.Unmarshal(raw, &index); err != nil || !reflect.DeepEqual(index.Scope, current.Scope) {
		t.Fatalf("index scope differs from snapshot: %#v, %v", index, err)
	}
	current.Scope[0] = "caller mutation"
	if scopedSnapshot(t, sprint).Scope[0] != "chapter" {
		t.Fatal("snapshot exposed mutable scope metadata")
	}
}

func TestScopeRefreshesParentValuesAndPreservesPublishedSnapshots(t *testing.T) {
	parent := viewContext(t)
	goal := putContext(t, parent, "goal", "first")
	child := ownershipChild(t, parent, "research")
	before := scopedSnapshot(t, child)
	if err := SetContext(parent, goal, "second"); err != nil {
		t.Fatal(err)
	}
	putContext(t, parent, "new_key", "learned after child creation")
	after := scopedSnapshot(t, child)
	if value, _ := scopedValue(t, after, "goal"); value != `"second"` {
		t.Fatalf("child missed later parent update: %s", value)
	}
	if value, _ := scopedValue(t, after, "new_key"); value != `"learned after child creation"` {
		t.Fatalf("child missed new parent binding: %s", value)
	}
	if value, _ := readViewValue(t, before, "goal"); value != `"first"` {
		t.Fatalf("parent update changed a published child view: %s", value)
	}
	if before.View == after.View || len(readContextIndex(t, before).Entries) != 1 {
		t.Fatal("parent update mutated the earlier snapshot")
	}
}

func TestScopeParallelSiblingWritesAreIsolated(t *testing.T) {
	parent, err := NewContext(t.Context(), t.TempDir(), ContextLimits{ValueBytes: 256, PromptBytes: 2000})
	if err != nil {
		t.Fatal(err)
	}
	putContext(t, parent, "goal", "parent")
	group, ctx := errgroup.WithContext(parent)
	results := make([]ContextSnapshot, 3)
	for i := range results {
		group.Go(func() error {
			child, err := Scope(ctx, fmt.Sprintf("candidate-%d", i))
			if err != nil {
				return err
			}
			choice, err := DeclareContext(child, "choice")
			if err != nil {
				return err
			}
			if err := SetContext(child, choice, fmt.Sprintf("child-%d", i)); err != nil {
				return err
			}
			results[i], err = SnapshotContext(child)
			return err
		})
	}
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
	for i, result := range results {
		if value, _ := scopedValue(t, result, "choice"); value != fmt.Sprintf("\"child-%d\"", i) {
			t.Fatalf("sibling %d contains %s", i, value)
		}
	}
	if value, _ := scopedValue(t, scopedSnapshot(t, parent), "goal"); value != `"parent"` {
		t.Fatalf("parallel child writes leaked to parent: %s", value)
	}
}

func TestScopeRejectsMissingStoreAndInheritsCancellation(t *testing.T) {
	if _, err := Scope(t.Context(), "unattached"); err == nil {
		t.Fatal("accepted a scope without a parent store")
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	parent, err := NewContext(ctx, t.TempDir(), ContextLimits{ValueBytes: 256, PromptBytes: 2000})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Scope(parent, " \n "); err == nil {
		t.Fatal("accepted a blank scope")
	}
	child, err := Scope(parent, "research")
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := SnapshotContext(child); !errors.Is(err, context.Canceled) {
		t.Fatalf("child ignored parent cancellation: %v", err)
	}
	if _, err := Scope(parent, "too late"); !errors.Is(err, context.Canceled) {
		t.Fatalf("created child after cancellation: %v", err)
	}
}
