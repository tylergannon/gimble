// Package contextcheck preserves an unadopted, local-only analyzer experiment.
// It does not traverse the call graph and must not serve as ownership validation.
// Any adopted analyzer must cover calls; this prototype remains for reference.
//
// The experiment catches locally visible context ownership conflicts.
// It follows constant keys and scope relationships within one function, including
// ordinary context wrappers and SSA aliases. It skips unknown calls, dynamic
// keys, merged scope values, and writes in mutually exclusive branches. It does
// not prove ownership across helpers, goroutines, or iterator-created scopes;
// SetContext's runtime check remains authoritative.
package contextcheck

import (
	"go/constant"
	"go/token"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const programPath = "github.com/tylergannon/gimble/examples/go-workflows/internal/program"

var Analyzer = &analysis.Analyzer{
	Name:     "contextcheck",
	Doc:      "experimental local-only diagnostics; not a workflow ownership validator",
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
	Run:      run,
}

type scope struct{ parent *scope }

type write struct {
	key   string
	scope *scope
	block *ssa.BasicBlock
	pos   token.Pos
}

func run(pass *analysis.Pass) (any, error) {
	for _, fn := range pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA).SrcFuncs {
		scopes := make(map[ssa.Value]*scope)
		var writes []write
		for _, block := range fn.Blocks {
			for _, instruction := range block.Instrs {
				call, ok := instruction.(*ssa.Call)
				if !ok || !calls(call, programPath, "SetContext") {
					continue
				}
				key, ok := call.Call.Args[1].(*ssa.Const)
				if !ok || key.Value == nil || key.Value.Kind() != constant.String {
					continue
				}
				owner := resolve(call.Call.Args[0], scopes)
				if owner != nil {
					writes = append(writes, write{constant.StringVal(key.Value), owner, block, call.Pos()})
				}
			}
		}
		slices.SortFunc(writes, func(a, b write) int { return int(a.pos - b.pos) })
		for i, current := range writes {
			for _, earlier := range writes[:i] {
				if current.key != earlier.key || current.scope == earlier.scope {
					continue
				}
				if !ancestor(current.scope, earlier.scope) && !ancestor(earlier.scope, current.scope) {
					continue
				}
				// Dominance conservatively excludes writes on alternative paths.
				if !current.block.Dominates(earlier.block) && !earlier.block.Dominates(current.block) {
					continue
				}
				pass.Reportf(current.pos, "context key %q is also written in an ancestor or descendant scope at %s", current.key, pass.Fset.Position(earlier.pos))
				break
			}
		}
	}
	return nil, nil
}

func ancestor(parent, child *scope) bool {
	for candidate := child.parent; candidate != nil; candidate = candidate.parent {
		if candidate == parent {
			return true
		}
	}
	return false
}

func resolve(value ssa.Value, scopes map[ssa.Value]*scope) *scope {
	if known, ok := scopes[value]; ok {
		return known
	}
	// An unresolved Phi or other value stays unknown; don't infer an owner from
	// a variable name after reassignment or from one of several possible paths.
	var resolved *scope
	switch value := value.(type) {
	case *ssa.Parameter:
		resolved = &scope{}
	case *ssa.ChangeInterface:
		resolved = resolve(value.X, scopes)
	case *ssa.MakeInterface:
		resolved = resolve(value.X, scopes)
	case *ssa.Extract:
		if value.Index == 0 {
			resolved = resolve(value.Tuple, scopes)
		}
	case *ssa.Call:
		switch {
		case calls(value, programPath, "NewContext"):
			resolved = &scope{}
		case calls(value, programPath, "Scope"):
			if parent := resolve(value.Call.Args[0], scopes); parent != nil {
				resolved = &scope{parent: parent}
			}
		case contextWrapper(value):
			resolved = resolve(value.Call.Args[0], scopes)
		}
	}
	scopes[value] = resolved
	return resolved
}

func calls(call *ssa.Call, path, name string) bool {
	callee := call.Call.StaticCallee()
	return callee != nil && callee.Pkg != nil && callee.Pkg.Pkg.Path() == path && callee.Name() == name
}

func contextWrapper(call *ssa.Call) bool {
	for _, name := range []string{"WithCancel", "WithCancelCause", "WithDeadline", "WithDeadlineCause", "WithTimeout", "WithTimeoutCause", "WithValue", "WithoutCancel"} {
		if calls(call, "context", name) {
			return true
		}
	}
	return false
}
