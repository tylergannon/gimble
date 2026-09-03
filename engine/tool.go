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

func toolHandler(node graph.Node, offered []graph.Edge, scope ExecutionScope, _ *graph.Graph) (harness.Outcome, *harness.Error) {
	tool, ok := node.(*graph.ToolNode)
	if !ok {
		return harness.Outcome{}, terminalError("tool handler received a non-tool node")
	}
	if strings.TrimSpace(tool.ToolCommand) == "" {
		return harness.Outcome{}, terminalError("no tool_command on " + tool.ID)
	}
	if scope.Stop != nil && scope.Stop.IsSet() {
		return harness.Outcome{}, interruptedError("tool command stopped by operator")
	}

	timeout, timeoutErr := toolTimeout(tool)
	if timeoutErr != nil {
		return harness.Outcome{}, terminalError(timeoutErr.Error())
	}
	logPath := filepath.Join(scope.StageDir, "tool.log")
	exitCode, err := runShell(tool.ToolCommand, scope.Workdir, logPath, timeout, tool.Timeout.Present, scope.Stop)
	switch {
	case errors.Is(err, errShellStopped):
		return harness.Outcome{}, interruptedError("tool command stopped by operator")
	case errors.Is(err, errShellTimedOut):
		return harness.Outcome{}, interruptedError("tool command timed out")
	case err != nil:
		return harness.Outcome{}, terminalError(shellFailureMessage("tool", err))
	}

	route, routeErr := toolRoute(tool, exitCode)
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

func toolTimeout(node *graph.ToolNode) (time.Duration, error) {
	if !node.Timeout.Present {
		return 0, nil
	}
	timeout, err := node.Timeout.Value.Parse()
	if err != nil {
		return 0, fmt.Errorf("parse tool timeout: %w", err)
	}
	return timeout, nil
}

func toolRoute(node *graph.ToolNode, exitCode int) (string, *harness.Error) {
	if exitCode != 0 {
		if !node.OnError.Present {
			return "", terminalError(fmt.Sprintf("%s exited %d", node.ToolCommand, exitCode))
		}
		return node.OnError.Value, nil
	}
	return node.OnSuccess, nil
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
