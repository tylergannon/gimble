package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tylergannon/tractor/checklist"
)

const (
	BriefFile          = "brief.md"
	ChecklistFile      = "checklist.md"
	RecommendationFile = "recommendation.md"
)

// Recommendation is the parsed handoff written by the plan workflow.
type Recommendation struct {
	Size      string
	Rationale string
	Next      string
}

// ValidatePlanArtifacts mechanically validates the three planning outputs.
// It returns the parsed recommendation for the CLI handoff on success.
func ValidatePlanArtifacts(workdir, project string) (*Recommendation, error) {
	root, err := ProjectDir(workdir, project)
	if err != nil {
		return nil, err
	}
	if err := requireNonEmpty(filepath.Join(root, BriefFile)); err != nil {
		return nil, err
	}

	recommendationPath := filepath.Join(root, RecommendationFile)
	if err := requireNonEmpty(recommendationPath); err != nil {
		return nil, err
	}
	recommendation, err := ParseRecommendation(recommendationPath)
	if err != nil {
		return nil, err
	}
	if err := recommendation.validate(project); err != nil {
		return nil, fmt.Errorf("validate %s: %w", RecommendationFile, err)
	}

	checklistPath := filepath.Join(root, ChecklistFile)
	if err := requireNonEmpty(checklistPath); err != nil {
		return nil, err
	}
	list, err := checklist.Load(checklistPath)
	if err != nil {
		return nil, fmt.Errorf("validate %s: %w", ChecklistFile, err)
	}
	for _, item := range list.Items {
		if item.DonePresent {
			return nil, fmt.Errorf("validate %s: item %q contains engine-owned done field", ChecklistFile, item.Name)
		}
	}
	if err := validateExecutionShape(workdir, root, recommendation.Size, list); err != nil {
		return nil, fmt.Errorf("validate %s for %s: %w", ChecklistFile, recommendation.Size, err)
	}
	return recommendation, nil
}

func validateExecutionShape(workdir, root, size string, list *checklist.Checklist) error {
	switch size {
	case "SIMPLE":
		if len(list.Items) > 1 {
			return fmt.Errorf("SIMPLE must contain at most one item, got %d", len(list.Items))
		}
		return requireFlatItems(list)
	case "MEDIUM":
		if len(list.Items) <= 1 {
			return fmt.Errorf("MEDIUM must contain more than one item, got %d", len(list.Items))
		}
		return requireFlatItems(list)
	case "LARGE":
		if len(list.Items) <= 1 {
			return fmt.Errorf("LARGE must contain more than one chapter item, got %d", len(list.Items))
		}
		return validateLargeItems(workdir, root, list)
	default:
		return fmt.Errorf("unsupported recommendation size %q", size)
	}
}

func requireFlatItems(list *checklist.Checklist) error {
	for _, item := range list.Items {
		if item.Checklist != "" {
			return fmt.Errorf("item %q must not carry a nested checklist", item.Name)
		}
	}
	return nil
}

func validateLargeItems(workdir, root string, list *checklist.Checklist) error {
	seen := make(map[string]string, len(list.Items)*2)
	for _, item := range list.Items {
		docPath, err := resolveSupportingPath(workdir, root, item.Doc)
		if err != nil {
			return fmt.Errorf("chapter %q doc: %w", item.Name, err)
		}
		if err := requireUniqueSupportingPath(seen, docPath, item.Name+" doc"); err != nil {
			return err
		}
		if err := requireNonEmpty(docPath); err != nil {
			return fmt.Errorf("chapter %q doc: %w", item.Name, err)
		}

		checklistPath, err := resolveSupportingPath(workdir, root, item.Checklist)
		if err != nil {
			return fmt.Errorf("chapter %q checklist: %w", item.Name, err)
		}
		if err := requireUniqueSupportingPath(seen, checklistPath, item.Name+" checklist"); err != nil {
			return err
		}
		sprints, err := checklist.Load(checklistPath)
		if err != nil {
			return fmt.Errorf("chapter %q checklist: %w", item.Name, err)
		}
		for _, sprint := range sprints.Items {
			if sprint.DonePresent {
				return fmt.Errorf("chapter %q checklist item %q contains engine-owned done field", item.Name, sprint.Name)
			}
		}
		if len(sprints.Items) != 0 {
			return fmt.Errorf("chapter %q checklist must initially contain no sprint items, got %d", item.Name, len(sprints.Items))
		}
	}
	return nil
}

func resolveSupportingPath(workdir, root, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("path %q must be relative to the repository workdir", path)
	}
	if slices.Contains(strings.Split(filepath.ToSlash(path), "/"), "..") {
		return "", fmt.Errorf("path %q must not contain traversal", path)
	}

	rootPath, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	resolved := filepath.Join(workdir, filepath.Clean(path))
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	if !pathWithin(rootPath, resolved) {
		return "", fmt.Errorf("path %q resolves outside project root", path)
	}

	realRoot, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}
	if !pathWithin(realRoot, realPath) {
		return "", fmt.Errorf("path %q resolves outside project root through a symbolic link", path)
	}
	return realPath, nil
}

func pathWithin(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != "." && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func requireUniqueSupportingPath(seen map[string]string, path, owner string) error {
	if previous, exists := seen[path]; exists {
		return fmt.Errorf("%s duplicates supporting path used by %s", owner, previous)
	}
	seen[path] = owner
	return nil
}

func requireNonEmpty(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("require planning artifact %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("require planning artifact %s: not a regular file", path)
	}
	if info.Size() == 0 {
		return fmt.Errorf("require planning artifact %s: file is empty", path)
	}
	return nil
}

// ParseRecommendation parses the fixed recommendation.md contract.
func ParseRecommendation(path string) (*Recommendation, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", RecommendationFile, err)
	}

	var lines []string
	for rawLine := range strings.SplitSeq(string(contents), "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) != 4 || lines[0] != "# Recommendation" {
		return nil, errors.New("parse recommendation.md: expected heading and exactly Size, Rationale, and Next fields")
	}

	size, err := field(lines[1], "Size")
	if err != nil {
		return nil, err
	}
	rationale, err := field(lines[2], "Rationale")
	if err != nil {
		return nil, err
	}
	next, err := field(lines[3], "Next")
	if err != nil {
		return nil, err
	}
	return &Recommendation{Size: size, Rationale: rationale, Next: next}, nil
}

func field(line, name string) (string, error) {
	prefix := name + ":"
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("parse %s: expected %s field", RecommendationFile, name)
	}
	value := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if value == "" {
		return "", fmt.Errorf("parse %s: %s must not be empty", RecommendationFile, name)
	}
	return value, nil
}

func (recommendation *Recommendation) validate(project string) error {
	var next string
	switch recommendation.Size {
	case "SIMPLE":
		next = "Execute the plan yourself."
	case "MEDIUM":
		next = "tractor workflow run medium --project " + project
	case "LARGE":
		next = "tractor workflow run large --project " + project
	default:
		return fmt.Errorf("size must be SIMPLE, MEDIUM, or LARGE, got %q", recommendation.Size)
	}
	if recommendation.Next != next {
		return fmt.Errorf("next for %s must be %q, got %q", recommendation.Size, next, recommendation.Next)
	}
	return nil
}
