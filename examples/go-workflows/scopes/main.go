package main

import (
	"context"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
	"golang.org/x/sync/errgroup"
)

const (
	sswe   program.Role = "sswe"
	tester program.Role = "tester"
	engMgr program.Role = "eng-mgr"
)

func NestedScopes(ctx context.Context, runtime *program.Runtime, input Input) ([]program.Item, error) {
	chapters, sprints, err := input.ledgers()
	if err != nil {
		return nil, err
	}
	goal, err := program.DeclareContext(ctx, "goal")
	if err != nil {
		return nil, err
	}
	focus, err := program.DeclareContext(ctx, "focus")
	if err != nil {
		return nil, err
	}
	if err := program.SetContext(ctx, goal, input.Goal); err != nil {
		return nil, err
	}
	if err := program.SetContext(ctx, focus, "Deliver the declared quote behavior."); err != nil {
		return nil, err
	}
	if _, err := program.Codergen[string](ctx, runtime, "start", engMgr, "Frame the work."); err != nil {
		return nil, err
	}

	reviewed := false
	for chapter, err := range program.Chapters(ctx, chapters, loopOptions(runtime)) {
		if err != nil {
			return nil, err
		}
		chapterFocus, err := program.DeclareContext(chapter.Context, "chapter_focus")
		if err != nil {
			return nil, err
		}
		if err := program.SetContext(chapter.Context, chapterFocus, "Complete "+chapter.Item.Name+" without expanding scope."); err != nil {
			return nil, err
		}
		for sprint, err := range program.Sprints(chapter.Context, sprints[chapter.Item.Name], loopOptions(runtime)) {
			if err != nil {
				return nil, err
			}
			if _, err := program.Codergen[string](sprint.Context, runtime, "build", sswe, "Implement the current sprint."); err != nil {
				return nil, err
			}
			if !reviewed {
				if err := parallelReviews(sprint.Context, runtime); err != nil {
					return nil, err
				}
				if _, err := program.Codergen[string](sprint.Context, runtime, "after-children", sswe, "Summarize the sprint's next step."); err != nil {
					return nil, err
				}
				reviewed = true
			}
			if err := program.SetContext(chapter.Context, chapterFocus, "Build on the completed sprint: "+sprint.Item.Name); err != nil {
				return nil, err
			}
		}
		if _, err := program.Codergen[string](chapter.Context, runtime, "chapter-restored", engMgr, "Summarize the chapter's progress."); err != nil {
			return nil, err
		}
	}
	_, err = program.Codergen[string](ctx, runtime, "root-restored", engMgr, "Summarize the overall outcome.")
	return chapters, err
}

func parallelReviews(ctx context.Context, runtime *program.Runtime) error {
	group, reviewCtx := errgroup.WithContext(ctx)
	for _, review := range reviewAssignments() {
		group.Go(func() error {
			child, err := program.Scope(reviewCtx, review.Name)
			if err != nil {
				return err
			}
			focus, err := program.DeclareContext(child, "review_focus")
			if err != nil {
				return err
			}
			if err := program.SetContext(child, focus, review.Focus); err != nil {
				return err
			}
			_, err = program.Codergen[string](child, runtime, "review", review.Role, "Review the current sprint.")
			return err
		})
	}
	return group.Wait()
}

func loopOptions(runtime *program.Runtime) program.LoopOptions {
	return program.LoopOptions{
		MaxIterations: 12,
		Validate: func(ctx context.Context, item program.Item) (bool, error) {
			check, err := runtime.Command(ctx, item.Command)
			return check.Passed, err
		},
	}
}
