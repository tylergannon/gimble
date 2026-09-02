# Handoff: Tractor documentation & presentation work

Mission: **update Tractor's documentation and presentation** — for humans
(site, visuals, examples) and for agents (skill/MCP descriptions, llms.txt).
Documenting Tractor **as it exists today**. No new engine features. If an idea
requires new API surface, note it in a parking lot and move on.

## Read first (the accepted state)

1. `ephemeral/projects/tractor/adoption-design/brief.md` — the adoption
   design brief: two audiences (humans convert on demos + a drawable mental
   model; agents act on skill descriptions, tool text, and zero-authoring
   entry points), the four-stage agent funnel, the named patterns, the site
   information architecture, build order. Accepted by Tyler ("this is
   great").
2. `ephemeral/projects/tractor/adoption-design/wisdom.md` — distillation of
   the tylergannon/agents skills (proof-of-work, adversarial review,
   consensus, nlspec, slice design, prompt craft): 14 principles with
   verbatim quotes and file citations. The goal is baking this wisdom into
   Tractor's docs and guidance.

Visual artifact of the brief: https://claude.ai/code/artifact/e8191720-2c11-4b64-a252-cb24ba6a355a
(update it via its URL, don't create a new one).

## One settled question

Tyler questioned nlspec's total language-agnosticism for write-once software.
A three-model critique circle (run through Tractor itself; evidence in
`critique-circle/`) examined it and converged **in Tyler's direction**:

> Specs and docs written against a single existing repo should be
> target-language-aware — real paths, signatures, idiomatic examples.
> Language-agnostic, behavior-only phrasing is reserved for durable
> contracts and acceptance claims (so they stay checkable from outside the
> source), not used as a general style. Specificity is licensed by how much
> of the referenced code already exists.

Treat that as the writing rule for Tractor documentation and for any
spec-writing guidance the docs give pipeline authors.

## Explicitly out of scope — do not resurrect

The critique-circle materials (`critique-circle/REVIEW.md`, `SYNTHESIS.md`)
and the "Idea Ledger" artifact contain **feature proposals that were never
requested or discussed** (a `blocked` terminal / failure taxonomy, a
`claims`/demonstrate parameter schema, token budget guards, human-approval
primitives, run export). They contaminated a documentation thread with
feature design. They are not decisions. Ignore them except as review
evidence for the language question above.

## Status update (2026-08-20)

Shipped to main: `examples/loops/` (fix-until-green, critique-circle,
bake-off, milestone-loop) + the `tractor` Claude Code skill with bundled
examples and a drift-tripwire test (#22); moment-based MCP server/tool
copy, llms.txt when-to-use + examples sections, and the Loops docs page
(#23). The skill text was produced by a Tractor collab-edit run; evidence
in `wisdom-circle/skill-collab/`. The wisdom circle's proposals, critiques,
merge notes, and SYNTHESIS.md (take 2 — examples-first, loops are two
nodes) live in `wisdom-circle/` (its own git repo). The agy artifact-declaration bug is fixed and merged (#24); the shipped
bake-off example then ran end-to-end live — three providers incl. the
Gemini lane with declared artifacts, judge ran each script itself, winner
merged, mechanical gates green (evidence:
~/.cache/tractor-bakeoff-proof/.tractor/proof-run). Caveat learned: nested
Claude CLI runs inside Claude Code's sandboxed /tmp scratchpad get their
worktree paths remapped — run Tractor workspaces from normal filesystem
locations. Still deferred: site rebuild beyond the Loops page; a live
critique-circle demo run (poem demo remains the reference).

## The original open work (for context)

Per brief.md's build order, the documentation-shaped items:

- Site: expand beyond the current 3 pages — patterns/examples pages using
  the real runs as material (`examples/`, the poem-critique demo, the
  critique-circle run), a diagram-led "how it works" page, surface the
  existing spec.md / 27 lint rules / CLI / MCP reference on-site.
- Agent-facing text: rewrite MCP server instructions and tool descriptions
  around user moments, not mechanics; llms.txt / llms-full.txt.
- Claude Code skill description (the highest-leverage string): written from
  trigger phrases ("have another model check this", "try a couple approaches
  in parallel", "keep working after I close my laptop").
- The two milestone loop patterns Tyler asked for (plan-the-steps-first vs
  choose-your-own-adventure) are described in brief.md §03 as Ladder and
  Pathfinder — as **documentation of pipeline patterns** authors can build
  with today's five node types, not as new engine features.

Repo context: Tractor repo at ~/src/tractor (Astro site lives at repo root;
docs content in src/content/docs/). Tyler's agents repo at ~/src/agents.
Tyler wants terse, kernel-first answers to direct questions.
