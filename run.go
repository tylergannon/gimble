package gimble

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type projectKey struct{}

// Project puts the project's directory in the root ctx. Runs are made
// under its runs directory.
func Project(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, projectKey{}, dir)
}

type run struct {
	dir      string // <project>/runs/<id>
	writer   *eventWriter
	project  *eventWriter
	mu       sync.Mutex
	sessions map[string]*eventWriter
}

// Run starts one run of a workflow and blocks until the body returns. The
// run's ctx derives from the caller's, so main can put a deadline on it.
// The run is the root scope: when the body returns, its sessions are
// closed and its ctx is cancelled.
func Run(ctx context.Context, name string, body func(ctx context.Context) error) error {
	project, _ := ctx.Value(projectKey{}).(string)
	if project == "" {
		return errors.New("gimble: Run needs gimble.Project in its ctx")
	}
	project, err := filepath.Abs(project)
	if err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	id := time.Now().Format("20060102-150405") + "." + name
	dir := filepath.Join(project, "runs", id)
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	w, err := newEventWriter(filepath.Join(dir, "run.jsonl"))
	if err != nil {
		return fmt.Errorf("gimble: %w", err)
	}
	pw, _ := newEventWriter(filepath.Join(project, "project.jsonl"))
	r := &run{dir: dir, writer: w, project: pw, sessions: make(map[string]*eventWriter)}
	if pw != nil {
		_ = pw.write(Event{Kind: "run_started", Name: name})
	}
	r.event(Event{Kind: "run_started", Name: name})
	logf("run %s started in %s", id, dir)
	err = (&scope{run: r}).do(ctx, body)
	if ctx.Err() != nil {
		r.event(Event{Kind: "run_cancelled", Error: ctx.Err().Error()})
		if pw != nil {
			_ = pw.write(Event{Kind: "run_cancelled", Name: name, Error: ctx.Err().Error()})
		}
	}
	r.event(Event{Kind: "run_ended", Name: name, Error: errString(err)})
	if pw != nil {
		_ = pw.write(Event{Kind: "run_ended", Name: name, Error: errString(err)})
	}
	r.event(Event{Kind: "complete"})
	_ = w.close()
	if pw != nil {
		_ = pw.close()
	}
	logf("run %s ended: %v", id, orNone(err))
	return err
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// logf traces what the runtime does on stderr until the run log exists.
func logf(format string, args ...any) {
	log.Printf("gimble: "+format, args...)
}

func orNone(err error) any {
	if err == nil {
		return "ok"
	}
	return err
}
