package gimble

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// fake is a HarnessAdapter whose turns are answered by a function.
type fake struct {
	answer func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error)

	mu      sync.Mutex
	made    int
	steers  []string
	running map[string]func(Event)
}

func (f *fake) CreateSession(ctx context.Context, model, workdir string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.made++
	return "native-" + string(rune('0'+f.made)), nil
}

func (f *fake) RunTurn(ctx context.Context, session, prompt string, schema json.RawMessage, onEvent func(Event)) (json.RawMessage, error) {
	f.mu.Lock()
	if f.running == nil {
		f.running = map[string]func(Event){}
	}
	f.running[session] = onEvent
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		delete(f.running, session)
		f.mu.Unlock()
	}()
	onEvent(Event{Kind: "user", Text: prompt})
	out, err := f.answer(ctx, session, prompt, schema, onEvent)
	if err != nil {
		return nil, err
	}
	if len(schema) == 0 {
		return json.Marshal(out)
	}
	return json.RawMessage(out), nil
}

func (f *fake) Steer(ctx context.Context, session, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if emit := f.running[session]; emit != nil {
		f.steers = append(f.steers, message)
		emit(Event{Kind: "user", Text: message})
	}
	return nil
}

func (f *fake) Fork(ctx context.Context, session string) (string, error) {
	return session + "-fork", nil
}

func runTest(t *testing.T, body func(ctx context.Context) error) error {
	t.Helper()
	return Run(Project(t.Context(), t.TempDir()), "test", body)
}

func TestScopeData(t *testing.T) {
	err := runTest(t, func(ctx context.Context) error {
		if err := Set(ctx, "language", "go"); err != nil {
			return err
		}
		if err := Set(ctx, "language", "rust"); err == nil {
			t.Error("a second Set of a key in one scope succeeded")
		}
		if err := SetJSON(ctx, "review", Review{Objections: []string{"too big"}}); err != nil {
			return err
		}
		var inner context.Context
		err := Scope(ctx, "lap", func(ctx context.Context) error {
			inner = ctx
			return Set(ctx, "language", "zig")
		})
		if err != nil {
			return err
		}
		got := ScopeText(inner)
		want := "## review\n\n{\n  \"objections\": [\n    \"too big\"\n  ]\n}\n\n## language\n\nzig"
		if got != want {
			t.Errorf("ScopeText = %q, want %q", got, want)
		}
		if err := Set(inner, "late", "x"); err == nil {
			t.Error("Set on an ended scope succeeded")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerate(t *testing.T) {
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error) {
		switch {
		case len(schema) == 0:
			return "hello", nil
		case prompt == "bad":
			return `{"objections": "not a list"}`, nil
		default:
			return `{"objections": ["one"]}`, nil
		}
	}}
	var escaped *Session
	err := runTest(t, func(ctx context.Context) error {
		s := NewSession(ctx, "coder", f, "m", "/w")
		text, err := s.Generate[Text](ctx, "hi")
		if err != nil || text != "hello" {
			t.Errorf("Generate[Text] = %q, %v", text, err)
		}
		review, err := s.Generate[Review](ctx, "review")
		if err != nil || len(review.Objections) != 1 {
			t.Errorf("Generate[Review] = %v, %v", review, err)
		}
		if _, err := s.Generate[Review](ctx, "bad"); err == nil {
			t.Error("a result that does not validate was accepted")
		}
		if s.id != "coder.1" || NewSession(ctx, "coder", f, "m", "/w").id != "coder.2" {
			t.Errorf("session ids: %q", s.id)
		}
		return Scope(ctx, "lap", func(ctx context.Context) error {
			escaped = NewSession(ctx, "coder", f, "m", "/w")
			if escaped.id != "lap.1/coder.1" {
				t.Errorf("id = %q", escaped.id)
			}
			return nil
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := escaped.Generate[Text](t.Context(), "zombie"); err == nil {
		t.Error("a session whose scope ended ran a turn")
	}
}

func TestGroupFirstErrorCancelsTheRest(t *testing.T) {
	boom := errors.New("boom")
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error) {
		if prompt == "fail" {
			return "", boom
		}
		<-ctx.Done()
		return "", ctx.Err()
	}}
	var slow error
	err := runTest(t, func(ctx context.Context) error {
		g := Group(ctx, "bakeoff")
		g.Go("attempt", func(ctx context.Context) error {
			_, slow = NewSession(ctx, "candidate", f, "m", "/w").Generate[Text](ctx, "wait")
			return nil
		})
		g.Go("attempt", func(ctx context.Context) error {
			_, err := NewSession(ctx, "candidate", f, "m", "/w").Generate[Text](ctx, "fail")
			return err
		})
		return g.Wait()
	})
	if !errors.Is(err, boom) {
		t.Fatalf("Wait = %v, want boom", err)
	}
	if !errors.Is(slow, context.Canceled) {
		t.Fatalf("the other turn ended with %v, want context.Canceled", slow)
	}
}

func TestFork(t *testing.T) {
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error) {
		return session, nil
	}}
	err := runTest(t, func(ctx context.Context) error {
		researcher := NewSession(ctx, "researcher", f, "m", "/w")
		if _, err := researcher.Generate[Text](ctx, "prime"); err != nil {
			return err
		}
		judge, err := researcher.Fork(ctx, "judge")
		if err != nil {
			return err
		}
		native, err := judge.Generate[Text](ctx, "who")
		if native != "native-1-fork" || judge.id != "judge.1" || judge.workdir != "/w" {
			t.Errorf("fork ran on %q as %q in %q", native, judge.id, judge.workdir)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSupervise(t *testing.T) {
	var workerTurns, looks int
	f := &fake{}
	f.answer = func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error) {
		if session == "native-1" { // the worker, which runs until the supervisor's steer lands
			workerTurns++
			emit(Event{Kind: "tool_call", CallID: "1", Tool: "shell", Data: json.RawMessage(`{"command":"make plugins"}`)})
			emit(Event{Kind: "tool_result", CallID: "1", Data: json.RawMessage(`"built a plugin system"`)})
			for range 200 {
				f.mu.Lock()
				n := len(f.steers)
				f.mu.Unlock()
				if n > 0 {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			return "done", nil
		}
		f.mu.Lock() // the supervisor
		looks++
		f.mu.Unlock()
		if strings.Contains(prompt, "no plugin systems") && strings.Contains(prompt, "built a plugin system") {
			return `{"objections": ["remove the plugin system"]}`, nil
		}
		return `{"objections": []}`, nil
	}
	err := runTest(t, func(ctx context.Context) error {
		worker := NewSession(ctx, "coder", f, "m", "/w")
		supervisor := NewSession(ctx, "taste", f, "m", "/w")
		res, err := worker.Generate[Text](ctx, "build it", WithSupervisor(supervisor, "no plugin systems", WithInterval(10*time.Millisecond)))
		if res != "done" {
			t.Errorf("result %q", res)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.steers) == 0 || !strings.Contains(f.steers[0], "remove the plugin system") {
		t.Errorf("steers: %q", f.steers)
	}
	if workerTurns != 1 || looks != 1 {
		t.Errorf("worker ran %d turns and the supervisor looked %d times, want 1 and 1: a look waits for something new", workerTurns, looks)
	}
}

func TestSuperviseASupervisor(t *testing.T) {
	f := &fake{}
	steered := func(text string) bool {
		for range 400 {
			f.mu.Lock()
			found := slices.ContainsFunc(f.steers, func(s string) bool { return strings.Contains(s, text) })
			f.mu.Unlock()
			if found {
				return true
			}
			time.Sleep(5 * time.Millisecond)
		}
		return false
	}
	f.answer = func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error) {
		switch {
		case session == "native-1": // the worker, until its supervisor objects
			emit(Event{Kind: "tool_result", CallID: "1", Data: json.RawMessage(`"built a plugin system"`)})
			steered("remove the plugin system")
			return "done", nil
		case session == "native-2" && strings.Contains(prompt, "no plugin systems"): // the supervisor's first look, until its own supervisor objects
			emit(Event{Kind: "assistant", Text: "I object to the variable names"})
			if !steered("object only to plugin systems") {
				return `{"objections": []}`, nil
			}
			return `{"objections": ["remove the plugin system"]}`, nil
		case session == "native-3" && strings.Contains(prompt, "no nitpicking"): // the supervisor's supervisor
			return `{"objections": ["object only to plugin systems"]}`, nil
		}
		return `{"objections": []}`, nil
	}
	err := runTest(t, func(ctx context.Context) error {
		worker := NewSession(ctx, "coder", f, "m", "/w")
		supervisor := NewSession(ctx, "taste", f, "m", "/w")
		lead := NewSession(ctx, "lead", f, "m", "/w")
		_, err := worker.Generate[Text](ctx, "build it", WithSupervisor(supervisor, "no plugin systems",
			WithInterval(10*time.Millisecond),
			WithSupervisor(lead, "no nitpicking", WithInterval(10*time.Millisecond)),
		))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.steers) < 2 || !strings.Contains(f.steers[0], "object only to plugin systems") || !strings.Contains(f.steers[1], "remove the plugin system") {
		t.Errorf("steers: %q, want the lead's steer to the supervisor, then the supervisor's to the worker", f.steers)
	}
}

func TestLoop(t *testing.T) {
	var prompts []string
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(Event)) (string, error) {
		prompts = append(prompts, prompt)
		file := strings.Fields(prompt[strings.Index(prompt, "Its backlog is the file ")+len("Its backlog is the file "):])[0]
		file = strings.TrimSuffix(file, ",")
		if len(prompts) == 1 {
			front := "---\ngoal: ship\nsteps:\n  - step: it builds\n    command: echo checked; exit 3\n---\nnotes\n"
			if err := os.WriteFile(file, []byte(front), 0o644); err != nil {
				return "", err
			}
			return `{"next": "write the code"}`, nil
		}
		return `{"next": ""}`, nil
	}}
	var tasks []Task
	var lapKeys []string
	err := runTest(t, func(ctx context.Context) error {
		planner := NewSession(ctx, "planner", f, "m", t.TempDir())
		loop := Loop(ctx, "sprint", "ship", planner)
		for ctx, task := range loop.Laps {
			tasks = append(tasks, task)
			s, _ := current(ctx)
			lapKeys = append(lapKeys, s.key)
		}
		return loop.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0] != (Task{Lap: 1, Text: "write the code"}) || lapKeys[0] != "sprint.1/lap.1" {
		t.Fatalf("tasks %v in %v", tasks, lapKeys)
	}
	if !strings.Contains(prompts[0], "goal: ship") || !strings.Contains(prompts[1], "$ echo checked; exit 3\nexit 3\nchecked") {
		t.Errorf("planner prompts:\n%s\n---\n%s", prompts[0], prompts[1])
	}
}
