# Tractor loop-guidance research

Research snapshot: 2026-08-18.

This project gathers public Attractor implementations and practical agent-loop
advice, then derives a small proposed Codex skill for choosing, writing, and
running ordinary Tractor pipelines.

## Contents

- [attractor-implementations.md](attractor-implementations.md) inventories 44
  public implementation, adaptation, and demonstration repositories inspected
  at pinned commits, plus canonical upstream and target Tractor context.
- [workflow-corpus.md](workflow-corpus.md) extracts reusable workflow patterns
  from implementation examples, practitioner articles, open-source agent
  methods, research, and official OpenAI guidance.
- [synthesis.md](synthesis.md) turns the corpus into Tractor-specific authoring
  advice and identifies where sources disagree.
- [proposal/use-tractor/SKILL.md](proposal/use-tractor/SKILL.md) is a lean skill
  proposal, not an installed or released plugin change.
- [proposal/use-tractor/references/patterns.md](proposal/use-tractor/references/patterns.md)
  contains native Tractor JSON examples.

## Discovery method and limits

Discovery used GitHub repository and code search for combinations of
`strongdm attractor`, `attractor-spec.md`, `coding-agent-loop-spec.md`,
`DOT-based pipeline runner`, and exact phrases from the upstream NLSpec.
Candidates were shallow-cloned and inspected rather than accepted from search
snippets. Repositories were classified as direct implementations, embedded or
adapted systems, demonstrations, or adjacent/spec-only material.

This is a broad public-web snapshot, not a proof that no private, unindexed, or
new implementation exists. The 196 upstream forks were not counted merely for
being forks; a repository needed implementation code, distinct examples, or a
documented adaptation to enter the inventory. Search results with unrelated
mathematical or neural-network uses of “attractor” were excluded.

## Main conclusion

The most reusable Tractor shape is not a grand software factory. It is one
worker, one independently observable check, and a bounded back-edge. Add a
planning, diagnosis, or review node only when that boundary creates a useful
artifact, fresh perspective, or different kind of judgment.
