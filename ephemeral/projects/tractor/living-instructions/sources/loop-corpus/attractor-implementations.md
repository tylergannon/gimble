# Public Attractor implementation inventory

Snapshot: 2026-08-18. Commit IDs identify the inspected source, not a claim
about latest releases after this date.

## Direct implementations and close ports

| Repository | Inspected commit | Stack / form | Material useful to Tractor guidance |
|---|---:|---|---|
| [strongdm/attractor](https://github.com/strongdm/attractor) | `fb57a55ed973` | Normative NLSpecs, not an implementation | Canonical plan/implement/validate loop, goal gates, retries, context fidelity, and human-gate examples. |
| [tylergannon/tractor](https://github.com/tylergannon/tractor) | `37df42b` | Go; JSON/YAML; Codex plugin | Target system. Simpler node union, chooser prose, tool exit-code routing, isolated parallel worktrees, supervision. |
| [samueljklee/attractor](https://github.com/samueljklee/attractor) | `cf7822c45fd6` | Python | Twenty readable examples: research-then-build, code review, spec-driven development, parallel approaches, and supervisor loops. |
| [microsoft/amplifier-bundle-attractor](https://github.com/microsoft/amplifier-bundle-attractor) | `701edc7e7795` | Python / Amplifier bundle | Largest curated example set; convergence loop, bug fix, refactor, PR review, multi-lens review, feature build, and detailed design principles. |
| [calebmchenry/nectar](https://github.com/calebmchenry/nectar) | `9c176716a66a` | TypeScript | Local-first DOT runner with an agent loop, checkpoints, human gates, retries, and restarts; its garden-authoring guide is unusually approachable. |
| [jhugman/attractor-pi-dev](https://github.com/jhugman/attractor-pi-dev) | `f289d3b20505` | TypeScript / pi.dev | Backlog loop: list ready work, plan, implement, validate, close, commit, repeat; also a deliberately minimal Ralph loop. |
| [jmccarthy/attractor-c](https://github.com/jmccarthy/attractor-c) | `144e20137e5d` | C11 | Small linear examples plus self-audit/fix pipelines; useful evidence that the workflow model is not language-specific. |
| [martinemde/attractor](https://github.com/martinemde/attractor) | `36c256fc5e83` | Go | Compact `read spec -> implement -> test -> implement` development loop. |
| [anishkny/attractor](https://github.com/anishkny/attractor) | `43acbd8d9cea` | Python | Minimal linear, branching, tool, human-gate, and stylesheet examples. |
| [anishkny/attractor-nodejs](https://github.com/anishkny/attractor-nodejs) | `e912be67615b` | Node.js | Same core concepts in a small JavaScript implementation; useful simple-linear authoring examples. |
| [bkrabach/attractor](https://github.com/bkrabach/attractor) | `6fbf54d79608` | Rust | Concise library quick start with plan and implementation stages, events, human gates, and checkpointing. |
| [bborn/attractor-ruby](https://github.com/bborn/attractor-ruby) | `8ede7ddc6477` | Ruby | Small embedded pipeline example and clear handler/routing summary. |
| [bencivjan/attractor-scala](https://github.com/bencivjan/attractor-scala) | `6480d8383c55` | Scala 3 | Built-in developer, evaluator, factory, megaplan, and sprint pipelines. |
| [jaytaylor/attractor-php](https://github.com/jaytaylor/attractor-php) | `bcd348182723` | PHP | Minimal CLI and basic pipeline with verification evidence. |
| [maciejgryka/attractor](https://github.com/maciejgryka/attractor) | `9130a27fbed5` | Elixir | Early compact implementation with one simple DOT example. |
| [Alezrik/attractor-phoenix](https://github.com/Alezrik/attractor-phoenix) | `6755d25bdd17` | Elixir / Phoenix | Standalone engine plus HTTP operation and multiple graph renderings. |
| [siarhei-belavus/attractor-kotlin](https://github.com/siarhei-belavus/attractor-kotlin) | `0f8bd7ba8bb9` | Kotlin | Simple, branching, review, and real workspace examples with a dashboard/API. |
| [TheFellow/fkyeah](https://github.com/TheFellow/fkyeah) | `875d56ee12b0` | F# | Large fixture/example corpus; emphasizes separate threads for late stages to avoid cumulative context exhaustion. |
| [bphansg/attractor](https://github.com/bphansg/attractor) | `69d056f7450d` | TypeScript | Five-minute examples, natural-language pipeline generation, test/fix loops, and goal gates. |
| [bromanko/attractor](https://github.com/bromanko/attractor) | `e268580e1006` | TypeScript / KDL adaptation | Replaces DOT with a KDL workflow language; shows that the useful abstraction is typed stages and routes, not DOT itself. |
| [talmage89/attractor](https://github.com/talmage89/attractor) | `1e5d94ef5df0` | TypeScript / `.dag` | Real sprint workflow: plan, audit, implement, review, parallel tests, fix, wrap up. |
| [arikWaisman/klaus](https://github.com/arikWaisman/klaus) | `340d04834af0` | TypeScript | Clear plan/review/implement/test/fix example and a much heavier multi-model consensus example useful as a contrast. |
| [mhingston/factorial](https://github.com/mhingston/factorial) | `cbaff572dffe` | TypeScript | Twenty-plus examples grouped by basic, quality, governance, and reliability patterns. |
| [tgoodwin/tractor](https://github.com/tgoodwin/tractor) | `1a3fc15a31ec` | Elixir | Feedback loops, parallel audits, conditional forks, recovery, and a pipeline-authoring skill. Unrelated to this repository despite the same name. |
| [Industrial/streamweave-attractor](https://github.com/Industrial/streamweave-attractor) | `5f78f0983081` | Rust / StreamWeave | `run tests -> report` minimal example and a pre-push fix-and-retry workflow. |
| [citadelgrad/pascals-discrete-attractor](https://github.com/citadelgrad/pascals-discrete-attractor) | `80e130b024b3` | Rust | Strong “backpressure” framing, planning-to-execution template, validation layers, and bounded goal-gate retries. |
| [dallumnz/software-factory](https://github.com/dallumnz/software-factory) | `a232e14b1fa4` | Python | Six compact DOT examples and simulation-first execution. |
| [point-labs-dev/arc](https://github.com/point-labs-dev/arc) | `ecbce6c94e5f` | TypeScript | A single default convergence pipeline; useful proof that one good loop can be the product surface. |
| [ArgaLabs/attractor-visual-builder](https://github.com/ArgaLabs/attractor-visual-builder) | `2394d4c3244f` | Python + TypeScript UI | Direct implementation paired with a visual builder; reinforces graph legibility as an authoring concern. |
| [jawhnycooke/attractor](https://github.com/jawhnycooke/attractor) | `b2601b2add67` | Python | Full pipeline, LLM, agent-loop, checkpoint, and HTTP implementation. |
| [apridachin/attractor](https://github.com/apridachin/attractor) | `fb7cc458cc40` | Python demo | Small human-gate implementation and logs. |
| [rkinder/attractor](https://github.com/rkinder/attractor) | `5ac78e903226` | JavaScript | Implemented engine plus development-lifecycle, testing, documentation, and code-review workflows; README foregrounds feature specs. |
| [az9713/attractor-software-factory](https://github.com/az9713/attractor-software-factory) | `0d5c732f4757` | Python | Six end-to-end project-building pipelines and practical authoring documentation. |
| [jleechanorg/dark-factory](https://github.com/jleechanorg/dark-factory) | `92c71077fecf` | Python + Rust | Large operational corpus, including intentionally slim two-node, research, feature, bug-fix, red/green, and review pipelines. |
| [johnnyhchen/soulcaster](https://github.com/johnnyhchen/soulcaster) | `84dbfbbeacd4` | C# | Full runner and unusually varied workflows: critique, UI ideation/implementation, multimodal editing, QA, provider validation, and parallel queues. |
| [wcraigjones/attractor-factory](https://github.com/wcraigjones/attractor-factory) | `277141cf888e` | TypeScript | Production-oriented factory, review attractors, self-cycle scripts, and many generated/stored pipeline records. |
| [EndlessCommerce/orchestra](https://github.com/EndlessCommerce/orchestra) | `bd51b2b2441a` | Python | Staged implementation of the full spec with a capstone adversarial PR-review pipeline. |
| [danshapiro/kilroy](https://github.com/danshapiro/kilroy) | `b55fb0f2b3d5` | Go | Local-first Attractor runner with isolated worktrees, commit-per-node, resume, and a substantial example set. |
| [2389-research/smasher](https://github.com/2389-research/smasher) | `e3736a233805` | Rust | Reimplementation from scratch with nested pipelines, web operation, and 30-plus DOT examples. |
| [2389-research/mammoth](https://github.com/2389-research/mammoth) | `65c8294e1216` | Go | Attractor-style runner built on Tracker with many examples and run persistence. |
| [fabro-sh/fabro](https://github.com/fabro-sh/fabro) | `400be9f2dcbf` | Rust + TypeScript | Mature Attractor-descended platform emphasizing verification gates, cross-model critique, reusable workflows, and durable run inspection. |

## Adaptations, demonstrations, and adjacent attractor systems

| Repository | Inspected commit | Classification | Why it belongs in the corpus |
|---|---:|---|---|
| [strongdm/agate](https://github.com/strongdm/agate) | `ea95448fb317` | Adjacent, file-state orchestrator | Interview, design, sprint-plan, implement/review, assess, and next-sprint cycle. It is simpler to operate than a user-authored graph and is useful evidence about common lifecycle stages. |
| [foundatron/octopusgarden](https://github.com/foundatron/octopusgarden) | `1414be0be9b9` | Adjacent dark factory | Explicit `generate -> test -> score -> feedback -> regenerate` loop with stall diagnosis. |
| [2389-research/tracker](https://github.com/2389-research/tracker) | `c6ea59d98dc1` | Adapted graph system | Uses Dippin rather than DOT; ships competitive implementation and multi-lens review workflows. |
| [amolstrongdm/attractor](https://github.com/amolstrongdm/attractor) | `bbfc95176dac` | Scenario-driven factory | Not a close pipeline-spec port; contributes satisfaction scoring and digital-twin/scenario ideas. |
| [914ash/attractor-workflow-demo](https://github.com/914ash/attractor-workflow-demo) | `71bc7bda0da9` | Demonstration | One concrete requirements -> human review -> storage/UI -> tests/browser -> human review -> walkthrough workflow. |

## Excluded or low-signal candidates

- Pure copies of the three NLSpecs with no implementation or distinct workflow
  material were not counted as implementations.
- Repositories that only mention Attractor in a comparison, research note, star
  list, or dependency were used for discovery but excluded from the table.
- Upstream forks with no independent commits or examples were not treated as
  separate evidence.

## Licensing and reuse boundary

The proposal paraphrases patterns and writes fresh Tractor JSON; it does not
copy implementation code. Several inspected repositories have no detectable
top-level license, so their text and examples should be treated as reference
only. Notable declared licenses include MIT
(`microsoft/amplifier-bundle-attractor`, `TheFellow/fkyeah`,
`bborn/attractor-ruby`, `bkrabach/attractor`, `calebmchenry/nectar`,
`dallumnz/software-factory`, and `tgoodwin/tractor`), Apache-2.0
(`strongdm/attractor`, `strongdm/agate`,
`jhugman/attractor-pi-dev`, `jmccarthy/attractor-c`,
`martinemde/attractor`, and others), dual MIT/Apache-2.0
(`citadelgrad/pascals-discrete-attractor`), and CC BY-SA 4.0
(`Industrial/streamweave-attractor`). Reusing verbatim examples requires a
fresh license check and the applicable attribution/share-alike treatment.
