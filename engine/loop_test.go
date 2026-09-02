package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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

// Claim 2: a failing command re-enters the item with the failure and its
// log path in the frame; once the command passes the item is marked, the
// loop advances, and the next item's frame carries neither.
func TestLoopReentersFailedItemWithLastValidation(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Flag
    check: The flag file exists
    command: test -f flag
  - name: Next
    check: Nothing to check
    command: "true"
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
	if result.Status != RunCompleted || laps != 3 {
		t.Fatalf("result = %#v after %d laps", result, laps)
	}
	first := readFile(t, filepath.Join(root, "stages", "000002-implement", "prompt.md"))
	if strings.Contains(first, "last validation") || !strings.Contains(first, `<iterate loop="items" checklist="sprint.md" item="1/2" lap="1">`) {
		t.Fatalf("first lap prompt = %q", first)
	}
	second := readFile(t, filepath.Join(root, "stages", "000004-implement", "prompt.md"))
	failedLog := filepath.Join(root, "stages", "000003-items", "validation.log")
	for _, want := range []string{`lap="2"`, "last validation: failed — exit 1", "\n  validation log: " + failedLog + "\n", "name: Flag", "command: test -f flag", "\n\nimplement"} {
		if !strings.Contains(second, want) {
			t.Fatalf("second lap prompt %q lacks %q", second, want)
		}
	}
	third := readFile(t, filepath.Join(root, "stages", "000006-implement", "prompt.md"))
	if strings.Contains(third, "last validation") || strings.Contains(third, "validation log:") || !strings.Contains(third, `item="2/2" lap="1"`) {
		t.Fatalf("prompt after the pass = %q", third)
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
	outerAt := strings.Index(work, `<iterate loop="outer" checklist="outer.md" item="1/2" lap="1">`)
	innerAt := strings.Index(work, `<iterate loop="inner" checklist="inner-a.md" item="2/2" lap="1">`)
	if outerAt < 0 || innerAt < 0 || outerAt > innerAt {
		t.Fatalf("work prompt = %q", work)
	}
	if !strings.Contains(work, "checklist: inner-a.md") || !strings.Contains(work, "name: Sprint 2") {
		t.Fatalf("work prompt = %q", work)
	}
	plan := readFile(t, filepath.Join(root, "stages", "000009-plan", "prompt.md"))
	if strings.Contains(plan, `loop="inner"`) || !strings.Contains(plan, `<iterate loop="outer" checklist="outer.md" item="2/2" lap="1">`) {
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
		{name: "readable doc", writeDoc: true, want: "  --- doc: notes.md ---\n  Read me carefully.\n</iterate>"},
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

func TestLoopValidationCommandStoppedByOperator(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Slow
    check: Finishes eventually
    command: touch validating && sleep 30
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {}))
	runner := newLoopRunner(t, pipeline, registry, root, workdir, nil)

	type runResponse struct {
		result RunResult
		err    error
	}
	finished := make(chan runResponse, 1)
	go func() {
		result, err := runner.Run()
		finished <- runResponse{result, err}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(workdir, "validating")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("validation command never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
	stopped := time.Now()
	runner.Stop()
	select {
	case response := <-finished:
		if response.err != nil {
			t.Fatal(response.err)
		}
		if response.result.Status != RunFailed || response.result.FailureReason != "validation command stopped by operator" {
			t.Fatalf("result = %#v", response.result)
		}
		if elapsed := time.Since(stopped); elapsed > 5*time.Second {
			t.Fatalf("stop took %s", elapsed)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run did not return after stop")
	}
}

// Finding 1(a): a checkpoint whose continuation is a body node (a failure
// checkpoint with retry_visit) resumes at the loop node, so the body runs
// with a frame.
func TestResumeInsideLoopBodyRewindsToLoop(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Only
    check: It works
    command: "true"
---
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	crashing := NewRegistry()
	crashing.Register("codergen", HandlerFunc(func(graph.Node, []graph.Edge, ExecutionScope, *graph.Graph) (harness.Outcome, *harness.Error) {
		return harness.Outcome{}, terminalError("harness died")
	}))
	result, err := newLoopRunner(t, pipeline, crashing, root, workdir, nil).Run()
	if err != nil || result.Status != RunFailed {
		t.Fatalf("initial result=%#v err=%v", result, err)
	}
	checkpoint := mustCheckpoint(t, root)
	if checkpoint.CurrentNode != "implement" || checkpoint.NextNode != "implement" || !checkpoint.RetryVisit {
		t.Fatalf("failure checkpoint = %#v", checkpoint)
	}

	var dispatched []string
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(scope ExecutionScope) {
		dispatched = append(dispatched, filepath.Base(scope.StageDir))
	}))
	result, err = newLoopResumeRunner(t, pipeline, registry, root, workdir).Run()
	if err != nil || result.Status != RunCompleted {
		t.Fatalf("resumed result=%#v err=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(root, "stages", "000003-items")); err != nil {
		t.Fatalf("resumed run did not start at the loop node: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "stages", "000003-items", "validation.json")); !os.IsNotExist(err) {
		t.Fatalf("rewound arrival validated something: %v", err)
	}
	if len(dispatched) != 1 || dispatched[0] != "000004-implement" {
		t.Fatalf("body dispatches = %v", dispatched)
	}
	prompt := readFile(t, filepath.Join(root, "stages", "000004-implement", "prompt.md"))
	if !strings.Contains(prompt, `<iterate loop="items" checklist="sprint.md" item="1/1" lap="1">`) {
		t.Fatalf("resumed body prompt = %q", prompt)
	}
	checkpoint = mustCheckpoint(t, root)
	if checkpoint.NodeVisits["items"] != 3 || checkpoint.NodeVisits["implement"] != 2 || checkpoint.RetryVisit {
		t.Fatalf("resumed counters = %#v", checkpoint)
	}
	if item, _, _ := mustChecklist(t, filepath.Join(workdir, "sprint.md")).Find("Only"); !item.Done {
		t.Fatal("item was not marked done")
	}
	assertResumeRewound(t, root, "implement", "items")
}

// Finding 1(b): a continuation at a nested loop node rewinds to the outer
// loop, whose arrival resolves the inner checklist again.
func TestResumeAtNestedLoopRewindsToOuterLoop(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "outer.md"), `---
items:
  - name: Chapter A
    check: Chapter A is delivered
    checklist: inner-a.md
---
`)
	writeFile(t, filepath.Join(workdir, "inner-a.md"), `---
items:
  - name: Sprint 1
    check: Sprint 1 holds
    command: "true"
---
`)
	pipeline := testGraph(
		startNode("start", "outer"),
		loopNode("outer", "outer.md", "plan", graph.Success, 0),
		customNode("plan", "task", []graph.Edge{{To: "inner"}}, 0),
		loopNode("inner", "", "work", "outer", 0),
		customNode("work", "task", []graph.Edge{{To: "inner"}}, 0),
	)
	var initial *Runner
	stopping := NewRegistry()
	stopping.Register("codergen", HandlerFunc(func(node graph.Node, _ []graph.Edge, _ ExecutionScope, _ *graph.Graph) (harness.Outcome, *harness.Error) {
		if node.Base().ID != "plan" {
			t.Fatalf("unexpected initial dispatch: %s", node.Base().ID)
		}
		initial.Stop()
		return harness.Outcome{Notes: "planned"}, nil
	}))
	initial = newLoopRunner(t, pipeline, stopping, root, workdir, nil)
	result, err := initial.Run()
	if err != nil || result.Status != RunFailed {
		t.Fatalf("initial result=%#v err=%v", result, err)
	}
	checkpoint := mustCheckpoint(t, root)
	if checkpoint.CurrentNode != "plan" || checkpoint.NextNode != "inner" || checkpoint.RetryVisit {
		t.Fatalf("checkpoint = %#v", checkpoint)
	}

	var dispatched []string
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(scope ExecutionScope) {
		dispatched = append(dispatched, filepath.Base(scope.StageDir))
	}))
	result, err = newLoopResumeRunner(t, pipeline, registry, root, workdir).Run()
	if err != nil || result.Status != RunCompleted {
		t.Fatalf("resumed result=%#v err=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(root, "stages", "000003-outer")); err != nil {
		t.Fatalf("resumed run did not start at the outer loop: %v", err)
	}
	// 3 outer, 4 plan, 5 inner, 6 work, 7 inner, 8 outer -> success
	if want := []string{"000004-plan", "000006-work"}; !reflect.DeepEqual(dispatched, want) {
		t.Fatalf("body dispatches = %v, want %v", dispatched, want)
	}
	work := readFile(t, filepath.Join(root, "stages", "000006-work", "prompt.md"))
	if !strings.Contains(work, `<iterate loop="outer" checklist="outer.md" item="1/1" lap="1">`) ||
		!strings.Contains(work, `<iterate loop="inner" checklist="inner-a.md" item="1/1" lap="1">`) {
		t.Fatalf("work prompt = %q", work)
	}
	for _, file := range []string{"outer.md", "inner-a.md"} {
		for _, item := range mustChecklist(t, filepath.Join(workdir, file)).Items {
			if !item.Done {
				t.Fatalf("%s item %q is not done", file, item.Name)
			}
		}
	}
	assertResumeRewound(t, root, "inner", "outer")
}

func TestResumeOutsideLoopDoesNotRewind(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Only
    check: It works
    command: "true"
---
`)
	pipeline := testGraph(
		startNode("start", "pre"),
		customNode("pre", "task", []graph.Edge{{To: "items"}}, 0),
		loopNode("items", "sprint.md", "implement", graph.Success, 0),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	var initial *Runner
	stopping := NewRegistry()
	stopping.Register("codergen", HandlerFunc(func(node graph.Node, _ []graph.Edge, _ ExecutionScope, _ *graph.Graph) (harness.Outcome, *harness.Error) {
		initial.Stop()
		return harness.Outcome{Notes: "prepared"}, nil
	}))
	initial = newLoopRunner(t, pipeline, stopping, root, workdir, nil)
	if result, err := initial.Run(); err != nil || result.Status != RunFailed {
		t.Fatalf("initial result=%#v err=%v", result, err)
	}
	registry := NewRegistry()
	registry.Register("codergen", bodyHandler(t, func(ExecutionScope) {}))
	if result, err := newLoopResumeRunner(t, pipeline, registry, root, workdir).Run(); err != nil || result.Status != RunCompleted {
		t.Fatalf("resumed result=%#v err=%v", result, err)
	}
	for _, event := range mustTimeline(t, root) {
		if event["type"] == "ResumeRewound" {
			t.Fatalf("resume outside a loop rewound: %v", event)
		}
	}
}

// Finding 4: evidence globs match relative to the workdir, so metacharacters
// in the workdir path are inert; directories are skipped; duplicates across
// globs collapse; absolute and escaping patterns are reported, not matched.
func TestMatchEvidence(t *testing.T) {
	workdir := filepath.Join(t.TempDir(), "run[1]")
	writeFile(t, filepath.Join(workdir, "notes.md"), "notes")
	writeFile(t, filepath.Join(workdir, "captures", "a.png"), "a")
	writeFile(t, filepath.Join(workdir, "captures", "b.png"), "b")
	writeFile(t, filepath.Join(workdir, "captures", "sub", "c.png"), "c")

	tests := []struct {
		name        string
		globs       []string
		wantFiles   []string
		wantInvalid []string
	}{
		{name: "metacharacter workdir", globs: []string{"notes.md"}, wantFiles: []string{"notes.md"}},
		{name: "directories skipped", globs: []string{"captures/*"}, wantFiles: []string{filepath.Join("captures", "a.png"), filepath.Join("captures", "b.png")}},
		{name: "overlapping globs deduplicate", globs: []string{"captures/*.png", "captures/a.png"}, wantFiles: []string{filepath.Join("captures", "a.png"), filepath.Join("captures", "b.png")}},
		{name: "no match", globs: []string{"missing/*.png"}, wantFiles: []string{}},
		{name: "invalid patterns", globs: []string{"/etc/*", "../notes.md", "captures/../notes.md", "[", "notes.md"}, wantFiles: []string{"notes.md"}, wantInvalid: []string{"/etc/*", "../notes.md", "captures/../notes.md", "["}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files, invalid := matchEvidence(workdir, test.globs)
			if !reflect.DeepEqual(files, test.wantFiles) || !reflect.DeepEqual(invalid, test.wantInvalid) {
				t.Fatalf("matchEvidence = %v, %v; want %v, %v", files, invalid, test.wantFiles, test.wantInvalid)
			}
		})
	}
}

func TestLoopInferRecordsInvalidPatternAsFailure(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Screens
    check: Every screen looks usable
    infer:
      files: ../captures/*.png
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
	if result.Status != RunFailed || len(backend.turns) != 0 {
		t.Fatalf("result = %#v after %d judge turns", result, len(backend.turns))
	}
	var record validationRecord
	readJSON(t, filepath.Join(root, "stages", "000003-items", "validation.json"), &record)
	if record.Passed || record.Summary != "invalid evidence pattern (must be relative to the workdir): ../captures/*.png" || record.Infer == nil || record.Infer.Verdict != "fail" {
		t.Fatalf("validation = %#v infer=%#v", record, record.Infer)
	}
}

func assertResumeRewound(t *testing.T, root, from, to string) {
	t.Helper()
	for _, event := range mustTimeline(t, root) {
		if event["type"] != "ResumeRewound" {
			continue
		}
		if event["from"] != from || event["to"] != to {
			t.Fatalf("ResumeRewound = %v, want from %q to %q", event, from, to)
		}
		return
	}
	t.Fatal("timeline has no ResumeRewound event")
}

func newLoopResumeRunner(t *testing.T, pipeline graph.Graph, registry *Registry, root, workdir string) *Runner {
	t.Helper()
	runner, err := ResumeRunner(pipeline, registry, RunnerConfig{
		LogsRoot:     root,
		Workdir:      workdir,
		Validate:     func(graph.Graph) error { return nil },
		DefaultModel: "gpt-5.3-codex",
	})
	if err != nil {
		t.Fatal(err)
	}
	return runner
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
