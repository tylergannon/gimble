// Run with go run ./ephemeral/attest/issue126. The agents work in a temporary
// directory; the durable record is printed for inspection after the run.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tylergannon/gimble"
	"github.com/tylergannon/gimble/claude"
	"github.com/tylergannon/gimble/codex"
)

func main() {
	dir, err := os.MkdirTemp("", "gimble-issue126-")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Evidence directory:", dir)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	err = gimble.Run(gimble.Project(ctx, dir), "join", func(ctx context.Context) error {
		worker := gimble.NewSession(ctx, "worker", codex.New(), "gpt-5.6-luna", dir)
		reviewer := gimble.NewSession(ctx, "reviewer", claude.New(), "haiku", dir)
		result, err := worker.Generate[gimble.Text](ctx,
			"Run this shell command: while [ ! -f reviewer-started ]; do sleep 1; done; sleep 2. Then reply with exactly DONE. Do nothing else.",
			gimble.WithSupervisor(reviewer,
				"This is a cancellation test. Immediately run this shell command: touch reviewer-started; sleep 60. Wait for it to finish, then return no objections. Do nothing else.",
				gimble.WithInterval(time.Second)))
		fmt.Printf("Generate returned: result=%q error=%v\n", result, err)
		return err
	})
	fmt.Printf("Run returned: error=%v parent=%v elapsed=%s\n", err, ctx.Err(), time.Since(started).Round(time.Millisecond))
	if err != nil || ctx.Err() != nil {
		os.Exit(1)
	}
}
