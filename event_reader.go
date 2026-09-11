package gimble

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ReadRun replays the run log at dir and follows it until the complete event
// is observed. It is safe to read while the run writer is appending: a record
// is delivered only after its newline has been written.
func ReadRun(ctx context.Context, dir string, yield func(Event) error) error {
	f, err := os.Open(filepath.Join(dir, "run.jsonl"))
	if err != nil {
		return err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var pending []byte
	for {
		line, readErr := r.ReadBytes('\n')
		pending = append(pending, line...)
		if len(pending) != 0 && pending[len(pending)-1] == '\n' {
			var event Event
			if err := json.Unmarshal(pending, &event); err != nil {
				return err
			}
			pending = pending[:0]
			if err := yield(event); err != nil {
				return err
			}
			if event.Kind == "complete" {
				return nil
			}
		}
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return readErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
