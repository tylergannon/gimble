package runlog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadDrainsAvailableHistory(t *testing.T) {
	dir := t.TempDir()
	const count = 1000
	data := strings.Repeat("{\"event\":{\"kind\":\"run_started\"}}\n", count) + "{\"event\":{\"kind\":\"complete\"}}\n"
	if err := os.WriteFile(filepath.Join(dir, "run.jsonl"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	// A completed file must drain without the tail-following poll delay on
	// every record (which would take more than ten seconds for this input).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	seen := 0
	err := Read[map[string]any](ctx, dir, func(map[string]any) error {
		seen++
		return nil
	})
	if err != nil || seen != count+1 {
		t.Fatalf("drain: records=%d, error=%v", seen, err)
	}
}
