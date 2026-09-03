package lint

import (
	"fmt"

	"github.com/tylergannon/tractor/graph"
)

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
	analysis := newAnalysis(g, options{})
	var containing []*loopBlock
	for _, block := range analysis.loopBlocks() {
		if _, inside := block.nodes[nodeID]; inside {
			containing = append(containing, block)
		}
	}
	for _, candidate := range containing {
		nested := false
		for _, other := range containing {
			if other == candidate {
				continue
			}
			if _, inside := other.nodes[candidate.node.ID]; inside {
				nested = true
				break
			}
		}
		if !nested {
			return candidate.node.ID, true
		}
	}
	return "", false
}
