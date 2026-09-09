package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// loopFrame is one active loop's current item. The stack lives only in
// memory: the checklist file is the truth, and nothing here is checkpointed.
type loopFrame struct {
	loopID        string
	checklist     string // as authored or resolved from the enclosing item, relative to the workdir
	item          string
	itemChecklist string // the item's own checklist field, for a nested loop without one
	rendered      string // Item.Render() output captured at push time
	doc           string // the item's doc path, read at render time
	index         int    // 1-based
	count         int
	lap           int
	lastFailures  []loopValidationFailure
}

type loopValidationFailure struct {
	item    string
	summary string
	logPath string
}

// frameRecord is the observer-facing shape written to frames.json.
type frameRecord struct {
	Loop      string `json:"loop"`
	Checklist string `json:"checklist"`
	Item      string `json:"item"`
	Index     int    `json:"index"`
	Count     int    `json:"count"`
	Lap       int    `json:"lap"`
}

// frameFor returns the frame belonging to loopID, popping every frame above
// it: an inner loop bypassed by an escalation edge is abandoned.
func (r *Runner) frameFor(loopID string) (loopFrame, bool) {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	for index, frame := range r.frames {
		if frame.loopID == loopID {
			r.frames = r.frames[:index+1]
			return frame, true
		}
	}
	return loopFrame{}, false
}

// topFrame returns the innermost frame, whichever loop owns it.
func (r *Runner) topFrame() (loopFrame, bool) {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	if len(r.frames) == 0 {
		return loopFrame{}, false
	}
	return r.frames[len(r.frames)-1], true
}

// pushOrReplaceFrame replaces the top frame when it belongs to the same
// loop and pushes a new one otherwise.
func (r *Runner) pushOrReplaceFrame(frame loopFrame) {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	if top := len(r.frames) - 1; top >= 0 && r.frames[top].loopID == frame.loopID {
		r.frames[top] = frame
		return
	}
	r.frames = append(r.frames, frame)
}

// setFrameFailures records every failed validation and its log path on the
// loop's frame.
func (r *Runner) setFrameFailures(loopID string, failures []loopValidationFailure) {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	for index := range r.frames {
		if r.frames[index].loopID == loopID {
			r.frames[index].lastFailures = append([]loopValidationFailure(nil), failures...)
			return
		}
	}
}

// popFrame removes loopID's frame and everything above it.
func (r *Runner) popFrame(loopID string) {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	for index, frame := range r.frames {
		if frame.loopID == loopID {
			r.frames = r.frames[:index]
			return
		}
	}
}

func (r *Runner) snapshotFrames() []loopFrame {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	return append([]loopFrame(nil), r.frames...)
}

func frameRecords(frames []loopFrame) []frameRecord {
	records := make([]frameRecord, len(frames))
	for index, frame := range frames {
		records[index] = frameRecord{
			Loop:      frame.loopID,
			Checklist: frame.checklist,
			Item:      frame.item,
			Index:     frame.index,
			Count:     frame.count,
			Lap:       frame.lap,
		}
	}
	return records
}

// renderFrames renders every active frame for prompt injection, nested
// outermost to innermost so the block structure mirrors the loops. Doc
// files are read relative to workdir at render time so edits made during a
// lap reach the next turn. Empty when no loop is active.
func (r *Runner) renderFrames(workdir string) string {
	frames := r.snapshotFrames()
	if len(frames) == 0 {
		return ""
	}
	return framePreamble + "\n" + renderNested(frames, workdir, 0)
}

// framePreamble tells the agent what the iterate blocks are. It states
// engine facts only and never restates the node's own prompt.
const framePreamble = `<system-message>
This is one step of a Gimble run inside a checklist loop. The iterate
blocks below are the engine's record of where you are: the item selected
for this lap, the check it must satisfy, the command and judge that will
validate it when this step ends, and what the previous validation reported.
Outer blocks enclose inner ones. Paths are relative to the working
directory. Do only what your prompt asks; the engine marks items done.
</system-message>`

func renderNested(frames []loopFrame, workdir string, depth int) string {
	indent := strings.Repeat("  ", depth)
	frame := frames[0]
	var block strings.Builder
	fmt.Fprintf(&block, "%s<iterate loop=%q checklist=%q item=\"%d/%d\" lap=\"%d\">\n",
		indent, frame.loopID, frame.checklist, frame.index, frame.count, frame.lap)
	for line := range strings.SplitSeq(renderFrameBody(frame, workdir), "\n") {
		block.WriteString(indent)
		block.WriteString("  ")
		block.WriteString(line)
		block.WriteByte('\n')
	}
	if len(frames) > 1 {
		block.WriteString(renderNested(frames[1:], workdir, depth+1))
		block.WriteByte('\n')
	}
	block.WriteString(indent)
	block.WriteString("</iterate>")
	return block.String()
}

// renderFrameBody renders the lines between a frame's tags: the item, the
// last failed validations and their log paths if any, and the item's doc.
func renderFrameBody(frame loopFrame, workdir string) string {
	var body strings.Builder
	body.WriteString(frame.rendered)
	if len(frame.lastFailures) > 0 {
		body.WriteString("\nlast validation: failed")
		for _, failure := range frame.lastFailures {
			fmt.Fprintf(&body, "\n  - item: %s\n    summary: %s\n    validation log: %s",
				failure.item, strings.ReplaceAll(failure.summary, "\n", "\n      "), failure.logPath)
		}
	}
	if frame.doc != "" {
		contents, err := os.ReadFile(resolveWorkdirPath(workdir, frame.doc))
		if err != nil {
			fmt.Fprintf(&body, "\ndoc: %s (unreadable: %v)", frame.doc, err)
		} else {
			fmt.Fprintf(&body, "\n--- doc: %s ---\n%s", frame.doc, strings.TrimRight(string(contents), "\n"))
		}
	}
	return body.String()
}

// writeFrames rewrites {logs_root}/frames.json from the current stack.
func (r *Runner) writeFrames() error {
	return writeJSON(filepath.Join(r.config.LogsRoot, "frames.json"), frameRecords(r.snapshotFrames()))
}

// resolveWorkdirPath joins a checklist-relative path onto the workdir,
// leaving absolute paths alone.
func resolveWorkdirPath(workdir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workdir, path)
}
