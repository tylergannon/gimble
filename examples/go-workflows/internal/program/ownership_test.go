package program

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"
)

func ownershipChild(t *testing.T, parent context.Context, name string) context.Context {
	t.Helper()
	child, err := Scope(parent, name)
	if err != nil {
		t.Fatal(err)
	}
	return child
}

func rejectedOwnershipWrite(t *testing.T, ctx context.Context, key string, value any) {
	t.Helper()
	before := scopedSnapshot(t, ctx)
	filesBefore, err := filepath.Glob(filepath.Join(filepath.Dir(before.Index), "*"))
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, key, value); err == nil || !strings.Contains(err.Error(), "owned by scope") {
		t.Fatalf("expected ownership error for %q, got %v", key, err)
	}
	if after := scopedSnapshot(t, ctx); !reflect.DeepEqual(before, after) {
		t.Fatalf("rejected write changed the published snapshot: before=%#v after=%#v", before, after)
	}
	filesAfter, err := filepath.Glob(filepath.Join(filepath.Dir(before.Index), "*"))
	if err != nil || !reflect.DeepEqual(filesBefore, filesAfter) {
		t.Fatalf("rejected write changed the scope directory: before=%v after=%v err=%v", filesBefore, filesAfter, err)
	}
}

func TestOwnershipRejectsInheritedWritesAndAllowsOwnerReplacement(t *testing.T) {
	root := viewContext(t)
	if err := SetContext(root, "goal", "first"); err != nil {
		t.Fatal(err)
	}
	child := ownershipChild(t, root, "chapter")
	grandchild := ownershipChild(t, child, "sprint")
	for _, ctx := range []context.Context{child, grandchild} {
		// Equal JSON is still a write from the wrong scope.
		rejectedOwnershipWrite(t, ctx, "goal", "first")
		rejectedOwnershipWrite(t, ctx, "goal", "replacement")
		rejectedOwnershipWrite(t, ctx, "Goal", "case alias")
	}
	derived, cancel := context.WithCancel(root)
	defer cancel()
	if err := SetContext(derived, "goal", "second"); err != nil {
		t.Fatalf("ordinary derived context lost its owning scope: %v", err)
	}
	if value, _ := scopedValue(t, scopedSnapshot(t, root), "goal"); value != `"second"` {
		t.Fatalf("owner could not replace its value: %s", value)
	}
	if value, _ := scopedValue(t, scopedSnapshot(t, grandchild), "goal"); value != `"first"` {
		t.Fatalf("ownership changed frozen inheritance: %s", value)
	}
}

func TestOwnershipChecksNewAncestorKeysMissingFromFrozenChild(t *testing.T) {
	root := viewContext(t)
	child := ownershipChild(t, root, "chapter")
	grandchild := ownershipChild(t, child, "sprint")
	if err := SetContext(root, "goal", "created after children"); err != nil {
		t.Fatal(err)
	}
	for _, ctx := range []context.Context{child, grandchild} {
		if len(readContextIndex(t, scopedSnapshot(t, ctx)).Entries) != 0 {
			t.Fatal("later ancestor claim changed a frozen snapshot")
		}
		rejectedOwnershipWrite(t, ctx, "goal", "cannot claim unseen ancestor key")
	}
}

func TestOwnershipRejectsAncestorClaimAfterDescendantWrites(t *testing.T) {
	root := viewContext(t)
	chapter := ownershipChild(t, root, "chapter")
	sprint := ownershipChild(t, chapter, "sprint")
	if err := SetContext(sprint, "findings", "sprint findings"); err != nil {
		t.Fatal(err)
	}
	rejectedOwnershipWrite(t, chapter, "findings", "chapter claim")
	rejectedOwnershipWrite(t, root, "Findings", "root alias claim")
	sibling := ownershipChild(t, chapter, "another sprint")
	if err := SetContext(sibling, "findings", "independent findings"); err != nil {
		t.Fatalf("sibling could not own an independent binding: %v", err)
	}
	if err := SetContext(sprint, "findings", "updated sprint findings"); err != nil {
		t.Fatal(err)
	}
	if value, _ := scopedValue(t, scopedSnapshot(t, sibling), "findings"); value != `"independent findings"` {
		t.Fatalf("owner replacement leaked across siblings: %s", value)
	}
}

func TestOwnershipConcurrentComparableClaimsHaveOneWinner(t *testing.T) {
	for trial := range 12 {
		t.Run(fmt.Sprint(trial), func(t *testing.T) {
			root := viewContext(t)
			child := ownershipChild(t, root, "child")
			contexts := []context.Context{root, child}
			keys := []string{"claim", "Claim"}
			results := make([]error, len(contexts))
			start := make(chan struct{})
			var group errgroup.Group
			for i, ctx := range contexts {
				group.Go(func() error {
					<-start
					results[i] = SetContext(ctx, keys[i], "winner")
					return nil
				})
			}
			close(start)
			if err := group.Wait(); err != nil {
				t.Fatal(err)
			}
			successes := 0
			for i, err := range results {
				snapshot := scopedSnapshot(t, contexts[i])
				if err == nil {
					successes++
					if value, _ := scopedValue(t, snapshot, keys[i]); value != `"winner"` {
						t.Fatalf("successful claim was not published: %s", value)
					}
				} else if !strings.Contains(err.Error(), "owned by scope") || snapshot.Revision != 0 || len(readContextIndex(t, snapshot).Entries) != 0 {
					t.Fatalf("losing claim mutated state or failed unexpectedly: %#v, %v", snapshot, err)
				}
			}
			if successes != 1 {
				t.Fatalf("got %d successful comparable claims: %v", successes, results)
			}
		})
	}
}

func TestOwnershipFailedFileWriteDoesNotReserveKey(t *testing.T) {
	root := viewContext(t)
	child := ownershipChild(t, root, "child")
	store := child.Value(contextKey{}).(*contextStore)
	if err := os.Remove(store.dir); err != nil {
		t.Fatal(err)
	}
	if err := SetContext(child, "goal", "cannot persist"); err == nil {
		t.Fatal("setter unexpectedly wrote to a removed directory")
	}
	if err := SetContext(root, "goal", "parent may still claim"); err != nil {
		t.Fatalf("failed child write reserved ownership: %v", err)
	}
	if value, _ := scopedValue(t, scopedSnapshot(t, root), "goal"); value != `"parent may still claim"` {
		t.Fatalf("recovered claim lost its value: %s", value)
	}
}
