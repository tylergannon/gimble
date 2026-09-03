package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	errShellStopped  = errors.New("shell command stopped")
	errShellTimedOut = errors.New("shell command timed out")
)

// shellFailure is an infrastructure failure around a shell command: the log
// could not be opened or closed, or the command could not be started. A
// nonzero exit is not a shellFailure; it is reported through the exit code.
type shellFailure struct {
	Op  string // "open log", "close log", or "run"
	Err error
}

func (f *shellFailure) Error() string { return f.Op + ": " + f.Err.Error() }
func (f *shellFailure) Unwrap() error { return f.Err }

// runShell runs command through /bin/sh -c in workdir with stdout and stderr
// written to a fresh log file at logPath. The command runs in its own
// process group, which is killed when stop is set or, when hasTimeout is
// true, when timeout elapses. The returned error is errShellStopped,
// errShellTimedOut, or a *shellFailure; otherwise the exit code is returned.
func runShell(command, workdir, logPath string, timeout time.Duration, hasTimeout bool, stop *StopSignal) (int, error) {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, &shellFailure{Op: "open log", Err: err}
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	commandDone := make(chan struct{})
	if stop != nil {
		go func() {
			select {
			case <-stop.done:
				cancel(errShellStopped)
			case <-commandDone:
			}
		}()
	}
	var timer *time.Timer
	if hasTimeout {
		timer = time.AfterFunc(timeout, func() { cancel(errShellTimedOut) })
	}

	process := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	process.Dir = workdir
	process.Stdout = logFile
	process.Stderr = logFile
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	process.Cancel = func() error {
		if process.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-process.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	runErr := process.Run()
	close(commandDone)
	if timer != nil {
		timer.Stop()
	}
	cause := context.Cause(ctx)
	cancel(nil)
	closeErr := logFile.Close()

	switch {
	case errors.Is(cause, errShellStopped):
		return 0, errShellStopped
	case errors.Is(cause, errShellTimedOut):
		return 0, errShellTimedOut
	case closeErr != nil:
		return 0, &shellFailure{Op: "close log", Err: closeErr}
	}
	if runErr != nil {
		var exitError *exec.ExitError
		if !errors.As(runErr, &exitError) {
			return 0, &shellFailure{Op: "run", Err: runErr}
		}
		return exitError.ExitCode(), nil
	}
	return 0, nil
}

// logTail returns the last maxRunes runes of the trimmed file contents.
func logTail(path string, maxRunes int) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	tail := []rune(strings.TrimSpace(string(contents)))
	if len(tail) > maxRunes {
		tail = tail[len(tail)-maxRunes:]
	}
	return string(tail), nil
}

// logExcerpt returns the trimmed file contents when they fit in head+tail
// runes, and otherwise the first head runes and the last tail runes joined
// by a line saying how many were omitted.
func logExcerpt(path string, head, tail int) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := []rune(strings.TrimSpace(string(contents)))
	if len(text) <= head+tail {
		return string(text), nil
	}
	return fmt.Sprintf("%s\n… (%d runes omitted) …\n%s", string(text[:head]), len(text)-head-tail, string(text[len(text)-tail:])), nil
}

// shellFailureMessage phrases a runShell infrastructure error for a
// terminal error, naming the kind of command (tool, validation) that ran.
func shellFailureMessage(what string, err error) string {
	if failure, ok := errors.AsType[*shellFailure](err); ok {
		switch failure.Op {
		case "open log":
			return fmt.Sprintf("open %s log: %v", what, failure.Err)
		case "close log":
			return fmt.Sprintf("close %s log: %v", what, failure.Err)
		}
		return fmt.Sprintf("run %s command: %v", what, failure.Err)
	}
	return fmt.Sprintf("run %s command: %v", what, err)
}
