package gimble

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tylergannon/gimble/internal/runlog"
)

// fake is a HarnessAdapter whose turns are answered by a function.
type fake struct {
	answer func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error)

	mu      sync.Mutex
	made    int
	steers  []string
	running map[string]func(AgentEvent)
}

func (f *fake) CreateSession(ctx context.Context, model, workdir string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.made++
	return "native-" + string(rune('0'+f.made)), nil
}

func (f *fake) RunTurn(ctx context.Context, session, prompt string, schema json.RawMessage, onEvent func(AgentEvent)) (json.RawMessage, error) {
	f.mu.Lock()
	if f.running == nil {
		f.running = map[string]func(AgentEvent){}
	}
	f.running[session] = onEvent
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		delete(f.running, session)
		f.mu.Unlock()
	}()
	onEvent(UserMessage{Text: prompt})
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
		emit(UserMessage{Text: message})
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

func lifecycleKind(event LifecycleEvent) string {
	switch event.(type) {
	case RunStarted:
		return "run_started"
	case RunEnded:
		return "run_ended"
	case RunCancelled:
		return "run_cancelled"
	case ScopeBegan:
		return "scope_began"
	case ScopeEnded:
		return "scope_ended"
	case LoopCommand:
		return "loop_command"
	case PlannerDecision:
		return "planner_decision"
	case ValueSet:
		return "value_set"
	case SessionCreated:
		return "session_created"
	case SessionClosed:
		return "session_closed"
	case TurnStarted:
		return "turn_started"
	case TurnEnded:
		return "turn_ended"
	case SuperviseAttached:
		return "supervise_attached"
	case Steer:
		return "steer"
	case Interrupt:
		return "interrupt"
	case Complete:
		return "complete"
	default:
		return "unknown"
	}
}

func agentKind(event AgentEvent) string {
	switch event.(type) {
	case UserMessage:
		return "user_message"
	case AssistantMessage:
		return "assistant_message"
	case AssistantMessageDelta:
		return "assistant_message_delta"
	case Thinking:
		return "thinking"
	case ThinkingDelta:
		return "thinking_delta"
	case ToolCall:
		return "tool_call"
	case ToolInputDelta:
		return "tool_input_delta"
	case ToolResult:
		return "tool_result"
	case ToolResultDelta:
		return "tool_result_delta"
	case Usage:
		return "usage"
	case HarnessError:
		return "harness_error"
	case Retry:
		return "retry"
	case ApprovalRequest:
		return "approval_request"
	case NestedTranscript:
		return "nested_transcript"
	default:
		return "unknown"
	}
}

func TestRunLogCanBeRead(t *testing.T) {
	project := t.TempDir()
	var dir string
	if err := Run(Project(t.Context(), project), "reader", func(ctx context.Context) error {
		dir = runDir(ctx)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		t.Fatal("run directory was empty inside a run")
	}
	if runDir(t.Context()) != "" {
		t.Fatal("run directory was present outside a run")
	}
	var got []string
	if err := runlog.Read[LifecycleRecord](t.Context(), dir, func(e LifecycleRecord) error {
		got = append(got, lifecycleKind(e.Event))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[len(got)-1] != "complete" {
		t.Fatalf("read events = %v", got)
	}

	liveDir := t.TempDir()
	w, err := newEventWriter(filepath.Join(liveDir, "run.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer w.close()
	seen := make(chan string, 2)
	readErr := make(chan error, 1)
	go func() {
		readErr <- runlog.Read[LifecycleRecord](t.Context(), liveDir, func(e LifecycleRecord) error {
			seen <- lifecycleKind(e.Event)
			return nil
		})
	}()
	if err := w.writeLifecycle("", "", "", RunStarted{}); err != nil {
		t.Fatal(err)
	}
	if err := w.writeLifecycle("", "", "", Complete{}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-readErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("live reader did not stop at complete")
	}
	if first, second := <-seen, <-seen; first != "run_started" || second != "complete" {
		t.Fatalf("live events = %q, %q", first, second)
	}
}

func TestScopeData(t *testing.T) {
	err := runTest(t, func(ctx context.Context) error {
		if err := Set(ctx, "language", "go"); err != nil {
			return err
		}
		if err := Set(ctx, "language", "rust"); err == nil {
			t.Error("a second Set of a key in one scope succeeded")
		}
		if err := SetJSON(ctx, "review", review{Objections: []string{"too big"}}); err != nil {
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
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
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
		gotReview, err := s.Generate[review](ctx, "review")
		if err != nil || len(gotReview.Objections) != 1 {
			t.Errorf("Generate[review] = %v, %v", gotReview, err)
		}
		if _, err := s.Generate[review](ctx, "bad"); err == nil {
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
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
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
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
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
	f.answer = func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
		if session == "native-1" { // the worker, which runs until the supervisor's steer lands
			workerTurns++
			emit(ToolCall{CallID: "1", Tool: "shell", Input: JSONText(`{"command":"make plugins"}`)})
			emit(ToolResult{CallID: "1", Output: JSONText(`"built a plugin system"`)})
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
	f.answer = func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
		switch {
		case session == "native-1": // the worker, until its supervisor objects
			emit(ToolResult{CallID: "1", Output: JSONText(`"built a plugin system"`)})
			steered("remove the plugin system")
			return "done", nil
		case session == "native-2" && strings.Contains(prompt, "no plugin systems"): // the supervisor's first look, until its own supervisor objects
			emit(AssistantMessage{ID: "message-1", Text: "I object to the variable names"})
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
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
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

func TestLoopShowsThePlannerABadBacklog(t *testing.T) {
	var prompts []string
	f := &fake{answer: func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
		prompts = append(prompts, prompt)
		file := strings.Fields(prompt[strings.Index(prompt, "Its backlog is the file ")+len("Its backlog is the file "):])[0]
		file = strings.TrimSuffix(file, ",")
		front := "---\ngoal: ship\nsteps:\n  - step: build: the code\n    command: echo checked\n---\n"
		if len(prompts) == 2 {
			front = "---\ngoal: ship\nsteps:\n  - step: \"build: the code\"\n    command: echo checked\n---\n"
		}
		if err := os.WriteFile(file, []byte(front), 0o644); err != nil {
			return "", err
		}
		if len(prompts) == 3 {
			return `{"next": ""}`, nil
		}
		return `{"next": "write the code"}`, nil
	}}
	laps := 0
	err := runTest(t, func(ctx context.Context) error {
		planner := NewSession(ctx, "planner", f, "m", t.TempDir())
		loop := Loop(ctx, "sprint", "ship", planner)
		for range loop.Laps {
			laps++
		}
		return loop.Err()
	})
	if err != nil {
		t.Fatalf("a bad backlog ended the loop: %v", err)
	}
	if laps != 2 || !strings.Contains(prompts[1], "does not parse") || strings.Contains(prompts[1], "$ echo checked") || !strings.Contains(prompts[2], "$ echo checked\nexit 0") {
		t.Errorf("%d laps; planner prompts:\n%s\n---\n%s", laps, prompts[1], prompts[2])
	}
}

func TestAttestEventFixture(t *testing.T) {
	project := t.TempDir()
	f := &fake{}
	reviewerStarted := make(chan struct{})
	var reviewerStartedOnce sync.Once
	f.answer = func(ctx context.Context, session, prompt string, schema json.RawMessage, emit func(AgentEvent)) (string, error) {
		emit(AssistantMessage{ID: "message-1", Text: "working"})
		if prompt == "build" {
			emit(ToolCall{CallID: "call-1", Tool: "shell", Input: JSONText(`{"command":"test"}`)})
			emit(ToolResult{CallID: "call-1", Output: JSONText(`"ok"`)})
			for {
				f.mu.Lock()
				steered := len(f.steers) > 0
				f.mu.Unlock()
				if steered || ctx.Err() != nil {
					break
				}
				time.Sleep(time.Millisecond)
			}
		}
		if strings.Contains(prompt, "You are supervising another agent") {
			reviewerStartedOnce.Do(func() { close(reviewerStarted) })
			return `{"objections": []}`, nil
		}
		return "done", nil
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- Run(Project(t.Context(), project), "attest", func(ctx context.Context) error {
			if runDir(ctx) == "" {
				return errors.New("run directory was empty inside a run")
			}
			if err := Set(ctx, "root", "value"); err != nil {
				return err
			}
			researcher := NewSession(ctx, "researcher", f, "model", project)
			if _, err := researcher.Generate[Text](ctx, "prime"); err != nil {
				return err
			}
			fork, err := researcher.Fork(ctx, "planner")
			if err != nil {
				return err
			}
			worker := NewSession(ctx, "worker", f, "model", project)
			reviewer := NewSession(ctx, "reviewer", f, "review", project)
			turnDone := make(chan error, 1)
			go func() {
				_, err := worker.Generate[Text](ctx, "build", WithSupervisor(reviewer, "watch", WithInterval(time.Millisecond)))
				turnDone <- err
			}()
			select {
			case <-reviewerStarted:
			case <-time.After(time.Second):
				return errors.New("reviewer turn did not start")
			}
			if err := worker.Steer(withSteerSource(ctx, reviewer.id), "continue"); err != nil {
				return err
			}
			if err := <-turnDone; err != nil {
				return err
			}
			if err := worker.Steer(ctx, "too late"); err != nil {
				return err
			}
			if _, err := fork.Generate[Text](ctx, "plan"); err != nil {
				return err
			}
			return Scope(ctx, "nested", func(ctx context.Context) error {
				return Set(ctx, "child", "value")
			})
		})
	}()

	runs := filepath.Join(project, "runs")
	var runDir string
	for range 1000 {
		entries, _ := os.ReadDir(runs)
		if len(entries) == 1 {
			runDir = filepath.Join(runs, entries[0].Name())
			break
		}
		time.Sleep(time.Millisecond)
	}
	if runDir == "" {
		t.Fatal("run directory was not created")
	}
	var tailed []LifecycleRecord
	readDone := make(chan error, 1)
	go func() {
		readDone <- runlog.Read[LifecycleRecord](t.Context(), runDir, func(e LifecycleRecord) error {
			tailed = append(tailed, e)
			return nil
		})
	}()
	if err := <-runDone; err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-readDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("public reader did not stop at the final event")
	}

	if len(tailed) < 2 || lifecycleKind(tailed[0].Event) != "run_started" || lifecycleKind(tailed[len(tailed)-1].Event) != "complete" {
		t.Fatalf("tail did not replay through completion: %v", tailed)
	}
	seq := uint64(0)
	seen := map[string]bool{}
	intervals := map[string][2]time.Time{}
	scopes := map[string]bool{}
	var sessionIDs, turnIDs []string
	for _, e := range tailed {
		if e.Seq <= seq || e.Time.IsZero() {
			t.Fatalf("invalid event ordering: %+v", e)
		}
		seq = e.Seq
		kind := lifecycleKind(e.Event)
		seen[kind] = true
		switch e.Event.(type) {
		case ScopeBegan:
			scopes[e.Scope] = true
			intervals[e.Scope] = [2]time.Time{e.Time}
		case ScopeEnded:
			pair := intervals[e.Scope]
			pair[1] = e.Time
			intervals[e.Scope] = pair
		case SessionCreated:
			sessionIDs = append(sessionIDs, e.Session.Value)
		case TurnStarted:
			turnIDs = append(turnIDs, e.Turn.Value)
		}
	}
	for _, e := range tailed {
		if e.Scope == "" {
			continue
		}
		contained := false
		for scope := range scopes {
			if e.Scope == scope || strings.HasPrefix(e.Scope, scope+"/") {
				contained = true
				break
			}
		}
		if !contained {
			t.Fatalf("event %q has scope %q outside the began-scope prefix tree", lifecycleKind(e.Event), e.Scope)
		}
	}
	for scope, pair := range intervals {
		if pair[1].IsZero() || !pair[0].Before(pair[1]) {
			t.Fatalf("scope %q lacks a valid began/ended interval: %v", scope, pair)
		}
	}
	if len(sessionIDs) < 4 || !slices.Contains(sessionIDs, "researcher.1") || !slices.Contains(sessionIDs, "planner.1") || !slices.Contains(sessionIDs, "worker.1") || !slices.Contains(sessionIDs, "reviewer.1") {
		t.Fatalf("ordinal session ids = %v", sessionIDs)
	}
	for _, id := range turnIDs {
		if !strings.Contains(id, "/turn.") {
			t.Fatalf("turn id lacks ordinal = %q", id)
		}
	}
	for _, kind := range []string{"scope_began", "scope_ended", "value_set", "session_created", "session_closed", "turn_started", "turn_ended", "supervise_attached", "steer"} {
		if !seen[kind] {
			t.Errorf("run log lacks %s", kind)
		}
	}
	if !seen["complete"] {
		t.Error("run log lacks completion event")
	}
	if got := tailed[0].Scope; got != "" {
		t.Errorf("run_started scope = %q, want root scope", got)
	}
	for _, want := range []string{"nested.1", ""} {
		if !slices.ContainsFunc(tailed, func(e LifecycleRecord) bool {
			_, began := e.Event.(ScopeBegan)
			return began && e.Scope == want
		}) {
			t.Errorf("scope tree lacks began event for %q", want)
		}
	}
	for _, e := range tailed {
		if value, ok := e.Event.(ValueSet); ok {
			if value.Key == "root" && value.Value != JSONText(`"value"`) {
				t.Errorf("root set value = %s, want JSON string", value.Value)
			}
			if value.Key == "child" && e.Scope != "nested.1" {
				t.Errorf("child set scope = %q, want nested.1", e.Scope)
			}
		}
	}
	var forked, supervised, landed, dropped bool
	var workerTurn, reviewerTurn [2]time.Time
	for _, e := range tailed {
		switch event := e.Event.(type) {
		case SessionCreated:
			if e.Session.Value == "planner.1" && event.Parent == "researcher.1" {
				forked = true
			}
		case SuperviseAttached:
			if event.Reviewer == "reviewer.1" && strings.HasPrefix(event.Worker, "worker.1/turn.") {
				supervised = true
			}
		case Steer:
			if event.Source == "reviewer.1" && event.Landed {
				landed = true
			}
			if event.Message == "too late" && !event.Landed {
				dropped = true
			}
		case TurnStarted:
			if strings.HasPrefix(e.Turn.Value, "worker.1/turn.") {
				workerTurn[0] = e.Time
			}
			if e.Session.Value == "reviewer.1" && reviewerTurn[0].IsZero() {
				reviewerTurn[0] = e.Time
			}
		case TurnEnded:
			if strings.HasPrefix(e.Turn.Value, "worker.1/turn.") {
				workerTurn[1] = e.Time
			}
			if e.Session.Value == "reviewer.1" && reviewerTurn[1].IsZero() {
				reviewerTurn[1] = e.Time
			}
		}
	}
	if !forked || !supervised || !landed || !dropped {
		t.Fatalf("relations: fork=%v supervise=%v landed=%v dropped=%v", forked, supervised, landed, dropped)
	}
	if workerTurn[0].IsZero() || workerTurn[1].IsZero() || reviewerTurn[0].IsZero() || !workerTurn[0].Before(reviewerTurn[0]) || !reviewerTurn[0].Before(workerTurn[1]) {
		t.Fatalf("worker/reviewer turns did not overlap: worker=%v reviewer=%v", workerTurn, reviewerTurn)
	}
	persisted := readRecords[LifecycleRecord](t, filepath.Join(runDir, "run.jsonl"))
	if len(persisted) != len(tailed) {
		t.Fatalf("persisted lifecycle records = %d, tailed %d", len(persisted), len(tailed))
	}

	projectEvents := readRecords[LifecycleRecord](t, filepath.Join(project, "project.jsonl"))
	if len(projectEvents) != 2 || lifecycleKind(projectEvents[0].Event) != "run_started" || lifecycleKind(projectEvents[1].Event) != "run_ended" {
		t.Fatalf("project lifecycle: %+v", projectEvents)
	}
	files, _ := filepath.Glob(filepath.Join(runDir, "sessions", "*.jsonl"))
	if len(files) < 3 {
		t.Fatalf("session transcripts = %d, want at least 3", len(files))
	}
	transcriptKinds := map[string]map[string]bool{}
	for _, file := range files {
		events := readRecords[AgentRecord](t, file)
		if len(events) == 0 || agentKind(events[0].Event) != "user_message" {
			t.Errorf("%s is not a transcript: %+v", file, events)
		}
		seenKinds := map[string]bool{}
		for _, e := range events {
			if e.Seq == 0 || e.Time.IsZero() || e.Session == "" || e.Turn == "" {
				t.Errorf("%s has incomplete agent event placement: %+v", file, e)
			}
			seenKinds[agentKind(e.Event)] = true
		}
		transcriptKinds[filepath.Base(file)] = seenKinds
	}
	if !slices.ContainsFunc(files, func(file string) bool {
		kinds := transcriptKinds[filepath.Base(file)]
		return kinds["tool_call"] && kinds["tool_result"]
	}) {
		t.Error("no session transcript contains the tool call/result pair")
	}
}

func readRecords[T any](t *testing.T, file string) []T {
	t.Helper()
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var records []T
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		var e T
		if validator, ok := any(e).(interface{ ValidateJSON([]byte) error }); ok {
			if err := validator.ValidateJSON([]byte(line)); err != nil {
				t.Fatalf("validate %s: %v\n%s", file, err, line)
			}
		}
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("decode %s: %v", file, err)
		}
		records = append(records, e)
	}
	return records
}
