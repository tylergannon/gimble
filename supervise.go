package gimble

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
)

//go:generate go tool polytype --validate

// Option is an argument to Generate. The only one is Supervise.
type Option struct {
	reviewer    *Session
	instruction string
}

// Supervise attaches a reviewer to a turn: a session of its own and an
// instruction saying what to watch for. The reviewer looks each time the
// worker produces a tool result and steers it on objection; when the turn
// ends it looks at the result, and its objections become a follow-up turn
// on the worker's session, up to three rounds.
func Supervise(reviewer *Session, instruction string) Option {
	return Option{reviewer: reviewer, instruction: instruction}
}

// Review is a reviewer's answer to one look at the work.
type Review struct {
	// Each objection is one thing the agent under review must change or stop doing, written as an instruction to that agent. Leave the list empty when you have no objection.
	Objections []string `json:"objections"`
}

const maxRounds = 3

// supervise is Generate with reviewers attached.
func supervise[T Output](ctx context.Context, s *Session, prompt string, reviewers []Option) (T, error) {
	briefed := make([]bool, len(reviewers))
	for round := 1; ; round++ {
		t := &transcript{wake: make(chan struct{})}

		// The worker's turn, in the background so the reviewers can run beside it.
		var res T
		var err error
		done := make(chan struct{})
		go func() {
			defer close(done)
			res, err = generate[T](ctx, s, prompt, t.append)
		}()

		// Each reviewer, self-paced: look whenever there is a new tool result,
		// steer on objection, stop looking when the turn ends.
		seen := make([]int, len(reviewers))
		var wg sync.WaitGroup
		for i, r := range reviewers {
			wg.Go(func() {
				for {
					events, ok := t.next(done, seen[i])
					if !ok {
						return
					}
					seen[i] += len(events)
					look := lookPrompt(r, prompt, briefed[i], events, nil)
					briefed[i] = true
					review, err := r.reviewer.Generate[Review](ctx, look)
					if err != nil {
						logf("%s: review of %s failed: %v", r.reviewer.id, s.id, err)
						continue
					}
					if len(review.Objections) > 0 {
						_ = s.Steer(ctx, "A reviewer objects:\n\n- "+strings.Join(review.Objections, "\n- "))
					}
				}
			})
		}
		<-done
		wg.Wait()
		if err != nil {
			return res, err
		}

		// A final look at the finished answer.
		answer, _ := json.Marshal(res)
		var objections []string
		for i, r := range reviewers {
			look := lookPrompt(r, prompt, briefed[i], t.since(seen[i]), answer)
			briefed[i] = true
			review, err := r.reviewer.Generate[Review](ctx, look)
			if err != nil {
				return res, err
			}
			objections = append(objections, review.Objections...)
		}
		if len(objections) == 0 {
			return res, nil // every reviewer approved
		}
		if round == maxRounds {
			return res, fmt.Errorf("supervise: %d rounds without approval", round)
		}
		logf("%s: round %d ended with %d objections", s.id, round, len(objections))
		prompt = "Your reviewers object to the finished work:\n\n- " + strings.Join(objections, "\n- ") +
			"\n\nAddress every objection, then answer again as you were first asked to."
	}
}

// lookPrompt asks a reviewer for objections: to the work so far while the
// turn runs, or to the finished work when answer is set.
func lookPrompt(r Option, prompt string, briefed bool, events []Event, answer []byte) string {
	var b strings.Builder
	if !briefed {
		fmt.Fprintf(&b, "You are reviewing the work of another agent in %s while it happens. What you watch for:\n\n%s\n\n", r.reviewer.workdir, r.instruction)
		fmt.Fprintf(&b, "The agent was asked:\n\n%s\n\n", prompt)
		b.WriteString("You will be shown what it does as it goes. Object only when what it does goes against what you watch for; each objection is sent to the agent at once, as an instruction. You may read the files yourself; do not change any. When you have no objection, answer with an empty list.\n\n")
	}
	if len(events) > 0 {
		b.WriteString("What the agent did since your last look:\n\n")
		b.WriteString(renderEvents(events))
		b.WriteString("\n")
	}
	if answer != nil {
		fmt.Fprintf(&b, "The agent has finished its turn. Its answer:\n\n%s\n\n", render(answer))
		b.WriteString("Object to anything in the finished work that goes against what you watch for; the agent will be asked to fix it. Answer with an empty list to approve.")
	} else {
		b.WriteString("Any objections?")
	}
	return b.String()
}

func renderEvents(events []Event) string {
	var b strings.Builder
	for _, e := range events {
		switch e.Kind {
		case "user":
			fmt.Fprintf(&b, "[message to the agent] %s\n", clip(e.Text))
		case "assistant":
			fmt.Fprintf(&b, "[agent] %s\n", clip(e.Text))
		case "tool_call":
			fmt.Fprintf(&b, "[tool call: %s] %s\n", e.Tool, clip(string(e.Data)))
		case "tool_result":
			fmt.Fprintf(&b, "[tool result] %s\n", clip(string(e.Data)))
		}
	}
	return b.String()
}

func clip(text string) string {
	const limit = 2000
	if len(text) <= limit {
		return text
	}
	return strings.ToValidUTF8(text[:limit], "") + fmt.Sprintf(" [... %d more bytes]", len(text)-limit)
}

// transcript is the worker's events in one turn, appended as they arrive.
type transcript struct {
	mu     sync.Mutex
	events []Event
	wake   chan struct{} // closed and replaced whenever an event arrives
}

func (t *transcript) append(e Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
	close(t.wake)
	t.wake = make(chan struct{})
}

// next blocks until a tool result has arrived after the first seen events
// and returns every event after seen, or returns false once the turn is done.
func (t *transcript) next(done <-chan struct{}, seen int) ([]Event, bool) {
	for {
		t.mu.Lock()
		events := slices.Clone(t.events[seen:])
		wake := t.wake
		t.mu.Unlock()
		if slices.ContainsFunc(events, func(e Event) bool { return e.Kind == "tool_result" }) {
			return events, true
		}
		select {
		case <-done:
			return nil, false
		case <-wake:
		}
	}
}

func (t *transcript) since(seen int) []Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return slices.Clone(t.events[seen:])
}
