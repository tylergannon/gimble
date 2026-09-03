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
	"github.com/tylergannon/tractor/lint"
)

const (
	defaultLoopJudgeModel           = "flash"
	defaultLoopJudgeReasoningEffort = "medium"
)

// loopHandler iterates a checklist file: after every lap it validates the
// framed item and every done item, reconciles done in both directions, and
// asks an evaluator whether the checklist body's definition of done is met.
// A not-done verdict re-reads the ledger and dispatches its first open item.
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
	var evaluationRecords []validationRecord
	evaluateAfterPass := false
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
				evaluationRecords = records
				evaluateAfterPass = true
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

	_, _, open := list.Open()
	if evaluateAfterPass || !open {
		if err := h.runner.writeFrames(); err != nil {
			return harness.Outcome{}, terminalError(fmt.Sprintf("write frames: %v", err))
		}
		verdict, evaluateErr := h.evaluate(loop, listPath, list, evaluationRecords, scope, pipeline)
		if evaluateErr != nil {
			return harness.Outcome{}, evaluateErr
		}
		if err := h.store.appendTimeline(timelineEvent{
			"type": "LoopEvaluated", "node": loop.ID, "verdict": verdict.Next, "notes": verdict.Notes,
		}); err != nil {
			return harness.Outcome{}, terminalError(err.Error())
		}

		// The evaluator is allowed to edit the checklist, so no decision uses
		// the copy loaded before its turn.
		list, err = checklist.Load(absolutePath)
		if err != nil {
			return harness.Outcome{}, terminalError(err.Error())
		}
		switch verdict.Next {
		case "done":
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
			return harness.Outcome{Next: loop.OnDone, Notes: fmt.Sprintf("%sevaluator decided done: %s", notes, verdict.Notes)}, nil
		case "not_done":
			if _, _, open = list.Open(); !open {
				return harness.Outcome{}, terminalError(fmt.Sprintf("loop evaluator returned not_done but checklist %q has no open item", listPath))
			}
		case "":
			return harness.Outcome{}, terminalError("loop evaluator returned no verdict")
		default:
			return harness.Outcome{}, terminalError(fmt.Sprintf("loop evaluator returned invalid verdict %q", verdict.Next))
		}
	}

	next, index, open := list.Open()
	if !open {
		// Defensive invariant: either this arrival already had an open item or a
		// not-done evaluator verdict was checked above.
		h.runner.popFrame(loop.ID)
		if err := h.runner.writeFrames(); err != nil {
			return harness.Outcome{}, terminalError(fmt.Sprintf("write frames: %v", err))
		}
		return harness.Outcome{}, terminalError(fmt.Sprintf("loop evaluator returned not_done but checklist %q has no open item", listPath))
	}

	current, stillFramed := h.runner.frameFor(loop.ID)
	selectedNewItem := hasFrame && (evaluateAfterPass || frame.item != next.Name)
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
	if selectedNewItem {
		bodyNodes, ok := lint.LoopBodyNodes(h.runner.graph, loop.ID)
		if !ok {
			return harness.Outcome{}, terminalError(fmt.Sprintf("loop body node-set not found for %q", loop.ID))
		}
		h.state.resetVisits(bodyNodes)
		offered, err = h.runner.offeredSuccessors(loop, h.state)
		if err != nil {
			return harness.Outcome{}, terminalError(err.Error())
		}
	}
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
		outcome, runErr := h.judge(loop, item, files, strings.TrimSuffix(logPath, ".log"), scope, pipeline)
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
func (h *loopHandler) judge(loop *graph.LoopNode, item checklist.Item, files []string, artifactPrefix string, scope ExecutionScope, pipeline *graph.Graph) (harness.Outcome, *harness.Error) {
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
	return handler.executeTurnAt(loop, fields, offered, turnScope, &judgePipeline, judgePrompt(item, files),
		artifactPrefix+"-prompt.md", artifactPrefix+"-response.md")
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

// evaluate runs one coding-harness turn that compares the checklist body's
// definition of done with the workspace. Unlike the infer judge, its model
// selection follows the pipeline defaults unless the evaluator-specific loop
// fields override them.
func (h *loopHandler) evaluate(loop *graph.LoopNode, listPath string, list *checklist.Checklist, records []validationRecord, scope ExecutionScope, pipeline *graph.Graph) (harness.Outcome, *harness.Error) {
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
		LLMModel:        loop.EvaluatorLLMModel,
		LLMProvider:     loop.EvaluatorLLMProvider,
		ReasoningEffort: loop.EvaluatorReasoningEffort,
	}
	offered := []graph.Edge{
		{To: "done", Condition: "The checklist definition of done is satisfied in the workspace."},
		{To: "not_done", Condition: "The definition of done is not yet satisfied; continue with the first open checklist item."},
	}
	if config.Backend == nil {
		// Simulation preserves the old useful traversal behavior while still
		// exercising the real two-target choice schema and artifact writes.
		if _, _, open := list.Open(); open {
			offered[0], offered[1] = offered[1], offered[0]
		}
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
	return handler.executeTurnAt(loop, fields, offered, turnScope, pipeline,
		evaluatorPrompt(listPath, list, records),
		filepath.Join(scope.StageDir, "evaluator-prompt.md"),
		filepath.Join(scope.StageDir, "evaluator-response.md"))
}

func evaluatorPrompt(listPath string, list *checklist.Checklist, records []validationRecord) string {
	byItem := make(map[string]validationRecord, len(records))
	for _, record := range records {
		byItem[record.Item] = record
	}

	var prompt strings.Builder
	prompt.WriteString("You are evaluating a checklist loop. Decide from the workspace and the checklist's prose definition of done whether the loop is done.\n")
	fmt.Fprintf(&prompt, "You may inspect the workspace and edit %s as ordinary work. If the definition is not met, leave at least one open item in the file by appending, reordering, or rewriting open items. Never rewrite a done item. If the existing first open item is already the right next step, leave the checklist unchanged.\n\n", listPath)
	prompt.WriteString("Checklist body (the definition of done):\n<definition-of-done>")
	prompt.WriteString(list.Body)
	if !strings.HasSuffix(list.Body, "\n") {
		prompt.WriteByte('\n')
	}
	prompt.WriteString("</definition-of-done>\n\nItems and their last validation results:\n")
	if len(list.Items) == 0 {
		prompt.WriteString("(no items)\n")
	}
	for index, item := range list.Items {
		status := "open"
		if item.Done {
			status = "done"
		}
		fmt.Fprintf(&prompt, "\nItem %d (%s):\n%s\n", index+1, status, item.Render())
		if record, ok := byItem[item.Name]; ok {
			fmt.Fprintf(&prompt, "last validation: passed=%t; summary=%s; log=%s\n", record.Passed, record.Summary, record.LogPath)
			if record.Infer != nil {
				fmt.Fprintf(&prompt, "infer verdict: %s; notes=%s; files=%s\n", record.Infer.Verdict, record.Infer.Notes, strings.Join(record.Infer.Files, ", "))
			}
		} else {
			prompt.WriteString("last validation: not run on this arrival\n")
		}
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
