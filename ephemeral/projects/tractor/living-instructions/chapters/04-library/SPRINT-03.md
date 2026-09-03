# Sprint 3: doctrine pages and skeletons

Written by Claude, not the coder (interview 0013 round, question 5),
before the chapter's run starts, and committed under
`chapters/04-library/content/doctrine/` and `content/templates/`. The
coder's turn copies them into `workflow/library/doctrine/` and
`workflow/library/templates/`, adds the `doctrine` includes to the
three migrated prompts, updates the snapshot, and runs the proof
script. The pages are the design; the sprint is the wiring.

## Pages

Only pages the three existing prompts can cite now. Pages for nodes that
do not exist yet (validation archetypes, reviewer independence, prior
art, research leaf) come with their prompts in chapter 5, so the orphan
walk stays green.

| Page | Cited by | Distilled from |
|---|---|---|
| `doctrine/promises.md` | planner | decision 37; `sources/diffusioninc/.claude/skills/df-promise/SKILL.md` |
| `doctrine/elicit-then-prune.md` | planner | decision 38; spec-authoring stopping rule; grilling |
| `doctrine/question-files.md` | planner, both implement prompts | decisions 26–29, 39; `BUILD.md` |
| `doctrine/promise-adjacent-seams.md` | planner | decision 40; nlspec methodology (Parnas test) |
| `doctrine/proof-not-theater.md` | both implement prompts | decision 41; proof-of-work; research R2 findings |
| `doctrine/vertical-slices.md` | planner, large plan | slice-design; wisdom.md |
| `doctrine/chapter-doc.md` | large plan | df-chapter-create; chapters 1–4 `CHAPTER.md` |
| `doctrine/sprint-doc.md` | large plan, planner | df-sprint-plan; chapter 1 sprint docs |
| `doctrine/pyramid-index.md` | large plan | df-chapter-create |
| `doctrine/ledger.md` | planner, both implement prompts | `loop-node.md` §2; decision 15, 44 |

Each page: under about sixty lines, one idea, positive phrasing, a
`Source:` line pointing into `sources/` or `decisions.md`. Written in
the library's voice, which is the voice of `BUILD.md`: tells the agent
what to do and why in as few words as hold. A page is rendered as a
template, so it must not contain the library's delimiters except as
actions; `{{` is literal under non-default delimiters and allowed.

## Skeletons

`templates/brief.md`, `templates/promises.md`, `templates/recommendation.md`,
`templates/CHAPTER.md`, `templates/SPRINT.md`, `templates/ledger.md`. Each
is the artifact with its fixed parts filled and its variable parts as
prose placeholders in angle brackets. The planner prompt cites the first
three; the large plan prompt cites the rest.

## Prompt edits

The three migrated prompts gain `doctrine` includes where they currently
paraphrase a page, and lose the paraphrase. This changes prompt text, so
the sprint 1 byte-equality snapshot is updated in the same commit, with
the diff reviewed as content: nothing an agent is told may be lost, only
moved.

## Proof script

`prove/doctrine-pages.sh`: every page in the table exists with a
`Source:` line; every skeleton exists;
`TestLibraryNoOrphans` and `TestLibraryRendersAll` run and pass (the
render test is what catches a delimiter misuse in a page). This script
demonstrates the sprint; it is not part of P8's proof, which makes no
claim about page names. The `infer` judge on the ledger item reads the
pages against decisions 37–59.
