package hostwake

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CodexQueue wakes a Codex session through the public `codex queue` command.
// The Codex app-server starts the queued user message when a loaded thread is
// idle and retains it while the thread is busy.
type CodexQueue struct {
	// Binary is the Codex executable. Empty means resolve codex from PATH.
	Binary string
	// Timeout bounds one queue request. Zero means ten seconds.
	Timeout time.Duration
	// Run overrides command execution for tests.
	Run func(context.Context, string, ...string) ([]byte, error)
}

// NewCodexQueue returns a channel using the Codex CLI on PATH.
func NewCodexQueue() *CodexQueue { return &CodexQueue{} }

// Capture deliberately does nothing. A shared MCP server cannot infer which
// desktop thread called it; the launching parent supplies that identity.
func (*CodexQueue) Capture() (Session, bool) { return Session{}, false }

// Available validates the durable address and verifies that the CLI exists.
// Actual app-server reachability is reported by Wake without failing the run.
func (c *CodexQueue) Available(session Session) bool {
	if session.Host != HostCodexDesktop || strings.TrimSpace(session.ThreadID) == "" {
		return false
	}
	_, err := exec.LookPath(c.binary())
	return err == nil
}

// Wake queues message as user input for the explicitly identified thread.
func (c *CodexQueue) Wake(session Session, message string) error {
	if session.Host != HostCodexDesktop || strings.TrimSpace(session.ThreadID) == "" {
		return errors.New("no Codex desktop thread captured")
	}
	if len(message) > MaxMessageBytes {
		return fmt.Errorf("message is %d bytes, over the host's %d byte cap", len(message), MaxMessageBytes)
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	run := c.Run
	if run == nil {
		run = func(ctx context.Context, binary string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, binary, args...).CombinedOutput()
		}
	}
	output, err := run(ctx, c.binary(), "queue", "--thread", session.ThreadID, "--message", message)
	if err != nil {
		failure := strings.TrimSpace(string(output))
		if failure == "" {
			failure = err.Error()
		}
		return fmt.Errorf("queue Codex parent turn: %s", failure)
	}
	return nil
}

func (c *CodexQueue) binary() string {
	if binary := strings.TrimSpace(c.Binary); binary != "" {
		return binary
	}
	return "codex"
}
