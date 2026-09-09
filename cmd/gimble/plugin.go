package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	gimbleMarketplaceName   = "gimble"
	gimbleMarketplaceSource = "tylergannon/gimble"
	gimblePluginSelector    = "gimble@gimble"
)

type pluginInstaller struct {
	run     func(string, ...string) ([]byte, error)
	homeDir func() (string, error)
	output  func(string, ...any)
}

type codexMarketplaceListing struct {
	Marketplaces []struct {
		Name string `json:"name"`
	} `json:"marketplaces"`
}

type codexPluginListing struct {
	Installed []struct {
		PluginID string `json:"pluginId"`
	} `json:"installed"`
}

type processSnapshot struct {
	PID     int
	PPID    int
	Command string
}

func newPluginCommand() *cobra.Command {
	plugin := &cobra.Command{Use: "plugin", Short: "Manage the Gimble Codex plugin"}
	plugin.AddCommand(newPluginInstallCommand())
	return plugin
}

func newPluginInstallCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "install",
		Short: "Install or cleanly reinstall the Gimble Codex plugin",
		Args:  cobra.NoArgs,
	}
	command.RunE = func(command *cobra.Command, _ []string) error {
		installer := pluginInstaller{
			run:     runPluginCommand,
			homeDir: os.UserHomeDir,
			output: func(format string, arguments ...any) {
				_, _ = fmt.Fprintf(command.OutOrStdout(), format, arguments...)
			},
		}
		return installer.install()
	}
	return command
}

func runPluginCommand(name string, arguments ...string) ([]byte, error) {
	command := exec.Command(name, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s %s: %w: %s", name, strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func (i pluginInstaller) install() error {
	marketplacesRaw, err := i.run("codex", "plugin", "marketplace", "list", "--json")
	if err != nil {
		return err
	}
	var marketplaces codexMarketplaceListing
	if err := json.Unmarshal(marketplacesRaw, &marketplaces); err != nil {
		return fmt.Errorf("decode Codex marketplace list: %w", err)
	}
	marketplaceExists := false
	for _, marketplace := range marketplaces.Marketplaces {
		if marketplace.Name == gimbleMarketplaceName {
			marketplaceExists = true
		}
	}
	if marketplaceExists {
		if _, err := i.run("codex", "plugin", "marketplace", "upgrade", gimbleMarketplaceName, "--json"); err != nil {
			return err
		}
	} else if _, err := i.run("codex", "plugin", "marketplace", "add", gimbleMarketplaceSource, "--json"); err != nil {
		return err
	}

	pluginsRaw, err := i.run("codex", "plugin", "list", "--json")
	if err != nil {
		return err
	}
	var plugins codexPluginListing
	if err := json.Unmarshal(pluginsRaw, &plugins); err != nil {
		return fmt.Errorf("decode Codex plugin list: %w", err)
	}
	for _, plugin := range plugins.Installed {
		if plugin.PluginID == gimblePluginSelector {
			if _, err := i.run("codex", "plugin", "remove", gimblePluginSelector, "--json"); err != nil {
				return err
			}
		}
	}
	if _, err := i.run("codex", "plugin", "add", gimblePluginSelector, "--json"); err != nil {
		return err
	}

	registeredPIDs := registeredMCPPIDs()
	retired, err := retireDetachedMCPServers()
	if err != nil {
		return err
	}
	idleRetired, activePreserved, err := retireUnregisteredMCPServers(registeredPIDs)
	if err != nil {
		return err
	}
	home, err := i.homeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory for plugin cache cleanup: %w", err)
	}
	cache := filepath.Join(home, ".cache", "gimble", "codex-plugin")
	if err := os.RemoveAll(cache); err != nil {
		return fmt.Errorf("remove Gimble plugin cache %q: %w", cache, err)
	}
	i.output("installed %s; retired %d superseded and %d idle MCP server(s); preserved detached runs and %d active run owner(s)\n",
		gimblePluginSelector, retired, idleRetired, activePreserved)
	i.output("start a new Codex task to use Gimble %s\n", gimbleMCPVersion)
	return nil
}

func registeredMCPPIDs() map[int]bool {
	registered := make(map[int]bool)
	store, err := defaultMCPRunStore()
	if err != nil {
		return registered
	}
	entries, err := os.ReadDir(filepath.Join(filepath.Dir(store.dir), "mcp-instances"))
	if err != nil {
		return registered
	}
	for _, entry := range entries {
		pid, err := strconv.Atoi(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		if err == nil && pid > 0 {
			registered[pid] = true
		}
	}
	return registered
}

func retireUnregisteredMCPServers(registered map[int]bool) (int, int, error) {
	output, err := exec.Command("ps", "-axo", "pid=,ppid=,command=").Output()
	if err != nil {
		return 0, 0, fmt.Errorf("list processes for MCP cleanup: %w", err)
	}
	processes := parseProcessSnapshot(string(output))
	return retireUnregisteredMCPProcesses(processes, registered, func(pid int) error {
		return syscall.Kill(pid, syscall.SIGKILL)
	})
}

func parseProcessSnapshot(output string) []processSnapshot {
	var processes []processSnapshot
	for line := range strings.SplitSeq(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, pidErr := strconv.Atoi(fields[0])
		ppid, ppidErr := strconv.Atoi(fields[1])
		if pidErr != nil || ppidErr != nil {
			continue
		}
		commandStart := strings.Index(line, fields[2])
		if commandStart < 0 {
			continue
		}
		processes = append(processes, processSnapshot{PID: pid, PPID: ppid, Command: strings.TrimSpace(line[commandStart:])})
	}
	return processes
}

func retireUnregisteredMCPProcesses(processes []processSnapshot, registered map[int]bool, kill func(int) error) (int, int, error) {
	children := make(map[int][]int)
	for _, process := range processes {
		children[process.PPID] = append(children[process.PPID], process.PID)
	}
	retired := 0
	preserved := 0
	var retireErrors []error
	for _, process := range processes {
		if registered[process.PID] || !isGimbleMCPCommand(process.Command) {
			continue
		}
		if hasProcessDescendants(process.PID, children) {
			preserved++
			continue
		}
		if err := kill(process.PID); err != nil && !errors.Is(err, syscall.ESRCH) {
			retireErrors = append(retireErrors, fmt.Errorf("retire idle MCP server %d: %w", process.PID, err))
			continue
		}
		retired++
	}
	return retired, preserved, errors.Join(retireErrors...)
}

func isGimbleMCPCommand(command string) bool {
	fields := strings.Fields(command)
	if len(fields) != 2 || fields[1] != "mcp" {
		return false
	}
	executable := filepath.Base(fields[0])
	return executable == "gimble" || strings.HasPrefix(executable, "gimble-")
}

func hasProcessDescendants(pid int, children map[int][]int) bool {
	return len(children[pid]) > 0
}

func retireDetachedMCPServers() (int, error) {
	store, err := defaultMCPRunStore()
	if err != nil {
		return 0, err
	}
	dir := filepath.Join(filepath.Dir(store.dir), "mcp-instances")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read MCP instance directory: %w", err)
	}
	retired := 0
	var cleanupErrors []error
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			cleanupErrors = append(cleanupErrors, readErr)
			continue
		}
		var instance mcpInstanceRecord
		if decodeErr := json.Unmarshal(raw, &instance); decodeErr != nil {
			_ = os.Remove(path)
			continue
		}
		if instance.PID <= 0 || instance.PID == os.Getpid() || !instance.DetachedRuns {
			continue
		}
		if !isProcessAlive(instance.PID) {
			_ = os.Remove(path)
			continue
		}
		if err := processOwnsMCPInstance(instance); err != nil {
			_ = os.Remove(path)
			continue
		}
		if err := syscall.Kill(instance.PID, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("stop superseded MCP server %d: %w", instance.PID, err))
			continue
		}
		deadline := time.Now().Add(time.Second)
		for isProcessAlive(instance.PID) && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
		if isProcessAlive(instance.PID) {
			if err := syscall.Kill(instance.PID, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("force-stop superseded MCP server %d: %w", instance.PID, err))
				continue
			}
		}
		_ = os.Remove(path)
		retired++
	}
	return retired, errors.Join(cleanupErrors...)
}

func processOwnsMCPInstance(instance mcpInstanceRecord) error {
	output, err := exec.Command("ps", "-p", fmt.Sprint(instance.PID), "-o", "command=").Output()
	if err != nil {
		return fmt.Errorf("inspect MCP server process %d: %w", instance.PID, err)
	}
	fields := strings.Fields(string(output))
	executableMatches := len(fields) >= 2 && (fields[0] == instance.ExecutablePath || filepath.Base(fields[0]) == filepath.Base(instance.ExecutablePath))
	if !executableMatches || fields[len(fields)-1] != "mcp" {
		return fmt.Errorf("process %d no longer matches its Gimble MCP instance record", instance.PID)
	}
	return nil
}
