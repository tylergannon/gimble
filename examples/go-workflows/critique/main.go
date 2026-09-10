package main

import (
	"context"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
	"golang.org/x/sync/errgroup"
)

const (
	sswe   program.Role = "sswe"
	engMgr program.Role = "eng-mgr"
	tester program.Role = "tester"
)

type Proposal struct {
	Author program.Role `json:"author"`
	Text   string       `json:"text"`
}

type Critique struct {
	Author   program.Role   `json:"author"`
	Subjects []program.Role `json:"subjects"`
	Text     string         `json:"text"`
}

type Result struct {
	Proposals []Proposal `json:"proposals"`
	Critiques []Critique `json:"critiques"`
}

func CritiqueCircle(ctx context.Context, runtime *program.Runtime, input Input) (Result, error) {
	if err := input.validate(); err != nil {
		return Result{}, err
	}
	authors := []program.Role{sswe, engMgr, tester}
	proposals := make([]Proposal, len(authors))
	drafts, draftCtx := errgroup.WithContext(ctx)
	for i, author := range authors {
		drafts.Go(func() error {
			branch, err := runtime.Worktree(draftCtx, program.Workspace(ctx), "proposal-"+string(author))
			if err != nil {
				return err
			}
			text, err := program.Codergen[string](branch, runtime, "propose", author, proposalPrompt(input))
			proposals[i] = Proposal{author, text}
			return err
		})
	}
	if err := drafts.Wait(); err != nil {
		return Result{}, err
	}

	// Collected proposals remain unchanged throughout the peer-review phase.
	critiques := make([]Critique, len(authors))
	reviews, reviewCtx := errgroup.WithContext(ctx)
	for i, author := range authors {
		reviews.Go(func() error {
			branch, err := runtime.Worktree(reviewCtx, program.Workspace(ctx), "critique-"+string(author))
			if err != nil {
				return err
			}
			var peers []Proposal
			var subjects []program.Role
			for _, proposal := range proposals {
				if proposal.Author != author {
					peers = append(peers, proposal)
					subjects = append(subjects, proposal.Author)
				}
			}
			text, err := program.Codergen[string](branch, runtime, "critique", author, critiquePrompt(input, proposals[i], peers))
			critiques[i] = Critique{author, subjects, text}
			return err
		})
	}
	if err := reviews.Wait(); err != nil {
		return Result{}, err
	}
	return Result{proposals, critiques}, nil
}
