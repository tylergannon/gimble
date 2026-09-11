package gimble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tylergannon/gimble"
)

type exampleAdapter struct{}

func (*exampleAdapter) CreateSession(context.Context, string, string) (string, error) {
	return "example-session", nil
}

func (a *exampleAdapter) RunTurn(_ context.Context, _ string, prompt string, schema json.RawMessage, emit func(gimble.Event)) (json.RawMessage, error) {
	emit(gimble.Event{Kind: "user", Text: prompt})
	return json.Marshal("done")
}

func (*exampleAdapter) Steer(context.Context, string, string) error { return nil }
func (*exampleAdapter) Fork(context.Context, string) (string, error) {
	return "example-fork", nil
}

func exampleContext() (context.Context, func()) {
	dir, err := os.MkdirTemp("", "gimble-example-")
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return gimble.Project(ctx, dir), func() {
		cancel()
		_ = os.RemoveAll(dir)
	}
}

func Example() {
	ctx, closeProject := exampleContext()
	defer closeProject()

	err := gimble.Run(ctx, "example", func(ctx context.Context) error {
		if err := gimble.Set(ctx, "goal", "demonstrate the public API"); err != nil {
			return err
		}
		worker := gimble.NewSession(ctx, "worker", &exampleAdapter{}, "example", ".")
		answer, err := worker.Generate[gimble.Text](ctx,
			"Complete the goal.\n\n"+gimble.ScopeText(ctx))
		if err != nil {
			return err
		}
		fmt.Println(answer)
		return nil
	})
	fmt.Println(err)

	// Output:
	// done
	// <nil>
}

func ExampleGroup() {
	ctx, closeProject := exampleContext()
	defer closeProject()

	err := gimble.Run(ctx, "parallel", func(ctx context.Context) error {
		group := gimble.Group(ctx, "drafts")
		group.Go("draft", func(ctx context.Context) error {
			return gimble.Set(ctx, "approach", "first")
		})
		group.Go("draft", func(ctx context.Context) error {
			return gimble.Set(ctx, "approach", "second")
		})
		return group.Wait()
	})
	fmt.Println(err)

	// Output: <nil>
}
