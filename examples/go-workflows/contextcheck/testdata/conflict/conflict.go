// Package conflict is an intentionally invalid workflow for exercising the CLI.
// It is excluded from ordinary ./... package discovery by the testdata directory.
package conflict

import (
	"context"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func Workflow(ctx context.Context) error {
	if err := program.SetContext(ctx, "goal", "ship the change"); err != nil {
		return err
	}
	chapter, err := program.Scope(ctx, "chapter")
	if err != nil {
		return err
	}
	return program.SetContext(chapter, "goal", "replace the outer goal")
}
