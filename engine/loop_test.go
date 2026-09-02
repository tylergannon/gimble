package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/checklist"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

const threeItemChecklist = `---
items:
  - name: First
    check: The first thing works
    command: "true"
  - name: Second
    check: The second thing works
    command: "true"
  - name: Third
    check: The third thing works
    command: "true"
---

# Sprint

Definition of done, in prose the engine never reads.
`

// Claim 1: three commands that exit 0 complete in three laps, every item ends
// done, and the markdown body survives byte for byte.
func TestLoopCompletesThreeItemChecklistInThreeLaps(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), threeItemChecklist)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	dispatches := 0
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) { dispatches++ }))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted {
		t.Fatalf("result = %#v", result)
	}
	if dispatches != 3 {
		t.Fatalf("body dispatched %d times, want 3", dispatches)
	}
	list := mustChecklist(t, filepath.Join(workdir, "sprint.md"))
	for _, item := range list.Items {
		if !item.Done {
			t.Fatalf("item %q is not done", item.Name)
		}
	}
	wantBody := threeItemChecklist[strings.LastIndex(threeItemChecklist, "---")+3:]
	if list.Body != wantBody {
		t.Fatalf("body = %q, want %q", list.Body, wantBody)
	}
	if _, err := os.Stat(filepath.Join(root, "stages", "000001-items", "validation.log")); !os.IsNotExist(err) {
		t.Fatalf("first arrival wrote a validation log: %v", err)
	}
	for _, stage := range []string{"000003-items", "000005-items", "000007-items"} {
		for _, name := range []string{"validation.log", "validation.json"} {
			if _, err := os.Stat(filepath.Join(root, "stages", stage, name)); err != nil {
				t.Fatalf("%s/%s: %v", stage, name, err)
			}
		}
	}
	assertTextFile(t, filepath.Join(root, "frames.json"), "[]\n")
	events := mustTimeline(t, root)
	validated, completed := 0, 0
	for _, event := range events {
		switch event["type"] {
		case "LoopValidated":
			if event["passed"] != true {
				t.Fatalf("validation did not pass: %v", event)
			}
			validated++
		case "LoopCompleted":
			completed++
		}
	}
	if validated != 3 || completed != 1 {
		t.Fatalf("timeline has %d LoopValidated and %d LoopCompleted", validated, completed)
	}
}

// Claim 2: a failing command re-enters the item with the failure in the
// frame; once the command passes the item is marked and the loop advances.
func TestLoopReentersFailedItemWithLastValidation(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Flag
    check: The flag file exists
    command: test -f flag
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	laps := 0
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(scope ExecutionScope) {
		laps++
		if laps == 2 {
			writeFile(t, filepath.Join(scope.Workdir, "flag"), "")
		}
	}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted || laps != 2 {
		t.Fatalf("result = %#v after %d laps", result, laps)
	}
	first := readFile(t, filepath.Join(root, "stages", "000002-implement", "prompt.md"))
	if strings.Contains(first, "last validation") || !strings.HasPrefix(first, `<tractor loop="items" checklist="sprint.md" item="1/1" lap="1">`) {
		t.Fatalf("first lap prompt = %q", first)
	}
	second := readFile(t, filepath.Join(root, "stages", "000004-implement", "prompt.md"))
	for _, want := range []string{`lap="2"`, "last validation: failed — exit 1", "name: Flag", "command: test -f flag", "\n\nimplement"} {
		if !strings.Contains(second, want) {
			t.Fatalf("second lap prompt %q lacks %q", second, want)
		}
	}
	var failed validationRecord
	readJSON(t, filepath.Join(root, "stages", "000003-items", "validation.json"), &failed)
	if failed.Passed || failed.ExitCode != 1 || !strings.HasPrefix(failed.Summary, "exit 1") || failed.Item != "Flag" {
		t.Fatalf("failed validation = %#v", failed)
	}
	var passed validationRecord
	readJSON(t, filepath.Join(root, "stages", "000005-items", "validation.json"), &passed)
	if !passed.Passed || passed.ExitCode != 0 {
		t.Fatalf("passed validation = %#v", passed)
	}
	if item, _, _ := mustChecklist(t, filepath.Join(workdir, "sprint.md")).Find("Flag"); !item.Done {
		t.Fatal("item was not marked done")
	}
}

// Claim 3: with a judge that answers fail then pass, an infer item is marked
// only after the pass.
func TestLoopInferItemIsMarkedOnlyAfterJudgePasses(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Screens
    check: Every screen looks usable
    infer:
      files: captures/*.png
      prompt: Judge whether the screens look usable
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	backend := &scriptedBackend{outcomes: []harness.Outcome{
		{Next: "fail", Notes: "the button is missing"},
		{Next: "pass", Notes: "all screens usable"},
	}}
	laps := 0
	registry := NewRegistry()
	registry.Register("codergen", HandlerFunc(func(_ graph.Node, _ []graph.Edge, scope ExecutionScope, _ *graph.Graph) (harness.Outcome, *harness.Error) {
		laps++
		writeFile(t, filepath.Join(scope.Workdir, "captures", "shot.png"), "png")
		return harness.Outcome{Notes: "worked"}, nil
	}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, backend).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted || laps != 2 || len(backend.turns) != 2 {
		t.Fatalf("result = %#v after %d laps and %d judge turns", result, laps, len(backend.turns))
	}
	var first validationRecord
	readJSON(t, filepath.Join(root, "stages", "000003-items", "validation.json"), &first)
	if first.Passed || first.Infer == nil || first.Infer.Verdict != "fail" || first.Summary != "judge: the button is missing" {
		t.Fatalf("first validation = %#v infer=%#v", first, first.Infer)
	}
	if len(first.Infer.Files) != 1 || first.Infer.Files[0] != filepath.Join("captures", "shot.png") {
		t.Fatalf("evidence files = %v", first.Infer.Files)
	}
	var second validationRecord
	readJSON(t, filepath.Join(root, "stages", "000005-items", "validation.json"), &second)
	if !second.Passed || second.Infer == nil || second.Infer.Verdict != "pass" || second.Infer.Notes != "all screens usable" {
		t.Fatalf("second validation = %#v infer=%#v", second, second.Infer)
	}
	judgePrompt := readFile(t, filepath.Join(root, "stages", "000003-items", "prompt.md"))
	for _, want := range []string{"item: Screens", "check: Every screen looks usable", "Judge whether the screens look usable", "\n- captures/shot.png"} {
		if !strings.Contains(judgePrompt, want) {
			t.Fatalf("judge prompt %q lacks %q", judgePrompt, want)
		}
	}
	turn := backend.turns[0]
	if turn.Fidelity != harness.FidelityNone || turn.ThreadKey != "" || turn.NodeID != "items" || turn.RunLog == "" || turn.Workdir != workdir {
		t.Fatalf("judge turn = %#v", turn)
	}
	if !strings.Contains(string(turn.OutputSchema), `"enum":["pass","fail"]`) {
		t.Fatalf("judge schema = %s", turn.OutputSchema)
	}
	if item, _, _ := mustChecklist(t, filepath.Join(workdir, "sprint.md")).Find("Screens"); !item.Done {
		t.Fatal("item was not marked done")
	}
}

func TestLoopInferFailsWhenNoEvidenceMatches(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Screens
    check: Every screen looks usable
    infer:
      files: [captures/*.png, shots/*.jpg]
      prompt: Judge the screens
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 1),
	)
	backend := &scriptedBackend{}
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, backend).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunFailed || !strings.Contains(result.FailureReason, `loop body "implement" has exhausted its visit budget with 1 items open`) {
		t.Fatalf("result = %#v", result)
	}
	if len(backend.turns) != 0 {
		t.Fatalf("judge ran %d times without evidence", len(backend.turns))
	}
	var record validationRecord
	readJSON(t, filepath.Join(root, "stages", "000003-items", "validation.json"), &record)
	if record.Passed || record.Summary != "no evidence files matched captures/*.png, shots/*.jpg" {
		t.Fatalf("validation = %#v", record)
	}
}

// Claim 4: an outer checklist whose items carry checklist: drives an inner
// loop with no checklist field; the inner lap's prompt carries both frames.
func TestLoopNestedChecklistsDriveInnerLoop(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "outer.md"), `---
items:
  - name: Chapter A
    check: Chapter A is delivered
    checklist: inner-a.md
  - name: Chapter B
    check: Chapter B is delivered
    checklist: inner-b.md
---
`)
	for _, name := range []string{"a", "b"} {
		writeFile(t, filepath.Join(workdir, "inner-"+name+".md"), `---
items:
  - name: Sprint 1
    check: Sprint 1 holds
    command: "true"
  - name: Sprint 2
    check: Sprint 2 holds
    command: "true"
---
`)
	}
	pipeline := testGraph(
		startNode("start", "outer"),
		loopNode("outer", "outer.md", "plan", graph.Success, 0),
		customNode("plan", "task", []graph.Edge{{To: "inner"}}, 0),
		loopNode("inner", "", "work", "outer", 0),
		customNode("work", "task", []graph.Edge{{To: "inner"}}, 0),
	)
	plans, works := 0, 0
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(scope ExecutionScope) {
		if strings.HasSuffix(scope.StageDir, "-plan") {
			plans++
		} else {
			works++
		}
	}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted || plans != 2 || works != 4 {
		t.Fatalf("result = %#v with %d plans and %d works", result, plans, works)
	}
	for _, file := range []string{"outer.md", "inner-a.md", "inner-b.md"} {
		for _, item := range mustChecklist(t, filepath.Join(workdir, file)).Items {
			if !item.Done {
				t.Fatalf("%s item %q is not done", file, item.Name)
			}
		}
	}
	// Stages: 1 outer, 2 plan, 3 inner, 4 work, 5 inner, 6 work, 7 inner, 8 outer, ...
	work := readFile(t, filepath.Join(root, "stages", "000006-work", "prompt.md"))
	outerAt := strings.Index(work, `<tractor loop="outer" checklist="outer.md" item="1/2" lap="1">`)
	innerAt := strings.Index(work, `<tractor loop="inner" checklist="inner-a.md" item="2/2" lap="1">`)
	if outerAt < 0 || innerAt < 0 || outerAt > innerAt {
		t.Fatalf("work prompt = %q", work)
	}
	if !strings.Contains(work, "checklist: inner-a.md") || !strings.Contains(work, "name: Sprint 2") {
		t.Fatalf("work prompt = %q", work)
	}
	plan := readFile(t, filepath.Join(root, "stages", "000009-plan", "prompt.md"))
	if strings.Contains(plan, `loop="inner"`) || !strings.Contains(plan, `<tractor loop="outer" checklist="outer.md" item="2/2" lap="1">`) {
		t.Fatalf("second plan prompt = %q", plan)
	}
	if _, err := os.Stat(filepath.Join(root, "stages", "000008-outer", "validation.json")); err != nil {
		t.Fatalf("outer item was not validated on return: %v", err)
	}
}

func TestLoopFrameCarriesDocContents(t *testing.T) {
	tests := []struct {
		name     string
		writeDoc bool
		want     string
	}{
		{name: "readable doc", writeDoc: true, want: "--- doc: notes.md ---\nRead me carefully.\n</tractor>"},
		{name: "missing doc", writeDoc: false, want: "doc: notes.md (unreadable: "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, workdir := t.TempDir(), t.TempDir()
			writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Only
    check: It works
    doc: notes.md
---
`)
			if test.writeDoc {
				writeFile(t, filepath.Join(workdir, "notes.md"), "Read me carefully.\n")
			}
			pipeline := testGraph(
				startNode("start", "items"),
				loopNode("items", "sprint.md", "implement", graph.Success, 0),
				customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
			)
			registry := NewRegistry()
			registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {}))

			result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != RunCompleted {
				t.Fatalf("result = %#v", result)
			}
			prompt := readFile(t, filepath.Join(root, "stages", "000002-implement", "prompt.md"))
			if !strings.Contains(prompt, test.want) {
				t.Fatalf("prompt %q lacks %q", prompt, test.want)
			}
		})
	}
}

func TestLoopHonorsHandMarkedItemWithoutValidating(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	path := filepath.Join(workdir, "sprint.md")
	writeFile(t, path, `---
items:
  - name: Manual
    check: A person signed off
    command: "false"
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {
		writeFile(t, path, strings.Replace(readFile(t, path), "command: \"false\"\n", "command: \"false\"\n    done: true\n", 1))
	}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted {
		t.Fatalf("result = %#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "stages", "000003-items", "validation.log")); !os.IsNotExist(err) {
		t.Fatalf("hand-marked item was validated: %v", err)
	}
	for _, event := range mustTimeline(t, root) {
		if event["type"] == "LoopValidated" {
			t.Fatalf("hand-marked item produced %v", event)
		}
	}
}

func TestLoopFailsWhenBodyBudgetIsExhaustedWithItemsOpen(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: One
    check: One
    command: "true"
  - name: Two
    check: Two
    command: "true"
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 1),
	)
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunFailed || result.FailureReason != `loop body "implement" has exhausted its visit budget with 1 items open` {
		t.Fatalf("result = %#v", result)
	}
	list := mustChecklist(t, filepath.Join(workdir, "sprint.md"))
	if !list.Items[0].Done || list.Items[1].Done {
		t.Fatalf("items = %#v", list.Items)
	}
}

func TestLoopMaxVisitsBoundsLapsThroughOfferedSet(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Never
    check: Never passes
    command: "false"
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 2),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	laps := 0
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) { laps++ }))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	// Two arrivals are one lap: the second arrival re-selects the item, and
	// the body then finds its only successor exhausted before dispatch.
	if result.Status != RunFailed || !strings.Contains(result.FailureReason, "every successor of implement has exhausted") || laps != 1 {
		t.Fatalf("result = %#v after %d laps", result, laps)
	}
}

func TestLoopTerminalErrorsBeforeDispatch(t *testing.T) {
	tests := []struct {
		name      string
		checklist string
		want      string
	}{
		{name: "missing checklist file", checklist: "absent.md", want: "absent.md"},
		{name: "no checklist outside a loop", checklist: "", want: "loop items has no checklist and is not inside another loop's body"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, workdir := t.TempDir(), t.TempDir()
			pipeline := testGraph(
				startNode("start", "items"),
				loopNode("items", test.checklist, "implement", graph.Success, 0),
				customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
			)
			registry := NewRegistry()
			registry.Register("codergen", bodyHandler(t, func(ExecutionScope) { t.Fatal("body dispatched") }))

			result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != RunFailed || !strings.Contains(result.FailureReason, test.want) {
				t.Fatalf("result = %#v", result)
			}
		})
	}
}

func TestLoopValidationCommandRespectsTimeoutAndStop(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Slow
    check: Finishes quickly
    command: sleep 10
---
`)
	loop := loopNode("items", "sprint.md", "implement", graph.Success, 0)
	loop.Timeout = jsonschema.Optional[graph.Duration]{Present: true, Value: "50ms"}
	pipeline := testGraph(
		startNode("start", "items"),
		loop,
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunFailed || result.FailureReason != "validation command timed out" {
		t.Fatalf("result = %#v", result)
	}
}

func TestFanInPrependsFrame(t *testing.T) {
	if got := prependFrame("", "prompt"); got != "prompt" {
		t.Fatalf("empty frame produced %q", got)
	}
	if got := prependFrame("<tractor>", "prompt"); got != "<tractor>\n\nprompt" {
		t.Fatalf("frame produced %q", got)
	}
}

func loopNode(id, checklistPath, body, onDone string, maxVisits int) *graph.LoopNode {
	node := &graph.LoopNode{NodeBase: graph.NodeBase{ID: id}, Body: body, OnDone: onDone}
	if checklistPath != "" {
		node.Checklist = optional(checklistPath)
	}
	if maxVisits > 0 {
		node.MaxVisits = jsonschema.Optional[int]{Present: true, Value: maxVisits}
	}
	return node
}

// bodyHandler runs hook and then a real simulated codergen turn so that
// prompt.md in the body's stage directory reflects the injected frame.
func bodyHandler(t *testing.T, hook func(ExecutionScope)) Handler {
	t.Helper()
	real := NewCodergenHandler(CodergenConfig{DefaultModel: "gpt-5.3-codex"})
	return HandlerFunc(func(node graph.Node, offered []graph.Edge, scope ExecutionScope, pipeline *graph.Graph) (harness.Outcome, *harness.Error) {
		hook(scope)
		return real.Execute(node, offered, scope, pipeline)
	})
}

func newLoopRunner(t *testing.T, pipeline graph.Graph, registry *Registry, root, workdir string, backend harness.CodergenBackend) *Runner {
	t.Helper()
	runner, err := NewRunner(pipeline, registry, RunnerConfig{
		LogsRoot:     root,
		Workdir:      workdir,
		Validate:     func(graph.Graph) error { return nil },
		Backend:      backend,
		DefaultModel: "gpt-5.3-codex",
	})
	if err != nil {
		t.Fatal(err)
	}
	return runner
}

// scriptedBackend answers codergen turns from a fixed list of outcomes and
// records every turn it saw. Everything else behaves like fakeBackend.
type scriptedBackend struct {
	fakeBackend
	outcomes []harness.Outcome
	turns    []harness.CodergenTurn
}

func (b *scriptedBackend) Run(turn harness.CodergenTurn) (harness.Outcome, *harness.Error) {
	b.turns = append(b.turns, turn)
	if len(b.outcomes) == 0 {
		return harness.Outcome{}, terminalError("scripted backend has no outcome left")
	}
	outcome := b.outcomes[0]
	b.outcomes = b.outcomes[1:]
	return outcome, nil
}

func mustChecklist(t *testing.T, path string) *checklist.Checklist {
	t.Helper()
	list, err := checklist.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return list
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func readJSON(t *testing.T, path string, value any) {
	t.Helper()
	if err := json.Unmarshal([]byte(readFile(t, path)), value); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

func mustTimeline(t *testing.T, root string) []timelineEvent {
	t.Helper()
	events, err := readTimeline(filepath.Join(root, "timeline.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return events
}
