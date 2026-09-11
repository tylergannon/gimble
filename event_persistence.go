package gimble

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
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
func (w *eventWriter) write(e Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seq++
	e.Seq = w.seq
	e.Time = time.Now().UTC()
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.file.Write(b)
	return err
}
func (w *eventWriter) close() error { return w.file.Close() }
func (r *run) event(e Event) {
	if r != nil && r.writer != nil {
		_ = r.writer.write(e)
	}
}
func (r *run) sessionEvent(id string, e Event) {
	if r == nil {
		return
	}
	r.mu.Lock()
	w := r.sessions[id]
	if w == nil {
		w, _ = newEventWriter(filepath.Join(r.dir, "sessions", id+".jsonl"))
		if w != nil {
			r.sessions[id] = w
		}
	}
	r.mu.Unlock()
	if w != nil {
		_ = w.write(e)
	}
}
func projectEvent(dir string, e Event) {
	w, err := newEventWriter(filepath.Join(dir, "project.jsonl"))
	if err == nil {
		_ = w.write(e)
		_ = w.close()
	}
}
