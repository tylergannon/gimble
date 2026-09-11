package gimble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/tylergannon/gimble"
)

type exampleAdapter struct {
	mu    sync.Mutex
	turns int
}

func (*exampleAdapter) CreateSession(context.Context, string, string) (string, error) {
	return "example-session", nil
}

func (a *exampleAdapter) RunTurn(_ context.Context, _ string, prompt string, schema json.RawMessage, emit func(gimble.Event)) (json.RawMessage, error) {
	emit(gimble.Event{Kind: "user", Text: prompt})
	if len(schema) == 0 {
		return json.Marshal("done")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.turns++
	if a.turns == 1 {
		return json.RawMessage(`{"next":"write the implementation"}`), nil
	}
	return json.RawMessage(`{"next":""}`), nil
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

func ExampleLoop() {
	ctx, closeProject := exampleContext()
	defer closeProject()

	err := gimble.Run(ctx, "planned", func(ctx context.Context) error {
		planner := gimble.NewSession(ctx, "planner", &exampleAdapter{}, "example", ".")
		loop := gimble.Loop(ctx, "delivery", "the implementation works", planner)
		for _, task := range loop.Laps {
			fmt.Println(task.Text)
		}
		return loop.Err()
	})
	if err != nil {
		fmt.Println("error:", err)
	}

	// Output: write the implementation
}
