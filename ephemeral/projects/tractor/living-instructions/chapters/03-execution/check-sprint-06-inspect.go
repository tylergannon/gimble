package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tylergannon/tractor/checklist"
)

type timelineEvent struct {
	Type string `json:"type"`
	Node string `json:"node"`
	Item string `json:"item"`
}

type binding struct {
	Harness string `json:"harness"`
	Workdir string `json:"workdir"`
}

type checkpoint struct {
	Sessions map[string]binding `json:"sessions"`
}

type validationRecord struct {
	Item     string `json:"item"`
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
	Passed   bool   `json:"passed"`
}

type eventIndexEntry struct {
	NodeID string `json:"node_id"`
	Path   string `json:"path"`
}

func main() {
	if len(os.Args) != 5 || (os.Args[1] != "plan" && os.Args[1] != "execution") {
		fatalf("usage: check-sprint-06-inspect <plan|execution> <checklist> <logs> <workdir>")
	}
	list, err := checklist.Load(os.Args[2])
	if err != nil {
		fatalf("load checklist: %v", err)
	}
	logs, err := filepath.Abs(os.Args[3])
	if err != nil {
		fatalf("resolve logs: %v", err)
	}
	workdir, err := filepath.Abs(os.Args[4])
	if err != nil {
		fatalf("resolve workdir: %v", err)
	}

	thread := os.Args[1]
	if thread == "execution" {
		thread = "medium"
	}
	requireCompletedRun(logs, thread, workdir)
	if os.Args[1] == "plan" {
		inspectPlan(list, logs)
		return
	}
	inspectExecution(list, logs)
}

func inspectPlan(list *checklist.Checklist, logs string) {
	expected := map[string]bool{
		"go test ./greeting": false,
		"go test ./words":    false,
	}
	if len(list.Items) != len(expected) {
		fatalf("planned item count = %d, want %d", len(list.Items), len(expected))
	}
	for _, item := range list.Items {
		if item.Done || item.DonePresent {
			fatalf("planned item %q contains engine-owned done", item.Name)
		}
		if item.Checklist != "" {
			fatalf("planned item %q is nested through %q", item.Name, item.Checklist)
		}
		if _, ok := expected[item.Command]; !ok {
			fatalf("planned item %q has non-specific command %q", item.Name, item.Command)
		}
		if expected[item.Command] {
			fatalf("planned command %q is duplicated", item.Command)
		}
		expected[item.Command] = true
	}
	for command, found := range expected {
		if !found {
			fatalf("planned checklist is missing command %q", command)
		}
	}

	events := readJSONLines[timelineEvent](filepath.Join(logs, "timeline.jsonl"))
	questions := 0
	for _, event := range events {
		if event.Type == "QuestionAsked" {
			questions++
		}
	}
	if questions == 0 {
		fatalf("planning timeline has no QuestionAsked event")
	}
}

func inspectExecution(list *checklist.Checklist, logs string) {
	if len(list.Items) < 2 {
		fatalf("completed checklist has %d items, want at least 2", len(list.Items))
	}
	commands := make(map[string]string, len(list.Items))
	for _, item := range list.Items {
		if !item.Done || !item.DonePresent {
			fatalf("completed item %q is not engine-marked done", item.Name)
		}
		commands[item.Name] = item.Command
	}

	events := readJSONLines[timelineEvent](filepath.Join(logs, "timeline.jsonl"))
	selected := make(map[string]bool, len(list.Items))
	selectionCount := 0
	validatedCount := 0
	for _, event := range events {
		switch {
		case event.Type == "LoopItemSelected" && event.Node == "sprints":
			selectionCount++
			selected[event.Item] = true
		case event.Type == "LoopValidated" && event.Node == "sprints":
			validatedCount++
		}
	}
	if selectionCount < 2 || len(selected) != len(list.Items) {
		fatalf("item selections = %d across %d items, want every one of %d items", selectionCount, len(selected), len(list.Items))
	}

	paths, err := filepath.Glob(filepath.Join(logs, "stages", "*-sprints", "validation.json"))
	if err != nil {
		fatalf("glob validation records: %v", err)
	}
	if len(paths) != validatedCount {
		fatalf("validation records = %d, timeline validations = %d", len(paths), validatedCount)
	}
	passed := make(map[string]bool, len(list.Items))
	for _, path := range paths {
		var record validationRecord
		readJSON(path, &record)
		command, ok := commands[record.Item]
		if !ok || record.Command != command {
			fatalf("validation %s does not match a completed checklist item", path)
		}
		info, err := os.Stat(filepath.Join(filepath.Dir(path), "validation.log"))
		if err != nil || info.Size() == 0 {
			fatalf("validation log beside %s is missing or empty", path)
		}
		if record.Passed {
			if record.ExitCode != 0 {
				fatalf("passing validation %s has exit code %d", path, record.ExitCode)
			}
			passed[record.Item] = true
		}
	}
	for _, item := range list.Items {
		if !passed[item.Name] {
			fatalf("completed item %q has no passing validation record", item.Name)
		}
	}

	index := readJSONLines[eventIndexEntry](filepath.Join(logs, "events", "index.jsonl"))
	implementTurns := 0
	for _, entry := range index {
		if entry.NodeID != "implement" {
			continue
		}
		implementTurns++
		info, err := os.Stat(filepath.Join(logs, entry.Path))
		if err != nil || info.Size() == 0 {
			fatalf("real harness event log %q is missing or empty", entry.Path)
		}
	}
	if implementTurns != selectionCount {
		fatalf("real implement turns = %d, item selections = %d", implementTurns, selectionCount)
	}
}

func requireCompletedRun(logs, thread string, workdir string) {
	events := readJSONLines[timelineEvent](filepath.Join(logs, "timeline.jsonl"))
	completed := 0
	for _, event := range events {
		if event.Type == "PipelineCompleted" {
			completed++
		}
	}
	if completed != 1 {
		fatalf("%s timeline has %d PipelineCompleted events, want 1", thread, completed)
	}

	var state checkpoint
	readJSON(filepath.Join(logs, "checkpoint.json"), &state)
	got, ok := state.Sessions[thread]
	if !ok || got.Harness != "codex" || got.Workdir != workdir {
		fatalf("%s binding = %#v, want Codex in %s", thread, got, workdir)
	}
}

func readJSON(path string, value any) {
	raw, err := os.ReadFile(path)
	if err != nil {
		fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		fatalf("decode %s: %v", path, err)
	}
}

func readJSONLines[T any](path string) []T {
	file, err := os.Open(path)
	if err != nil {
		fatalf("open %s: %v", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fatalf("close %s: %v", path, err)
		}
	}()

	var values []T
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var value T
		if err := json.Unmarshal(scanner.Bytes(), &value); err != nil {
			fatalf("decode %s: %v", path, err)
		}
		values = append(values, value)
	}
	if err := scanner.Err(); err != nil {
		fatalf("scan %s: %v", path, err)
	}
	return values
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
