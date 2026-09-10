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

// itemState keeps loop-owned metadata in one binding in the item's scope.
type itemState struct {
	Item
	Attempt int `json:"attempt"`
}

// Iteration offers one failed item to the workflow's ordinary Go loop body.
type Iteration struct {
	Item    Item
	Attempt int
	Context context.Context
}

type LoopOptions struct {
	Validate      func(context.Context, Item) (bool, error)
	MaxIterations int
	// Scope names the item's context key. Chapters and Sprints set it for you.
	Scope string
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
		contexts := make([]context.Context, len(items))
		keys := make([]Key, len(items))
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
				if contexts[index] == nil {
					itemContext := ctx
					if options.Scope != "" {
						var err error
						itemContext, err = Scope(ctx, options.Scope+": "+items[index].Name)
						if err != nil {
							yield(Iteration{}, err)
							return
						}
						keys[index], err = DeclareContext(itemContext, options.Scope)
						if err != nil {
							yield(Iteration{}, err)
							return
						}
					}
					contexts[index] = itemContext
				}
				itemContext := contexts[index]
				if options.Scope != "" {
					if err := SetContext(itemContext, keys[index], itemState{items[index], attempts[index]}); err != nil {
						yield(Iteration{}, err)
						return
					}
				}
				passed, err := options.Validate(itemContext, items[index])
				if err != nil {
					yield(Iteration{}, fmt.Errorf("validate %q: %w", items[index].Name, err))
					return
				}
				if err := ctx.Err(); err != nil {
					yield(Iteration{}, err)
					return
				}
				changed := items[index].Done != passed
				items[index].Done = passed
				if changed && options.Scope != "" {
					if err := SetContext(itemContext, keys[index], itemState{items[index], attempts[index]}); err != nil {
						yield(Iteration{}, err)
						return
					}
				}
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
			itemContext := contexts[firstFailed]
			if options.Scope != "" {
				if err := SetContext(itemContext, keys[firstFailed], itemState{items[firstFailed], attempts[firstFailed]}); err != nil {
					yield(Iteration{}, err)
					return
				}
			}
			if !yield(Iteration{Item: items[firstFailed], Attempt: attempts[firstFailed], Context: itemContext}, nil) {
				return
			}
		}
	}
}

// Chapters supplies an isolated context for each chapter, reused across retries
// and its validation calls. Nested Sprints inherit from chapter.Context.
func Chapters(ctx context.Context, items []Item, options LoopOptions) iter.Seq2[Iteration, error] {
	options.Scope = "chapter"
	return Loop(ctx, items, options)
}

// Sprints is the same checklist loop with automatic sprint context.
func Sprints(ctx context.Context, items []Item, options LoopOptions) iter.Seq2[Iteration, error] {
	options.Scope = "sprint"
	return Loop(ctx, items, options)
}
