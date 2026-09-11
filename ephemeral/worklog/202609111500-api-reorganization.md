correction: Godoc, not a published API.md, is the authoritative public API definition.
decision: Keep `codex.New` and `claude.New` as the explicitly requested public constructors returning `HarnessAdapter`; only their concrete implementation types become private.
design_bug: Event currently combines unrelated variants in a comprehensive struct; issue #122 owns replacement with Polytype sealed unions, and this task leaves Event unchanged.
decision: HarnessAdapter remains in the root package for this pass because moving it without Event would either create an import cycle or force a temporary duplicate event contract; issue #122 should move both along the final dependency boundary.
correction: Runtime-owned web serving and every related web, build, embedding, CLI, and generated-binding change are outside this API-reorganization PR; issue #124 owns that work.
correction: `Interrupt` was deliberately removed from the required `HarnessAdapter` contract and must remain optional; do not promote concrete harness capability into the public interface.
design_bug: The public Loop `Task` was reduced to `Lap` and opaque `Text`, which does not carry the structured information workflows need; issue #125 owns recovery of the intended contract, and this PR must not further document that shape.
friction: TestAttestEventFixture has repeatedly missed its intended overlap under `-race` because the reviewer starts just after the worker ends; ordinary uncached tests pass. This is separate from this API reorganization and the race result must not be reported as green.
