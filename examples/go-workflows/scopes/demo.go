package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

type Stage struct {
	Name    string                     `json:"name"`
	Role    program.Role               `json:"role"`
	Prompt  string                     `json:"prompt"`
	Context program.ContextSnapshot    `json:"context"`
	Values  map[string]json.RawMessage `json:"values"`
}

type Report struct {
	Directory string         `json:"directory"`
	Chapters  []program.Item `json:"chapters"`
	Stages    []Stage        `json:"stages"`
}

func main() {
	program.Main(exampleInput(), func(ctx context.Context, input Input) error {
		directory, err := os.MkdirTemp("", "gimble-scopes-")
		if err != nil {
			return err
		}
		ctx, err = program.NewContext(ctx, directory, program.ContextLimits{ValueBytes: 600, PromptBytes: 2400})
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Context files retained under %s\n", directory)
		report := Report{Directory: directory}
		report.Chapters, err = NestedScopes(ctx, demoRuntime(input, &report), input)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(report)
	})
}

// Only the context files are real. A build callback marks its scripted command
// complete; chapter checks require every scripted sprint in that chapter.
func demoRuntime(input Input, report *Report) *program.Runtime {
	var mu sync.Mutex
	completed := make(map[string]bool)
	chapterCommands := make(map[string][]string)
	for _, chapter := range input.Chapters {
		for _, sprint := range chapter.Sprints {
			chapterCommands[chapter.Command] = append(chapterCommands[chapter.Command], sprint.Command)
		}
	}
	return &program.Runtime{
		Output: os.Stderr,
		Agent: func(ctx context.Context, call program.Call) (any, error) {
			snapshot, err := program.SnapshotContext(ctx)
			if err != nil {
				return nil, err
			}
			if !strings.HasPrefix(call.Prompt, snapshot.Prompt) {
				return nil, fmt.Errorf("%s did not receive its captured context", call.Name)
			}
			values, err := readValues(snapshot.Index)
			if err != nil {
				return nil, err
			}
			mu.Lock()
			defer mu.Unlock()
			if call.Name == "build" {
				var sprint program.Item
				if err := json.Unmarshal(values["sprint"], &sprint); err != nil {
					return nil, fmt.Errorf("build needs automatic sprint metadata: %w", err)
				}
				completed[sprint.Command] = true
			}
			report.Stages = append(report.Stages, Stage{call.Name, call.Role, call.Prompt, snapshot, values})
			return "Canned response; no agent or software command ran.", nil
		},
		RunCommand: func(_ context.Context, _, command string) (program.Check, error) {
			mu.Lock()
			defer mu.Unlock()
			passed := completed[command]
			if sprints, chapter := chapterCommands[command]; chapter {
				passed = true
				for _, sprint := range sprints {
					passed = passed && completed[sprint]
				}
			}
			return program.Check{Passed: passed, Output: "Scripted completion state; no command executed."}, nil
		},
	}
}

func readValues(path string) (map[string]json.RawMessage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var index struct {
		Entries []struct{ Key, Path string }
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return nil, err
	}
	values := make(map[string]json.RawMessage)
	for _, entry := range index.Entries {
		raw, err := os.ReadFile(entry.Path)
		if err != nil {
			return nil, err
		}
		values[entry.Key] = json.RawMessage(raw)
	}
	return values, nil
}
