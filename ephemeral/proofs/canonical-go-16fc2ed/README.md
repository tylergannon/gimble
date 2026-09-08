# Canonical Go workflow proof

Passed on 2026-09-08 UTC against clean source commit
`16fc2edd7b38a3902f30653bf8dd5b1ff7899ac8` and candidate SHA-256
`47a0e8f6fbbb5710e5106c3731b0049acf4780c8e86f010e9232935ecaaf2dc6`.

```sh
go test -tags=integration ./internal/workflows -run '^TestCanonicalLoop$' -count=1 -v -timeout=25m -args -proof-dir /tmp/tractor-canonical-go-16fc2ed
```

The Go acceptance runner launched the real candidate's `run sprint-execute`
command with real native agents against a fresh, deliberately broken Go CLI.
The acceptance run passed in 317.51 seconds; the workflow itself took 310.71
seconds. Full native transcripts, application repository, candidate binary,
and baseline/final acceptance logs remain at `/tmp/tractor-canonical-go-16fc2ed`.

Observed sequence in [timeline.jsonl](timeline.jsonl):

1. Standard shipping selected, implemented and independently reviewed.
2. Engine command and evidence judge passed standard shipping. Goal evaluation
   returned `not_done`; expedited shipping was selected.
3. Expedited shipping implemented and independently reviewed.
4. Engine commands and evidence judges passed both sprints, including standard
   shipping revalidation. Goal evaluation returned `done`; pipeline completed.
5. The unchanged acceptance executable compiled before agent execution rebuilt
   the final application and passed all six public CLI scenarios.

The baseline exposed the incorrect standard charge at exactly 50.00 and all
three missing expedited cases. Final CLI stdout, exit codes, and expected
values are retained in [standard evidence](evidence-standard.json) and
[expedited evidence](evidence-expedited.json). Both independent reviewer
responses are retained alongside the [result receipt](result.json).
Acceptance files and source provenance checks passed; Codex memories were
disabled for this proof process.

This proves the flat two-item checklist loop. It does not demonstrate nested
loops, service lifecycle, or an engine validation failure/retry: the negative
baseline observations happened before workflow execution.

Additional checks passed: full repository race-enabled Go tests, integration
package vet, and integration-tagged golangci-lint. This evidence-only closeout
does not change the source inputs of the proved build.
