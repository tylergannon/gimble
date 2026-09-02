// Package workflow provides Tractor's named workflows and their artifact
// contracts. It deliberately has no dependency on the CLI that exposes it.
package workflow

import (
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/tylergannon/tractor/graph"
)

const (
	PlanName   = "plan"
	MediumName = "medium"
	LargeName  = "large"
)

//go:embed plan.yaml
var planDefinition []byte

// Definition is the caller-facing identity of an embedded workflow.
type Definition struct {
	Name        string
	Description string
}

// Parameters are the explicit values used to materialize a workflow graph.
type Parameters struct {
	Project    string
	Seed       string
	Workdir    string
	Executable string
}

var definitions = map[string]Definition{
	PlanName: {
		Name:        PlanName,
		Description: "Interview the caller and write a planning brief, checklist, and size recommendation.",
	},
}

// List returns the available embedded workflows in stable name order.
func List() []Definition {
	listed := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		listed = append(listed, definition)
	}
	slices.SortFunc(listed, func(left, right Definition) int {
		return strings.Compare(left.Name, right.Name)
	})
	return listed
}

// Build materializes one embedded workflow using explicit parameters.
func Build(name string, params Parameters) (*graph.Graph, error) {
	if _, ok := definitions[name]; !ok {
		return nil, fmt.Errorf("unknown built-in workflow %q", name)
	}
	if name != PlanName {
		return nil, fmt.Errorf("built-in workflow %q is not runnable", name)
	}
	if err := validateParameters(params); err != nil {
		return nil, err
	}

	pipeline, err := graph.ParseYAML(planDefinition)
	if err != nil {
		return nil, fmt.Errorf("parse embedded workflow %q: %w", name, err)
	}
	planner, ok := pipeline.NodeByID("planner")
	if !ok {
		return nil, errors.New("embedded workflow plan has no planner node")
	}
	codergen, ok := planner.(*graph.CodergenNode)
	if !ok {
		return nil, errors.New("embedded workflow plan planner is not a codergen node")
	}
	validator, ok := pipeline.NodeByID("validate")
	if !ok {
		return nil, errors.New("embedded workflow plan has no validate node")
	}
	tool, ok := validator.(*graph.ToolNode)
	if !ok {
		return nil, errors.New("embedded workflow plan validate is not a tool node")
	}

	codergen.Prompt.Value = plannerPrompt(params)
	tool.ToolCommand = validatorCommand(params)
	return pipeline, nil
}

func validateParameters(params Parameters) error {
	if err := ValidateProject(params.Project); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"seed": params.Seed, "workdir": params.Workdir, "executable": params.Executable,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

// ValidateProject requires a single safe directory name. Workflow output is
// always rooted below ephemeral/projects, never at a caller-controlled path.
func ValidateProject(project string) error {
	if project == "" || project == "." || project == ".." {
		return fmt.Errorf("project %q is not a safe directory name", project)
	}
	for _, char := range project {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return fmt.Errorf("project %q is not a safe directory name", project)
	}
	return nil
}

// ProjectDir returns the fixed output root for a validated project name.
func ProjectDir(workdir, project string) (string, error) {
	if err := ValidateProject(project); err != nil {
		return "", err
	}
	if strings.TrimSpace(workdir) == "" {
		return "", errors.New("workdir is required")
	}
	return filepath.Join(workdir, "ephemeral", "projects", project), nil
}

func plannerPrompt(params Parameters) string {
	projectDir := filepath.Join(params.Workdir, "ephemeral", "projects", params.Project)
	questionCommand := shellQuote(params.Executable) + " ask <question-file>"
	return fmt.Sprintf(`You are Tractor's built-in planning workflow. Own the entire interview and planning write in this one long-lived turn.

Inputs:
- Project: %s
- Seed file: %s
- Repository workdir: %s
- Tractor executable: %s
- Planning output root: %s

Read the seed file and inspect relevant repository context before interviewing. Work from the repository workdir. Ask the caller only when an answer can change the contract. Cover intent, scope and non-goals, constraints, and the Definition of success explicitly. Keep the depth proportional to the apparent size. After those dimensions are clear, make a second pass over them; stop asking once two passes surface only details derivable from the seed, prior answers, or repository.

Ask one Markdown or HTML question at a time. Write each question beneath the planning output root and run %s. The command uses the interview directory configured by the workflow runner and blocks until the answer arrives. Read each returned answer and continue in this same context.

When the contract is clear, write exactly these planning artifacts beneath the output root and nowhere outside it:

1. brief.md — the agreed intent, scope, non-goals, constraints, and observable Definition of success.
2. checklist.md — Markdown with YAML frontmatter in the loop checklist format. The frontmatter is an items list in this shape:
   items:
     - name: <stable identity>
       check: <observable behavior>
       command: <optional real shell command>
       infer:
         files: <optional evidence glob or list of globs>
         prompt: <what a judge decides from those files>
       doc: <optional prose document path>
       checklist: <optional nested checklist path>
   Every item has name and check. Add a real command and/or infer wherever the repository makes one knowable. Paths are relative to the repository workdir. Never write a done field; that field belongs to the loop engine. Keep every item small enough for one agent turn. The Markdown body may hold context and notes that the engine does not parse.
3. recommendation.md — exactly this four-line non-empty field contract:
   # Recommendation
   Size: SIMPLE|MEDIUM|LARGE
   Rationale: <short rationale>
   Next: <next action>

Use SIMPLE for at most one sprint, MEDIUM for more than one sprint but fewer than two chapters, and LARGE for multiple chapters. The exact Next value is:
- SIMPLE: Execute the plan yourself.
- MEDIUM: tractor workflow run medium --project %s
- LARGE: tractor workflow run large --project %s

Before finishing, reread all three artifacts against this contract. If the following validator sends you back, retain the interview context, inspect its mechanical failure, and repair the artifacts without restarting the interview.`,
		strconv.Quote(params.Project), strconv.Quote(params.Seed), strconv.Quote(params.Workdir),
		strconv.Quote(params.Executable), strconv.Quote(projectDir), questionCommand, params.Project, params.Project)
}

func validatorCommand(params Parameters) string {
	return strings.Join([]string{
		shellQuote(params.Executable),
		"workflow", "validate-plan",
		"--project", shellQuote(params.Project),
		"--workdir", shellQuote(params.Workdir),
	}, " ")
}

// shellQuote returns one POSIX shell word with no interpolation surface.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
