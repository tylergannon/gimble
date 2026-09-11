// Package runlog reads Gimble's persisted run records for the runtime and its
// web application.
package runlog

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

// Read replays the run log at dir and follows it until a record whose kind is
// "complete" is observed. T must decode the records written to the log.
func Read[T any](ctx context.Context, dir string, yield func(T) error) error {
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
			var envelope struct {
				Event struct {
					Kind string `json:"kind"`
				} `json:"event"`
			}
			if err := json.Unmarshal(pending, &envelope); err != nil {
				return err
			}
			var event T
			if err := json.Unmarshal(pending, &event); err != nil {
				return err
			}
			pending = pending[:0]
			if err := yield(event); err != nil {
				return err
			}
			if envelope.Event.Kind == "complete" {
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
