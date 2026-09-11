package gimble

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Task is one lap's work: the lap number, from 1, and the planner's text.
type Task struct {
	// Lap is the one-based number of this task in the loop.
	Lap int
	// Text is the complete next task supplied by the planner.
	Text string
}

// plan is the planner's answer each lap.
type plan struct {
	// The next task: one self-contained piece of work that moves the goal forward, written as complete instructions for an agent who has not seen the backlog. Leave it empty when the goal is met and nothing is left to do.
	Next string `json:"next"`
}

type loop struct {
	ctx     context.Context
	name    string
	goal    string
	planner *Session
	err     error
}

// Loop creates a planner-owned backlog loop. Range over its Laps field, then
// return its Err method's result:
//
//	loop := gimble.Loop(ctx, "delivery", goal, planner)
//	for ctx, task := range loop.Laps {
//		// Perform task.Text in the lap's ctx.
//	}
//	return loop.Err()
//
// The backlog is Markdown with YAML frontmatter containing the goal and steps
// with optional commands. It lives beneath the loop's scope in the run
// directory. Each lap reloads the file, runs its commands as facts for the
// planner, asks for one next task, and yields that task in a child scope. The
// loop ends when the planner returns no task.
func Loop(ctx context.Context, name, goal string, planner *Session) *loop {
	return &loop{ctx: ctx, name: name, goal: goal, planner: planner}
}

// Laps yields each lap's task with a ctx for the lap, a child scope of the
// loop's that ends when the body returns. Each lap it reloads the file,
// runs every step's command, and asks the planner what is next; when the
// planner names nothing, it returns. A file whose frontmatter does not
// parse runs no commands, and the planner is shown why so it can fix it.
// It never edits the file after writing the goal into it.
func (l *loop) Laps(yield func(context.Context, Task) bool) {
	parent, err := current(l.ctx)
	if err != nil {
		l.err = err
		return
	}
	l.err = parent.child(l.name).do(l.ctx, func(ctx context.Context) error {
		loopScope, _ := current(ctx)
		dir := filepath.Join(loopScope.run.dir, "scopes", filepath.FromSlash(loopScope.key))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("gimble: %w", err)
		}
		file := filepath.Join(dir, "backlog.md")
		front, err := yaml.Marshal(backlog{Goal: l.goal, Steps: []step{}})
		if err != nil {
			return fmt.Errorf("gimble: %w", err)
		}
		if err := os.WriteFile(file, []byte("---\n"+string(front)+"---\n"), 0o644); err != nil {
			return fmt.Errorf("gimble: %w", err)
		}
		for lap := 1; ; lap++ {
			raw, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("gimble: %w", err)
			}
			commands, bad := readBacklog(raw)
			if bad != nil {
				logf("%s: the backlog does not parse, so the planner is asked to fix it: %v", loopScope.key, bad)
			}
			results, err := runCommands(ctx, loopScope.run, loopScope.key, l.planner.workdir, commands)
			if err != nil {
				return err
			}
			p, err := l.planner.Generate[plan](ctx, planPrompt(l.name, file, l.planner.workdir, lap, string(raw), bad, results))
			if err != nil {
				return err
			}
			if strings.TrimSpace(p.Next) == "" {
				loopScope.run.event(Event{Kind: "planner_decision", Scope: loopScope.key, Decision: ""})
				logf("%s: the planner named nothing after %d laps", loopScope.key, lap-1)
				return nil
			}
			logf("%s: lap %d: %s", loopScope.key, lap, oneLine(p.Next))
			loopScope.run.event(Event{Kind: "planner_decision", Scope: loopScope.key, Decision: p.Next})
			more := true
			lapCtx := context.WithValue(ctx, taskKey{}, p.Next)
			_ = loopScope.child("lap").do(lapCtx, func(ctx context.Context) error {
				more = yield(ctx, Task{Lap: lap, Text: p.Next})
				return nil
			})
			if !more {
				return nil
			}
		}
	})
}

// Err returns the error that ended the laps, if any: a file that cannot be
// read or written, a command that would not start, or the planner's harness.
func (l *loop) Err() error {
	return l.err
}

type backlog struct {
	Goal  string `yaml:"goal"`
	Steps []step `yaml:"steps"`
}

type step struct {
	Step    string `yaml:"step"`
	Command string `yaml:"command,omitempty"`
}

// readBacklog returns the commands of the file's steps, or why its
// frontmatter does not parse.
func readBacklog(raw []byte) ([]string, error) {
	rest, ok := strings.CutPrefix(string(raw), "---\n")
	front, _, closed := strings.Cut(rest, "\n---")
	if !ok || !closed {
		return nil, errors.New("the file does not start with YAML frontmatter between --- lines")
	}
	var b backlog
	if err := yaml.Unmarshal([]byte(front), &b); err != nil {
		return nil, err
	}
	var commands []string
	for _, s := range b.Steps {
		if strings.TrimSpace(s.Command) != "" {
			commands = append(commands, s.Command)
		}
	}
	return commands, nil
}

type commandResult struct {
	command string
	code    int
	output  string
}

// runCommands runs each command in dir. A non-zero exit is a result, not an
// error; a command that cannot run is an error.
func runCommands(ctx context.Context, r *run, scope, dir string, commands []string) ([]commandResult, error) {
	var results []commandResult
	for _, command := range commands {
		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		code := 0
		if err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) || ctx.Err() != nil {
				return nil, fmt.Errorf("gimble: command %q: %w", command, err)
			}
			code = exit.ExitCode()
		}
		logf("$ %s: exit %d", command, code)
		r.event(Event{Kind: "loop_command", Scope: scope, Command: command, ExitCode: code})
		results = append(results, commandResult{command: command, code: code, output: tail(string(out), 3000)})
	}
	return results, nil
}

func tail(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return "[...]" + strings.ToValidUTF8(text[len(text)-limit:], "")
}

func planPrompt(name, file, workdir string, lap int, text string, bad error, results []commandResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are the planner of the loop %q, working in %s. Its backlog is the file %s, markdown with YAML frontmatter:\n\n", name, workdir, file)
	b.WriteString("- `goal` is the Definition of Done. Never change it.\n")
	b.WriteString("- `steps` is the plan: a list of items, each with `step:` saying what, and optionally `command:`, a shell command that checks it. Before asking you each lap, the loop runs every step's command in " + workdir + " and shows you the exit code and the end of the output.\n")
	b.WriteString("- Below the frontmatter, keep whatever notes help you plan.\n\n")
	b.WriteString("The loop never edits the file; you do. Rewrite it however the work shows it should be: add, remove, reorder, and rewrite steps. Keep the frontmatter valid YAML.\n\n")
	fmt.Fprintf(&b, "This is lap %d. The file now:\n\n%s\n\n", lap, text)
	if bad != nil {
		fmt.Fprintf(&b, "The frontmatter does not parse, so no command ran: %v\n\nFix the file first; quote any value that contains a colon.\n\n", bad)
	} else if len(results) == 0 {
		b.WriteString("No step has a command yet.\n\n")
	} else {
		b.WriteString("Command results:\n\n")
		for _, r := range results {
			fmt.Fprintf(&b, "$ %s\nexit %d\n%s\n\n", r.command, r.code, r.output)
		}
	}
	b.WriteString("Read the file and the workspace, update the file, then answer with the next task, or with an empty `next` when the goal is met and nothing is left to do.")
	return b.String()
}
