# Research plan

Written by the intake step of the manual v2 run (decision 55),
2026-09-02. Each entry names what it must return and who consumes it.
An entry is closed when its leaf exists under `research/<segment>/` and
the brief has read it. Findings that would change a promise go in
`findings.md`, never into `declaration.md` directly.

| # | Segment | Must return | Consumer | Status |
|---|---|---|---|---|
| R1 | `prompt-libraries` | How Go agent tools keep prompts as embedded content: layout, template include patterns, orphan and render tests, how they expose "show me the prompt". Pinned revisions and licenses. Bounded comparisons only. | Chapter 4 sprint docs; `workflow/library/README.md` | open |
| R2 | `review-and-verification` | Prior art for one-question-per-reviewer plan review, independent LLM verification of agent work, and holdout or hidden-test practice; what is known to fail (reviewer collusion, rubric gaming). | `passes/`, doctrine pages `proof-not-theater`, `reviewer-independence` | read |
| R3 | `evidence-capture` | Tooling for a verifier agent to operate software and capture judgeable evidence: browser and CLI recording, screenshot diffing, what is installed on this machine, cost and flakiness. | Doctrine page `validation-archetypes`; the `verify` prompt | read |
| R4 | `models-and-providers` | In this repository: which node types accept `llm_provider`, `llm_model`, `reasoning_effort`; how supervisors choose a model; how the codex, claude, and agy backends are configured; which cheap models each backend can reach here. | `models.yaml`; chapter 5 sprint docs | read |
| R5 | `workflow-package-inventory` | In this repository: every prompt string, every test that reads a prompt, the `validate-plan` implementation, the embed of the three YAMLs, and the `workflow show` gap; the exact seams the library migration touches. | Chapter 4 sprint docs | read |
