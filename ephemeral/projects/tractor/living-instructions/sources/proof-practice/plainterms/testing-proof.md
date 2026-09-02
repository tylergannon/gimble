# Recipe: Testing And Proof

Use this file when answering: "what verification is enough?", "do I need BDD?", "what proof should I capture?", or "how do I prove this guide still routes correctly?"

## Product Code Proof

| change                                       | proof                                                                                                                                                                |
| -------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Svelte component/UI behavior                 | focused browser/BDD when browser-visible, `pnpm run check`, screenshots if visible                                                                                   |
| SvelteKit remote form/query                  | focused unit/integration plus browser/BDD only for the visible user flow                                                                                             |
| API endpoint or server route                 | focused route/unit/integration tests plus direct HTTP proof; no Playwright BDD unless the behavior is browser-visible UI                                             |
| Webhook delivery or external callback        | focused workflow/step tests, receiver/API proof, stored payload/DB assertions, signature/error assertions; no Playwright BDD for the endpoint contract               |
| Workflow v5 orchestration                    | real Workflow v5 run plus observed steps/state and DB/UI result; BDD only for visible UI behavior around the workflow                                                |
| Supabase/RLS/storage                         | migration tests and real full-stack path through RLS/storage; no Playwright BDD for the policy/RPC/storage contract itself                                           |
| AI extraction/prompt/schema                  | focused unit test plus fixture extraction/render artifact                                                                                                            |
| Document extraction/rendering transform      | focused unit/fixture proof plus rendered artifact; no Playwright BDD solely for extraction/rendering correctness                                                     |
| Customer feedback about extraction/rendering | start from complaint artifacts, state source truth, compare extracted data and rendered PDF, apply current invariants, then add fixture or complained-document proof |
| docs/coding-guide                            | targeted link/path checks plus review against the changed routing surface                                                                                            |

Default PR E2E runs use `pnpm run test:e2e:required`, which excludes
`@experimental-kb`. Keep experimental product-KB/product-contract scenarios
under `@experimental-kb` and prove them explicitly with
`pnpm run test:e2e:experimental-kb` when the changed behavior depends on that
slice.

For changes to Preview E2E, local E2E loop scripts, Vercel preview env cleanup,
or local bless policy, use targeted script/workflow proof instead of full
application Preview E2E. The expected proof is classifier/unit coverage for the
changed helper, `git diff --check`, and a static workflow review. Hosted
Preview E2E itself must fail early with a named class for Supabase branch target
recovery, Vercel branch env absence, runtime identity mismatch, stale dispatch,
or browser-test failure. New or changed preview/local proof scripts should also
be syntax-checked directly (`node --check` for `.mjs`, YAML parsing for
workflow files when touched) and covered by focused tests for every emitted
failure class.

## E2E / Storybook Split

`pnpm test` now runs both `vitest --project=unit` and
`vitest --project=storybook`. Storybook browser tests are an executable proof
layer when they live in the `storybook` Vitest project; static stories or
`storybook:build` alone are not enough to delete E2E behavior coverage.

During every Proof of Work run, the `proof-of-work` CI Confidence Pass runs two
requests: coverage design before material implementation
(`.agents/skills/ci-confidence-test-placement/SKILL.md`) and test repair after
behavior proof (`.agents/skills/ci-confidence-test-maintenance/SKILL.md`). In
Codex, use the project-scoped `ci-confidence-engineer` agent from
`.codex/agents/ci-confidence-engineer.toml` when a separate subagent run is
useful; otherwise run the tracked charter and skills directly. The charter
`.agents/workspace-agents/ci-confidence-engineer/CHARTER.md` is canonical for
layer judgment; the split below is the local testing guide.

Use this split:

- Keep at least one true browser upload lane for file input, client hashing,
  upload reservation, signed storage PUT, storage callback, extraction, and
  rendered/downloadable result.
- Convert slow setup to seeded app E2E when the behavior under test is a real
  dashboard, route, remote mutation, download, or refresh path for a known
  document. The feature text should stay product-language, such as "signed in
  as an agency admin" or "with an at-threshold quote ready for branded PDF
  download"; step code owns session reuse and seeded records.
- Move pure component states to Storybook browser tests only after the matching
  Storybook play assertions exist and are in the gated `test:storybook` path.
  This is the right layer for modal close controls, copy controls, loading and
  error variants, responsive fit, controlled input synchronization, validation
  display, and component-only visual states.
- Keep E2E assertions for persistence, routing, auth, webhooks, API/UI
  integration, and server mutations. A Storybook prop fixture cannot prove that
  persisted state is reloaded from the database after navigation.

Dashboard and upload E2E should address the scenario's record instead of a
pristine page. Use `spec/helpers/dashboard-selectors.ts` helpers such as
`quoteRowForDocument(page, documentId)` and
`uploadTrackerForDocument(page, documentId)`, backed by row/tracker attributes
like `data-document-id`, `data-status`, `data-error-code`, and
`data-requires-ack`. Avoid `.first()`, `nth(0)`, and global row counts unless
the product behavior is actually about the whole list.

For product contract UI changes, the focused `@product-contract` scenario is
the expected browser proof. Hydration regressions should be asserted against a
real hydrated control after navigation, then the toast/download path should be
asserted after `Download PDF`; do not wait for Sonner's toast-list element
before a toast exists.

## Development Parity Guardrail

Do not make local or dev application code bypass the production contract just to
get E2E green. If production capture submits a PDX `file_url` that PDX fetches
from Supabase Storage, local proof for that path must exercise the same
externally reachable HTTPS `file_url` contract.

The local `cloudflared` tunnel publishes the app origin for PDX callbacks. Local
PDX file URLs use that same HTTPS origin plus the Vite `/__pdx-supabase` proxy to
reach local Supabase Kong/API. A local signed URL built from
`SUPABASE_URL=http://127.0.0.1:*` is not equivalent until it is rewritten to that
PDX-facing HTTPS proxy origin. Do not hide this with a dev-only fallback that
uploads bytes directly to PDX; PDX proof must still send and fetch a `file_url`.
`pnpm run dev:e2e:loop` runs this as a preflight before Playwright so Docker
socket drift, Workflow route overlays, and invalid callback/file URL shape fail
once with an actionable classification instead of many downstream browser
timeouts.

When Supabase production/preview reads are needed for proof, use
Doppler-backed exported credentials or an approved connector path. Do not add
agent-facing instructions that rely on raw `supabase link` or
`supabase db query --linked`; those paths can store native CLI credentials and
trigger local Keychain prompts. `pnpm run verify:supabase-credential-guard`
checks repo guidance for this failure mode.

## E2E Tenant Setup

Plainterms is multi-tenant. For browser E2E, the isolation proof is that
generated users in one organization cannot see another organization's rows
through normal authenticated paths. A fresh empty database is useful for schema
work, but it is not the product security boundary.

- Prefer product-path setup for tenant lifecycle behavior: signup, invite email
  receipt/acceptance, set-password, password reset, role assignment, manager
  visibility, and cross-org isolation.
- Use run-scoped emails, organization names, document labels, and storage paths
  so stable shared preview databases are safe for non-schema PRs. Canonical
  browser and Supabase integration test orgs use `pt-e2e-*` slugs.
- For non-auth behavior, prefer worker-keyed shared accounts plus
  scenario-owned records over per-scenario fresh users. Establish the browser
  session through the shared E2E helper path and keep scenario text in product
  language. Preserve fresh-user setup only when auth, invite, suspension,
  role/permission, or deliberately first-run behavior is under test.
- Hosted preview tests that observe email through Resend must use addresses
  under `@guys.dev.guilde.ai`; build those recipients with `buildE2EEmail(...)`.
  Local-only addresses such as `example.test` can be delivered to
  Mailpit/Inbucket but will not be visible to the hosted Resend receiver. The
  deployed preview app needs `RESEND_API_KEY` to send. The hosted runner should
  query Resend directly with `RESEND_API_KEY` when the workflow exports it; for
  repository-dispatch checks still running an older default-branch workflow,
  use the preview app's protected `/api/preview-e2e/resend-email` probe with
  `PDX_CALLBACK_SECRET` and the Vercel automation bypass header.
- Treat service-role Supabase setup as fixture plumbing. It can create
  deterministic run-scoped org graphs, but it should not be the proof for
  signup, invite receipt/acceptance, set-password, password reset, role
  authorization, or cross-tenant visibility.
- The preview vacuum workflow removes stale `pt-e2e-*` organizations after the
  configured TTL. Do not create DB-backed test data outside a vacuum-owned
  naming contract. If a product surface must show data that cannot use
  `pt-e2e-*`, extend the vacuum job and naming contract first.
- E2E tests must not delete application data during scenarios or `After` hooks.
  Do not call Supabase `.delete()` or Storage `.remove()` from `spec/**` to
  clean product rows. Create canonical vacuum-owned `pt-e2e-*` data, leave
  recent rows for failure diagnosis, and let
  `.github/workflows/vacuum-preview-env.yml` plus
  `scripts/e2e-test-data-vacuum.ts` own stale cleanup. Browser/runtime cleanup
  such as closing pages, contexts, and local HTTP servers is still allowed.
- Use Supabase branch databases when the changed-file scope includes
  branch-scoped Supabase artifacts such as `supabase/migrations/**`,
  `supabase/config.toml`, `supabase/seeds/**`, `supabase/templates/**`, or other
  non-test files under `supabase/`. For ordinary app changes and
  `supabase/tests/**`-only changes, Vercel preview plus stable Supabase can be
  enough when the runtime identity guard confirms the deployed app and test
  harness use the same Supabase ref.

## Artifact Conventions

- Do not add repo-local guide verification scripts, probe banks, or audit
  journals. Use global skills and the current session worklog for process
  evidence.
- For PR proof artifacts, prefer `pnpm pr-proof:upload` when the shell has the
  required Supabase proof-bucket credentials. It uploads files to the private
  `pr-proof` bucket, emits signed URLs plus JSON/Markdown manifests, and can
  rewrite bare Markdown filenames in a PR body. If credentials are missing,
  record that blocker and keep artifacts reproducible from the worktree.
- For product proof, prefer the existing task/issue verification path if one exists, such as `verification/issue-####/` or `spec/test-results/issue-####/`.
- Name artifacts by behavior, not tool: `workflow-v5-run-state.json`, `fixture-0094-savvy-page-1.png`, `remote-form-red.log`.
- For visual proof, capture before/after screenshots or rendered PDF rasters and upload when publishing PR/issue handoff.
- For non-visual proof, keep command output summaries and generated JSON/text artifacts with enough context to reproduce.

## Docs Guide Proof

Run targeted checks that match the changed files. At minimum:

```sh
git diff --check
```

When deleting or moving a guide surface, grep for the exact old paths and
artifact names across `AGENTS.md`, `CLAUDE.md`, `README.md`, `docs/`,
`.claude/`, `scripts/`, and `package.json`.

For routing changes, manually verify that the root index points to existing
leaves and that the answer can be recovered from the root plus the chosen leaf.

## Stop Rule For This Guide

Stop when the root index routes the changed concepts, all touched links or paths
exist, and no repo-local process harness was added to prove the docs.

## PR Closeout

When the user asks for auto-merge or the repo task normally includes
auto-merge, closeout proof must verify one of these states before stopping:

- GitHub auto-merge is enabled for the PR (`autoMergeRequest` is set).
- The PR has already merged.
- The user explicitly opted out of auto-merge for this turn.

Do not treat green checks, a pushed branch, or an open PR alone as closeout when
auto-merge was requested.

## Citations

Repo:

- `AGENTS.md`
- `docs/README.md`
- `sprints/worklog/README.md`
- `CHANGELOG.md`
