package program

import (
	"context"
	"errors"
	"fmt"
	"iter"
)

// Item is an in-memory checklist entry for these examples.
type Item struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Done    bool   `json:"done"`
}

// Iteration offers one failed item to the workflow's ordinary Go loop body.
type Iteration struct {
	Item    Item
	Attempt int
}

type LoopOptions struct {
	Validate      func(context.Context, Item) (bool, error)
	MaxIterations int
}

// Loop rechecks every item before choosing work, including after the last
// allowed body execution. Only these checks update Done in the supplied slice.
// Breaking the range stops the iterator; it does not declare completion.
// This example has no ledger persistence or concurrent mutation support.
func Loop(ctx context.Context, items []Item, options LoopOptions) iter.Seq2[Iteration, error] {
	return func(yield func(Iteration, error) bool) {
		if options.Validate == nil {
			yield(Iteration{}, errors.New("checklist: validator is required"))
			return
		}
		if options.MaxIterations <= 0 {
			yield(Iteration{}, errors.New("checklist: max iterations must be positive"))
			return
		}
		attempts := make([]int, len(items))
		for completed := 0; ; completed++ {
			if err := ctx.Err(); err != nil {
				yield(Iteration{}, err)
				return
			}
			firstFailed := -1
			for index := range items {
				if err := ctx.Err(); err != nil {
					yield(Iteration{}, err)
					return
				}
				passed, err := options.Validate(ctx, items[index])
				if err != nil {
					yield(Iteration{}, fmt.Errorf("validate %q: %w", items[index].Name, err))
					return
				}
				if err := ctx.Err(); err != nil {
					yield(Iteration{}, err)
					return
				}
				items[index].Done = passed
				if !passed && firstFailed == -1 {
					firstFailed = index
				}
			}
			if firstFailed == -1 {
				return
			}
			if completed == options.MaxIterations {
				yield(Iteration{}, fmt.Errorf("checklist: reached max iterations (%d)", options.MaxIterations))
				return
			}
			attempts[firstFailed]++
			if !yield(Iteration{Item: items[firstFailed], Attempt: attempts[firstFailed]}, nil) {
				return
			}
		}
	}
}
