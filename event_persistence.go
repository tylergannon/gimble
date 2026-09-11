package gimble

import (
	"encoding/json"
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
func (w *eventWriter) writeLifecycle(scope, session, turn string, event lifecycleEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seq++
	record := lifecycleRecord{
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
	record := agentRecord{
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
func (r *run) event(scope, session, turn string, event lifecycleEvent) {
	if r != nil && r.writer != nil {
		_ = r.writer.writeLifecycle(scope, session, turn, event)
	}
}
func (r *run) sessionEvent(scope, session, turn string, event AgentEvent) {
	if r == nil {
		return
	}
	r.mu.Lock()
	w := r.sessions[session]
	if w == nil {
		w, _ = newEventWriter(filepath.Join(r.dir, "sessions", session+".jsonl"))
		if w != nil {
			r.sessions[session] = w
		}
	}
	r.mu.Unlock()
	if w != nil {
		_ = w.writeAgent(scope, session, turn, event)
	}
}
func projectEvent(dir string, event lifecycleEvent) {
	w, err := newEventWriter(filepath.Join(dir, "project.jsonl"))
	if err == nil {
		_ = w.writeLifecycle("", "", "", event)
		_ = w.close()
	}
}
