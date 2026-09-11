// Package sprint is the sprint workflow. It builds one sprint of
// ephemeral/research/api/SPRINTS.md in the repository it runs in. A
// researcher reads the code once; a planner forked from it keeps the
// backlog; each lap a coder forked from it does the planner's task under a
// scope reviewer and a taste reviewer, and the lap is committed when the
// checks pass.
package sprint

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tylergannon/gimble"
	"github.com/tylergannon/gimble/claude"
	"github.com/tylergannon/gimble/codex"
)

//go:generate go tool polytype --validate

// Input starts a sprint.
type Input struct {
	// The sprint to build: its number in ephemeral/research/api/SPRINTS.md, e.g. 2.
	Sprint int `json:"sprint"`
	// Absolute path of the repository. Each lap is committed to the branch checked out there.
	Repo string `json:"repo"`
	// Codex model for the researcher, the planner, and the coders, e.g. "gpt-5.6-luna".
	Model string `json:"model"`
	// Claude Code model for the reviewers, e.g. "haiku".
	ReviewModel string `json:"review_model"`
	// The most laps to run, done or not.
	Laps int `json:"laps"`
}

var checks = []string{"go vet ./...", "go test ./..."}

// Sprint builds sprint in.Sprint.
func Sprint(ctx context.Context, in Input) error {
	if err := gimble.SetJSON(ctx, "input", in); err != nil {
		return err
	}
	plan, err := os.ReadFile(filepath.Join(in.Repo, "ephemeral/research/api/SPRINTS.md"))
	if err != nil {
		return err
	}
	goal, ok := section(string(plan), in.Sprint)
	if !ok {
		return fmt.Errorf("sprint: SPRINTS.md has no Sprint %d", in.Sprint)
	}
	goal += "\n\nDone when all of the above is committed in " + in.Repo + " and `go vet ./...` and `go test ./...` exit 0 there. Give those two checks steps of their own, with those commands."

	cx, cl := codex.New(), claude.New()
	researcher := gimble.NewSession(ctx, "researcher", cx, in.Model, in.Repo)
	if _, err := researcher.Generate[gimble.Text](ctx, researchPrompt+"\n\n"+goal); err != nil {
		return err
	}
	planner, err := researcher.Fork(ctx, "planner")
	if err != nil {
		return err
	}

	loop := gimble.Loop(ctx, "sprint", goal, planner)
	for ctx, task := range loop.Laps {
		if task.Lap > in.Laps {
			break
		}
		if err := lap(ctx, in, researcher, cl, task); err != nil {
			return err
		}
	}
	return loop.Err()
}

// lap does one task: a coder forked from the researcher, supervised, then
// a commit if the checks pass.
func lap(ctx context.Context, in Input, researcher *gimble.Session, cl gimble.HarnessAdapter, task gimble.Task) error {
	if err := gimble.Set(ctx, "task", task.Text); err != nil {
		return err
	}
	coder, err := researcher.Fork(ctx, "coder")
	if err != nil {
		return err
	}
	scope := gimble.NewSession(ctx, "scope", cl, in.ReviewModel, in.Repo)
	taste := gimble.NewSession(ctx, "taste", cl, in.ReviewModel, in.Repo)
	_, err = coder.Generate[gimble.Text](ctx, codePrompt+"\n\n"+gimble.ScopeText(ctx),
		gimble.Supervise(scope, scopeInstruction),
		gimble.Supervise(taste, tasteInstruction),
	)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		// An unapproved or failed turn is a fact for the planner, not the end
		// of the sprint: its work stays in the tree for the next lap to see.
		log.Printf("sprint: lap %d: %v", task.Lap, err)
	}

	// Work that fails the checks stays uncommitted; the planner sees the
	// failure in its command results next lap.
	for _, check := range checks {
		cmd := exec.CommandContext(ctx, "sh", "-c", check)
		cmd.Dir = in.Repo
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("sprint: lap %d: %s failed, so the lap stays uncommitted:\n%s", task.Lap, check, out)
			return nil
		}
	}
	if _, err := git(ctx, in.Repo, "add", "-A"); err != nil {
		return err
	}
	if status, _ := git(ctx, in.Repo, "status", "--porcelain"); status == "" {
		log.Printf("sprint: lap %d: nothing to commit", task.Lap)
		return nil
	}
	message := fmt.Sprintf("Sprint %d, lap %d: %s\n\n%s", in.Sprint, task.Lap, firstLine(task.Text), task.Text)
	if _, err := git(ctx, in.Repo, "commit", "-m", message); err != nil {
		return err
	}
	log.Printf("sprint: lap %d: committed", task.Lap)
	return nil
}

const researchPrompt = `You are about to lead the build of one sprint on this repository, Gimble, a Go library. Read AGENTS.md, ephemeral/research/api/API.md, ephemeral/research/api/SPRINTS.md, and the code the sprint below touches, until you know where everything it needs is. Change no files. Answer with a short summary of what exists and what the sprint needs.`

const codePrompt = `Do the task below in this repository, following AGENTS.md and ephemeral/research/api/API.md, and build only what it asks. Make "go vet ./..." and "go test ./..." pass. Do not commit: the workflow commits the lap when those pass. Answer with a short summary of what you changed.`

const (
	scopeInstruction = "Scope: the agent builds only what its task asks, and nothing that ephemeral/research/api/API.md does not name. Object to any exported name, file, option, or feature beyond that."
	tasteInstruction = "Taste: as simple as possible. Object to wrappers, needless indirection, speculative generality, dead code, and comments that narrate the code."
)

// section returns the "## Sprint n:" section of SPRINTS.md, up to the next
// heading of its level.
func section(doc string, n int) (string, bool) {
	heading := fmt.Sprintf("## Sprint %d:", n)
	_, rest, ok := strings.Cut(doc, "\n"+heading)
	if !ok {
		return "", false
	}
	body, _, _ := strings.Cut(rest, "\n## ")
	return heading + body, true
}

// git runs git in dir and returns its trimmed output.
func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	if len(line) > 72 {
		line = strings.ToValidUTF8(line[:72], "") + "..."
	}
	return line
}
