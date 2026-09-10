package program

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSprintsShareItemScopeWithValidationAndRetries(t *testing.T) {
	root, err := NewContext(t.Context(), t.TempDir(), ContextLimits{ValueBytes: 1024, PromptBytes: 4096})
	if err != nil {
		t.Fatal(err)
	}
	goal := putContext(t, root, "goal", "original")
	putContext(t, root, "attempt", 99)
	items := []Item{{Name: "first", Done: true}, {Name: "second"}}
	worked := map[string]int{}
	scopes := map[string]context.Context{}
	options := LoopOptions{MaxIterations: 3, Validate: func(ctx context.Context, item Item) (bool, error) {
		if previous := scopes[item.Name]; previous != nil && previous != ctx {
			t.Fatal("validation did not reuse the item's scope")
		}
		scopes[item.Name] = ctx
		snapshot := scopedSnapshot(t, ctx)
		if !reflect.DeepEqual(snapshot.Scope, []string{"sprint: " + item.Name}) {
			t.Fatalf("wrong scope: %v", snapshot.Scope)
		}
		if worked[item.Name] == 0 {
			metadata, _ := scopedValue(t, snapshot, "sprint")
			var state itemState
			if err := json.Unmarshal([]byte(metadata), &state); err != nil || state.Attempt != 0 {
				t.Fatalf("initial sprint validation has wrong attempt: %s, %v", metadata, err)
			}
		}
		if item.Name == "first" {
			return worked[item.Name] == 2, nil
		}
		return worked[item.Name] == 1, nil
	}}
	for sprint, err := range Sprints(root, items, options) {
		if err != nil {
			t.Fatal(err)
		}
		if sprint.Context != scopes[sprint.Item.Name] {
			t.Fatal("workflow and validation received different scopes")
		}
		snapshot := scopedSnapshot(t, sprint.Context)
		metadata, _ := scopedValue(t, snapshot, "sprint")
		var state itemState
		if err := json.Unmarshal([]byte(metadata), &state); err != nil || state.Item != sprint.Item || state.Attempt != sprint.Attempt {
			t.Fatalf("scope kept stale checklist state: %s, %v", metadata, err)
		}
		wantGoal := `"original"`
		if worked["first"] > 0 {
			wantGoal = `"later parent update"`
		}
		if value, _ := scopedValue(t, snapshot, "goal"); value != wantGoal {
			t.Fatalf("retry or later sprint missed inherited update: got %s, want %s", value, wantGoal)
		}
		if sprint.Item.Name == "first" && sprint.Attempt == 1 {
			putContext(t, sprint.Context, "feedback", "repair this item")
			if err := SetContext(root, goal, "later parent update"); err != nil {
				t.Fatal(err)
			}
		} else if sprint.Item.Name == "first" {
			if value, _ := scopedValue(t, snapshot, "feedback"); value != `"repair this item"` {
				t.Fatalf("retry lost explicit feedback: %s", value)
			}
		} else {
			for _, entry := range readContextIndex(t, snapshot).Entries {
				if entry.Key == "feedback" {
					t.Fatal("previous sprint's feedback leaked into its sibling")
				}
			}
		}
		worked[sprint.Item.Name]++
	}
	if !items[0].Done || !items[1].Done {
		t.Fatal("scoped validation did not finish the checklist")
	}
	if value, _ := scopedValue(t, scopedSnapshot(t, root), "attempt"); value != "99" {
		t.Fatal("iterator wrote attempt state into its parent")
	}
}

func TestSprintsCanNestWithoutSharingMetadataBindings(t *testing.T) {
	root := viewContext(t)
	outerDone, innerDone := false, false
	outerOptions := LoopOptions{MaxIterations: 1, Validate: func(context.Context, Item) (bool, error) { return outerDone, nil }}
	for outer, err := range Sprints(root, []Item{{Name: "outer"}}, outerOptions) {
		if err != nil {
			t.Fatal(err)
		}
		innerOptions := LoopOptions{MaxIterations: 1, Validate: func(context.Context, Item) (bool, error) { return innerDone, nil }}
		for inner, err := range Sprints(outer.Context, []Item{{Name: "inner"}}, innerOptions) {
			if err != nil {
				t.Fatal(err)
			}
			names := map[string]bool{}
			for _, entry := range readContextIndex(t, scopedSnapshot(t, inner.Context)).Entries {
				value, _ := readViewValue(t, scopedSnapshot(t, inner.Context), entry.Key)
				var state itemState
				if err := json.Unmarshal([]byte(value), &state); err != nil {
					t.Fatal(err)
				}
				names[state.Name] = true
			}
			if len(names) != 2 || !names["outer"] || !names["inner"] {
				t.Fatalf("nested sprint metadata lost an owning scope: %v", names)
			}
			innerDone = true
		}
		outerDone = true
	}
}
