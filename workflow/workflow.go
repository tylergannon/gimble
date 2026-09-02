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

//go:embed medium.yaml
var mediumDefinition []byte

//go:embed large.yaml
var largeDefinition []byte

// Definition is the caller-facing identity of an embedded workflow.
type Definition struct {
	Name        string
	Description string
}

// Parameters are the common values used to materialize a workflow graph.
type Parameters struct {
	Project    string
	Workdir    string
	Executable string
	Plan       PlanParameters
}

// PlanParameters are inputs used only by the planning workflow.
type PlanParameters struct {
	Seed string
}

var definitions = map[string]Definition{
	LargeName: {
		Name:        LargeName,
		Description: "Plan and execute every chapter through nested engine-owned checklists.",
	},
	MediumName: {
		Name:        MediumName,
		Description: "Run every sprint in a planning checklist through engine-owned validation.",
	},
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
	if err := validateParameters(params); err != nil {
		return nil, err
	}
	switch name {
	case PlanName:
		return buildPlan(params)
	case MediumName:
		return buildMedium(params)
	case LargeName:
		return buildLarge(params)
	default:
		return nil, fmt.Errorf("built-in workflow %q is not runnable", name)
	}
}

func buildLarge(params Parameters) (*graph.Graph, error) {
	pipeline, err := graph.ParseYAML(largeDefinition)
	if err != nil {
		return nil, fmt.Errorf("parse embedded workflow %q: %w", LargeName, err)
	}
	chapters, err := pipelineLoopNode(pipeline, LargeName, "chapters")
	if err != nil {
		return nil, err
	}
	plan, err := pipelineCodergenNode(pipeline, LargeName, "plan")
	if err != nil {
		return nil, err
	}
	if _, err := pipelineLoopNode(pipeline, LargeName, "sprints"); err != nil {
		return nil, err
	}
	implement, err := pipelineCodergenNode(pipeline, LargeName, "implement")
	if err != nil {
		return nil, err
	}

	projectDir := filepath.Join(params.Workdir, "ephemeral", "projects", params.Project)
	chapters.Checklist.Value = filepath.Join(projectDir, ChecklistFile)
	plan.Prompt.Value = largePlanPrompt(params, projectDir)
	implement.Prompt.Value = largeImplementPrompt(params, projectDir)
	return pipeline, nil
}

func pipelineLoopNode(pipeline *graph.Graph, workflowName, id string) (*graph.LoopNode, error) {
	node, ok := pipeline.NodeByID(id)
	if !ok {
		return nil, fmt.Errorf("embedded workflow %s has no %s node", workflowName, id)
	}
	typed, ok := node.(*graph.LoopNode)
	if !ok {
		return nil, fmt.Errorf("embedded workflow %s %s is not a loop node", workflowName, id)
	}
	return typed, nil
}

func pipelineCodergenNode(pipeline *graph.Graph, workflowName, id string) (*graph.CodergenNode, error) {
	node, ok := pipeline.NodeByID(id)
	if !ok {
		return nil, fmt.Errorf("embedded workflow %s has no %s node", workflowName, id)
	}
	typed, ok := node.(*graph.CodergenNode)
	if !ok {
		return nil, fmt.Errorf("embedded workflow %s %s is not a codergen node", workflowName, id)
	}
	return typed, nil
}

func buildPlan(params Parameters) (*graph.Graph, error) {
	if strings.TrimSpace(params.Plan.Seed) == "" {
		return nil, errors.New("seed is required")
	}

	pipeline, err := graph.ParseYAML(planDefinition)
	if err != nil {
		return nil, fmt.Errorf("parse embedded workflow %q: %w", PlanName, err)
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

func buildMedium(params Parameters) (*graph.Graph, error) {
	pipeline, err := graph.ParseYAML(mediumDefinition)
	if err != nil {
		return nil, fmt.Errorf("parse embedded workflow %q: %w", MediumName, err)
	}
	loopNode, ok := pipeline.NodeByID("sprints")
	if !ok {
		return nil, errors.New("embedded workflow medium has no sprints node")
	}
	sprints, ok := loopNode.(*graph.LoopNode)
	if !ok {
		return nil, errors.New("embedded workflow medium sprints is not a loop node")
	}
	implementNode, ok := pipeline.NodeByID("implement")
	if !ok {
		return nil, errors.New("embedded workflow medium has no implement node")
	}
	implement, ok := implementNode.(*graph.CodergenNode)
	if !ok {
		return nil, errors.New("embedded workflow medium implement is not a codergen node")
	}

	projectDir := filepath.Join(params.Workdir, "ephemeral", "projects", params.Project)
	sprints.Checklist.Value = filepath.Join(projectDir, ChecklistFile)
	implement.Prompt.Value = mediumPrompt(params, projectDir)
	return pipeline, nil
}

func validateParameters(params Parameters) error {
	if err := ValidateProject(params.Project); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"workdir": params.Workdir, "executable": params.Executable,
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

When the contract is clear, write these three fixed top-level planning artifacts beneath the output root and nowhere outside it. A LARGE plan also writes the chapter documents and sprint ledgers described below:

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
   Every item has name and check. Add a real command and/or infer wherever the repository makes one knowable. Every generated path is relative to the repository workdir, resolves beneath the planning output root, and contains no traversal. Never write a done field; that field belongs to the loop engine. The Markdown body may hold context and notes that the engine does not parse.
3. recommendation.md — exactly this four-line non-empty field contract:
   # Recommendation
   Size: SIMPLE|MEDIUM|LARGE
   Rationale: <short rationale>
   Next: <next action>

Match checklist.md and its supporting files to the recommendation size:
- SIMPLE: zero or one flat implementation sprint. An item must not carry a nested checklist, and each item is small enough for one agent turn.
- MEDIUM: more than one flat implementation sprint. Items must not carry nested checklists, and each item is small enough for one agent turn.
- LARGE: more than one chapter item. Every chapter item carries a unique doc path to a non-empty chapter document and a unique checklist path to that chapter's initially empty sprint ledger. Write both supporting files beneath the planning output root. Each empty sprint ledger is a valid Markdown checklist with an items: [] frontmatter list. A chapter may require multiple agent turns; the implementation sprints later written to its ledger must each fit one agent turn.

Never write a done field in the top-level checklist or a chapter's sprint ledger. Use SIMPLE for at most one sprint, MEDIUM for more than one sprint but fewer than two chapters, and LARGE for multiple chapters. The exact Next value is:
- SIMPLE: Execute the plan yourself.
- MEDIUM: tractor workflow run medium --project %s
- LARGE: tractor workflow run large --project %s

Before finishing, reread all three artifacts against this contract. If the following validator sends you back, retain the interview context, inspect its mechanical failure, and repair the artifacts without restarting the interview.`,
		strconv.Quote(params.Project), strconv.Quote(params.Plan.Seed), strconv.Quote(params.Workdir),
		strconv.Quote(params.Executable), strconv.Quote(projectDir), questionCommand, params.Project, params.Project)
}

func mediumPrompt(params Parameters, projectDir string) string {
	briefPath := filepath.Join(projectDir, BriefFile)
	checklistPath := filepath.Join(projectDir, ChecklistFile)
	interviewDir := filepath.Join(projectDir, "interview")
	questionCommand := "TRACTOR_INTERVIEW_DIR=" + shellQuote(interviewDir) + " " + shellQuote(params.Executable) + " ask <question-file>"
	return fmt.Sprintf(`You are Tractor's built-in MEDIUM execution workflow. Complete the whole current sprint in this one turn.

Inputs:
- Project: %s
- Repository workdir: %s
- Project brief: %s
- Sprint checklist: %s
- Interview directory: %s

The loop engine has injected the current <iterate> frame above this prompt, including the selected checklist item and its item document when one exists. Read the project brief, the current frame, and that item document before editing. Work only on the current sprint and inspect enough repository context to complete its entire contract.

Implement the whole current sprint, including its code, tests, and named documentation. Run the sprint's validator command yourself before returning. The loop engine runs validation again after this turn and is the only owner of checklist marking: never add, remove, or edit a done field. Commit the completed sprint before returning.

Use a blocking reviewer question only when the current sprint has no runnable validator or when a repeated engine validation failure raises a material question. In that case write one Markdown or HTML question beneath the project root and run %s. Read the returned answer and continue in this same turn. Do not ask for routine implementation choices.

Do not start a child Tractor run and do not add a failure-escalation mechanism.`,
		strconv.Quote(params.Project), strconv.Quote(params.Workdir), strconv.Quote(briefPath),
		strconv.Quote(checklistPath), strconv.Quote(interviewDir), questionCommand)
}

func largePlanPrompt(params Parameters, projectDir string) string {
	briefPath := filepath.Join(projectDir, BriefFile)
	checklistPath := filepath.Join(projectDir, ChecklistFile)
	interviewDir := filepath.Join(projectDir, "interview")
	questionCommand := "TRACTOR_INTERVIEW_DIR=" + shellQuote(interviewDir) + " " + shellQuote(params.Executable) + " ask <question-file>"
	return fmt.Sprintf(`You are Tractor's built-in LARGE execution workflow. Plan the current chapter in this one turn.

Inputs:
- Project: %s
- Repository workdir: %s
- Project brief: %s
- Chapter checklist: %s
- Interview directory: %s

The loop engine has injected the current outer <iterate> frame above this prompt. It contains the chapter document and the path of that chapter's sprint checklist. Read the project brief, the current frame, and the chapter document before planning.

Load the sprint checklist named by the current chapter item. If it already contains open sprints with non-empty item documents that cover the chapter contract, leave it unchanged and finish. Otherwise plan the whole chapter as one-turn implementation sprints. Ask one blocking reviewer question at a time only when its answer changes sprint scope or validation; write a Markdown or HTML question beneath the project root and run %s. Read the returned answer and continue in this same turn. When the chapter document settles the work, plan silently.

Write one sprint document per sprint beside the chapter's sprint checklist, then append the sprint items to that checklist. Every sprint item must have a stable name, an observable check, a real command or infer validator, and a doc path. Each sprint must fit one implementation turn and collectively the open sprints must cover the chapter. Never add, remove, or edit a done field: the loop engine alone owns checklist marking. Commit the completed chapter plan before returning.

Do not implement a sprint, start a child Tractor run, or add a repeated-failure escalation route.`,
		strconv.Quote(params.Project), strconv.Quote(params.Workdir), strconv.Quote(briefPath),
		strconv.Quote(checklistPath), strconv.Quote(interviewDir), questionCommand)
}

func largeImplementPrompt(params Parameters, projectDir string) string {
	briefPath := filepath.Join(projectDir, BriefFile)
	checklistPath := filepath.Join(projectDir, ChecklistFile)
	interviewDir := filepath.Join(projectDir, "interview")
	questionCommand := "TRACTOR_INTERVIEW_DIR=" + shellQuote(interviewDir) + " " + shellQuote(params.Executable) + " ask <question-file>"
	return fmt.Sprintf(`You are Tractor's built-in LARGE execution workflow. Complete the whole current sprint in this one turn.

Inputs:
- Project: %s
- Repository workdir: %s
- Project brief: %s
- Chapter checklist: %s
- Interview directory: %s

The loop engine has injected nested <iterate> frames above this prompt: the outer frame is the current chapter and the inner frame is the current sprint. Read the project brief, both frames, the chapter document, and the sprint document before editing. Work only on the current sprint and inspect enough repository context to complete its entire contract.

Implement the whole current sprint, including its code, tests, and named documentation. Run the sprint's validator command yourself before returning. The loop engine runs validation again after this turn and is the only owner of checklist marking: never add, remove, or edit a done field. Commit the completed sprint before returning.

Use a blocking reviewer question only when the current sprint has no runnable validator or when a repeated engine validation failure raises a material question. In that case write one Markdown or HTML question beneath the project root and run %s. Read the returned answer and continue in this same turn. Do not ask for routine implementation choices.

Do not start a child Tractor run and do not add a failure-escalation mechanism.`,
		strconv.Quote(params.Project), strconv.Quote(params.Workdir), strconv.Quote(briefPath),
		strconv.Quote(checklistPath), strconv.Quote(interviewDir), questionCommand)
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
