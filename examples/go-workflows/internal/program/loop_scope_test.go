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
	if err := SetContext(root, "goal", "original"); err != nil {
		t.Fatal(err)
	}
	if err := SetContext(root, "attempt", 99); err != nil {
		t.Fatal(err)
	}
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
		if value, _ := scopedValue(t, snapshot, "goal"); value != `"original"` {
			t.Fatalf("item scope changed after its first validation: %s", value)
		}
		if sprint.Item.Name == "first" && sprint.Attempt == 1 {
			if err := SetContext(sprint.Context, "feedback", "repair this item"); err != nil {
				t.Fatal(err)
			}
			if err := SetContext(root, "goal", "later parent update"); err != nil {
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
