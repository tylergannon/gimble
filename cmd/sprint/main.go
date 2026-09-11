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
	"strconv"

	"github.com/tylergannon/gimble"
	"github.com/tylergannon/gimble/sprint"
)

func main() {
	in := sprint.Input{Issues: []int{}}
	flag.IntVar(&in.Sprint, "sprint", 0, "the sprint of SPRINTS.md to build")
	flag.StringVar(&in.Model, "model", "gpt-5.6-luna", "Codex model for the researchers, the planners, and the coders")
	flag.StringVar(&in.ReviewModel, "review-model", "haiku", "Claude Code model for the supervisors and the validators")
	flag.IntVar(&in.Laps, "laps", 10, "the most laps one piece of work may run")
	flag.Func("issue", "a GitHub issue to work beside the sprint; repeat for more", func(s string) error {
		n, err := strconv.Atoi(s)
		in.Issues = append(in.Issues, n)
		return err
	})
	flag.Parse()
	repo, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	in.Repo = repo

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx = gimble.Project(ctx, filepath.Join(repo, ".gimble"))
	err = gimble.Run(ctx, "sprint", func(ctx context.Context) error {
		return sprint.Sprint(ctx, in)
	})
	if err != nil {
		log.Fatal(err)
	}
}
