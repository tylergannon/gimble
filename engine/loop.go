package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/checklist"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

const (
	defaultLoopJudgeModel           = "flash"
	defaultLoopJudgeReasoningEffort = "medium"
)

// loopHandler iterates a checklist file: after every lap it validates the
// framed item and every done item, reconciles done in both directions, selects
// the first open item as the next lap's frame, and dispatches the body. With no
// open item left it routes to on_done.
type loopHandler struct {
	runner *Runner
	state  *engineState
	store  *runStore
}

// validationReport is stages/{seq}-{loop}/validation.json.
type validationReport struct {
	Validations []validationRecord `json:"validations"`
}

type validationRecord struct {
	Item     string       `json:"item"`
	Command  string       `json:"command"`
	LogPath  string       `json:"log_path"`
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

// validationLogTailRunes bounds the record's log_tail; the head and tail
// bounds shape the failure summary the next lap's frame carries.
const (
	validationLogTailRunes     = 2000
	validationSummaryHeadRunes = 600
	validationSummaryTailRunes = 1400
)

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
		default:
			records := make([]validationRecord, 0, len(list.Items))
			allPassed := true
			for index, candidate := range list.Items {
				if !candidate.Done && candidate.Name != frame.item {
					continue
				}
				logPath := filepath.Join(scope.StageDir, fmt.Sprintf("validation-%03d.log", index+1))
				record, validateErr := h.validate(loop, candidate, logPath, scope, pipeline)
				if validateErr != nil {
					return harness.Outcome{}, validateErr
				}
				records = append(records, record)
				if !record.Passed {
					allPassed = false
				}
			}
			report := validationReport{Validations: records}
			if err := writeJSON(filepath.Join(scope.StageDir, "validation.json"), report); err != nil {
				return harness.Outcome{}, terminalError(fmt.Sprintf("write validation record: %v", err))
			}
			if err := h.store.appendTimeline(timelineEvent{
				"type": "LoopValidated", "node": loop.ID, "passed": allPassed, "validations": records,
			}); err != nil {
				return harness.Outcome{}, terminalError(err.Error())
			}

			failures := make([]loopValidationFailure, 0, len(records))
			for _, record := range records {
				if record.Passed {
					continue
				}
				if err := checklist.UnmarkDone(absolutePath, record.Item); err != nil {
					return harness.Outcome{}, terminalError(err.Error())
				}
				failures = append(failures, loopValidationFailure{
					item: record.Item, summary: record.Summary, logPath: record.LogPath,
				})
			}
			if allPassed {
				if err := checklist.MarkDone(absolutePath, item.Name); err != nil {
					return harness.Outcome{}, terminalError(err.Error())
				}
				h.runner.popFrame(loop.ID)
			} else {
				// A hand-marked framed item is only trusted for this lap. Keep it
				// open unless the whole validation set passed.
				if err := checklist.UnmarkDone(absolutePath, item.Name); err != nil {
					return harness.Outcome{}, terminalError(err.Error())
				}
				h.runner.setFrameFailures(loop.ID, failures)
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
	lap := 1
	var lastFailures []loopValidationFailure
	if stillFramed {
		lastFailures = current.lastFailures
		if current.item == next.Name {
			lap = current.lap + 1
		}
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
		lastFailures:  lastFailures,
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

// validate runs the item's command and then its infer judge. An item with
// neither passes. logPath is unique to this item within the loop stage.
func (h *loopHandler) validate(loop *graph.LoopNode, item checklist.Item, logPath string, scope ExecutionScope, pipeline *graph.Graph) (validationRecord, *harness.Error) {
	record := validationRecord{Item: item.Name, Command: item.Command, LogPath: logPath}
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
			excerpt, err := logExcerpt(logPath, validationSummaryHeadRunes, validationSummaryTailRunes)
			if err != nil {
				return record, terminalError(fmt.Sprintf("read validation log: %v", err))
			}
			record.Summary = fmt.Sprintf("exit %d — %s", exitCode, excerpt)
			return record, nil
		}
	} else if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		return record, terminalError(fmt.Sprintf("open validation log: %v", err))
	}

	if item.Infer != nil {
		files, invalid := matchEvidence(scope.Workdir, item.Infer.Files)
		if len(invalid) > 0 {
			record.Infer = &inferRecord{Files: files, Verdict: "fail", Notes: "invalid evidence pattern: " + strings.Join(invalid, ", ")}
			record.Summary = "invalid evidence pattern (must be relative to the workdir): " + strings.Join(invalid, ", ")
			return record, nil
		}
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
		DefaultModel:           defaultLoopJudgeModel,
		DefaultReasoningEffort: defaultLoopJudgeReasoningEffort,
	})
	judgePipeline := *pipeline
	judgePipeline.Defaults.LLMModel = jsonschema.Optional[string]{}
	judgePipeline.Defaults.LLMProvider = jsonschema.Optional[string]{}
	judgePipeline.Defaults.ReasoningEffort = jsonschema.Optional[string]{}
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
	return handler.executeTurn(loop, fields, offered, turnScope, &judgePipeline, judgePrompt(item, files))
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

// matchEvidence expands each glob relative to workdir (doublestar syntax;
// ** matches zero or more directories) and returns the regular files it
// matched as workdir-relative paths, in glob order, without duplicates. A
// leading ./ is ignored. Patterns that are absolute, escape the workdir, or
// fail to parse are returned as invalid rather than matched. Matching through
// an fs.FS keeps metacharacters in the workdir path itself inert.
func matchEvidence(workdir string, globs []string) (files, invalid []string) {
	files = []string{}
	seen := map[string]struct{}{}
	fsys := os.DirFS(workdir)
	for _, pattern := range globs {
		normalized := strings.TrimPrefix(pattern, "./")
		if filepath.IsAbs(pattern) || !fs.ValidPath(normalized) {
			invalid = append(invalid, pattern)
			continue
		}
		matches, err := doublestar.Glob(fsys, normalized)
		if err != nil {
			invalid = append(invalid, pattern)
			continue
		}
		for _, match := range matches {
			info, err := fs.Stat(fsys, match)
			if err != nil || info.IsDir() {
				continue
			}
			relative := filepath.FromSlash(match)
			if _, dup := seen[relative]; dup {
				continue
			}
			seen[relative] = struct{}{}
			files = append(files, relative)
		}
	}
	return files, invalid
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
