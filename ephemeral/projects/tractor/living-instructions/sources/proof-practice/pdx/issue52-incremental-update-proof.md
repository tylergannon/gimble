# Issue 52 Incremental KB Update Proof

- Pass: `True`
- Service: `https://pdx-extraction-service.fly.dev`
- KB: `kbt_7d121e29c1d2400e0f7884d1e6911baf`
- Cleanup: `deleted`
- Scenario: start from a built 5-document KB, replace `rules.md`, run the non-forced build path, then chat for the changed fact.

## Build Metrics

| Build | Mode | Changed docs | Tool calls | Total tokens | Wall time ms | Lint |
|---|---:|---|---:|---:|---:|---|
| Initial forced full | full | doc_claims, doc_coverage, doc_deadlines, doc_glossary, doc_rules | 20 | 108989 | 81917 | True |
| Incremental after replacing `rules.md` | incremental | doc_rules | 16 | 66070 | 31157 | True |
| Forced full after same replacement | full | doc_claims, doc_coverage, doc_deadlines, doc_glossary, doc_rules | 22 | 103966 | 61635 | True |

## Cost Comparison

- Incremental / forced-full-after-change token ratio: `0.6354962199180502`
- Incremental / forced-full-after-change tool-call ratio: `0.7272727272727273`
- Incremental / forced-full-after-change wall-time ratio: `0.5055082339579784`

## Product Proof

Initial chat answered with `$75` before replacement. After replacing exactly one source document and running the incremental build, production chat answered:

> The specialty deductible pilot amount is `$125` for qualified glass repair claims. [[doc:doc_rules]]

Citation count after update: `1`. The first updated citation includes `source_document_id` `kbd_8924da6824b807b2565f9b1416b3265d` and source URL `pdx-upload://uploads/pdx/kb-proof-prod-20260613175601-codex/kb-upload/kbt_7d121e29c1d2400e0f7884d1e6911baf/kbu_935b2fe9b7fee83c04b3b1b2b280fade-rules.md`. The full JSON proof preserves the citation payloads, source document metadata, build records, and cleanup result.
