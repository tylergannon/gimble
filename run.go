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

func runDir(ctx context.Context) string {
	s, _ := ctx.Value(scopeKey{}).(*scope)
	if s == nil || s.run == nil {
		return ""
	}
	return s.run.dir
}

// Project puts the project's directory in the root ctx. Runs are made
// under its runs directory.
func Project(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, projectKey{}, dir)
}

type run struct {
	dir       string // <project>/runs/<id>
	writer    *eventWriter
	project   *eventWriter
	mu        sync.Mutex
	sessions  map[string]*eventWriter
	errMu     sync.Mutex
	recordErr error
}

// Run starts one run of a workflow and blocks until the body returns. The
// run's ctx derives from the caller's, so main can put a deadline on it.
// The run is the root scope: when the body returns, its sessions are
// closed and its ctx is cancelled.
// The body must join its concurrent work before returning. Run then finishes
// every log it owns and returns the body's error joined with the first recording
// failure, if any; cancellation alone is not completion.
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
	r := &run{dir: dir, writer: w, sessions: make(map[string]*eventWriter)}
	pw, pwErr := newEventWriter(filepath.Join(project, "project.jsonl"))
	if pwErr != nil {
		r.recordFailure("open project log", pwErr)
	} else {
		r.project = pw
	}
	r.projectEvent(RunStarted{Name: name})
	r.event("", "", "", RunStarted{Name: name})
	logf("run %s started in %s", id, dir)
	err = (&scope{run: r}).do(ctx, body)
	if ctx.Err() != nil {
		cancelled := RunCancelled{Name: name, Source: steerSource(ctx), Error: ctx.Err().Error()}
		r.event("", "", "", cancelled)
		r.projectEvent(cancelled)
	}
	r.event("", "", "", RunEnded{Name: name, Error: errString(err)})
	r.projectEvent(RunEnded{Name: name, Error: errString(err)})
	r.closeSessions()
	if r.project != nil {
		r.recordFailure("close project log", r.project.close())
	}
	r.event("", "", "", Complete{RecordingError: errString(r.recordingError())})
	r.recordFailure("close run log", w.close())
	err = errors.Join(err, r.recordingError())
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
