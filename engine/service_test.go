package engine

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tylergannon/tractor/checklist"
	"github.com/tylergannon/tractor/graph"
)

// writeServiceFile puts a service command in a fresh workdir and returns it.
func writeServiceFile(t *testing.T, command string) string {
	t.Helper()
	workdir := t.TempDir()
	writeServiceFileIn(t, workdir, command)
	return workdir
}

func writeServiceFileIn(t *testing.T, workdir, command string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(workdir, ".tractor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workdir, ServiceFile), []byte(command), 0o644); err != nil {
		t.Fatal(err)
	}
}

// stubServer is a minimal HTTP server that honours PORT. It deliberately
// avoids http.server, whose server_bind calls getfqdn and can block on DNS for
// tens of seconds — which the readiness wait would correctly report as an
// application that had not come up.
const stubServer = `import os, socket
sock = socket.socket()
sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
sock.bind(("127.0.0.1", int(os.environ["PORT"])))
sock.listen(8)
while True:
    conn, _ = sock.accept()
    conn.recv(4096)
    conn.sendall(b"HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nok")
    conn.close()
`

// newStubApp writes a repository that starts the stub server and returns its
// workdir and the service command.
func newStubApp(t *testing.T) (string, string) {
	t.Helper()
	requirePython(t)
	workdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workdir, "server.py"), []byte(stubServer), 0o644); err != nil {
		t.Fatal(err)
	}
	command := "exec python3 server.py"
	writeServiceFileIn(t, workdir, command)
	return workdir, command
}

// requirePython skips when no python3 is available to stand in for a real
// application under test.
func requirePython(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available to stand in for an application")
	}
	return path
}

// A repository that declares nothing runs exactly as it always has.
func TestReadServiceCommandAbsentIsNotAnError(t *testing.T) {
	t.Parallel()
	command, present, err := readServiceCommand(t.TempDir())
	if err != nil {
		t.Fatalf("absent service file should not error: %v", err)
	}
	if present || command != "" {
		t.Errorf("got present=%v command=%q, want absent", present, command)
	}
}

// An empty file is a mistake worth naming rather than silently ignoring.
func TestReadServiceCommandRejectsEmptyFile(t *testing.T) {
	t.Parallel()
	_, _, err := readServiceCommand(writeServiceFile(t, "   \n\t\n"))
	if err == nil {
		t.Fatal("expected an error for an empty service file")
	}
	if !strings.Contains(err.Error(), ServiceFile) {
		t.Errorf("error should name the file, got %q", err)
	}
}

func TestFreePortReturnsABindablePort(t *testing.T) {
	t.Parallel()
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatalf("port %d was not bindable after allocation: %v", port, err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
}

// The engine picks the port, hands it over as PORT, and the application that
// honours it is reachable at TRACTOR_URL. This is the whole guarantee: the
// target was chosen here, not found by whoever validates it.
func TestStartServiceRunsTheApplicationOnTheEnginesPort(t *testing.T) {
	workdir, _ := newStubApp(t)
	logPath := filepath.Join(t.TempDir(), "service.log")

	command, present, err := readServiceCommand(workdir)
	if err != nil || !present {
		t.Fatalf("read service command: %v present=%v", err, present)
	}

	app, err := startService(command, workdir, logPath, nil)
	if err != nil {
		t.Fatalf("start service: %v", err)
	}
	t.Cleanup(app.stop)

	if app.port == 0 {
		t.Fatal("service has no port")
	}
	if want := fmt.Sprintf("http://127.0.0.1:%d", app.port); app.url != want {
		t.Errorf("url = %q, want %q", app.url, want)
	}

	response, err := http.Get(app.url)
	if err != nil {
		t.Fatalf("the application the engine started did not answer: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", response.StatusCode)
	}

	env := app.env()
	wantPort := "PORT=" + strconv.Itoa(app.port)
	if len(env) != 2 || env[0] != wantPort || env[1] != "TRACTOR_URL="+app.url {
		t.Errorf("env = %v, want [%s TRACTOR_URL=%s]", env, wantPort, app.url)
	}
}

// Stopping releases the port. A survivor would become the next lap's stale
// target, which is the failure the whole design exists to prevent.
func TestStopReleasesThePort(t *testing.T) {
	workdir, command := newStubApp(t)

	app, err := startService(command, workdir, filepath.Join(t.TempDir(), "service.log"), nil)
	if err != nil {
		t.Fatalf("start service: %v", err)
	}
	port := app.port
	app.stop()

	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	deadline := time.Now().Add(5 * time.Second)
	for {
		listener, err := net.Listen("tcp", address)
		if err == nil {
			if closeErr := listener.Close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("port %d still held after stop: %v", port, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// stop is called from a defer and again on teardown, so it must not block or
// panic the second time.
func TestStopIsIdempotent(t *testing.T) {
	workdir, command := newStubApp(t)
	app, err := startService(command, workdir, filepath.Join(t.TempDir(), "service.log"), nil)
	if err != nil {
		t.Fatalf("start service: %v", err)
	}
	app.stop()
	app.stop()
}

// An application that never accepts is a failure, and the reason it failed
// reaches the record rather than staying in a log nobody opens.
func TestStartServiceFailsWhenTheApplicationExits(t *testing.T) {
	t.Parallel()
	logPath := filepath.Join(t.TempDir(), "service.log")
	_, err := startService("echo 'boom: port already in use' >&2; exit 1", t.TempDir(), logPath, nil)
	if !errors.Is(err, errServiceNotReady) {
		t.Fatalf("err = %v, want errServiceNotReady", err)
	}

	// The retry drew a second port, so the last attempt's log carries the reason.
	summary := serviceFailureSummary(err, fmt.Sprintf("%s.retry%d", logPath, serviceStartTries-1))
	if !strings.Contains(summary, "boom") {
		t.Errorf("summary should carry the service's own output, got %q", summary)
	}
}

// A readiness failure must never be reported as a passing or inconclusive
// validation: every item in the set fails, and no judge is consulted.
func TestReadinessFailureIsNotReadyNotAnAbsenceOfFindings(t *testing.T) {
	t.Parallel()
	_, err := startService("exit 3", t.TempDir(), filepath.Join(t.TempDir(), "service.log"), nil)
	if err == nil {
		t.Fatal("a service that exits immediately must not be reported ready")
	}
	if !errors.Is(err, errServiceNotReady) {
		t.Fatalf("err = %v, want errServiceNotReady so the loop fails the set", err)
	}
}

// An operator stop reaches the readiness wait rather than blocking for the
// full timeout.
func TestStartServiceHonoursTheStopSignal(t *testing.T) {
	t.Parallel()
	stop := NewStopSignal()
	stop.Stop()

	start := time.Now()
	_, stoppedErr := startService("sleep 60", t.TempDir(), filepath.Join(t.TempDir(), "service.log"), stop)
	if !errors.Is(stoppedErr, errShellStopped) {
		t.Fatalf("err = %v, want errShellStopped", stoppedErr)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("stop took %s to take effect", elapsed)
	}
}

// The whole point, end to end: the engine starts the application, and the
// item's own command reaches it at the URL the engine chose. Nothing in the
// checklist names a port, and nothing had to find the application.
func TestLoopValidatesAgainstTheApplicationTheEngineStarted(t *testing.T) {
	requirePython(t)
	root, workdir := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(workdir, "server.py"), stubServer)
	writeServiceFileIn(t, workdir, "exec python3 server.py")
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Serves
    check: The application answers on the port the engine allocated
    command: python3 -c "import os,urllib.request; assert urllib.request.urlopen(os.environ['TRACTOR_URL']).read() == b'ok'"
---
Done when the application answers.
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 4),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	registry := NewRegistry()
	registry.Register("agent", bodyHandler(t, func(ExecutionScope) {}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted {
		t.Fatalf("status = %v (%s), want success — the item command should have reached the service", result.Status, result.FailureReason)
	}

	list, err := checklist.Load(filepath.Join(workdir, "sprint.md"))
	if err != nil {
		t.Fatal(err)
	}
	if item, _, ok := list.Find("Serves"); !ok || !item.Done {
		t.Errorf("item Serves done = %v, want true", ok && item.Done)
	}
}

// An application that never comes up fails the item. The command would have
// passed on its own; what fails it is that there was nothing to observe.
func TestLoopFailsEveryItemWhenTheApplicationNeverStarts(t *testing.T) {
	root, workdir := t.TempDir(), t.TempDir()
	writeServiceFileIn(t, workdir, "echo 'could not start' >&2; exit 1")
	writeFile(t, filepath.Join(workdir, "sprint.md"), `---
items:
  - name: Would pass alone
    check: A command that says nothing about the application
    command: "true"
---
Done when the application answers.
`)
	pipeline := testGraph(
		startNode("start", "items"),
		loopNode("items", "sprint.md", "implement", graph.Success, 2),
		customNode("implement", "task", []graph.Edge{{To: "items"}}, 0),
	)
	registry := NewRegistry()
	registry.Register("agent", bodyHandler(t, func(ExecutionScope) {}))

	result, err := newLoopRunner(t, pipeline, registry, root, workdir, nil).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status == RunCompleted {
		t.Fatal("a run whose application never started must not succeed")
	}

	list, err := checklist.Load(filepath.Join(workdir, "sprint.md"))
	if err != nil {
		t.Fatal(err)
	}
	if item, _, ok := list.Find("Would pass alone"); ok && item.Done {
		t.Error("item was marked done although the application never ran")
	}
}
