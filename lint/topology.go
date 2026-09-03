package lint

import (
	"fmt"
	"slices"

	"github.com/tylergannon/tractor/graph"
)

// LoopBodyNodes returns the derived body node-set for loopID in graph order.
// The loop node itself is not part of its body node-set.
func LoopBodyNodes(g graph.Graph, loopID string) ([]string, bool) {
	analysis := newAnalysis(g, options{})
	var body map[string]struct{}
	for _, block := range analysis.loopBlocks() {
		if block.node.ID == loopID {
			body = block.nodes
			break
		}
	}
	if body == nil {
		return nil, false
	}
	nodes := make([]string, 0, len(body))
	for _, node := range g.Nodes {
		if _, ok := body[node.Base().ID]; ok {
			nodes = append(nodes, node.Base().ID)
		}
	}
	return nodes, true
}

// EnclosingLoops returns every loop whose body node-set contains nodeID,
// ordered outermost first.
func EnclosingLoops(g graph.Graph, nodeID string) []string {
	analysis := newAnalysis(g, options{})
	var containing []*loopBlock
	for _, block := range analysis.loopBlocks() {
		if _, inside := block.nodes[nodeID]; inside {
			containing = append(containing, block)
		}
	}
	ordered := make([]string, 0, len(containing))
	for len(containing) > 0 {
		index := slices.IndexFunc(containing, func(candidate *loopBlock) bool {
			for _, other := range containing {
				if other == candidate {
					continue
				}
				if _, inside := other.nodes[candidate.node.ID]; inside {
					return false
				}
			}
			return true
		})
		if index < 0 {
			break
		}
		ordered = append(ordered, containing[index].node.ID)
		containing = slices.Delete(containing, index, index+1)
	}
	return ordered
}

// ParallelForFanIn returns the sole converged parallel node that owns fanInID.
func ParallelForFanIn(g graph.Graph, fanInID string) (*graph.ParallelNode, error) {
	analysis := newAnalysis(g, options{})
	var owner *graph.ParallelNode
	count := 0
	for _, block := range analysis.parallelBlocks() {
		if block.converged && block.candidate == fanInID {
			owner = block.node
			count++
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("fan-in %q must have exactly one owning parallel node; found %d", fanInID, count)
	}
	return owner, nil
}

// OutermostLoop returns the outermost loop node whose body node-set contains
// nodeID. A nested loop node lies inside the outer loop's body, so it
// resolves to the outer loop; a top-level loop node is inside no body and
// yields ok=false for itself.
func OutermostLoop(g graph.Graph, nodeID string) (loopID string, ok bool) {
	loops := EnclosingLoops(g, nodeID)
	if len(loops) > 0 {
		return loops[0], true
	}
	return "", false
}
