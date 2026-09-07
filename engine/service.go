package engine

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ServiceFile is the repository-level file naming how to start the software
// under test. Its contents go to /bin/sh, the same as any command node, so it
// can be a single line or a script. It is written once by a person and
// committed; nothing in a run authors it.
const ServiceFile = ".tractor/run"

// serviceLogName is the service's log within a loop stage directory.
const serviceLogName = "service.log"

const (
	serviceReadyTimeout = 90 * time.Second
	serviceReadyPoll    = 150 * time.Millisecond
	serviceDialTimeout  = 500 * time.Millisecond
	serviceTermGrace    = 5 * time.Second
	serviceStartTries   = 2
)

// errServiceNotReady is returned when the application never accepted a
// connection on the port the engine allocated for it. It is a validation
// failure, never a question put to a model.
var errServiceNotReady = errors.New("service never became ready")

// service is one running application under test: a process group the engine
// started, on a port the engine chose. Nothing about what is inside the group
// is known here, which is what keeps Tractor out of process supervision —
// `run` may be a single binary, `overmind start`, or `docker compose up`, and
// the engine cannot tell the difference.
type service struct {
	port    int
	url     string
	logPath string

	process *exec.Cmd
	logFile *os.File
	exited  chan struct{}
}

// env returns the variables handed to every command validated against this
// service. PORT is the twelve-factor spelling that most frameworks already
// honour; a tool with its own spelling is adapted in the service file itself,
// as in `OVERMIND_PORT=$PORT overmind start`.
func (s *service) env() []string {
	return []string{
		"PORT=" + strconv.Itoa(s.port),
		"TRACTOR_URL=" + s.url,
	}
}

// readServiceCommand returns the contents of the service file under workdir.
// The second result reports whether the file exists; a repository with no
// service file simply runs without one.
func readServiceCommand(workdir string) (string, bool, error) {
	raw, err := os.ReadFile(filepath.Join(workdir, ServiceFile))
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "", false, nil
	case err != nil:
		return "", false, fmt.Errorf("read %s: %w", ServiceFile, err)
	}
	command := strings.TrimSpace(string(raw))
	if command == "" {
		return "", false, fmt.Errorf("%s is empty", ServiceFile)
	}
	return command, true, nil
}

// freePort asks the kernel for an unused port and releases it. The window
// between that release and the application's own bind is a race, which is why
// startService retries: a lost race shows up as the process exiting before it
// ever accepts, and a fresh port is drawn for the next attempt.
func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("allocate port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return 0, fmt.Errorf("release allocated port: %w", err)
	}
	return port, nil
}

// startService runs command on a freshly allocated port and waits for that
// port to accept a connection. It retries on a new port when the process
// exits before becoming ready, which is how a lost port race recovers. A
// returned service is running and must be stopped.
func startService(command, workdir, logPath string, stop *StopSignal) (*service, error) {
	var lastErr error
	for attempt := range serviceStartTries {
		if stop != nil && stop.IsSet() {
			return nil, errShellStopped
		}
		started, err := launchService(command, workdir, logPath, attempt)
		if err != nil {
			return nil, err
		}
		readyErr := started.waitReady(serviceReadyTimeout, stop)
		if readyErr == nil {
			return started, nil
		}
		started.stop()
		if errors.Is(readyErr, errShellStopped) {
			return nil, readyErr
		}
		lastErr = readyErr
	}
	return nil, lastErr
}

// launchService allocates a port and starts one attempt in its own process
// group. attempt only distinguishes the log file, so a retry does not erase
// the reason the first try failed.
func launchService(command, workdir, logPath string, attempt int) (*service, error) {
	port, err := freePort()
	if err != nil {
		return nil, err
	}
	if attempt > 0 {
		logPath = fmt.Sprintf("%s.retry%d", logPath, attempt)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, &shellFailure{Op: "open log", Err: err}
	}

	started := &service{
		port:    port,
		url:     fmt.Sprintf("http://127.0.0.1:%d", port),
		logPath: logPath,
		logFile: logFile,
		exited:  make(chan struct{}),
	}

	process := exec.Command("/bin/sh", "-c", command)
	process.Dir = workdir
	process.Stdout = logFile
	process.Stderr = logFile
	process.Env = append(os.Environ(), started.env()...)
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := process.Start(); err != nil {
		_ = logFile.Close()
		return nil, &shellFailure{Op: "run", Err: err}
	}
	started.process = process

	go func() {
		_ = process.Wait()
		close(started.exited)
	}()
	return started, nil
}

// waitReady polls until the port accepts a connection. An exit before that
// happens is a failure in its own right: an application that could not bind
// almost always dies, so watching the process closes the port race far more
// cheaply than inspecting who holds the listener.
func (s *service) waitReady(timeout time.Duration, stop *StopSignal) error {
	deadline := time.Now().Add(timeout)
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port))
	for {
		select {
		case <-s.exited:
			return fmt.Errorf("%w: %s exited before accepting on port %d", errServiceNotReady, ServiceFile, s.port)
		default:
		}
		if stop != nil && stop.IsSet() {
			return errShellStopped
		}
		conn, err := net.DialTimeout("tcp", address, serviceDialTimeout)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: nothing accepted on port %d within %s", errServiceNotReady, s.port, timeout)
		}
		time.Sleep(serviceReadyPoll)
	}
}

// stop signals the whole process group and waits for it to go. SIGTERM first,
// because a supervisor such as `docker compose up` tears its own children
// down on it and would leave containers behind under SIGKILL; SIGKILL after
// the grace period, because leaving a survivor makes it the next lap's stale
// target. Safe to call more than once.
func (s *service) stop() {
	if s.process == nil || s.process.Process == nil {
		return
	}
	group := -s.process.Process.Pid
	if err := syscall.Kill(group, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		_ = syscall.Kill(group, syscall.SIGKILL)
	}
	select {
	case <-s.exited:
	case <-time.After(serviceTermGrace):
		_ = syscall.Kill(group, syscall.SIGKILL)
		<-s.exited
	}
	if s.logFile != nil {
		_ = s.logFile.Close()
		s.logFile = nil
	}
}

// serviceFailureSummary phrases a readiness failure for a validation record,
// with the tail of the service log so the reason is in front of whoever reads
// the run rather than buried in a file.
func serviceFailureSummary(err error, logPath string) string {
	summary := err.Error()
	if tail, tailErr := logTail(logPath, validationLogTailRunes); tailErr == nil && tail != "" {
		return summary + " — " + tail
	}
	return summary
}
