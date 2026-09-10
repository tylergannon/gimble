package program

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func ownershipChild(t *testing.T, parent context.Context, name string) context.Context {
	t.Helper()
	child, err := Scope(parent, name)
	if err != nil {
		t.Fatal(err)
	}
	return child
}

func rejectedOwnershipWrite(t *testing.T, ctx context.Context, key Key) {
	t.Helper()
	before := scopedSnapshot(t, ctx)
	if err := SetContext(ctx, key, "wrong scope"); err == nil || !strings.Contains(err.Error(), "owned by scope") {
		t.Fatalf("expected ownership error, got %v", err)
	}
	if after := scopedSnapshot(t, ctx); !reflect.DeepEqual(before, after) {
		t.Fatalf("rejected write changed the published snapshot: before=%#v after=%#v", before, after)
	}
}

func TestOwnershipIsDeclaredBeforeAnyScopeWrites(t *testing.T) {
	root := viewContext(t)
	goal, err := DeclareContext(root, "goal")
	if err != nil {
		t.Fatal(err)
	}
	child := ownershipChild(t, root, "chapter")
	grandchild := ownershipChild(t, child, "sprint")
	// Descendants attempt the first write, before the owner assigns a value.
	for _, ctx := range []context.Context{child, grandchild} {
		rejectedOwnershipWrite(t, ctx, goal)
	}
	if err := SetContext(root, goal, "first"); err != nil {
		t.Fatal(err)
	}
	before := scopedSnapshot(t, grandchild)
	derived, cancel := context.WithCancel(root)
	defer cancel()
	if err := SetContext(derived, goal, "second"); err != nil {
		t.Fatalf("ordinary derived context lost its owning scope: %v", err)
	}
	for _, ctx := range []context.Context{root, child, grandchild} {
		if value, _ := scopedValue(t, scopedSnapshot(t, ctx), "goal"); value != `"second"` {
			t.Fatalf("owner replacement was not visible: %s", value)
		}
	}
	if value, _ := readViewValue(t, before, "goal"); value != `"first"` {
		t.Fatalf("owner replacement changed an earlier view: %s", value)
	}
	rejectedOwnershipWrite(t, grandchild, goal)
}

func TestOwnershipDistinctBindingsWithSameLabelRemainVisible(t *testing.T) {
	root := viewContext(t)
	chapter := ownershipChild(t, root, "chapter")
	sprint := ownershipChild(t, chapter, "sprint")
	// The descendant writes first; it cannot acquire the ancestor's binding.
	sprintFocus, err := DeclareContext(sprint, "focus")
	if err != nil {
		t.Fatal(err)
	}
	chapterFocus, err := DeclareContext(chapter, "focus")
	if err != nil {
		t.Fatal(err)
	}
	if sprintFocus == chapterFocus {
		t.Fatal("same display label aliased distinct scope bindings")
	}
	if err := SetContext(sprint, sprintFocus, "sprint findings"); err != nil {
		t.Fatal(err)
	}
	if err := SetContext(chapter, chapterFocus, "chapter direction"); err != nil {
		t.Fatalf("descendant write took ownership from the chapter: %v", err)
	}
	rejectedOwnershipWrite(t, sprint, chapterFocus)
	rejectedOwnershipWrite(t, chapter, sprintFocus)

	snapshot := scopedSnapshot(t, sprint)
	entries := readContextIndex(t, snapshot).Entries
	if len(entries) != 2 || strings.EqualFold(entries[0].Key, entries[1].Key) {
		t.Fatalf("colliding labels lost a binding or aliased native paths: %#v", entries)
	}
	values := map[string]bool{}
	for _, entry := range entries {
		value, target := readViewValue(t, snapshot, entry.Key)
		values[value] = true
		if target != entry.Path || !strings.Contains(snapshot.Prompt, entry.Key+" = "+value) {
			t.Fatalf("index, prompt, and view disagree for %s", entry.Key)
		}
	}
	if !values[`"chapter direction"`] || !values[`"sprint findings"`] {
		t.Fatalf("nested context did not retain both bindings: %v", values)
	}
}
