package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

func commandHandler(node graph.Node, offered []graph.Edge, scope ExecutionScope, _ *graph.Graph) (harness.Outcome, *harness.Error) {
	command, ok := node.(*graph.CommandNode)
	if !ok {
		return harness.Outcome{}, terminalError("command handler received a non-command node")
	}
	if strings.TrimSpace(command.Command) == "" {
		return harness.Outcome{}, terminalError("no command on " + command.ID)
	}
	if scope.Stop != nil && scope.Stop.IsSet() {
		return harness.Outcome{}, interruptedError("command stopped by operator")
	}

	timeout, timeoutErr := commandTimeout(command)
	if timeoutErr != nil {
		return harness.Outcome{}, terminalError(timeoutErr.Error())
	}
	logPath := filepath.Join(scope.StageDir, "tool.log")
	exitCode, err := runShell(command.Command, scope.Workdir, logPath, timeout, command.Timeout.Present, scope.Stop)
	switch {
	case errors.Is(err, errShellStopped):
		return harness.Outcome{}, interruptedError("command stopped by operator")
	case errors.Is(err, errShellTimedOut):
		return harness.Outcome{}, interruptedError("command timed out")
	case err != nil:
		return harness.Outcome{}, terminalError(shellFailureMessage("command", err))
	}

	route, routeErr := commandRoute(command, exitCode)
	if routeErr != nil {
		return harness.Outcome{}, routeErr
	}
	if !offeredTarget(offered, route) {
		return harness.Outcome{}, terminalError(fmt.Sprintf("exit-code route %s has exhausted its visit budget", route))
	}
	tail, err := toolLogTail(logPath)
	if err != nil {
		return harness.Outcome{}, terminalError(fmt.Sprintf("read tool log: %v", err))
	}
	return harness.Outcome{
		Next:  route,
		Notes: fmt.Sprintf("exit %d: %s", exitCode, tail),
	}, nil
}

func commandTimeout(node *graph.CommandNode) (time.Duration, error) {
	if !node.Timeout.Present {
		return 0, nil
	}
	timeout, err := node.Timeout.Value.Parse()
	if err != nil {
		return 0, fmt.Errorf("parse command timeout: %w", err)
	}
	return timeout, nil
}

func commandRoute(node *graph.CommandNode, exitCode int) (string, *harness.Error) {
	if exitCode != 0 {
		if !node.Edges.Error.Present {
			return "", terminalError(fmt.Sprintf("%s exited %d", node.Command, exitCode))
		}
		return node.Edges.Error.Value, nil
	}
	return node.Edges.Success, nil
}

func offeredTarget(offered []graph.Edge, target string) bool {
	for _, edge := range offered {
		if edge.To == target {
			return true
		}
	}
	return false
}

func toolLogTail(path string) (string, error) {
	return logTail(path, 200)
}
