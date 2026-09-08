package engine

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

const (
	ServicePortEnv = "TRACTOR_SERVICE_PORT"

	serviceReadyTimeout = 90 * time.Second
	serviceReadyPoll    = 150 * time.Millisecond
	serviceDialTimeout  = 500 * time.Millisecond
	serviceDownTimeout  = 10 * time.Second
	serviceStateFile    = "services.json"
)

// ServiceState is the durable view of the one service a workflow may name.
type ServiceState struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Running bool   `json:"running"`
}

// ProcessManager is the lifecycle boundary between the engine and a real
// process manager. The initial adapter is Overmind over one Procfile.
type ProcessManager interface {
	Ensure(context.Context) (ServiceState, error)
	Down(context.Context) error
}

type procfile struct {
	path      string
	processes []string
}

func readProcfile(workdir, configuredPath, service string) (procfile, int, error) {
	path := filepath.Join(workdir, filepath.Clean(configuredPath))
	file, err := os.Open(path)
	if err != nil {
		return procfile{}, 0, fmt.Errorf("open system_file %s: %w", configuredPath, err)
	}
	defer func() { _ = file.Close() }()

	definition := procfile{path: path}
	selected := -1
	seen := map[string]struct{}{}
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		name, command, ok := strings.Cut(text, ":")
		name, command = strings.TrimSpace(name), strings.TrimSpace(command)
		if !ok || name == "" || command == "" {
			return procfile{}, 0, fmt.Errorf("parse system_file %s line %d: expected name: command", configuredPath, line)
		}
		if _, duplicate := seen[name]; duplicate {
			return procfile{}, 0, fmt.Errorf("parse system_file %s line %d: duplicate process %q", configuredPath, line, name)
		}
		seen[name] = struct{}{}
		if name == service {
			selected = len(definition.processes)
		}
		definition.processes = append(definition.processes, name)
	}
	if err := scanner.Err(); err != nil {
		return procfile{}, 0, fmt.Errorf("read system_file %s: %w", configuredPath, err)
	}
	if len(definition.processes) == 0 {
		return procfile{}, 0, fmt.Errorf("system_file %s declares no processes", configuredPath)
	}
	if selected < 0 {
		return procfile{}, 0, fmt.Errorf("service %q is not declared in system_file %s", service, configuredPath)
	}
	return definition, selected, nil
}

type overmindManager struct {
	workdir     string
	definition  procfile
	service     string
	servicePort int
	basePort    int
	socket      string
	logPath     string
	readyFor    time.Duration
}

func newOvermindManager(workdir, logsRoot, configuredPath, service string) (*overmindManager, error) {
	if _, err := exec.LookPath("overmind"); err != nil {
		return nil, fmt.Errorf("workflow declares a service but overmind is not installed: %w", err)
	}
	definition, selected, err := readProcfile(workdir, configuredPath, service)
	if err != nil {
		return nil, err
	}
	basePort, err := freePortRange(len(definition.processes))
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(logsRoot))
	socket := filepath.Join(os.TempDir(), fmt.Sprintf("tractor-overmind-%x.sock", digest[:8]))
	return &overmindManager{
		workdir: workdir, definition: definition, service: service,
		basePort: basePort, servicePort: basePort + selected, socket: socket,
		logPath: filepath.Join(logsRoot, "services.log"), readyFor: serviceReadyTimeout,
	}, nil
}

func freePortRange(size int) (int, error) {
	for range 20 {
		seed, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return 0, fmt.Errorf("allocate service port: %w", err)
		}
		base := seed.Addr().(*net.TCPAddr).Port
		listeners := []net.Listener{seed}
		available := base+size <= 65535
		for offset := 1; available && offset < size; offset++ {
			listener, listenErr := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(base+offset)))
			if listenErr != nil {
				available = false
				break
			}
			listeners = append(listeners, listener)
		}
		for _, listener := range listeners {
			_ = listener.Close()
		}
		if available {
			return base, nil
		}
	}
	return 0, errors.New("allocate consecutive ports for Procfile processes")
}

func (m *overmindManager) Ensure(ctx context.Context) (ServiceState, error) {
	running := m.instanceRunning(ctx)
	if !running {
		_ = os.Remove(m.socket)
		if err := m.start(ctx); err != nil {
			return ServiceState{Name: m.service, Port: m.servicePort}, err
		}
	}
	if m.portAccepts() {
		return ServiceState{Name: m.service, Port: m.servicePort, Running: true}, nil
	}
	if running {
		if output, err := m.command(ctx, "restart", m.service); err != nil {
			return ServiceState{Name: m.service, Port: m.servicePort}, fmt.Errorf("restart service %q: %w: %s", m.service, err, strings.TrimSpace(output))
		}
		if err := m.waitReady(ctx); err != nil {
			return ServiceState{Name: m.service, Port: m.servicePort}, err
		}
	}
	return ServiceState{Name: m.service, Port: m.servicePort, Running: true}, nil
}

func (m *overmindManager) start(ctx context.Context) error {
	logFile, err := os.OpenFile(m.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open service log: %w", err)
	}
	defer func() { _ = logFile.Close() }()
	command := exec.CommandContext(ctx, "overmind", "start", "--daemonize", "--any-can-die",
		"--socket", m.socket, "--procfile", m.definition.path, "--root", m.workdir,
		"--port", strconv.Itoa(m.basePort), "--port-step", "1")
	command.Dir = m.workdir
	command.Stdout, command.Stderr = logFile, logFile
	if err := command.Run(); err != nil {
		return fmt.Errorf("start Overmind: %w (see %s)", err, m.logPath)
	}
	return m.waitReady(ctx)
}

func (m *overmindManager) waitReady(ctx context.Context) error {
	started := time.Now()
	deadline := time.Now().Add(m.readyFor)
	for {
		if m.portAccepts() {
			return nil
		}
		if time.Since(started) > 2*time.Second && !m.instanceRunning(ctx) {
			return fmt.Errorf("service %q exited before accepting connections on port %d (see %s)", m.service, m.servicePort, m.logPath)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("service %q did not accept connections on port %d within %s (see %s)", m.service, m.servicePort, m.readyFor, m.logPath)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(serviceReadyPoll):
		}
	}
}

func (m *overmindManager) portAccepts() bool {
	connection, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(m.servicePort)), serviceDialTimeout)
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func (m *overmindManager) instanceRunning(ctx context.Context) bool {
	_, err := m.command(ctx, "status")
	return err == nil
}

func (m *overmindManager) command(ctx context.Context, verb string, args ...string) (string, error) {
	arguments := []string{verb, "--socket", m.socket}
	arguments = append(arguments, args...)
	command := exec.CommandContext(ctx, "overmind", arguments...)
	command.Dir = m.workdir
	output, err := command.CombinedOutput()
	return string(output), err
}

func (m *overmindManager) Down(ctx context.Context) error {
	if !m.instanceRunning(ctx) {
		_ = os.Remove(m.socket)
		return nil
	}
	output, err := m.command(ctx, "quit")
	if err != nil {
		_ = os.Remove(m.socket)
		return fmt.Errorf("quit Overmind: %w: %s", err, strings.TrimSpace(output))
	}
	for m.portAccepts() {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for service %q to stop: %w", m.service, ctx.Err())
		case <-time.After(serviceReadyPoll):
		}
	}
	_ = os.Remove(m.socket)
	return nil
}

func serviceManagerFor(pipeline graph.Graph, workdir, logsRoot string) (ProcessManager, error) {
	if len(pipeline.Services) == 0 {
		return nil, nil
	}
	return newOvermindManager(workdir, logsRoot, pipeline.SystemFile.Value, pipeline.Services[0])
}

func (r *Runner) prepareService(store *runStore) (func(), error) {
	manager := r.config.ProcessManager
	if manager == nil {
		var err error
		manager, err = serviceManagerFor(r.graph, r.config.Workdir, r.config.LogsRoot)
		if err != nil {
			return func() {}, err
		}
	}
	if manager == nil {
		return func() {}, nil
	}
	r.processManager = manager
	state, err := r.ensureServiceState(store, r.startID, 0)
	if err != nil {
		return func() {}, err
	}
	previous, existed := os.LookupEnv(ServicePortEnv)
	if err := os.Setenv(ServicePortEnv, strconv.Itoa(state.Port)); err != nil {
		return func() {}, fmt.Errorf("set %s: %w", ServicePortEnv, err)
	}
	return func() {
		if existed {
			_ = os.Setenv(ServicePortEnv, previous)
		} else {
			_ = os.Unsetenv(ServicePortEnv)
		}
	}, nil
}

func (r *Runner) ensureService(nodeID string, stage uint64, store *runStore) *harness.Error {
	if r.processManager == nil {
		return nil
	}
	state, err := r.ensureServiceState(store, nodeID, stage)
	if err != nil {
		return terminalError(fmt.Sprintf("ensure service for %s: %v", nodeID, err))
	}
	if current := os.Getenv(ServicePortEnv); current != strconv.Itoa(state.Port) {
		if err := os.Setenv(ServicePortEnv, strconv.Itoa(state.Port)); err != nil {
			return terminalError(fmt.Sprintf("set %s: %v", ServicePortEnv, err))
		}
	}
	return nil
}

func (r *Runner) ensureServiceState(store *runStore, nodeID string, stage uint64) (ServiceState, error) {
	r.serviceMu.Lock()
	defer r.serviceMu.Unlock()
	started := time.Now()
	ctx, cancel := r.serviceContext()
	defer cancel()
	state, err := r.processManager.Ensure(ctx)
	if writeErr := writeServiceState(store.root, state); writeErr != nil && err == nil {
		err = writeErr
	}
	event := timelineEvent{
		"type": "ServiceChecked", "node": nodeID, "stage": stage,
		"service": state.Name, "port": state.Port, "running": state.Running,
		"duration": time.Since(started).String(),
	}
	if err != nil {
		event["error"] = err.Error()
	}
	if eventErr := store.appendTimeline(event); eventErr != nil && err == nil {
		err = eventErr
	}
	return state, err
}

func (r *Runner) downService(store *runStore) error {
	if r.processManager == nil {
		return nil
	}
	r.serviceMu.Lock()
	defer r.serviceMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), serviceDownTimeout)
	defer cancel()
	err := r.processManager.Down(ctx)
	state, loadErr := LoadServiceState(store.root)
	if loadErr == nil && err == nil {
		state.Running = false
		if writeErr := writeServiceState(store.root, state); writeErr != nil {
			err = writeErr
		}
	}
	event := timelineEvent{"type": "ServiceDown"}
	if err != nil {
		event["error"] = err.Error()
	}
	if eventErr := store.appendTimeline(event); eventErr != nil && err == nil {
		err = eventErr
	}
	return err
}

func (r *Runner) serviceContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-r.stop.done:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

func writeServiceState(root string, state ServiceState) error {
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode service state: %w", err)
	}
	raw = append(raw, '\n')
	temporary := filepath.Join(root, ".services.json.tmp")
	if err := os.WriteFile(temporary, raw, 0o644); err != nil {
		return fmt.Errorf("write service state: %w", err)
	}
	if err := os.Rename(temporary, filepath.Join(root, serviceStateFile)); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("replace service state: %w", err)
	}
	return nil
}

// LoadServiceState reads the most recently observed service state for status
// surfaces. A workflow without a service has no state file.
func LoadServiceState(root string) (ServiceState, error) {
	raw, err := os.ReadFile(filepath.Join(root, serviceStateFile))
	if err != nil {
		return ServiceState{}, err
	}
	var state ServiceState
	if err := json.Unmarshal(raw, &state); err != nil {
		return ServiceState{}, fmt.Errorf("decode service state: %w", err)
	}
	return state, nil
}
