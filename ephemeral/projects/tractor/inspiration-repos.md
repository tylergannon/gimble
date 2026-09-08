# Reference repositories to clone into inspiration

`inspiration/` is an untracked directory used for shallow clones of outside
projects worth reading. It is excluded through `.git/info/exclude`, so it stays
untracked in every worktree and can be deleted and rebuilt at any time. This
file is the list it should be rebuilt from.

Clone shallow, since none of these are read for their history:

```
git clone --depth 1 <url> inspiration/<dir>
```

## Orchestration engines and languages

Read for engine design: how a workflow is described, what is checked before a
run starts, how loops are bounded, how routing decisions are made and recorded,
and how a run survives interruption. Findings from the first pass over these
four are in `upstream-orchestrators/findings.md`.

| Directory | Source | Read it for |
| --- | --- | --- |
| `virtuslab-orca` | https://github.com/VirtusLab/orca | The closest competitor. Commit-anchored resumable stages, roles as an indirection over backends, capability as a type, and a long series of architecture decision records. The strongest argument on the other side of the declarative question. |
| `orca-lang` | https://github.com/jascal/orca-lang | Static verification. Declared properties checked by bounded search, deadlock and reachability errors, exhaustive event handling with an explicit ignore, and a guard language deliberately restricted so the verifier can be complete. |
| `orca-dsl` | https://github.com/ThakeeNathees/orca | Language surface. Schemas expressed in the language itself driving diagnostics, editor hover, and documentation from one source; and a workflow block that separates a graph position from the definition filling it. Note the project is stalled partway through a rewrite. |
| `stably-orca` | https://github.com/stablyai/orca | Interface reference only, not engine reference. Worktree management, cross-worktree status views, notification routing, a phone client acting as a remote control, and line-anchored diff comments batched into one revision prompt. |

## Agent tooling and skills

Read for skill and plugin authoring rather than engine design. These were cloned
in an earlier pass. The agents repository carries a pre-digested map of what to
borrow and what to reject from the rest.

| Directory | Source |
| --- | --- |
| `agents` | https://github.com/tylergannon/agents |
| `skills` | https://github.com/diffusioninc/skills |
| `mattpocock-skills` | https://github.com/mattpocock/skills |
| `superpowers` | https://github.com/obra/superpowers |
| `caveman` | https://github.com/JuliusBrussee/caveman |
| `gstack` | https://github.com/garrytan/gstack |
| `gbrain-evals` | https://github.com/garrytan/gbrain-evals |
| `compound-engineering-plugin` | https://github.com/EveryInc/compound-engineering-plugin |
