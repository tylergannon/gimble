package observation

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tylergannon/gimble/internal/sessionstate"
)

// Frame is one SSE frame: the event name and its data. The store produces
// them in the order it reduced them, and every subscriber sees that order.
type Frame struct {
	Name string
	Data json.RawMessage
}

// Frame names. A connection receives one snapshot, then an ordered suffix.
const (
	FrameSnapshot  = "snapshot"
	FrameEvent     = "event"
	FrameLifecycle = "lifecycle"
)

// Lifecycle is one Gimble lifecycle record handed to the store. Record is
// the exact LifecycleRecord JSON the run log holds, republished unchanged.
// The fields beside it are the run/session relationships the caller already
// knows from the record's variant, so this package need not decode a sealed
// union it cannot import.
type Lifecycle struct {
	Placement
	Record  json.RawMessage
	Name    string
	Status  string
	Error   string
	Session *SessionInfo
}

// invocation is one turn's reduction. The projection is mutated in place;
// only a snapshot, a checkpoint or a test observation detaches it.
type invocation struct {
	Placement
	projection *sessionstate.Projection
	provenance map[string]json.RawMessage
}

// Store is one run's observation. Every mutation and every detachment goes
// through mu, so a snapshot and the suffix behind it come from one cut.
type Store struct {
	id       string
	dir      string
	registry *Registry

	mu          sync.Mutex
	run         RunInfo
	invocations map[string]*invocation
	subs        map[*Subscription]struct{}
	closed      bool

	maxFrames int
	maxBytes  int
}

// Open returns the store for one run and registers it, if there is a
// registry. A run started without the web runtime still gets a store: it
// owns it privately and writes its final observation.
func Open(registry *Registry, id, name, dir string) *Store {
	s := &Store{
		id:       id,
		dir:      dir,
		registry: registry,
		run: RunInfo{
			ID:       id,
			Name:     name,
			Status:   StatusRunning,
			Sessions: map[string]SessionInfo{},
		},
		invocations: map[string]*invocation{},
		subs:        map[*Subscription]struct{}{},
		maxFrames:   defaultMaxFrames,
		maxBytes:    defaultMaxBytes,
	}
	if registry != nil {
		registry.add(s)
	}
	return s
}

// ID is the run's id.
func (s *Store) ID() string { return s.id }

// Lifecycle folds one lifecycle record into the run and publishes it.
func (s *Store) Lifecycle(entry Lifecycle) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	if entry.Name != "" {
		s.run.Name = entry.Name
	}
	switch entry.Status {
	case StatusRunning:
		// The run's own start record. Nothing to decide.
	case StatusCancelled:
		s.run.Status, s.run.Error = StatusCancelled, entry.Error
	case StatusCompleted, StatusFailed:
		// A cancelled run that then reports its body's error stays
		// cancelled: cancellation is the terminal fact, and the error it
		// produced is not a second outcome.
		if s.run.Status == StatusRunning {
			s.run.Status, s.run.Error = entry.Status, entry.Error
		}
	}
	if id := entry.Placement.Session; entry.Session != nil && id != "" {
		s.run.Sessions[id] = *entry.Session
	}
	if len(entry.Record) > 0 {
		s.publishLocked(Frame{Name: FrameLifecycle, Data: entry.Record})
	}
	s.mu.Unlock()
}

// eventFrame is the `event` frame's body.
type eventFrame struct {
	Scope     string          `json:"scope"`
	Session   string          `json:"session"`
	Turn      string          `json:"turn"`
	Event     json.RawMessage `json:"event"`
	NativeRef json.RawMessage `json:"nativeRef,omitempty"`
}

// Event applies one stamped native event to its invocation's projection,
// then publishes that same event to subscribers.
//
// envelope is the native event exactly as the runtime recorded it.
// nativeRef is the placement sidecar and is never inserted into the event.
//
// The returned error is malformed input, and the caller propagates it: an
// invalidly observed event must not silently continue a provider turn.
func (s *Store) Event(at Placement, envelope, nativeRef json.RawMessage) error {
	if s == nil {
		return nil
	}
	decoded, err := sessionstate.DecodeValue(envelope)
	if err != nil {
		return fmt.Errorf("observation: decode event in turn %s: %w", at.Turn, err)
	}
	event, ok := decoded.(*sessionstate.Obj)
	if !ok {
		return fmt.Errorf("observation: event in turn %s is not a JSON object", at.Turn)
	}
	ref, err := decodeRef(nativeRef)
	if err != nil {
		return fmt.Errorf("observation: decode native ref in turn %s: %w", at.Turn, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("observation: run %s is closed", s.id)
	}
	inv := s.invocationLocked(at)
	inv.projection.Apply(event)
	s.foldProvenanceLocked(inv, ref, nativeRef)

	s.publishLocked(Frame{Name: FrameEvent, Data: mustMarshal(eventFrame{
		Scope: at.Scope, Session: at.Session, Turn: at.Turn,
		Event: envelope, NativeRef: nativeRef,
	})})
	return nil
}

func (s *Store) invocationLocked(at Placement) *invocation {
	if inv, ok := s.invocations[at.Turn]; ok {
		return inv
	}
	inv := &invocation{
		Placement:  at,
		projection: sessionstate.New(sessionstate.NewProjectionState()),
		provenance: map[string]json.RawMessage{},
	}
	s.invocations[at.Turn] = inv
	return inv
}

// Snapshot detaches the complete public observation.
func (s *Store) Snapshot() RunSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshotLocked()
}

func (s *Store) snapshotLocked() RunSnapshot {
	out := RunSnapshot{
		Run:         RunInfo{ID: s.run.ID, Name: s.run.Name, Status: s.run.Status, Error: s.run.Error, Sessions: make(map[string]SessionInfo, len(s.run.Sessions))},
		Invocations: make(map[string]Invocation, len(s.invocations)),
	}
	for id, info := range s.run.Sessions {
		out.Run.Sessions[id] = info
	}
	for turn, inv := range s.invocations {
		out.Invocations[turn] = Invocation{
			Scope:      inv.Scope,
			Session:    inv.Session,
			Turn:       inv.Turn,
			Snapshot:   inv.projection.Snapshot(),
			Provenance: cloneProvenance(inv.provenance),
		}
	}
	return out
}

func cloneProvenance(in map[string]json.RawMessage) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(in))
	for key, value := range in {
		copied := make(json.RawMessage, len(value))
		copy(copied, value)
		out[key] = copied
	}
	return out
}

func mustMarshal(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Errorf("observation: encode frame: %w", err))
	}
	return raw
}
