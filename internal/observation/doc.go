// Package observation holds one run's live observation: the ordered
// reduction of its native session events into per-invocation projections,
// the snapshots detached from that state, and the bounded
// subscribers the web server streams to.
//
// It imports neither the root workflow package nor web. The root package
// publishes into a store; web reads and subscribes to one. A store is found
// through a Registry the web runtime puts in its context, so there is no
// global map of runs and no new workflow-facing name.
//
// The store is the producer. It never blocks on a subscriber and never reads
// a log to answer a live connection. A completed run is served from its one
// final snapshot file.
package observation
