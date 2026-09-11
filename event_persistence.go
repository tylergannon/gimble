package gimble

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tylergannon/polytype"
)

type eventWriter struct {
	mu   sync.Mutex
	file *os.File
	seq  uint64
}

func newEventWriter(name string) (*eventWriter, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &eventWriter{file: f}, nil
}
func (w *eventWriter) writeLifecycle(scope, session, turn string, event LifecycleEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seq++
	record := LifecycleRecord{
		Seq:     w.seq,
		Time:    time.Now().UTC(),
		Scope:   scope,
		Session: optionalString(session),
		Turn:    optionalString(turn),
		Event:   event,
	}
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.file.Write(b)
	return err
}

func (w *eventWriter) writeAgent(scope, session, turn string, event AgentEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seq++
	record := AgentRecord{
		Seq:     w.seq,
		Time:    time.Now().UTC(),
		Scope:   scope,
		Session: session,
		Turn:    turn,
		Event:   event,
	}
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.file.Write(b)
	return err
}

func optionalString(value string) polytype.Optional[string] {
	return polytype.Optional[string]{Present: value != "", Value: value}
}

func (w *eventWriter) close() error { return w.file.Close() }

func (r *run) recordFailure(operation string, err error) {
	if r == nil || err == nil {
		return
	}
	r.errMu.Lock()
	defer r.errMu.Unlock()
	if r.recordErr == nil {
		r.recordErr = fmt.Errorf("gimble: record: %s: %w", operation, err)
	}
}

func (r *run) recordingError() error {
	if r == nil {
		return nil
	}
	r.errMu.Lock()
	defer r.errMu.Unlock()
	return r.recordErr
}

func (r *run) event(scope, session, turn string, event LifecycleEvent) {
	if r != nil && r.writer != nil {
		r.recordFailure("write run log", r.writer.writeLifecycle(scope, session, turn, event))
	}
}

func (r *run) projectEvent(event LifecycleEvent) {
	if r != nil && r.project != nil {
		r.recordFailure("write project log", r.project.writeLifecycle("", "", "", event))
	}
}

func (r *run) sessionEvent(scope, session, turn string, event AgentEvent) {
	if r == nil {
		return
	}
	r.mu.Lock()
	w := r.sessions[session]
	if w == nil {
		var err error
		w, err = newEventWriter(filepath.Join(r.dir, "sessions", session+".jsonl"))
		if err != nil {
			r.recordFailure("open session log "+session, err)
		}
		if w != nil {
			r.sessions[session] = w
		}
	}
	r.mu.Unlock()
	if w != nil {
		r.recordFailure("write session log "+session, w.writeAgent(scope, session, turn, event))
	}
}

func (r *run) closeSessions() {
	r.mu.Lock()
	writers := make(map[string]*eventWriter, len(r.sessions))
	for session, writer := range r.sessions {
		writers[session] = writer
	}
	r.mu.Unlock()
	for session, writer := range writers {
		r.recordFailure("close session log "+session, writer.close())
	}
}
