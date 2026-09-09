package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPluginInstallerUpgradesAndReinstallsBeforeCleaningLegacyCache(t *testing.T) {
	stateRoot := t.TempDir()
	t.Setenv(mcpRunStateEnv, filepath.Join(stateRoot, "mcp-runs"))
	home := t.TempDir()
	legacyCaches := []string{
		filepath.Join(home, ".cache", "tractor", "codex-plugin"),
		filepath.Join(home, ".cache", "gimble", "codex-plugin"),
	}
	for _, legacyCache := range legacyCaches {
		if err := os.MkdirAll(legacyCache, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(legacyCache, "old-wrapper"), []byte("old"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	var calls []string
	run := func(name string, arguments ...string) ([]byte, error) {
		call := strings.Join(append([]string{name}, arguments...), " ")
		calls = append(calls, call)
		switch call {
		case "codex plugin marketplace list --json":
			return []byte(`{"marketplaces":[{"name":"gimble"},{"name":"tractor"}]}`), nil
		case "codex plugin marketplace upgrade gimble --json":
			return []byte(`{}`), nil
		case "codex plugin list --json":
			return []byte(`{"installed":[{"pluginId":"gimble@gimble"},{"pluginId":"tractor@tractor"}]}`), nil
		case "codex plugin remove gimble@gimble --json", "codex plugin remove tractor@tractor --json", "codex plugin add gimble@gimble --json", "codex plugin marketplace remove tractor --json":
			return []byte(`{}`), nil
		default:
			return nil, fmt.Errorf("unexpected command %q", call)
		}
	}
	var output strings.Builder
	installer := pluginInstaller{
		run: run, homeDir: func() (string, error) { return home, nil },
		output: func(format string, arguments ...any) { _, _ = fmt.Fprintf(&output, format, arguments...) },
	}
	if err := installer.install(); err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{
		"codex plugin marketplace list --json",
		"codex plugin marketplace upgrade gimble --json",
		"codex plugin list --json",
		"codex plugin remove gimble@gimble --json",
		"codex plugin remove tractor@tractor --json",
		"codex plugin add gimble@gimble --json",
		"codex plugin marketplace remove tractor --json",
	}
	if !slices.Equal(calls, wantCalls) {
		t.Fatalf("commands = %#v, want %#v", calls, wantCalls)
	}
	for _, legacyCache := range legacyCaches {
		if _, err := os.Stat(legacyCache); !os.IsNotExist(err) {
			t.Fatalf("legacy cache %q still exists: %v", legacyCache, err)
		}
	}
	if !strings.Contains(output.String(), "preserved detached runs") {
		t.Fatalf("installer output = %q", output.String())
	}
}

func TestRetireDetachedMCPServersDoesNotTouchUnregisteredRuns(t *testing.T) {
	stateRoot := t.TempDir()
	t.Setenv(mcpRunStateEnv, filepath.Join(stateRoot, "mcp-runs"))
	instanceDir := filepath.Join(stateRoot, "mcp-instances")
	if err := os.MkdirAll(instanceDir, 0o700); err != nil {
		t.Fatal(err)
	}

	server := exec.Command("/bin/sh", "-c", "trap 'exit 0' TERM; while :; do sleep 1; done", "mcp")
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	serverDone := make(chan struct{})
	go func() { _ = server.Wait(); close(serverDone) }()
	t.Cleanup(func() { _ = server.Process.Kill() })
	instance := mcpInstanceRecord{
		PID: server.Process.Pid, Version: gimbleMCPVersion, StartedAt: time.Now().UTC(),
		DetachedRuns: true, ExecutablePath: "/bin/sh",
	}
	raw := fmt.Appendf(nil, "{\"pid\":%d,\"version\":%q,\"started_at\":%q,\"detached_runs\":true,\"executable_path\":%q}\n",
		instance.PID, instance.Version, instance.StartedAt.Format(time.RFC3339Nano), instance.ExecutablePath)
	if err := os.WriteFile(filepath.Join(instanceDir, fmt.Sprintf("%d.json", instance.PID)), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	run := exec.Command("/bin/sh", "-c", "while :; do sleep 1; done")
	run.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := run.Start(); err != nil {
		t.Fatal(err)
	}
	runDone := make(chan struct{})
	go func() { _ = run.Wait(); close(runDone) }()
	t.Cleanup(func() { _ = syscall.Kill(-run.Process.Pid, syscall.SIGKILL) })

	retired, err := retireDetachedMCPServers()
	if err != nil {
		t.Fatal(err)
	}
	if retired != 1 {
		t.Fatalf("retired = %d, want 1", retired)
	}
	select {
	case <-serverDone:
	case <-time.After(2 * time.Second):
		t.Fatal("registered MCP server was not retired")
	}
	if !processExists(run.Process.Pid) {
		t.Fatal("unregistered detached run was stopped during MCP cleanup")
	}
	_ = syscall.Kill(-run.Process.Pid, syscall.SIGKILL)
	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("detached run was not reaped after test cleanup")
	}
}

func TestRetireLegacyMCPProcessesKillsOnlyIdleUnregisteredServers(t *testing.T) {
	processes := []processSnapshot{
		{PID: 10, PPID: 1, Command: "/old/cache/gimble-abc mcp"},
		{PID: 11, PPID: 1, Command: "gimble mcp"},
		{PID: 12, PPID: 11, Command: "gimble run pipeline.json"},
		{PID: 13, PPID: 1, Command: "/usr/local/bin/gimble mcp"},
		{PID: 14, PPID: 1, Command: "gimble run pipeline.json"},
		{PID: 15, PPID: 1, Command: "/old/cache/tractor-abc mcp"},
	}
	var killed []int
	retired, preserved, err := retireLegacyMCPProcesses(processes, map[int]bool{13: true}, func(pid int) error {
		killed = append(killed, pid)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if retired != 2 || preserved != 1 || !slices.Equal(killed, []int{10, 15}) {
		t.Fatalf("retired=%d preserved=%d killed=%v", retired, preserved, killed)
	}
}

func TestParseProcessSnapshotPreservesCommand(t *testing.T) {
	processes := parseProcessSnapshot("  42    1 gimble mcp\n  43   42 /bin/sh -c sleep 30\n")
	want := []processSnapshot{
		{PID: 42, PPID: 1, Command: "gimble mcp"},
		{PID: 43, PPID: 42, Command: "/bin/sh -c sleep 30"},
	}
	if !slices.Equal(processes, want) {
		t.Fatalf("processes = %#v, want %#v", processes, want)
	}
}
