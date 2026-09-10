package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func TestBakeOff(t *testing.T) {
	input := Input{Goal: "build a candidate", Acceptance: Acceptance{Command: "check integrated result"}}
	t.Run("concurrent builds join before judging", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		started, release := make(chan struct{}, 3), make(chan struct{})
		var completed, judged atomic.Int32
		runtime := &program.Runtime{
			Agent: func(callCtx context.Context, call program.Call) (any, error) {
				if call.Name == "build" {
					started <- struct{}{}
					select {
					case <-release:
						completed.Add(1)
						return "report", nil
					case <-callCtx.Done():
						return nil, callCtx.Err()
					}
				}
				if call.Name != "judge" || completed.Load() != 3 {
					return nil, fmt.Errorf("judge ran before three builds completed")
				}
				if err := callCtx.Err(); err != nil {
					return nil, fmt.Errorf("judge received canceled group context: %w", err)
				}
				judged.Add(1)
				return Decision{Winner: 0}, nil
			},
			RunCommand: func(context.Context, string, string) (program.Check, error) {
				return program.Check{Passed: true}, nil
			},
		}
		done := make(chan error, 1)
		go func() {
			_, err := BakeOff(ctx, runtime, input)
			done <- err
		}()
		// A sequential implementation cannot start all three before release.
		for range 3 {
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("three concurrent builders did not start")
			}
		}
		close(release)
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal("bake-off did not finish")
		}
		if judged.Load() != 1 {
			t.Fatalf("judge calls = %d, want 1", judged.Load())
		}
	})

	t.Run("failed build prevents judging", func(t *testing.T) {
		buildErr := errors.New("builder unavailable")
		var judged atomic.Bool
		runtime := &program.Runtime{Agent: func(_ context.Context, call program.Call) (any, error) {
			if call.Name == "judge" {
				judged.Store(true)
				return Decision{Winner: 0}, nil
			}
			if call.Role == sswe {
				return nil, buildErr
			}
			return "report", nil
		}}
		_, err := BakeOff(t.Context(), runtime, input)
		if !errors.Is(err, buildErr) {
			t.Fatalf("error = %v, want build failure", err)
		}
		if judged.Load() {
			t.Fatal("judge ran after a failed build")
		}
	})
}
