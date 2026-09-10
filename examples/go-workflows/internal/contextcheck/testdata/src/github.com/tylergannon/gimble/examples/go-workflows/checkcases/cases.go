package checkcases

import (
	"context"

	workflow "github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func descendant(ctx context.Context) {
	const key = "go" + "al"
	workflow.SetContext(ctx, key, "outer")
	child, _ := workflow.Scope(ctx, "chapter")
	workflow.SetContext(child, "goal", "inner") // want `context key "goal" is also written in an ancestor or descendant scope`
}

func reverseOrder(ctx context.Context) {
	child, _ := workflow.Scope(ctx, "chapter")
	workflow.SetContext(child, "findings", "inner")
	workflow.SetContext(ctx, "findings", "outer") // want `context key "findings" is also written in an ancestor or descendant scope`
}

func nestedAndAliased(ctx context.Context) {
	root, _ := workflow.NewContext(ctx, "dir", nil)
	workflow.SetContext(root, "goal", "outer")
	child, _ := workflow.Scope(root, "chapter")
	grandchild, _ := workflow.Scope(child, "sprint")
	alias := grandchild
	cancelled, cancel := context.WithCancel(alias)
	defer cancel()
	withValue := context.WithValue(cancelled, struct{}{}, true)
	workflow.SetContext(withValue, "goal", "inner") // want `context key "goal" is also written in an ancestor or descendant scope`
}

func allowed(ctx context.Context, dynamicKey string) {
	root, _ := workflow.NewContext(ctx, "dir", nil)
	workflow.SetContext(root, "goal", "first")
	alias := root
	workflow.SetContext(alias, "goal", "updated")
	left, _ := workflow.Scope(root, "left")
	right, _ := workflow.Scope(root, "right")
	workflow.SetContext(left, "findings", "left")
	workflow.SetContext(right, "findings", "right")
	workflow.SetContext(left, dynamicKey, "unknown key")
	independent, _ := workflow.NewContext(root, "dir", nil)
	workflow.SetContext(independent, "goal", "independent root")
}

func exclusive(ctx context.Context, choose bool) {
	child, _ := workflow.Scope(ctx, "chapter")
	if choose {
		workflow.SetContext(ctx, "goal", "outer")
	} else {
		workflow.SetContext(child, "goal", "inner")
	}
}

func reassigned(ctx context.Context, choose bool) {
	root, _ := workflow.NewContext(ctx, "dir", nil)
	workflow.SetContext(root, "goal", "outer")
	child, _ := workflow.Scope(root, "chapter")
	child = root
	workflow.SetContext(child, "goal", "same scope after assignment")
	if choose {
		child, _ = workflow.Scope(root, "another chapter")
	}
	workflow.SetContext(child, "goal", "merged scope is unknown")
}

func opaqueHelper(ctx context.Context) context.Context { return ctx }

func unknown(ctx context.Context) {
	workflow.SetContext(ctx, "goal", "outer")
	child, _ := workflow.Scope(ctx, "chapter")
	workflow.SetContext(opaqueHelper(child), "goal", "helper alias is unknown")
}

func SetContext(context.Context, string, any) {}

func unrelated(ctx context.Context) {
	workflow.SetContext(ctx, "goal", "outer")
	child, _ := workflow.Scope(ctx, "chapter")
	SetContext(child, "goal", "unrelated function")
}
