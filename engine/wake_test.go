package engine

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
	"github.com/tylergannon/tractor/internal/hostwake"
)

// fakeChannel is a host that records what it was told.
type fakeChannel struct {
	session   hostwake.Session
	captured  bool
	reachable bool
	failWith  error

	mu       sync.Mutex
	messages []string
	captures int
}

func (c *fakeChannel) Capture() (hostwake.Session, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.captures++
	return c.session, c.captured
}

func (c *fakeChannel) Available(hostwake.Session) bool { return c.reachable }

func (c *fakeChannel) Wake(_ hostwake.Session, message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failWith != nil {
		return c.failWith
	}
	c.messages = append(c.messages, message)
	return nil
}

func (c *fakeChannel) delivered() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.messages...)
}

func backgroundChannel() *fakeChannel {
	return &fakeChannel{
		session: hostwake.Session{
			Host: hostwake.HostClaudeCode, SessionID: "session-1",
			Kind: hostwake.KindBackground, Socket: "/tmp/host.sock", PID: 4242,
		},
		captured: true, reachable: true,
	}
}

func runWithWake(t *testing.T, root string, channel hostwake.Channel, config WakeConfig) RunResult {
	t.Helper()
	pipeline := testGraph(
		startNode("start", "work"),
		customNode("work", "agent", []graph.Edge{{To: "done"}}, 0),
		exitNode("done"),
	)
	registry := NewRegistry()
	registry.Register("agent", HandlerFunc(func(graph.Node, []graph.Edge, ExecutionScope, *graph.Graph) (harness.Outcome, *harness.Error) {
		return harness.Outcome{Notes: "did the work"}, nil
	}))
	config.Channel = channel
	runner, err := NewRunner(pipeline, registry, RunnerConfig{
		LogsRoot: root, Workdir: "/workspace",
		Validate: func(graph.Graph) error { return nil },
		Wake:     config,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustManifest(t *testing.T, root string) runManifest {
	t.Helper()
	var manifest runManifest
	if err := json.Unmarshal([]byte(mustReadFile(t, filepath.Join(root, "manifest.json"))), &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func wakeEvents(t *testing.T, root string) []timelineEvent {
	t.Helper()
	events, err := readTimeline(filepath.Join(root, "timeline.jsonl"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	wakes := []timelineEvent{}
	for _, event := range events {
		if event["type"] == "HostWake" {
			wakes = append(wakes, event)
		}
	}
	return wakes
}

func TestRunWakesBackgroundHostSessionWithItsNews(t *testing.T) {
	root := t.TempDir()
	channel := backgroundChannel()

	if result := runWithWake(t, root, channel, WakeConfig{Mode: WakeAuto}); result.Status != RunCompleted {
		t.Fatalf("result = %#v", result)
	}

	delivered := channel.delivered()
	if len(delivered) != 1 {
		t.Fatalf("delivered %d wakes: %v", len(delivered), delivered)
	}
	message := delivered[0]
	for _, want := range []string{"PipelineStarted", "PipelineCompleted", "StageCompleted", root} {
		if !strings.Contains(message, want) {
			t.Fatalf("wake message lacks %q:\n%s", want, message)
		}
	}
	wakes := wakeEvents(t, root)
	if len(wakes) != 1 || wakes[0]["delivered"] != true || wakes[0]["session_id"] != "session-1" {
		t.Fatalf("wake timeline events = %#v", wakes)
	}
}

func TestRunRecordsHostSessionInTheManifestWithoutACredential(t *testing.T) {
	root := t.TempDir()
	channel := backgroundChannel()
	runWithWake(t, root, channel, WakeConfig{Mode: WakeOff})

	manifest := mustManifest(t, root)
	if manifest.HostSession == nil {
		t.Fatal("manifest records no host session")
	}
	if manifest.HostSession.SessionID != "session-1" || manifest.HostSession.PID != 4242 {
		t.Fatalf("host session = %#v", *manifest.HostSession)
	}
	raw := mustReadFile(t, filepath.Join(root, "manifest.json"))
	if strings.Contains(raw, "token") {
		t.Fatalf("manifest carries a credential:\n%s", raw)
	}
	if delivered := channel.delivered(); len(delivered) != 0 {
		t.Fatalf("--wake=off woke the session: %v", delivered)
	}
}

func TestExplicitCodexParentIsRecordedWithoutProbingSharedMCPEnvironment(t *testing.T) {
	root := t.TempDir()
	channel := backgroundChannel()
	explicit := hostwake.Session{
		Host: hostwake.HostCodexDesktop, HostID: "local",
		ThreadID: "01a07eab-99d5-7923-84ac-f41374506751",
		Kind:     hostwake.KindInteractive,
	}
	runWithWake(t, root, channel, WakeConfig{Mode: WakeOff, Session: &explicit})

	manifest := mustManifest(t, root)
	if manifest.HostSession == nil || manifest.HostSession.Host != hostwake.HostCodexDesktop || manifest.HostSession.ThreadID != explicit.ThreadID {
		t.Fatalf("manifest host session = %#v", manifest.HostSession)
	}
	channel.mu.Lock()
	captures := channel.captures
	channel.mu.Unlock()
	if captures != 0 {
		t.Fatalf("explicit parent caused %d ambient captures", captures)
	}
}

func TestUnsetWakeModeNeverTouchesTheHost(t *testing.T) {
	root := t.TempDir()
	channel := backgroundChannel()
	runWithWake(t, root, channel, WakeConfig{})

	channel.mu.Lock()
	captures := channel.captures
	channel.mu.Unlock()
	if captures != 0 {
		t.Fatalf("captured %d times with no wake mode set", captures)
	}
	if manifest := mustManifest(t, root); manifest.HostSession != nil {
		t.Fatalf("manifest records %#v with no wake mode set", *manifest.HostSession)
	}
}

func TestAutoLeavesAnInteractiveSessionAloneAndOnWakesIt(t *testing.T) {
	channel := backgroundChannel()
	channel.session.Kind = hostwake.KindInteractive

	auto := t.TempDir()
	runWithWake(t, auto, channel, WakeConfig{Mode: WakeAuto})
	if delivered := channel.delivered(); len(delivered) != 0 {
		t.Fatalf("auto woke an interactive session: %v", delivered)
	}
	if wakes := wakeEvents(t, auto); len(wakes) != 0 {
		t.Fatalf("auto recorded wake events for an interactive session: %#v", wakes)
	}

	on := t.TempDir()
	runWithWake(t, on, channel, WakeConfig{Mode: WakeOn})
	if delivered := channel.delivered(); len(delivered) != 1 {
		t.Fatalf("--wake=on delivered %d wakes", len(delivered))
	}
}

func TestAnUnreachableOrFailingHostNeverFailsTheRun(t *testing.T) {
	gone := backgroundChannel()
	gone.reachable = false
	root := t.TempDir()
	if result := runWithWake(t, root, gone, WakeConfig{Mode: WakeAuto}); result.Status != RunCompleted {
		t.Fatalf("an unreachable host failed the run: %#v", result)
	}
	wakes := wakeEvents(t, root)
	if len(wakes) != 1 || wakes[0]["delivered"] != false || !strings.Contains(wakes[0]["error"].(string), "reachable") {
		t.Fatalf("wake timeline events = %#v", wakes)
	}

	broken := backgroundChannel()
	broken.failWith = errors.New("socket closed")
	brokenRoot := t.TempDir()
	if result := runWithWake(t, brokenRoot, broken, WakeConfig{Mode: WakeAuto}); result.Status != RunCompleted {
		t.Fatalf("a failing wake failed the run: %#v", result)
	}
	wakes = wakeEvents(t, brokenRoot)
	if len(wakes) != 1 || wakes[0]["delivered"] != false || wakes[0]["error"] != "socket closed" {
		t.Fatalf("wake timeline events = %#v", wakes)
	}
}

func TestUnknownWakeModeIsRejected(t *testing.T) {
	root := t.TempDir()
	pipeline := testGraph(startNode("start", "done"), exitNode("done"))
	runner, err := NewRunner(pipeline, NewRegistry(), RunnerConfig{
		LogsRoot: root, Workdir: "/workspace",
		Validate: func(graph.Graph) error { return nil },
		Wake:     WakeConfig{Mode: WakeMode("sometimes"), Channel: backgroundChannel()},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.Run(); err == nil || !strings.Contains(err.Error(), "unknown wake mode") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestTheCursorIsTheLastDeliveredWake(t *testing.T) {
	root := t.TempDir()
	store, err := openRunStore(root, newEngineState())
	if err != nil {
		t.Fatal(err)
	}
	appendAt := func(event timelineEvent, at time.Time) {
		event["ts"] = at.Format(time.RFC3339Nano)
		if err := store.appendTimeline(event); err != nil {
			t.Fatal(err)
		}
	}
	base := time.Now().UTC().Add(-time.Hour)
	appendAt(timelineEvent{"type": "PipelineStarted"}, base)
	appendAt(timelineEvent{"type": "HostWake", "delivered": true}, base.Add(time.Minute))
	appendAt(timelineEvent{"type": "StageCompleted"}, base.Add(2*time.Minute))
	appendAt(timelineEvent{"type": "HostWake", "delivered": false, "error": "gone"}, base.Add(3*time.Minute))

	cursor, err := store.lastWakeCursor()
	if err != nil {
		t.Fatal(err)
	}
	if !cursor.Equal(base.Add(time.Minute)) {
		t.Fatalf("cursor = %s, want the last delivered wake at %s", cursor, base.Add(time.Minute))
	}
	events, total, err := store.timelineSince(cursor)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(events) != 1 || events[0]["type"] != "StageCompleted" {
		t.Fatalf("news since the cursor = %#v", events)
	}
}

func TestARunWithNoNewsSendsNothing(t *testing.T) {
	root := t.TempDir()
	store, err := openRunStore(root, newEngineState())
	if err != nil {
		t.Fatal(err)
	}
	channel := backgroundChannel()
	service := &wakeService{
		runner:  newTestRunner(t, testGraph(startNode("start", "done"), exitNode("done")), NewRegistry(), root, nil),
		store:   store,
		channel: channel,
		session: channel.session,
		done:    make(chan struct{}),
	}
	service.deliver()
	if delivered := channel.delivered(); len(delivered) != 0 {
		t.Fatalf("an empty timeline produced a wake: %v", delivered)
	}
	if wakes := wakeEvents(t, root); len(wakes) != 0 {
		t.Fatalf("an empty timeline produced wake events: %#v", wakes)
	}
}

func TestRenderBoundsAndLabelsTruncation(t *testing.T) {
	root := t.TempDir()
	store, err := openRunStore(root, newEngineState())
	if err != nil {
		t.Fatal(err)
	}
	channel := backgroundChannel()
	service := &wakeService{
		runner: newTestRunner(t, testGraph(startNode("start", "done"), exitNode("done")), NewRegistry(), root, nil),
		store:  store, channel: channel, session: channel.session, done: make(chan struct{}),
	}
	events := make([]timelineEvent, 0, maxDigestEvents*3)
	for index := range cap(events) {
		events = append(events, timelineEvent{
			"type": "StageCompleted", "node_id": "work", "attempt": index,
			"ts": time.Now().UTC().Format(time.RFC3339Nano),
		})
	}

	message := service.render(events, len(events))
	if len(message) > maxDigestBytes+1024 {
		t.Fatalf("message is %d bytes", len(message))
	}
	if !strings.Contains(message, "truncated") || !strings.Contains(message, filepath.Join(root, "timeline.jsonl")) {
		t.Fatalf("a truncated message does not say so and name the artifact:\n%s", message)
	}
	if !strings.Contains(message, "120 total") {
		t.Fatalf("message does not report the true event count:\n%s", message)
	}
}
