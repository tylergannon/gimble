# Filesystem context and the prompt an agent receives

```sh
go run ./examples/go-workflows/context
```

The workflow declares scope-owned keys, sets a small goal and constraints,
adds oversized research, then
adds individually small acceptance notes that together exceed the context
budget. Its three agent instructions stay short. `Codergen` obtains a ready
context projection before handing each assembled prompt to the canned agent.
There is no explicit wait or prompt assembly in the workflow.

Context JSON files, versioned indices, and symlink views are real and retained
under the directory printed on stderr. Agent responses are canned: this example does
not implement or execute a quote command. `-dir /absolute/path` chooses a parent
directory; `-example` prints the typed JSON input, and `-input file.json` replaces it.

Stdout is one JSON report with the exact prompt delivered at each stage and
the matching context snapshot. The example uses a 240-byte individual value
limit and an 800-byte context projection budget, before the short instruction
is appended. Values remain in immutable JSON files even when projected inline.
An oversized value and aggregate overflow both move material out of the prompt
and into the index. `SetContext` invalidates that projection; the next agent
call synchronously obtains the updated index and complete view before proceeding.

Each snapshot's `view` is an ordinary revision directory. `values/<key>.json`
symlinks expose every effective value, and `index.json` links to the authoritative
`index` file. The context prompt points at `view/index.json`; native tools can
also list and read `view/values/`. No original value bytes are copied and no
FUSE mount is needed. Context is read-only by convention: declare a key with
`DeclareContext`, then use `SetContext` from its owning scope for updates.
Each update writes a new file and produces a new view at the next snapshot.
Writing through a symlink changes its target; it is not automatic copy-on-write.

Read `main.go` for the sequence, `input.go` for the argument shape, and
`fixtures.go` for the example task data. `demo.go` captures prompts at the actual
stub boundary. Generated reports and context files belong outside the repository.
