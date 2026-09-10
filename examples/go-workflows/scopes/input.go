package main

import (
	"fmt"
	"strings"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

type Input struct {
	Goal     string    `json:"goal"`
	Chapters []Chapter `json:"chapters"`
}

type Chapter struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Sprints []Sprint `json:"sprints"`
}

type Sprint struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

func (input Input) ledgers() ([]program.Item, map[string][]program.Item, error) {
	if strings.TrimSpace(input.Goal) == "" || len(input.Chapters) == 0 {
		return nil, nil, fmt.Errorf("goal and at least one chapter are required")
	}
	var chapters []program.Item
	sprints := make(map[string][]program.Item)
	commands := make(map[string]bool)
	for _, chapter := range input.Chapters {
		if strings.TrimSpace(chapter.Name) == "" || strings.TrimSpace(chapter.Command) == "" || len(chapter.Sprints) == 0 {
			return nil, nil, fmt.Errorf("each chapter needs a name, command, and sprints")
		}
		if _, exists := sprints[chapter.Name]; exists || commands[chapter.Command] {
			return nil, nil, fmt.Errorf("chapter names and all commands must be unique")
		}
		commands[chapter.Command] = true
		chapters = append(chapters, program.Item{Name: chapter.Name, Command: chapter.Command})
		names := make(map[string]bool)
		for _, sprint := range chapter.Sprints {
			if strings.TrimSpace(sprint.Name) == "" || strings.TrimSpace(sprint.Command) == "" || names[sprint.Name] || commands[sprint.Command] {
				return nil, nil, fmt.Errorf("sprints need distinct names within a chapter and unique nonempty commands")
			}
			names[sprint.Name], commands[sprint.Command] = true, true
			sprints[chapter.Name] = append(sprints[chapter.Name], program.Item{Name: sprint.Name, Command: sprint.Command})
		}
	}
	return chapters, sprints, nil
}
