package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/checklist"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

// loopHandler iterates a checklist file: it validates the item the previous
// lap worked on, marks it done when the validation passes, selects the first
// open item as the lap's frame, and dispatches the body. With no open item
// left it routes to on_done.
type loopHandler struct {
	runner *Runner
	state  *engineState
	store  *runStore
}

// validationRecord is stages/{seq}-{loop}/validation.json.
type validationRecord struct {
	Item     string       `json:"item"`
	Command  string       `json:"command"`
	ExitCode int          `json:"exit_code"`
	LogTail  string       `json:"log_tail"`
	Infer    *inferRecord `json:"infer,omitempty"`
	Passed   bool         `json:"passed"`
	Summary  string       `json:"summary"`
}

type inferRecord struct {
	Files   []string `json:"files"`
	Verdict string   `json:"verdict"`
	Notes   string   `json:"notes"`
}

const validationLogTailRunes = 400

func (h *loopHandler) Execute(node graph.Node, offered []graph.Edge, scope ExecutionScope, pipeline *graph.Graph) (harness.Outcome, *harness.Error) {
	loop, ok := node.(*graph.LoopNode)
	if !ok {
		return harness.Outcome{}, terminalError(fmt.Sprintf("loop handler cannot execute node type %s", node.NodeType()))
	}
	if scope.Stop != nil && scope.Stop.IsSet() {
		return harness.Outcome{}, interruptedError("stopped by operator")
	}

	frame, hasFrame := h.runner.frameFor(loop.ID)
	listPath, resolveErr := h.resolveChecklist(loop, frame, hasFrame)
	if resolveErr != nil {
		return harness.Outcome{}, resolveErr
	}
	absolutePath := resolveWorkdirPath(scope.Workdir, listPath)
	list, err := checklist.Load(absolutePath)
	if err != nil {
		return harness.Outcome{}, terminalError(err.Error())
	}

	notes := ""
	if hasFrame {
		item, _, found := list.Find(frame.item)
		switch {
		case !found:
			notes = fmt.Sprintf("item vanished: %s; ", frame.item)
			h.runner.popFrame(loop.ID)
		case item.Done:
			h.runner.popFrame(loop.ID)
		default:
			record, validateErr := h.validate(loop, item, scope, pipeline)
			if validateErr != nil {
				return harness.Outcome{}, validateErr
			}
			if err := writeJSON(filepath.Join(scope.StageDir, "validation.json"), record); err != nil {
				return harness.Outcome{}, terminalError(fmt.Sprintf("write validation record: %v", err))
			}
			if err := h.store.appendTimeline(timelineEvent{
				"type": "LoopValidated", "node": loop.ID, "item": item.Name, "passed": record.Passed, "summary": record.Summary,
			}); err != nil {
				return harness.Outcome{}, terminalError(err.Error())
			}
			if record.Passed {
				if err := checklist.MarkDone(absolutePath, item.Name); err != nil {
					return harness.Outcome{}, terminalError(err.Error())
				}
				h.runner.popFrame(loop.ID)
			} else {
				h.runner.setFrameFailure(loop.ID, record.Summary)
			}
		}
		list, err = checklist.Load(absolutePath)
		if err != nil {
			return harness.Outcome{}, terminalError(err.Error())
		}
	}

	next, index, open := list.Open()
	if !open {
		h.runner.popFrame(loop.ID)
		if err := h.runner.writeFrames(); err != nil {
			return harness.Outcome{}, terminalError(fmt.Sprintf("write frames: %v", err))
		}
		if !offeredTarget(offered, loop.OnDone) {
			return harness.Outcome{}, terminalError(fmt.Sprintf("loop on_done %q has exhausted its visit budget", loop.OnDone))
		}
		if err := h.store.appendTimeline(timelineEvent{"type": "LoopCompleted", "node": loop.ID, "count": len(list.Items)}); err != nil {
			return harness.Outcome{}, terminalError(err.Error())
		}
		return harness.Outcome{Next: loop.OnDone, Notes: fmt.Sprintf("%schecklist complete: %d items done", notes, len(list.Items))}, nil
	}

	current, stillFramed := h.runner.frameFor(loop.ID)
	lap, lastFailure := 1, ""
	if stillFramed && current.item == next.Name {
		lap, lastFailure = current.lap+1, current.lastFailure
	}
	h.runner.pushOrReplaceFrame(loopFrame{
		loopID:        loop.ID,
		checklist:     listPath,
		item:          next.Name,
		itemChecklist: next.Checklist,
		rendered:      next.Render(),
		doc:           next.Doc,
		index:         index + 1,
		count:         len(list.Items),
		lap:           lap,
		lastFailure:   lastFailure,
	})
	if err := h.runner.writeFrames(); err != nil {
		return harness.Outcome{}, terminalError(fmt.Sprintf("write frames: %v", err))
	}
	if !offeredTarget(offered, loop.Body) {
		openCount := 0
		for _, item := range list.Items {
			if !item.Done {
				openCount++
			}
		}
		return harness.Outcome{}, terminalError(fmt.Sprintf("loop body %q has exhausted its visit budget with %d items open", loop.Body, openCount))
	}
	if err := h.store.appendTimeline(timelineEvent{
		"type": "LoopItemSelected", "node": loop.ID, "item": next.Name, "index": index + 1, "count": len(list.Items), "lap": lap,
	}); err != nil {
		return harness.Outcome{}, terminalError(err.Error())
	}
	return harness.Outcome{
		Next:  loop.Body,
		Notes: fmt.Sprintf("%sitem %d/%d: %s (lap %d)", notes, index+1, len(list.Items), next.Name, lap),
	}, nil
}

// resolveChecklist returns the workdir-relative checklist path: the node's
// own, the one this loop is already iterating, or the enclosing loop's
// current item's checklist field.
func (h *loopHandler) resolveChecklist(loop *graph.LoopNode, frame loopFrame, hasFrame bool) (string, *harness.Error) {
	if loop.Checklist.Present && strings.TrimSpace(loop.Checklist.Value) != "" {
		return loop.Checklist.Value, nil
	}
	if hasFrame {
		return frame.checklist, nil
	}
	enclosing, ok := h.runner.topFrame()
	if !ok {
		return "", terminalError(fmt.Sprintf("loop %s has no checklist and is not inside another loop's body", loop.ID))
	}
	if strings.TrimSpace(enclosing.itemChecklist) == "" {
		return "", terminalError(fmt.Sprintf("loop %s has no checklist and the enclosing item %q of loop %s names none", loop.ID, enclosing.item, enclosing.loopID))
	}
	return enclosing.itemChecklist, nil
}

// validate runs the item's command and then its infer judge, in the loop
// node's stage directory. An item with neither passes.
func (h *loopHandler) validate(loop *graph.LoopNode, item checklist.Item, scope ExecutionScope, pipeline *graph.Graph) (validationRecord, *harness.Error) {
	record := validationRecord{Item: item.Name, Command: item.Command}
	logPath := filepath.Join(scope.StageDir, "validation.log")
	if strings.TrimSpace(item.Command) != "" {
		timeout, hasTimeout, timeoutErr := loopTimeout(loop)
		if timeoutErr != nil {
			return record, terminalError(timeoutErr.Error())
		}
		exitCode, err := runShell(item.Command, scope.Workdir, logPath, timeout, hasTimeout, scope.Stop)
		switch {
		case errors.Is(err, errShellStopped):
			return record, interruptedError("validation command stopped by operator")
		case errors.Is(err, errShellTimedOut):
			return record, interruptedError("validation command timed out")
		case err != nil:
			return record, terminalError(shellFailureMessage("validation", err))
		}
		tail, err := logTail(logPath, validationLogTailRunes)
		if err != nil {
			return record, terminalError(fmt.Sprintf("read validation log: %v", err))
		}
		record.ExitCode = exitCode
		record.LogTail = tail
		if exitCode != 0 {
			record.Summary = fmt.Sprintf("exit %d — %s", exitCode, tail)
			return record, nil
		}
	} else if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		return record, terminalError(fmt.Sprintf("open validation log: %v", err))
	}

	if item.Infer != nil {
		files := matchEvidence(scope.Workdir, item.Infer.Files)
		if len(files) == 0 {
			record.Infer = &inferRecord{Files: []string{}, Verdict: "fail", Notes: "no evidence files matched"}
			record.Summary = "no evidence files matched " + strings.Join(item.Infer.Files, ", ")
			return record, nil
		}
		outcome, runErr := h.judge(loop, item, files, scope, pipeline)
		if runErr != nil {
			return record, runErr
		}
		verdict := "fail"
		if outcome.Next == "pass" {
			verdict = "pass"
		}
		record.Infer = &inferRecord{Files: files, Verdict: verdict, Notes: outcome.Notes}
		if verdict != "pass" {
			record.Summary = "judge: " + outcome.Notes
			return record, nil
		}
	}

	record.Passed = true
	record.Summary = "passed"
	return record, nil
}

// judge runs one codergen turn that decides whether the evidence files
// demonstrate the item's check.
func (h *loopHandler) judge(loop *graph.LoopNode, item checklist.Item, files []string, scope ExecutionScope, pipeline *graph.Graph) (harness.Outcome, *harness.Error) {
	config := h.runner.config
	handler := NewCodergenHandler(CodergenConfig{
		Backend:                config.Backend,
		DefaultModel:           config.DefaultModel,
		DefaultProvider:        config.DefaultProvider,
		DefaultReasoningEffort: config.DefaultReasoningEffort,
	})
	fields := &graph.LLMNodeFields{
		Fidelity:        jsonschema.Optional[string]{Present: true, Value: string(harness.FidelityNone)},
		Timeout:         loop.Timeout,
		LLMModel:        loop.LLMModel,
		LLMProvider:     loop.LLMProvider,
		ReasoningEffort: loop.ReasoningEffort,
	}
	offered := []graph.Edge{
		{To: "pass", Condition: "The evidence demonstrates the check."},
		{To: "fail", Condition: "The evidence does not demonstrate the check, or is missing."},
	}
	turnScope := scope
	turnScope.RunLog = ""
	if config.Backend != nil {
		segment, err := h.runner.runLogs.Allocate(loop.ID)
		if err != nil {
			return harness.Outcome{}, terminalError(fmt.Sprintf("allocate run log: %v", err))
		}
		turnScope.RunLog = segment.Path
	}
	return handler.executeTurn(loop, fields, offered, turnScope, pipeline, judgePrompt(item, files))
}

func judgePrompt(item checklist.Item, files []string) string {
	var prompt strings.Builder
	prompt.WriteString("You are validating one checklist item. Judge only what the evidence shows; do not fix anything.\n\n")
	fmt.Fprintf(&prompt, "item: %s\ncheck: %s\n%s\n\n", item.Name, item.Check, item.Infer.Prompt)
	prompt.WriteString("Evidence, open and inspect every file:")
	for _, file := range files {
		prompt.WriteString("\n- ")
		prompt.WriteString(file)
	}
	return prompt.String()
}

// matchEvidence expands each glob relative to workdir and returns the
// matches as workdir-relative paths, in glob order.
func matchEvidence(workdir string, globs []string) []string {
	files := []string{}
	for _, pattern := range globs {
		matches, err := filepath.Glob(resolveWorkdirPath(workdir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			if relative, err := filepath.Rel(workdir, match); err == nil {
				files = append(files, relative)
			} else {
				files = append(files, match)
			}
		}
	}
	return files
}

func loopTimeout(loop *graph.LoopNode) (time.Duration, bool, error) {
	if !loop.Timeout.Present {
		return 0, false, nil
	}
	timeout, err := loop.Timeout.Value.Parse()
	if err != nil {
		return 0, false, fmt.Errorf("parse loop timeout: %w", err)
	}
	return timeout, true, nil
}
