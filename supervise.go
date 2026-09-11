package gimble

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

//go:generate go tool polytype --validate

// Option is an argument to Generate. The only one is Supervise.
type Option struct {
	supervisor  *Session
	instruction string
	every       time.Duration
}

// Supervise attaches a supervisor to a turn: a session of its own and an
// instruction saying what to watch for. Every three minutes while the turn
// runs, or at the interval given, the supervisor looks at what the worker
// did since its last look, and each objection it raises is steered into
// the turn. It never holds up the result.
func Supervise(supervisor *Session, instruction string, every ...time.Duration) Option {
	o := Option{supervisor: supervisor, instruction: instruction, every: 3 * time.Minute}
	if len(every) > 0 {
		o.every = every[0]
	}
	return o
}

// Review is a supervisor's answer to one look at the work.
type Review struct {
	// Each objection is one thing the agent under review must change or stop doing, written as an instruction to that agent. Leave the list empty when you have no objection.
	Objections []string `json:"objections"`
}

// supervise is Generate with supervisors attached.
func supervise[T Output](ctx context.Context, s *Session, prompt string, supervisors []Option) (T, error) {
	t := &transcript{}

	// The worker's turn, in the background so the supervisors can run beside it.
	var res T
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		res, err = generate[T](ctx, s, prompt, t.append)
	}()

	// Each supervisor, on its own clock: look at what is new, steer on
	// objection, stop when the turn ends.
	var wg sync.WaitGroup
	for _, o := range supervisors {
		wg.Go(func() {
			tick := time.NewTicker(o.every)
			defer tick.Stop()
			seen := 0
			for {
				select {
				case <-done:
					return
				case <-tick.C:
				}
				events := t.since(seen)
				if len(events) == 0 {
					continue
				}
				look := lookPrompt(o, prompt, seen == 0, events)
				seen += len(events)
				review, err := o.supervisor.Generate[Review](ctx, look)
				if err != nil {
					logf("%s: a look at %s failed: %v", o.supervisor.id, s.id, err)
					continue
				}
				if len(review.Objections) > 0 {
					_ = s.Steer(ctx, "Your supervisor objects:\n\n- "+strings.Join(review.Objections, "\n- "))
				}
			}
		})
	}
	<-done
	wg.Wait()
	return res, err
}

// lookPrompt asks a supervisor for objections to what the worker did since
// its last look; the first look also carries the instruction and the task.
func lookPrompt(o Option, prompt string, first bool, events []Event) string {
	var b strings.Builder
	if first {
		fmt.Fprintf(&b, "You are supervising another agent in %s while it works. What you watch for:\n\n%s\n\n", o.supervisor.workdir, o.instruction)
		fmt.Fprintf(&b, "The agent was asked:\n\n%s\n\n", prompt)
		b.WriteString("Every few minutes you are shown what it did since your last look; tool output is shortened, so read the files yourself when you need more, and change none. Object only when what it does goes against what you watch for; each objection is sent to the agent at once, as an instruction. When you have no objection, answer with an empty list.\n\n")
	}
	b.WriteString("What the agent did since your last look:\n\n")
	b.WriteString(renderEvents(events))
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
}

func (t *transcript) append(e Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
}

func (t *transcript) since(seen int) []Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return slices.Clone(t.events[seen:])
}
