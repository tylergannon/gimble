package gimble

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Session is one agent conversation on one harness, in one workdir. It
// belongs to the scope that created it and is closed when that scope ends.
type Session struct {
	adapter HarnessAdapter
	name    string
	model   string
	workdir string
	id      string // the creating scope's key, then name.ordinal, as in lap.3/coder.1

	mu      sync.Mutex
	native  string // the harness's session id, made on the first turn
	turns   int
	turnID  string
	running bool
	closed  bool
}

// NewSession creates a session in the scope the ctx is in, named for the
// graph. It cannot fail: the agent process starts on the first turn, and
// the adapter carries the harness-specific config. A session created outside
// Runtime.Run cannot generate turns or be forked.
func NewSession(ctx context.Context, name string, adapter HarnessAdapter, model, workdir string) *Session {
	s := &Session{adapter: adapter, name: name, model: model, workdir: workdir}
	if scope, err := current(ctx); err == nil {
		scope.adopt(s)
		scope.run.event(Event{Kind: "session_created", Scope: scope.key, Session: s.id, Name: name, Adapter: fmt.Sprintf("%T", adapter), Model: model, Workdir: workdir})
	}
	return s
}

// Text is the output of a prose turn: no schema is sent, and the result is
// the agent's final message.
type Text string

// Schema is empty: a prose turn sends no schema.
func (Text) Schema() json.RawMessage { return nil }

// ValidateJSON accepts the final message, a JSON string.
func (Text) ValidateJSON(raw []byte) error {
	var text string
	return json.Unmarshal(raw, &text)
}

// Generate runs one turn and blocks until it ends. T's schema is sent with
// the prompt, and the result is validated once, here, and decoded into T.
// A failed validation is an error. For Text no schema is sent and the
// result is the final message. The options attach supervisors.
func (s *Session) Generate[T Output](ctx context.Context, prompt string, opts ...AgentOption) (T, error) {
	if o := apply(opts); len(o.supervisors) > 0 {
		return supervise[T](ctx, s, prompt, o.supervisors)
	}
	var out T
	return generate[T](ctx, s, prompt, nil, fmt.Sprintf("%T", out))
}

func generate[T Output](ctx context.Context, s *Session, prompt string, onEvent func(Event), outputType string) (T, error) {
	var out T
	raw, err := s.turn(ctx, prompt, out.Schema(), onEvent, outputType)
	if err != nil {
		return out, err
	}
	if err := out.ValidateJSON(raw); err != nil {
		return out, fmt.Errorf("gimble: %s: the result does not validate: %w", s.id, err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("gimble: %s: decode the result: %w", s.id, err)
	}
	return out, nil
}

func (s *Session) turn(ctx context.Context, prompt string, schema json.RawMessage, onEvent func(Event), outputType string) (json.RawMessage, error) {
	s.mu.Lock()
	if err := s.usable(); err != nil {
		s.mu.Unlock()
		return nil, err
	}
	s.running = true
	s.turns++
	turnID := fmt.Sprintf("%s/turn.%d", s.id, s.turns)
	s.turnID = turnID
	native := s.native
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.turnID = ""
		s.mu.Unlock()
	}()

	if native == "" {
		id, err := s.adapter.CreateSession(ctx, s.model, s.workdir)
		if err != nil {
			return nil, fmt.Errorf("gimble: %s: %w", s.id, err)
		}
		s.mu.Lock()
		s.native, native = id, id
		s.mu.Unlock()
	}
	if onEvent == nil {
		onEvent = func(Event) {}
	}
	logf("%s: turn started (%s)", s.id, s.model)
	start := time.Now()
	scope, _ := current(ctx)
	if scope != nil {
		scope.run.event(Event{Kind: "turn_started", Scope: scope.key, Session: s.id, Turn: turnID, Text: prompt, OutputType: outputType})
	}
	var tokens []json.RawMessage
	wrapped := func(e Event) {
		e.Scope = scope.key
		e.Session = s.id
		e.Turn = turnID
		scope.run.sessionEvent(s.id, e)
		if e.Kind == "usage" && len(e.Data) != 0 {
			tokens = append(tokens, append(json.RawMessage(nil), e.Data...))
		}
		if onEvent != nil {
			onEvent(e)
		}
	}
	raw, err := s.adapter.RunTurn(ctx, native, prompt, schema, wrapped)
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	logf("%s: turn ended after %s: %v", s.id, time.Since(start).Round(time.Second), orNone(err))
	if ctx.Err() != nil {
		if scope != nil {
			scope.run.event(Event{Kind: "turn_ended", Scope: scope.key, Session: s.id, Turn: turnID, Error: ctx.Err().Error(), Tokens: tokens, Duration: time.Since(start), Interrupted: true})
		}
		return nil, ctx.Err()
	}
	if err != nil {
		if scope != nil {
			scope.run.event(Event{Kind: "turn_ended", Scope: scope.key, Session: s.id, Turn: turnID, Error: err.Error(), Tokens: tokens, Duration: time.Since(start)})
		}
		return nil, fmt.Errorf("gimble: %s: %w", s.id, err)
	}
	if scope != nil {
		scope.run.event(Event{Kind: "turn_ended", Scope: scope.key, Session: s.id, Turn: turnID, Result: raw, Tokens: tokens, Duration: time.Since(start)})
	}
	return raw, nil
}

// usable reports why s cannot start a turn. The caller holds s.mu.
func (s *Session) usable() error {
	switch {
	case s.id == "":
		return fmt.Errorf("gimble: session %q was not created in a run", s.name)
	case s.closed:
		return fmt.Errorf("gimble: session %s was used after its scope ended", s.id)
	case s.running:
		return fmt.Errorf("gimble: session %s is already running a turn", s.id)
	}
	return nil
}

// Steer injects a message into the turn that is running on this session.
// It is called from another goroutine while Generate blocks. The message
// lands at the worker's next model call. If no turn is running the message
// is dropped, and Steer returns nil either way. The error is for a harness
// that could not be reached.
func (s *Session) Steer(ctx context.Context, message string) error {
	s.mu.Lock()
	native, running, turn := s.native, s.running, s.turnID
	s.mu.Unlock()
	if !running || native == "" {
		logf("%s: steer dropped: %s", s.id, oneLine(message))
		if scope, err := current(ctx); err == nil {
			landed := false
			scope.run.event(Event{Kind: "steer", Scope: scope.key, Session: s.id, Turn: turn, Target: s.id, Source: steerSource(ctx), Message: message, Landed: &landed})
		}
		return nil
	}
	logf("%s: steer: %s", s.id, oneLine(message))
	if scope, err := current(ctx); err == nil {
		landed := true
		scope.run.event(Event{Kind: "steer", Scope: scope.key, Session: s.id, Turn: turn, Target: s.id, Source: steerSource(ctx), Message: message, Landed: &landed})
	}
	return s.adapter.Steer(ctx, native, message)
}

// Interrupt stops the running turn. It returns nil when no turn is running,
// since the caller can always race with a turn ending.
func (s *Session) Interrupt(ctx context.Context) error {
	s.mu.Lock()
	native, running, turn := s.native, s.running, s.turnID
	s.mu.Unlock()
	if !running || native == "" {
		return nil
	}
	if scope, err := current(ctx); err == nil {
		scope.run.event(Event{Kind: "interrupt", Scope: scope.key, Session: s.id, Turn: turn, Target: s.id, Source: steerSource(ctx)})
	}
	return s.adapter.Interrupt(ctx, native)
}

type steerSourceKey struct{}

func withSteerSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, steerSourceKey{}, source)
}
func steerSource(ctx context.Context) string {
	source, _ := ctx.Value(steerSourceKey{}).(string)
	return source
}

// Fork returns a new session, named for the graph, in the same workdir,
// with the same conversation so far. The two sessions are independent
// after that.
func (s *Session) Fork(ctx context.Context, name string) (*Session, error) {
	scope, err := current(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	err = s.usable()
	native := s.native
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	fork := &Session{adapter: s.adapter, name: name, model: s.model, workdir: s.workdir}
	if native != "" {
		if fork.native, err = s.adapter.Fork(ctx, native); err != nil {
			return nil, fmt.Errorf("gimble: fork %s: %w", s.id, err)
		}
	}
	scope.adopt(fork)
	scope.run.event(Event{Kind: "session_created", Scope: scope.key, Session: fork.id, Name: name, Adapter: fmt.Sprintf("%T", fork.adapter), Model: fork.model, Workdir: fork.workdir, Parent: s.id})
	logf("%s: forked from %s", fork.id, s.id)
	return fork, nil
}

func oneLine(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 160 {
		return text[:160] + "..."
	}
	return text
}
