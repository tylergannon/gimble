package program

import (
	"context"
	"iter"
)

type Item struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Done    bool   `json:"done"`
}
type Iteration struct {
	Item    Item
	Attempt int
	Context context.Context
}
type LoopOptions struct {
	Validate      func(context.Context, Item) (bool, error)
	MaxIterations int
	Scope         string
}

// Loop yields each supplied item once; validation and Done are intentionally stubbed.
func Loop(ctx context.Context, items []Item, _ LoopOptions) iter.Seq2[Iteration, error] {
	return func(y func(Iteration, error) bool) {
		for _, item := range items {
			if !y(Iteration{Item: item, Attempt: 1, Context: ctx}, nil) {
				return
			}
		}
	}
}
func Chapters(ctx context.Context, items []Item, o LoopOptions) iter.Seq2[Iteration, error] {
	return Loop(ctx, items, o)
}
func Sprints(ctx context.Context, items []Item, o LoopOptions) iter.Seq2[Iteration, error] {
	return Loop(ctx, items, o)
}
