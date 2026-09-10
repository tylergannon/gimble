package program

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodergenMaterializesContextBeforeCallingAgent(t *testing.T) {
	ctx, err := NewContext(t.Context(), t.TempDir(), ContextLimits{ValueBytes: 200, PromptBytes: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, "goal", "Demonstrate the quote command."); err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, "research", strings.Repeat("Boundary observations. ", 100)); err != nil {
		t.Fatal(err)
	}
	rt := &Runtime{Agent: func(callCtx context.Context, call Call) (any, error) {
		snapshot, err := SnapshotContext(callCtx)
		if err != nil {
			return nil, err
		}
		if snapshot.Revision != 2 || len(snapshot.External) == 0 {
			t.Errorf("agent did not receive both updates with spilled research: %+v", snapshot)
		}
		if call.Prompt != snapshot.Prompt+"\n\n## Task\nImplement the current task." {
			t.Errorf("agent prompt does not match the published context and task")
		}
		if _, err := os.ReadFile(snapshot.Index); err != nil {
			t.Errorf("agent started before its index was readable: %v", err)
		}
		if value, err := os.ReadFile(filepath.Join(snapshot.View, "values", "goal.json")); err != nil || string(value) != `"Demonstrate the quote command."` {
			t.Errorf("agent started before its filesystem view was readable: %q, %v", value, err)
		}
		return "scripted result", nil
	}}
	if _, err := Codergen[string](ctx, rt, "implement", "sswe", "Implement the current task."); err != nil {
		t.Fatal(err)
	}
}

func TestContextPublicationFailurePreventsAgentCall(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "context")
	ctx, err := NewContext(t.Context(), dir, ContextLimits{ValueBytes: 200, PromptBytes: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if err := SetContext(ctx, "goal", "Demonstrate the quote command."); err != nil {
		t.Fatal(err)
	}
	// Make the storage directory unavailable before pending indexing occurs.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir, []byte("unavailable"), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	rt := &Runtime{Agent: func(context.Context, Call) (any, error) {
		called = true
		return "unexpected", nil
	}}
	if _, err := Codergen[string](ctx, rt, "implement", "sswe", "Implement."); err == nil {
		t.Fatal("wanted context publication error")
	}
	if called {
		t.Fatal("agent ran without its pending context index")
	}
}
