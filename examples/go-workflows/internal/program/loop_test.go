package program

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestLoopRechecksStaleDoneAndLastAttempt(t *testing.T) {
	items := []Item{{Name: "shipping", Done: true}}
	ready, checks, bodies := false, 0, 0
	options := LoopOptions{MaxIterations: 1, Validate: func(context.Context, Item) (bool, error) {
		checks++
		return ready, nil
	}}
	for iteration, err := range Loop(context.Background(), items, options) {
		if err != nil {
			t.Fatal(err)
		}
		if iteration.Item.Done || iteration.Attempt != 1 || items[0].Done {
			t.Fatalf("stale done was trusted: %#v, %#v", iteration, items)
		}
		bodies++
		ready = true
	}
	if checks != 2 || bodies != 1 || !items[0].Done {
		t.Fatalf("checks=%d bodies=%d items=%#v", checks, bodies, items)
	}
}

func TestLoopReopensRegressionAfterOtherWork(t *testing.T) {
	items := []Item{{Name: "first"}, {Name: "second"}}
	ready := map[string]bool{"first": true}
	var worked []string
	checks := 0
	options := LoopOptions{MaxIterations: 2, Validate: func(_ context.Context, item Item) (bool, error) {
		checks++
		return ready[item.Name], nil
	}}
	for iteration, err := range Loop(context.Background(), items, options) {
		if err != nil {
			t.Fatal(err)
		}
		worked = append(worked, iteration.Item.Name)
		if iteration.Attempt != 1 {
			t.Fatalf("attempt is not per item: %#v", iteration)
		}
		if iteration.Item.Name == "second" {
			ready["first"] = false
		}
		ready[iteration.Item.Name] = true
	}
	if !reflect.DeepEqual(worked, []string{"second", "first"}) || checks != 6 || !items[0].Done || !items[1].Done {
		t.Fatalf("worked=%v checks=%d items=%#v", worked, checks, items)
	}
}

func TestLoopExhaustionStillChecksFinalAttempt(t *testing.T) {
	checks, bodies, failures := 0, 0, 0
	options := LoopOptions{MaxIterations: 2, Validate: func(context.Context, Item) (bool, error) {
		checks++
		return false, nil
	}}
	for iteration, err := range Loop(context.Background(), []Item{{Name: "broken"}}, options) {
		if err != nil {
			failures++
			if !strings.Contains(err.Error(), "max iterations (2)") {
				t.Fatal(err)
			}
			continue
		}
		bodies++
		if iteration.Attempt != bodies {
			t.Fatalf("attempt=%d bodies=%d", iteration.Attempt, bodies)
		}
	}
	if checks != 3 || bodies != 2 || failures != 1 {
		t.Fatalf("checks=%d bodies=%d errors=%d", checks, bodies, failures)
	}
}

func TestLoopCancellationAndEarlyBreak(t *testing.T) {
	for _, stop := range []string{"cancel", "break"} {
		t.Run(stop, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			checks, bodies, failures := 0, 0, 0
			items := []Item{{Name: "unfinished"}}
			options := LoopOptions{MaxIterations: 2, Validate: func(context.Context, Item) (bool, error) {
				checks++
				return false, nil
			}}
			for _, err := range Loop(ctx, items, options) {
				if err != nil {
					failures++
					if !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
					continue
				}
				bodies++
				if stop == "break" {
					break
				}
				cancel()
			}
			wantFailures := 0
			if stop == "cancel" {
				wantFailures = 1
			}
			if checks != 1 || bodies != 1 || failures != wantFailures || items[0].Done {
				t.Fatalf("checks=%d bodies=%d errors=%d items=%#v", checks, bodies, failures, items)
			}
		})
	}
}

func TestLoopRejectsInvalidOptionsAndValidatorErrors(t *testing.T) {
	failure := errors.New("validator unavailable")
	for _, options := range []LoopOptions{
		{MaxIterations: 1},
		{Validate: func(context.Context, Item) (bool, error) { t.Fatal("unexpected validation"); return false, nil }},
		{MaxIterations: 1, Validate: func(context.Context, Item) (bool, error) { return false, failure }},
	} {
		failures := 0
		for _, err := range Loop(context.Background(), []Item{{Name: "work"}}, options) {
			if err == nil {
				t.Fatal("invalid loop yielded work")
			}
			failures++
		}
		if failures != 1 {
			t.Fatalf("errors=%d, want one", failures)
		}
	}
}
