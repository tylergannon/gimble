// Package engine executes Tractor graphs and persists their run state.
package engine

import (
	"maps"
	"sync"
	"time"

	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/harness"
)

// ProofEvidenceRecord ties one proven contract case to this run's executed
// assertion, edge, and fingerprinted evidence artifacts.
type ProofEvidenceRecord struct {
	Status       graph.ProofStatus       `json:"status"`
	RunID        string                  `json:"run_id"`
	CaseID       string                  `json:"case_id"`
	NodeID       string                  `json:"node_id"`
	ExecutionRef string                  `json:"execution_ref"`
	EvidenceMode graph.ProofEvidenceMode `json:"evidence_mode"`
	Edge         ProofEdgeRef            `json:"edge"`
	Artifacts    []ProofArtifactRef      `json:"artifacts"`
}

// ProofEdgeRef records the exact successful routing edge of an assertion.
type ProofEdgeRef struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ProofArtifactRef fingerprints a current-run execution or workspace artifact.
type ProofArtifactRef struct {
	Source           string                  `json:"source"`
	Role             graph.ProofArtifactRole `json:"role"`
	Path             string                  `json:"path"`
	ArchitectureEdge string                  `json:"architecture_edge,omitempty"`
	StoredPath       string                  `json:"stored_path"`
	SHA256           string                  `json:"sha256"`
	CapturedAt       string                  `json:"captured_at"`
}

// Checkpoint is the durable top-level execution state.
type Checkpoint struct {
	Timestamp      time.Time                        `json:"timestamp"`
	CurrentNode    string                           `json:"current_node"`
	NextNode       string                           `json:"next_node"`
	CompletedNodes []string                         `json:"completed_nodes"`
	NodeVisits     map[string]int                   `json:"node_visits"`
	NodeAttempts   map[string]int                   `json:"node_attempts"`
	Seq            uint64                           `json:"seq"`
	RetryVisit     bool                             `json:"retry_visit"`
	LastStage      string                           `json:"last_stage"`
	LastResponse   string                           `json:"last_response"`
	Sessions       map[string]harness.ThreadBinding `json:"sessions"`
	ProofEvidence  map[string]ProofEvidenceRecord   `json:"proof_evidence,omitempty"`
}

type engineState struct {
	mu sync.Mutex

	completedNodes []string
	nodeVisits     map[string]int
	nodeAttempts   map[string]int
	proofEvidence  map[string]ProofEvidenceRecord
	lastStage      string
	lastResponse   string
	seq            uint64
	retryVisit     bool
}

type counterSnapshot struct {
	nodeVisits   map[string]int
	nodeAttempts map[string]int
}

func newEngineState() *engineState {
	return &engineState{
		completedNodes: []string{},
		nodeVisits:     make(map[string]int),
		nodeAttempts:   make(map[string]int),
		proofEvidence:  make(map[string]ProofEvidenceRecord),
	}
}

func stateFromCheckpoint(checkpoint Checkpoint) *engineState {
	return &engineState{
		completedNodes: cloneSlice(checkpoint.CompletedNodes),
		nodeVisits:     cloneMap(checkpoint.NodeVisits),
		nodeAttempts:   cloneMap(checkpoint.NodeAttempts),
		proofEvidence:  cloneProofEvidence(checkpoint.ProofEvidence),
		lastStage:      checkpoint.LastStage,
		lastResponse:   checkpoint.LastResponse,
		seq:            checkpoint.Seq,
		retryVisit:     checkpoint.RetryVisit,
	}
}

func (s *engineState) checkpoint(currentNode, nextNode string, retryVisit bool, sessions map[string]harness.ThreadBinding) Checkpoint {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Checkpoint{
		CurrentNode:    currentNode,
		NextNode:       nextNode,
		CompletedNodes: cloneSlice(s.completedNodes),
		NodeVisits:     cloneMap(s.nodeVisits),
		NodeAttempts:   cloneMap(s.nodeAttempts),
		ProofEvidence:  cloneProofEvidence(s.proofEvidence),
		Seq:            s.seq,
		RetryVisit:     retryVisit,
		LastStage:      s.lastStage,
		LastResponse:   s.lastResponse,
		Sessions:       cloneMap(sessions),
	}
}

func (s *engineState) recordProofEvidence(evidence ProofEvidenceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proofEvidence[evidence.CaseID] = evidence
}

func (s *engineState) proofEvidenceFor(caseID string) (ProofEvidenceRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	evidence, exists := s.proofEvidence[caseID]
	return evidence, exists
}

func (s *engineState) beginVisit(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.retryVisit {
		s.retryVisit = false
		return
	}
	s.nodeVisits[nodeID]++
}

func (s *engineState) beginAttempt(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeAttempts[nodeID]++
}

func (s *engineState) visits(nodeID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nodeVisits[nodeID]
}

func (s *engineState) snapshotCounters() counterSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return counterSnapshot{
		nodeVisits:   cloneMap(s.nodeVisits),
		nodeAttempts: cloneMap(s.nodeAttempts),
	}
}

func (s *engineState) restoreCounters(snapshot counterSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodeVisits = cloneMap(snapshot.nodeVisits)
	s.nodeAttempts = cloneMap(snapshot.nodeAttempts)
}

func (s *engineState) complete(nodeID, notes string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.completedNodes = append(s.completedNodes, nodeID)
	s.lastStage = nodeID
	s.lastResponse = truncate(notes, 200)
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func cloneSlice[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return append([]T(nil), values...)
}

func cloneMap[K comparable, V any](values map[K]V) map[K]V {
	result := make(map[K]V, len(values))
	maps.Copy(result, values)
	return result
}

func cloneProofEvidence(values map[string]ProofEvidenceRecord) map[string]ProofEvidenceRecord {
	result := make(map[string]ProofEvidenceRecord, len(values))
	for caseID, evidence := range values {
		evidence.Artifacts = cloneSlice(evidence.Artifacts)
		result[caseID] = evidence
	}
	return result
}
