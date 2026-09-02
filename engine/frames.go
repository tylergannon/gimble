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
	lastFailure   string
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

// setFrameFailure records the summary of a failed validation on the loop's frame.
func (r *Runner) setFrameFailure(loopID, summary string) {
	r.framesMu.Lock()
	defer r.framesMu.Unlock()
	for index := range r.frames {
		if r.frames[index].loopID == loopID {
			r.frames[index].lastFailure = summary
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

// renderFrames renders every active frame, outermost first, for prompt
// injection. Doc files are read relative to workdir at render time so edits
// made during a lap reach the next turn. Empty when no loop is active.
func (r *Runner) renderFrames(workdir string) string {
	frames := r.snapshotFrames()
	if len(frames) == 0 {
		return ""
	}
	blocks := make([]string, len(frames))
	for index, frame := range frames {
		blocks[index] = renderFrame(frame, workdir)
	}
	return strings.Join(blocks, "\n\n")
}

func renderFrame(frame loopFrame, workdir string) string {
	var block strings.Builder
	fmt.Fprintf(&block, "<tractor loop=%q checklist=%q item=\"%d/%d\" lap=\"%d\">\n",
		frame.loopID, frame.checklist, frame.index, frame.count, frame.lap)
	block.WriteString(frame.rendered)
	block.WriteByte('\n')
	if frame.lastFailure != "" {
		fmt.Fprintf(&block, "last validation: failed — %s\n", frame.lastFailure)
	}
	if frame.doc != "" {
		contents, err := os.ReadFile(resolveWorkdirPath(workdir, frame.doc))
		if err != nil {
			fmt.Fprintf(&block, "doc: %s (unreadable: %v)\n", frame.doc, err)
		} else {
			fmt.Fprintf(&block, "--- doc: %s ---\n", frame.doc)
			block.WriteString(strings.TrimRight(string(contents), "\n"))
			block.WriteByte('\n')
		}
	}
	block.WriteString("</tractor>")
	return block.String()
}

// writeFrames rewrites {logs_root}/frames.json from the current stack.
func (r *Runner) writeFrames() error {
	frames := r.snapshotFrames()
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
	return writeJSON(filepath.Join(r.config.LogsRoot, "frames.json"), records)
}

// resolveWorkdirPath joins a checklist-relative path onto the workdir,
// leaving absolute paths alone.
func resolveWorkdirPath(workdir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workdir, path)
}
