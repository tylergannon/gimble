package sessionstate

// Snapshot is a detached projection state at an event boundary.
type Snapshot struct {
	State ProjectionState `json:"state"`
}

func (p *Projection) Snapshot() Snapshot {
	return Snapshot{State: p.state.Clone()}
}
