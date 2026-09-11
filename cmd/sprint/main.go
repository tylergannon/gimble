// Command sprint runs the sprint workflow on the repository in the current
// directory, as in: go run ./cmd/sprint -sprint 2
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/tylergannon/gimble"
	"github.com/tylergannon/gimble/internal/workflows/sprint"
)

func main() {
	var in sprint.Input
	flag.IntVar(&in.Sprint, "sprint", 0, "the sprint of SPRINTS.md to build")
	flag.StringVar(&in.Model, "model", "gpt-5.6-luna", "Codex model for the researcher, the planner, and the coders")
	flag.StringVar(&in.ReviewModel, "review-model", "haiku", "Claude Code model for the supervisors and the validator")
	flag.IntVar(&in.Laps, "laps", 10, "the most laps to run in all")
	port := flag.Int("port", 8080, "loopback TCP port for the web application")
	uds := flag.String("uds", "", "Unix-domain socket for the web application instead of TCP")
	noWeb := flag.Bool("no-web", false, "run without the web application")
	flag.Parse()
	repo, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	in.Repo = repo

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var runtime *gimble.Runtime
	switch {
	case *noWeb:
		runtime, err = gimble.NewRuntime(ctx, filepath.Join(repo, ".gimble"), gimble.WithNoWeb())
	case *uds != "":
		runtime, err = gimble.NewRuntime(ctx, filepath.Join(repo, ".gimble"), gimble.WithUDS(*uds))
	default:
		runtime, err = gimble.NewRuntime(ctx, filepath.Join(repo, ".gimble"), gimble.WithPort(*port))
	}
	if err != nil {
		log.Fatal(err)
	}
	err = runtime.Run(ctx, "sprint", func(ctx context.Context) error {
		return sprint.Sprint(ctx, in)
	})
	if err != nil {
		log.Fatal(err)
	}
}
