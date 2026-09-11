correction: Godoc, not a published API.md, is the authoritative public API definition.
decision: Keep `codex.New` and `claude.New` as the explicitly requested public constructors returning `HarnessAdapter`; only their concrete implementation types become private.
design_bug: Event currently combines unrelated variants in a comprehensive struct; issue #122 owns replacement with Polytype sealed unions, and this task leaves Event unchanged.
decision: HarnessAdapter remains in the root package for this pass because moving it without Event would either create an import cycle or force a temporary duplicate event contract; issue #122 should move both along the final dependency boundary.
correction: Runtime-owned web serving and every related web, build, embedding, CLI, and generated-binding change are outside this API-reorganization PR; issue #124 owns that work.
friction: TestAttestEventFixture missed its intended overlap once during verification because its timing window closed before the reviewer end event; an uncached full rerun passed. This is separate from this API reorganization and should be treated as a flaky proof if it recurs.
