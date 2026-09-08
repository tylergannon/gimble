package engine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tylergannon/tractor/internal/hostwake"
)

// WakeMode selects when a run wakes the host agent session that launched it.
type WakeMode string

const (
	// WakeUnset is the zero value. An embedder that says nothing about waking
	// gets none of it: no host session is captured and no session is woken.
	// Tractor's own commands pass WakeAuto.
	WakeUnset WakeMode = ""
	// WakeAuto wakes a background host session and leaves an interactive one
	// alone. It is what `tractor run` uses unless told otherwise.
	WakeAuto WakeMode = "auto"
	// WakeOn wakes the launching session whatever its kind, including a
	// session a human is also using.
	WakeOn WakeMode = "on"
	// WakeOff records which session launched the run and never wakes it.
	WakeOff WakeMode = "off"
)

// DefaultWakeInterval is how often an armed run checks whether it has news
// for the session that launched it.
const DefaultWakeInterval = 4 * time.Minute

// MinimumWakeInterval floors the configured interval. The host refills its
// per-sender budget at half a message per second, so this keeps any number of
// concurrent runs far below the ceiling without a rate limiter of our own.
const MinimumWakeInterval = 30 * time.Second

// wakeEventPrefix marks the timeline events this service writes. They are the
// wake cursor and are never themselves news.
const wakeEventPrefix = "HostWake"

// WakeConfig configures host session wakes.
type WakeConfig struct {
	Mode     WakeMode
	Interval time.Duration
	// Channel overrides the channel selected from Session. Nil with no explicit
	// session means Claude Code.
	Channel hostwake.Channel
	// Session is an explicitly identified launching session. Codex desktop
	// parents use this because a shared MCP server cannot identify its caller.
	Session *hostwake.Session
}

// hostChannel is the channel this run captured its host session through.
func (r *Runner) hostChannel() hostwake.Channel {
	if r.config.Wake.Channel != nil {
		return r.config.Wake.Channel
	}
	if r.config.Wake.Session != nil && r.config.Wake.Session.Host == hostwake.HostCodexDesktop {
		return hostwake.NewCodexQueue()
	}
	return hostwake.NewClaudeCode()
}

// captureHostSession records the agent session that launched this run. It
// runs for every wake mode a caller asked for, WakeOff included: which
// session started a run is provenance, and a run that will never wake anybody
// still records it.
func (r *Runner) captureHostSession() (hostwake.Session, bool) {
	if r.config.Wake.Mode == WakeUnset {
		return hostwake.Session{}, false
	}
	if r.config.Wake.Session != nil && !r.config.Wake.Session.Empty() {
		session := *r.config.Wake.Session
		r.hostSession = session
		return session, true
	}
	session, ok := r.hostChannel().Capture()
	if !ok || session.Empty() {
		return hostwake.Session{}, false
	}
	r.hostSession = session
	return session, true
}

// wakeService wakes the launching session on an interval whenever the run has
// news it has not already been told about.
type wakeService struct {
	runner   *Runner
	store    *runStore
	channel  hostwake.Channel
	session  hostwake.Session
	interval time.Duration

	done     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup

	mu     sync.Mutex
	cursor time.Time
}

// newWakeService arms a wake service, or returns nil when this run has no
// host session to wake or was told not to wake it.
func newWakeService(runner *Runner, store *runStore) (*wakeService, error) {
	config := runner.config.Wake
	mode := config.Mode
	switch mode {
	case WakeUnset, WakeOff:
		return nil, nil
	case WakeAuto, WakeOn:
	default:
		return nil, fmt.Errorf("unknown wake mode %q", mode)
	}
	session := runner.hostSession
	if session.Empty() {
		return nil, nil
	}
	interval := config.Interval
	if interval <= 0 {
		interval = DefaultWakeInterval
	}
	if interval < MinimumWakeInterval {
		interval = MinimumWakeInterval
	}
	service := &wakeService{
		runner: runner, store: store, channel: runner.hostChannel(), session: session,
		interval: interval, done: make(chan struct{}),
	}
	// An interactive session belongs to a human who is using it. Waking one
	// is opt-in.
	if mode == WakeAuto && session.Interactive() {
		return nil, nil
	}
	cursor, err := store.lastWakeCursor()
	if err != nil {
		return nil, err
	}
	service.cursor = cursor
	return service, nil
}

func (s *wakeService) start() {
	if s == nil {
		return
	}
	s.wg.Go(func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
				return
			case <-s.runner.stop.done:
				return
			case <-ticker.C:
				s.deliver()
			}
		}
	})
}

// stopAndDrain ends the interval loop and delivers whatever news is left, so
// the run's own ending reaches the session that started it.
func (s *wakeService) stopAndDrain() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.done) })
	s.wg.Wait()
	s.deliver()
}

// deliver sends one wake if the run has news since the last delivered one.
// Nothing here can fail a run: every failure is recorded and swallowed.
func (s *wakeService) deliver() {
	s.mu.Lock()
	defer s.mu.Unlock()

	events, total, err := s.store.timelineSince(s.cursor)
	if err != nil {
		s.record(0, false, err.Error())
		return
	}
	if len(events) == 0 {
		return
	}
	if !s.channel.Available(s.session) {
		s.record(len(events), false, "host session is not reachable")
		return
	}
	message := s.render(events, total)
	if err := s.channel.Wake(s.session, message); err != nil {
		s.record(len(events), false, err.Error())
		return
	}
	s.cursor = time.Now().UTC()
	s.record(len(events), true, "")
}

// render builds the rollup the woken session receives.
//
// This is the interim renderer. The shared one being built for supervisor
// patrols replaces it: the host session is that renderer's second consumer,
// asking for the whole run where a supervisor asks for its own scope. Until
// then this stays deliberately thin, and keeps the two rules that renderer
// owes both consumers: every section is bounded, and a truncated section says
// so and names the artifact holding the rest.
func (s *wakeService) render(events []timelineEvent, total int) string {
	return renderObserverDigest(observerDigestInput{
		name: s.runner.graph.Name, goal: s.runner.graph.Goal, root: s.store.root,
		live: s.runner.liveSnapshot(), events: events, total: total,
	}).Message
}

// record writes the wake cursor event. Its own failures are dropped: a run
// does not fail because it could not tell a session about itself.
func (s *wakeService) record(events int, delivered bool, failure string) {
	event := timelineEvent{
		"type": wakeEventPrefix, "host": s.session.Host, "kind": s.session.Kind,
		"delivered": delivered, "events": events,
		"ts": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if s.session.SessionID != "" {
		event["session_id"] = s.session.SessionID
	}
	if failure != "" {
		event["error"] = failure
	}
	_ = s.store.appendTimeline(event)
}

// timelineSince returns the events written after the cursor, newest last,
// along with how many there were. The service's own wake events are excluded:
// they are the cursor, not news, and including them would make every wake
// cause the next one.
func (s *runStore) timelineSince(cursor time.Time) ([]timelineEvent, int, error) {
	file, err := os.Open(filepath.Join(s.root, "timeline.jsonl"))
	if os.IsNotExist(err) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("open timeline: %w", err)
	}
	defer func() { _ = file.Close() }()
	events := make([]timelineEvent, 0, 16)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event timelineEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if kind, _ := event["type"].(string); strings.HasPrefix(kind, wakeEventPrefix) {
			continue
		}
		if !cursor.IsZero() {
			stamp, ok := eventTime(event)
			if !ok || !stamp.After(cursor) {
				continue
			}
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("read timeline: %w", err)
	}
	return events, len(events), nil
}

// lastWakeCursor returns the timestamp of the last delivered wake, so a
// resumed run does not repeat news the session already has.
func (s *runStore) lastWakeCursor() (time.Time, error) {
	file, err := os.Open(filepath.Join(s.root, "timeline.jsonl"))
	if os.IsNotExist(err) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("open timeline: %w", err)
	}
	defer func() { _ = file.Close() }()
	cursor := time.Time{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var event timelineEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if kind, _ := event["type"].(string); kind != wakeEventPrefix {
			continue
		}
		if delivered, _ := event["delivered"].(bool); !delivered {
			continue
		}
		if stamp, ok := eventTime(event); ok {
			cursor = stamp
		}
	}
	if err := scanner.Err(); err != nil {
		return time.Time{}, fmt.Errorf("read timeline: %w", err)
	}
	return cursor, nil
}

func eventTime(event timelineEvent) (time.Time, bool) {
	raw, ok := event["ts"].(string)
	if !ok {
		return time.Time{}, false
	}
	stamp, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, false
	}
	return stamp, true
}
