package engine

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/lint"
)

func TestReadProcfileSelectsTheNamedProcessByName(t *testing.T) {
	workdir := t.TempDir()
	writeFile(t, filepath.Join(workdir, "Procfile.dev"), "worker: sleep 60\nweb: python3 server.py\n")

	definition, selected, err := readProcfile(workdir, "Procfile.dev", "web")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(definition.path, "Procfile.dev") || selected != 1 {
		t.Fatalf("definition = %#v, selected = %d", definition, selected)
	}
	if got := strings.Join(definition.processes, ","); got != "worker,web" {
		t.Fatalf("processes = %q", got)
	}
}

func TestReadProcfileRejectsAnUnknownNamedService(t *testing.T) {
	workdir := t.TempDir()
	writeFile(t, filepath.Join(workdir, "Procfile"), "web: sleep 60\n")
	_, _, err := readProcfile(workdir, "Procfile", "api")
	if err == nil || !strings.Contains(err.Error(), `service "api" is not declared`) {
		t.Fatalf("error = %v", err)
	}
}

type fakeProcessManager struct {
	state      ServiceState
	ensureErr  error
	downErr    error
	ensureCall int
	downCall   int
}

func (m *fakeProcessManager) Ensure(context.Context) (ServiceState, error) {
	m.ensureCall++
	return m.state, m.ensureErr
}

func (m *fakeProcessManager) Down(context.Context) error {
	m.downCall++
	return m.downErr
}

func TestServiceTeardownFailureDoesNotClaimTheServiceStopped(t *testing.T) {
	manager := &fakeProcessManager{
		state:   ServiceState{Name: "web", Port: 43123, Running: true},
		downErr: fmt.Errorf("still running"),
	}
	logsRoot := t.TempDir()
	runner, err := NewRunner(serviceTestGraph(&graph.CommandNode{
		ID: "probe", Command: "true",
		Edges: graph.CommandEdges{Success: graph.Success},
	}), NewRegistry(), RunnerConfig{
		LogsRoot: logsRoot, Workdir: t.TempDir(), Validate: validateTestGraph,
		ProcessManager: manager,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, runErr := runner.Run()
	if runErr == nil || result.Status != RunFailed || !strings.Contains(result.FailureReason, "still running") {
		t.Fatalf("result = %#v, error = %v", result, runErr)
	}
	state, err := LoadServiceState(logsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Running {
		t.Fatalf("service state falsely reports teardown success: %#v", state)
	}
}

func TestWorkflowServicePortReachesEveryCommandAndLivesForTheRun(t *testing.T) {
	manager := &fakeProcessManager{state: ServiceState{Name: "web", Port: 43123, Running: true}}
	pipeline := serviceTestGraph(
		&graph.CommandNode{ID: "first", Command: `test "$TRACTOR_SERVICE_PORT" = 43123`, Edges: graph.CommandEdges{Success: "second"}},
		&graph.CommandNode{ID: "second", Command: `test "$TRACTOR_SERVICE_PORT" = 43123`, Edges: graph.CommandEdges{Success: graph.Success}},
	)
	logsRoot := t.TempDir()
	runner, err := NewRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: logsRoot, Workdir: t.TempDir(), Validate: validateTestGraph,
		ProcessManager: manager,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted {
		serviceLog, _ := os.ReadFile(filepath.Join(logsRoot, "services.log"))
		t.Fatalf("status = %s: %s\n%s", result.Status, result.FailureReason, serviceLog)
	}
	if manager.ensureCall != 3 {
		t.Fatalf("Ensure calls = %d, want initial plus one per node", manager.ensureCall)
	}
	if manager.downCall != 1 {
		t.Fatalf("Down calls = %d, want 1", manager.downCall)
	}
	state, err := LoadServiceState(logsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if state.Name != "web" || state.Port != 43123 || state.Running {
		t.Fatalf("final service state = %#v", state)
	}
}

func TestServiceFailureStopsBeforeAWorkflowShellRuns(t *testing.T) {
	manager := &fakeProcessManager{
		state: ServiceState{Name: "web", Port: 43123}, ensureErr: fmt.Errorf("not healthy"),
	}
	workdir := t.TempDir()
	pipeline := serviceTestGraph(&graph.CommandNode{
		ID:      "must-not-run",
		Command: "touch ran",
		Edges:   graph.CommandEdges{Success: graph.Success},
	})
	runner, err := NewRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: t.TempDir(), Workdir: workdir, Validate: validateTestGraph,
		ProcessManager: manager,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunFailed || !strings.Contains(result.FailureReason, "not healthy") {
		t.Fatalf("result = %#v", result)
	}
	if _, err := os.Stat(filepath.Join(workdir, "ran")); !os.IsNotExist(err) {
		t.Fatalf("workflow shell ran despite service failure: %v", err)
	}
	if manager.downCall != 1 {
		t.Fatalf("Down calls = %d, want cleanup after failed readiness", manager.downCall)
	}
}

func TestOvermindRunsNamedProcfileServiceOnWorkflowPort(t *testing.T) {
	if _, err := exec.LookPath("overmind"); err != nil {
		t.Skip("overmind is not installed")
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux is not installed")
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed")
	}
	workdir := t.TempDir()
	server := `import os, socket
s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(("127.0.0.1", int(os.environ["PORT"])))
s.listen(8)
while True:
    c, _ = s.accept()
    c.recv(4096)
    c.sendall(b"HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nok")
    c.close()
`
	writeFile(t, filepath.Join(workdir, "server.py"), server)
	writeFile(t, filepath.Join(workdir, "Procfile.dev"), "worker: sleep 60\nweb: "+python+" server.py\n")

	pipeline := serviceTestGraph(&graph.CommandNode{
		ID:      "probe",
		Command: `python3 -c 'import os,urllib.request; p=os.environ["TRACTOR_SERVICE_PORT"]; assert urllib.request.urlopen("http://127.0.0.1:"+p).read() == b"ok"'`,
		Edges:   graph.CommandEdges{Success: graph.Success},
	})
	logsRoot := t.TempDir()
	runner, err := NewRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: logsRoot, Workdir: workdir, Validate: validateTestGraph,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != RunCompleted {
		serviceLog, _ := os.ReadFile(filepath.Join(logsRoot, "services.log"))
		t.Fatalf("status = %s: %s\n%s", result.Status, result.FailureReason, serviceLog)
	}
	state, err := LoadServiceState(logsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if state.Name != "web" || state.Port == 0 || state.Running {
		t.Fatalf("final service state = %#v", state)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(state.Port)))
	if err != nil {
		t.Fatalf("Overmind did not release service port %d: %v", state.Port, err)
	}
	_ = listener.Close()
}

func serviceTestGraph(nodes ...graph.Node) graph.Graph {
	return graph.Graph{
		SystemFile: jsonschema.Optional[string]{Present: true, Value: "Procfile.dev"},
		Services:   []string{"web"}, Start: nodes[0].Base().ID, Nodes: nodes,
	}
}

func validateTestGraph(candidate graph.Graph) error {
	_, err := lint.ValidateOrError(candidate)
	return err
}
