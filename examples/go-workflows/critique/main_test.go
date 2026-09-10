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

func TestCritiqueCircle(t *testing.T) {
	input := Input{Topic: "a shared topic"}
	t.Run("concurrent phases and peer identity", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		proposalsStarted, releaseProposals := make(chan struct{}, 3), make(chan struct{})
		reviewsStarted, releaseReviews := make(chan struct{}, 3), make(chan struct{})
		var proposalsCompleted atomic.Int32
		runtime := &program.Runtime{Agent: func(callCtx context.Context, call program.Call) (any, error) {
			started, release := proposalsStarted, releaseProposals
			if call.Name == "critique" {
				if proposalsCompleted.Load() != 3 {
					return nil, fmt.Errorf("critique began before all proposals completed")
				}
				started, release = reviewsStarted, releaseReviews
			}
			started <- struct{}{}
			select {
			case <-release:
				if call.Name == "propose" {
					proposalsCompleted.Add(1)
				}
				return string(call.Role), nil
			case <-callCtx.Done():
				return nil, callCtx.Err()
			}
		}}
		type outcome struct {
			result Result
			err    error
		}
		done := make(chan outcome, 1)
		go func() {
			result, err := CritiqueCircle(ctx, runtime, input)
			done <- outcome{result, err}
		}()
		for _, phase := range []struct {
			started <-chan struct{}
			release chan struct{}
		}{{proposalsStarted, releaseProposals}, {reviewsStarted, releaseReviews}} {
			// Each phase must start all three calls before any can finish.
			for range 3 {
				select {
				case <-phase.started:
				case <-ctx.Done():
					t.Fatal("three concurrent calls did not start")
				}
			}
			close(phase.release)
		}
		var result Result
		select {
		case out := <-done:
			if out.err != nil {
				t.Fatal(out.err)
			}
			result = out.result
		case <-ctx.Done():
			t.Fatal("critique circle did not finish")
		}
		if len(result.Proposals) != 3 || len(result.Critiques) != 3 {
			t.Fatal("expected three proposals and three critiques")
		}
		for _, critique := range result.Critiques {
			peers := make(map[program.Role]bool)
			for _, subject := range critique.Subjects {
				peers[subject] = true
			}
			if len(critique.Subjects) != 2 || len(peers) != 2 || peers[critique.Author] {
				t.Fatalf("%s did not critique exactly two distinct peers: %v", critique.Author, critique.Subjects)
			}
			for _, proposal := range result.Proposals {
				if proposal.Author != critique.Author && !peers[proposal.Author] {
					t.Fatalf("%s omitted peer %s", critique.Author, proposal.Author)
				}
			}
		}
	})

	t.Run("failed proposal prevents critique", func(t *testing.T) {
		proposalErr := errors.New("author unavailable")
		var critiqued atomic.Bool
		runtime := &program.Runtime{Agent: func(_ context.Context, call program.Call) (any, error) {
			if call.Name == "critique" {
				critiqued.Store(true)
			}
			if call.Role == sswe {
				return nil, proposalErr
			}
			return "proposal", nil
		}}
		_, err := CritiqueCircle(t.Context(), runtime, input)
		if !errors.Is(err, proposalErr) {
			t.Fatalf("error = %v, want proposal failure", err)
		}
		if critiqued.Load() {
			t.Fatal("critique ran after a failed proposal")
		}
	})
}
