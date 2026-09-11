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

// AgentOption is an argument to Generate. The same options are a
// supervisor's own where WithSupervisor attaches it, so a supervisor can
// have an interval and supervisors of its own.
type AgentOption func(*options)

type options struct {
	supervisors []supervisor
	every       time.Duration
}

type supervisor struct {
	session     *Session
	instruction string
	opts        []AgentOption
}

// WithSupervisor attaches a supervisor to a turn: a session of its own and
// an instruction saying what to watch for. While the turn runs, the
// supervisor looks at what the worker did since its last look, and each
// objection it raises is steered into the turn. It never holds up the
// result. opts are the supervisor's: WithInterval for how often it looks,
// and WithSupervisor for supervisors of its looks.
func WithSupervisor(session *Session, instruction string, opts ...AgentOption) AgentOption {
	return func(o *options) {
		o.supervisors = append(o.supervisors, supervisor{session: session, instruction: instruction, opts: opts})
	}
}

// WithInterval is how often a supervisor looks, given among its options to
// WithSupervisor. The default is three minutes.
func WithInterval(every time.Duration) AgentOption {
	return func(o *options) { o.every = every }
}

func apply(opts []AgentOption) options {
	o := options{every: 3 * time.Minute}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// review is a supervisor's answer to one look at the work.
type review struct {
	// Each objection is one thing the agent under review must change or stop doing, written as an instruction to that agent. Leave the list empty when you have no objection.
	Objections []string `json:"objections"`
}

// supervise is Generate with supervisors attached.
func supervise[T Output](ctx context.Context, s *Session, prompt string, supervisors []supervisor) (T, error) {
	t := &transcript{}
	for _, sup := range supervisors {
		s.mu.Lock()
		turn := fmt.Sprintf("%s/turn.%d", s.id, s.turns+1)
		s.mu.Unlock()
		if scope, err := current(ctx); err == nil {
			o := apply(sup.opts)
			scope.run.event(scope.key, s.id, turn, SuperviseAttached{Reviewer: sup.session.id, Worker: turn, Instruction: sup.instruction, Interval: o.every})
		}
	}

	// The worker's turn, in the background so the supervisors can run beside it.
	var res T
	var err error
	done := make(chan struct{})
	go func() {
		defer close(done)
		var out T
		res, err = generate[T](ctx, s, prompt, t.append, fmt.Sprintf("%T", out))
	}()

	// Each supervisor, on its own clock: look at what is new, steer on
	// objection, stop when the turn ends. A look is a turn with the
	// supervisor's own options, so it can be supervised in turn.
	var wg sync.WaitGroup
	for _, sup := range supervisors {
		wg.Go(func() {
			tick := time.NewTicker(apply(sup.opts).every)
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
				look := lookPrompt(sup, prompt, seen == 0, events)
				seen += len(events)
				review, err := sup.session.Generate[review](withSteerSource(ctx, sup.session.id), look, sup.opts...)
				if err != nil {
					logf("%s: a look at %s failed: %v", sup.session.id, s.id, err)
					continue
				}
				if len(review.Objections) > 0 {
					_ = s.Steer(withSteerSource(ctx, sup.session.id), "Your supervisor objects:\n\n- "+strings.Join(review.Objections, "\n- "))
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
func lookPrompt(sup supervisor, prompt string, first bool, events []AgentEvent) string {
	var b strings.Builder
	if first {
		fmt.Fprintf(&b, "You are supervising another agent in %s while it works. What you watch for:\n\n%s\n\n", sup.session.workdir, sup.instruction)
		fmt.Fprintf(&b, "The agent was asked:\n\n%s\n\n", prompt)
		b.WriteString("Every few minutes you are shown what it did since your last look; tool output is shortened, so read the files yourself when you need more, and change none. Object only when what it does goes against what you watch for; each objection is sent to the agent at once, as an instruction. When you have no objection, answer with an empty list.\n\n")
	}
	b.WriteString("What the agent did since your last look:\n\n")
	b.WriteString(renderEvents(events))
	return b.String()
}

func renderEvents(events []AgentEvent) string {
	var b strings.Builder
	for _, e := range events {
		switch e := e.(type) {
		case UserMessage:
			fmt.Fprintf(&b, "[message to the agent] %s\n", clip(e.Text))
		case AssistantMessage:
			fmt.Fprintf(&b, "[agent] %s\n", clip(e.Text))
		case AssistantMessageDelta:
			fmt.Fprintf(&b, "[agent fragment] %s\n", clip(e.Text))
		case ToolCall:
			fmt.Fprintf(&b, "[tool call: %s] %s\n", e.Tool, clip(string(e.Input)))
		case ToolResult:
			fmt.Fprintf(&b, "[tool result] %s\n", clip(string(e.Output)))
		case ToolResultDelta:
			fmt.Fprintf(&b, "[tool result fragment] %s\n", clip(string(e.Output)))
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
	events []AgentEvent
}

func (t *transcript) append(e AgentEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)
}

func (t *transcript) since(seen int) []AgentEvent {
	t.mu.Lock()
	defer t.mu.Unlock()
	return slices.Clone(t.events[seen:])
}
