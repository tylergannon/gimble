package engine

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Bounds on one observer digest. Host limits are much larger; these protect
// the parent or supervisor context that receives the result.
const (
	maxDigestEvents = 40
	maxDigestBytes  = 32 * 1024
)

type observerDigest struct {
	Message   string
	Truncated bool
	Artifact  string
}

type observerDigestInput struct {
	name   string
	goal   string
	root   string
	live   []liveExecution
	events []timelineEvent
	total  int
}

func renderObserverDigest(input observerDigestInput) observerDigest {
	artifact := filepath.Join(input.root, "timeline.jsonl")
	if len(input.events) == 0 {
		return observerDigest{Artifact: artifact}
	}

	var message strings.Builder
	fmt.Fprintf(&message, "Gimble run news: %s\n", input.name)
	if goal := strings.TrimSpace(input.goal); goal != "" {
		fmt.Fprintf(&message, "Goal: %s\n", goal)
	}
	fmt.Fprintf(&message, "Run directory: %s\n", input.root)
	writeCurrentWork(&message, input)

	shown := input.events
	truncated := len(shown) > maxDigestEvents
	if truncated {
		shown = shown[len(shown)-maxDigestEvents:]
	}
	sectionStart := message.Len()
	lines := make([]string, 0, len(shown))
	for _, event := range shown {
		line := "  " + summarizeEvent(event) + "\n"
		if sectionStart+len(strings.Join(lines, ""))+len(line)+512 > maxDigestBytes {
			truncated = true
			break
		}
		lines = append(lines, line)
	}
	completeness := "complete"
	if truncated {
		completeness = "truncated"
	}
	fmt.Fprintf(&message, "\nEvents (%s; %d total; full artifact: %s):\n", completeness, input.total, artifact)
	for _, line := range lines {
		message.WriteString(line)
	}

	return observerDigest{
		Message: message.String(), Truncated: truncated, Artifact: artifact,
	}
}

func writeCurrentWork(message *strings.Builder, input observerDigestInput) {
	if len(input.live) == 0 {
		message.WriteString("Running now: nothing\n")
		return
	}
	names := make([]string, 0, len(input.live))
	for _, entry := range input.live {
		names = append(names, fmt.Sprintf("%s (attempt %d)", entry.NodeID, entry.Attempt))
	}
	fmt.Fprintf(message, "Running now: %s\n", strings.Join(names, ", "))
}

// summarizeEvent renders one timeline event as a single line, type first.
func summarizeEvent(event timelineEvent) string {
	var line strings.Builder
	if timestamp, ok := event["ts"].(string); ok {
		line.WriteString(timestamp)
		line.WriteString(" ")
	}
	if kind, ok := event["type"].(string); ok {
		line.WriteString(kind)
	}
	keys := make([]string, 0, len(event))
	for key := range event {
		if key == "ts" || key == "type" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		raw, err := json.Marshal(event[key])
		if err != nil {
			continue
		}
		fmt.Fprintf(&line, " %s=%s", key, truncateValue(string(raw)))
	}
	return line.String()
}

func truncateValue(value string) string {
	const limit = 200
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
