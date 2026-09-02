package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	checklistPath := filepath.Join(root, ChecklistFile)
	if err := requireNonEmpty(checklistPath); err != nil {
		return nil, err
	}
	list, err := checklist.Load(checklistPath)
	if err != nil {
		return nil, fmt.Errorf("validate %s: %w", ChecklistFile, err)
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("validate %s: items must not be empty", ChecklistFile)
	}
	for _, item := range list.Items {
		if item.DonePresent {
			return nil, fmt.Errorf("validate %s: item %q contains engine-owned done field", ChecklistFile, item.Name)
		}
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
	return recommendation, nil
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
