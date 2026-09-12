// Ported from OpenCode c55ee2a8152603f04a409163bd3edf79c425fbd7. MIT,
// Copyright (c) 2025 opencode. See NOTICE.md beside this file.

package sessionstate

import (
	"encoding/json"
	"strconv"
	"testing"
)

func nativeEvent(id, kind string, fields ...any) *Obj {
	return obj("id", id, "type", kind, "created", json.Number("1"), "data", obj(fields...))
}

func sessionRow(id, title string) *Obj {
	return obj("id", id, "title", title, "time", obj("created", json.Number("1")))
}

// This port replaces every upstream findLast over a session transcript with a
// bounded index, so selection never walks cold history. These tests hold the
// indexes to the scans they replace: the naive functions below are literal
// transcriptions of the upstream selectors.

func naiveOpenAssistant(list []*Obj) *Obj {
	var found *Obj
	for _, item := range list {
		if str(item.Get("type")) == "assistant" && !truthy(objOf(item.Get("time")).Get("completed")) {
			found = item
		}
	}
	return found
}

func naiveRunningCompaction(list []*Obj) *Obj {
	var found *Obj
	for _, item := range list {
		if str(item.Get("type")) == "compaction" && str(item.Get("status")) == "running" {
			found = item
		}
	}
	return found
}

func naiveShell(list []*Obj, shellID string) *Obj {
	var found *Obj
	for _, item := range list {
		if str(item.Get("type")) == "shell" && str(item.Get("shellID")) == shellID {
			found = item
		}
	}
	return found
}

func naivePart(item *Obj, kind string, accept func(*Obj) bool) *Obj {
	var found *Obj
	for _, entry := range arrOf(item.Get("content")) {
		part := objOf(entry)
		if part == nil || str(part.Get("type")) != kind {
			continue
		}
		if accept == nil || accept(part) {
			found = part
		}
	}
	return found
}

func requireSame(t *testing.T, label string, got, want *Obj) {
	t.Helper()
	if got == want {
		return
	}
	t.Fatalf("%s: index selected %v, the upstream scan selects %v", label, got, want)
}

// TestIndexSelectorsMatchUpstreamScans replays every fixture and, after each
// accepted event, requires each index to name the exact row or part the
// corresponding upstream findLast would have returned.
func TestIndexSelectorsMatchUpstreamScans(t *testing.T) {
	for _, name := range oracleFixtures() {
		t.Run(name, func(t *testing.T) {
			fixture := loadFixture(t, name+".json")
			projection := newFromFixture(t, fixture)
			for cut, entry := range arrOf(fixture.Get("events")) {
				projection.Apply(objOf(entry))
				requireSelectionParity(t, name+" cut "+strconv.Itoa(cut+1), projection)
			}
		})
	}
}

func requireSelectionParity(t *testing.T, label string, projection *Projection) {
	t.Helper()
	for sessionID, list := range projection.state.Message {
		where := label + " " + sessionID
		requireSame(t, where+" open assistant", projection.openAssistant(sessionID), naiveOpenAssistant(list))
		running, _ := projection.runningCompaction(sessionID)
		requireSame(t, where+" running compaction", running, naiveRunningCompaction(list))
		shells := map[string]bool{}
		for _, item := range list {
			if str(item.Get("type")) == "shell" {
				shells[str(item.Get("shellID"))] = true
			}
		}
		for shellID := range shells {
			requireSame(t, where+" shell "+shellID, projection.shell(sessionID, shellID), naiveShell(list, shellID))
		}
		for _, item := range list {
			if str(item.Get("type")) != "assistant" {
				continue
			}
			id := str(item.Get("id"))
			key := partKey(sessionID, id, "")
			requireSame(t, where+" latest text of "+id, projection.latestText[key], naivePart(item, "text", nil))
			open := naivePart(item, "reasoning", func(part *Obj) bool {
				return !truthy(objOf(part.Get("time")).Get("completed"))
			})
			var top *Obj
			if stack := projection.openReasoning[key]; len(stack) > 0 {
				top = stack[len(stack)-1]
			}
			requireSame(t, where+" open reasoning of "+id, top, open)
			for _, content := range arrOf(item.Get("content")) {
				part := objOf(content)
				if part == nil || str(part.Get("type")) != "tool" {
					continue
				}
				toolID := str(part.Get("id"))
				want := naivePart(item, "tool", func(candidate *Obj) bool { return str(candidate.Get("id")) == toolID })
				requireSame(t, where+" tool "+toolID+" of "+id, projection.tools[partKey(sessionID, id, toolID)], want)
			}
		}
	}
}

// TestOverlappingReasoningOrdinalsFollowLatestOpen is the schema-valid
// counterexample the ordinal fields allow and the upstream updater ignores.
// Two reasoning blocks are open at once; upstream's findLast edits the latest
// still-open one, so completing the second uncovers the first and a later
// delta lands there. A single latest-reasoning slot would drop that delta.
// The producer invariant refuses to emit this shape; the reducer still has to
// reduce it the way upstream does when it arrives.
func TestOverlappingReasoningOrdinalsFollowLatestOpen(t *testing.T) {
	const sessionID, assistantID = "ses_a", "msg_a"
	seed := NewProjectionState()
	seed.Info[sessionID] = sessionRow(sessionID, "overlap")
	seed.Family[sessionID] = []string{sessionID}
	seed.Message[sessionID] = []*Obj{}
	projection := New(seed)

	start := func(id string) *Obj {
		return nativeEvent(id, "session.reasoning.started", "sessionID", sessionID, "assistantMessageID", assistantID)
	}
	projection.Apply(nativeEvent("evt_1", "session.step.started",
		"sessionID", sessionID, "assistantMessageID", assistantID, "agent", "build", "model", obj("id", "m")))

	projection.Apply(start("evt_2"))
	second := start("evt_3")
	projection.Apply(second)

	projection.Apply(nativeEvent("evt_4", "session.reasoning.delta",
		"sessionID", sessionID, "assistantMessageID", assistantID, "delta", "second"))
	projection.Apply(nativeEvent("evt_5", "session.reasoning.ended",
		"sessionID", sessionID, "assistantMessageID", assistantID, "text", "second done"))
	projection.Apply(nativeEvent("evt_6", "session.reasoning.delta",
		"sessionID", sessionID, "assistantMessageID", assistantID, "delta", "first"))

	assistant := projection.message(sessionID, assistantID)
	parts := arrOf(assistant.Get("content"))
	if len(parts) != 2 {
		t.Fatalf("assistant holds %d parts, want the two reasoning blocks", len(parts))
	}
	first, latest := objOf(parts[0]), objOf(parts[1])
	if text := str(first.Get("text")); text != "first" {
		t.Fatalf("the delta after the completion landed on %q, want the reopened earlier block", text)
	}
	if text := str(latest.Get("text")); text != "second done" {
		t.Fatalf("the completed block reads %q", text)
	}
	if truthy(objOf(first.Get("time")).Get("completed")) {
		t.Fatal("the earlier block was completed by the later block's end event")
	}
	requireSelectionParity(t, "overlapping reasoning", projection)

	encoded, err := json.Marshal(projection.Snapshot().State)
	if err != nil {
		t.Fatalf("encode state: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("empty state")
	}
}
