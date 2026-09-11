// Package sprint is the sprint workflow. It builds one sprint of
// ephemeral/research/api/SPRINTS.md and, beside it, works GitHub issues,
// each in a worktree of its own. Each piece of work goes the same way,
// gated as docs/definition-of-done.md says: a researcher reads the code
// once; a planner forked from it keeps the backlog; each lap a coder forked
// from it does the planner's task under a supervisor, and the lap is
// committed when the checks pass. When the planner is done, a validator
// checks that the work is demonstrated, and its objections are another
// loop. Then the planner files what is left and merges the branch.
package sprint

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tylergannon/gimble"
	"github.com/tylergannon/gimble/claude"
	"github.com/tylergannon/gimble/codex"
)

//go:generate go tool polytype --validate

// Input starts a sprint.
type Input struct {
	// The sprint to build: its number in ephemeral/research/api/SPRINTS.md, e.g. 3.
	Sprint int `json:"sprint"`
	// GitHub issues to work beside the sprint, each in a worktree and branch of its own, e.g. [120].
	Issues []int `json:"issues"`
	// Absolute path of the repository. The sprint's laps are committed to the branch checked out there, and the sprint ends by merging that branch.
	Repo string `json:"repo"`
	// Codex model for the researchers, the planners, and the coders, e.g. "gpt-5.6-luna".
	Model string `json:"model"`
	// Claude Code model for the supervisors and the validators, e.g. "haiku".
	ReviewModel string `json:"review_model"`
	// The most laps one piece of work may run. It fails if its planner is not done by then.
	Laps int `json:"laps"`
}

// checks gate every lap's commit.
var checks = []string{"just build", "go vet ./...", "go test ./..."}

// Sprint builds sprint in.Sprint in in.Repo and works in.Issues beside it.
func Sprint(ctx context.Context, in Input) error {
	if err := gimble.SetJSON(ctx, "input", in); err != nil {
		return err
	}
	goal, err := sprintGoal(in.Repo, in.Sprint)
	if err != nil {
		return err
	}
	base, err := git(ctx, in.Repo, "rev-parse", "HEAD")
	if err != nil {
		return err
	}

	// The sprint and each issue side by side. An issue that fails is
	// logged; the sprint failing stops them all.
	g := gimble.Group(ctx, "work")
	g.Go("sprint", func(ctx context.Context) error {
		return build(ctx, in, in.Repo, fmt.Sprintf("Sprint %d", in.Sprint), goal)
	})
	for _, n := range in.Issues {
		g.Go("issue", func(ctx context.Context) error {
			if err := issue(ctx, in, base, n); err != nil {
				log.Printf("sprint: issue %d: %v", n, err)
			}
			return nil
		})
	}
	return g.Wait()
}

// issue works GitHub issue n on its own branch, in a worktree that is
// removed when the work ends.
func issue(ctx context.Context, in Input, base string, n int) error {
	dir, err := os.MkdirTemp("", fmt.Sprintf("gimble-issue-%d-", n))
	if err != nil {
		return err
	}
	if _, err := git(ctx, in.Repo, "worktree", "add", "-b", fmt.Sprintf("issue-%d", n), dir, base); err != nil {
		return err
	}
	defer git(context.WithoutCancel(ctx), in.Repo, "worktree", "remove", "--force", dir)
	return build(ctx, in, dir, fmt.Sprintf("Issue #%d", n), fmt.Sprintf(issueGoal, n))
}

// build takes one piece of work in dir from its goal to a merge. name
// titles its commits.
func build(ctx context.Context, in Input, dir, name, goal string) error {
	cx, cl := codex.New(), claude.New()
	goal += "\n\n" + fmt.Sprintf(done, dir)
	researcher := gimble.NewSession(ctx, "researcher", cx, in.Model, dir)
	if _, err := researcher.Generate[gimble.Text](ctx, researchPrompt+"\n\n"+goal); err != nil {
		return err
	}
	planner, err := researcher.Fork(ctx, "planner")
	if err != nil {
		return err
	}
	supervisor := gimble.NewSession(ctx, "supervisor", cl, in.ReviewModel, dir)
	validator := gimble.NewSession(ctx, "validator", cl, in.ReviewModel, dir)
	referee := gimble.NewSession(ctx, "referee", cl, in.ReviewModel, dir)

	// Build until the planner is done, then validate. The validator's
	// objections are the goal of another loop, for three rounds at most.
	laps := 0
	for round := 1; ; round++ {
		loop := gimble.Loop(ctx, "build", goal, planner)
		for ctx, task := range loop.Laps {
			if laps++; laps > in.Laps {
				return fmt.Errorf("%s: the planner was not done after %d laps", name, in.Laps)
			}
			if err := lap(ctx, dir, name, researcher, supervisor, laps, task.Text); err != nil {
				return err
			}
		}
		if err := loop.Err(); err != nil {
			return err
		}
		review, err := validator.Generate[gimble.Review](ctx, fmt.Sprintf(validatePrompt, name)+"\n\n"+goal,
			gimble.WithSupervisor(referee, refereeInstruction),
		)
		if err != nil {
			return err
		}
		if len(review.Objections) == 0 {
			break
		}
		objections := "- " + strings.Join(review.Objections, "\n- ")
		log.Printf("sprint: %s, round %d: the validator objects:\n%s", name, round, objections)
		if round == 3 {
			return fmt.Errorf("%s: the validator still objects after %d rounds", name, round)
		}
		goal += "\n\nThe validator found the validation invalid, so it is also done only when these are answered:\n\n" + objections
	}

	// Exit and merge: the planner, who drove the work, files what is left
	// and merges the branch.
	summary, err := planner.Generate[gimble.Text](ctx, mergePrompt)
	if err != nil {
		return err
	}
	log.Printf("sprint: %s is done:\n%s", name, summary)
	return nil
}

// lap does one task: a coder forked from the researcher, supervised, then
// a commit if the checks pass.
func lap(ctx context.Context, dir, name string, researcher, supervisor *gimble.Session, n int, task string) error {
	if err := gimble.Set(ctx, "task", task); err != nil {
		return err
	}
	coder, err := researcher.Fork(ctx, "coder")
	if err != nil {
		return err
	}
	_, err = coder.Generate[gimble.Text](ctx, codePrompt+"\n\n"+gimble.ScopeText(ctx),
		gimble.WithSupervisor(supervisor, superviseInstruction, gimble.WithInterval(time.Minute)),
	)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		// A failed turn is a fact for the planner, not the end of the work:
		// what it did stays in the tree for the next lap to see.
		log.Printf("sprint: %s, lap %d: %v", name, n, err)
	}

	// Work that fails the checks stays uncommitted; the planner sees the
	// failure in its command results next lap.
	for _, check := range checks {
		cmd := exec.CommandContext(ctx, "sh", "-c", check)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("sprint: %s, lap %d: %s failed, so the lap stays uncommitted:\n%s", name, n, check, out)
			return nil
		}
	}
	if _, err := git(ctx, dir, "add", "-A"); err != nil {
		return err
	}
	if status, _ := git(ctx, dir, "status", "--porcelain"); status == "" {
		return nil
	}
	_, err = git(ctx, dir, "commit", "-m", fmt.Sprintf("%s, lap %d: %s\n\n%s", name, n, firstLine(task), task))
	return err
}

const issueGoal = "GitHub issue #%[1]d, which gh issue view %[1]d shows, is resolved as it asks, and the pull request that merges it says \"Fixes #%[1]d\"."

const done = "Definition of done (docs/definition-of-done.md): the software actually works, and what the goal above asks for is 90-95%% built and committed in %s, with `just build`, `go vet ./...` and `go test ./...` exiting 0 there and each part seen working in a test or a command. Gate on these requirements only, never on code quality. When this is true the goal is met, even with quirks left: they are filed as issues after the loop, not fixed in it. Give the checks steps of their own, with those commands."

const researchPrompt = `You are about to lead one piece of work on this repository, Gimble, a Go library for agent workflows with a web page. Read AGENTS.md, docs/definition-of-done.md, ephemeral/research/api/API.md, ephemeral/research/api/SPRINTS.md, and the code the goal below touches, until you know where everything it needs is. Change no files. Answer with a short summary of what exists and what the goal needs.`

const codePrompt = "Do the task below in this repository, following AGENTS.md and ephemeral/research/api/API.md, and build only what it asks. Make `just build`, `go vet ./...` and `go test ./...` pass. Do not commit: the workflow commits the lap when those pass. Answer with a short summary of what you changed."

const superviseInstruction = "Don't let it build what its task does not ask for, over-engineer what it does build, or break a rule in AGENTS.md. Object to nothing else: code quality and style are not yours to judge."

const refereeInstruction = "Keep it to validation. Don't let it object to anything but an invalid validation: a part of the goal nothing demonstrates, a demonstration that does not show what it claims, or a check that was tampered with. Enhancements, code quality, and bugs that do not stop the goal being demonstrated go to GitHub issues, never to objections. Don't let it change files."

const validatePrompt = `You are the validator of one piece of work on this repository, Gimble. Its goal is below, and the agents that did it say it is done. Read docs/definition-of-done.md, then decide whether the validation legitimately demonstrates the goal: that the software actually works and what the goal asks for is implemented and has been seen working.

- Find what demonstrates each part of the goal: the tests and the commands. Check that each really exercises what it claims to, and that none was weakened, skipped, or faked to pass. The work's commits are titled "%s, lap ..."; git log -p shows what they changed.
- Run the checks yourself, and run the software by hand where that shows more, the web page included.
- Change no files and commit nothing.

Object only where the validation is invalid: the software does not work, a large part of the goal is missing (the work is done at 90-95%%, and the rest is filed as issues), a demonstration does not show what it claims, or a check was tampered with. Code quality, style, enhancements, and bugs that do not stop the goal being demonstrated are not objections; file the ones that matter as GitHub issues with gh issue create. Each objection becomes work for the builders, so write it as an instruction to them. Answer with an empty list when the goal is demonstrated.`

const mergePrompt = `The validator found the work demonstrated, so finish it as docs/definition-of-done.md says. File each quirk and bug the work leaves as a GitHub issue with gh issue create, skipping any that gh issue list already has. Then push this branch, open a pull request for it with gh pr create that says what the work built, how it was seen working, and which issues it left, and merge it with gh pr merge --squash. If main has moved since the branch began, merge main into it and rerun the checks first. Answer with the pull request's URL and the issues you filed.`

// sprintGoal returns the "## Sprint n:" section of the repository's
// SPRINTS.md, up to the next heading of its level.
func sprintGoal(repo string, n int) (string, error) {
	doc, err := os.ReadFile(filepath.Join(repo, "ephemeral/research/api/SPRINTS.md"))
	if err != nil {
		return "", err
	}
	heading := fmt.Sprintf("## Sprint %d:", n)
	_, rest, ok := strings.Cut(string(doc), "\n"+heading)
	if !ok {
		return "", fmt.Errorf("sprint: SPRINTS.md has no Sprint %d", n)
	}
	body, _, _ := strings.Cut(rest, "\n## ")
	return heading + body, nil
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
