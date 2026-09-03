1. Yes. The cheapest game is to put a new artifact skeleton inline in an agent-facing prompt and omit it from the README’s skeleton list. Prompt provenance and sentinel checks still pass, while only listed skeletons are examined and unlisted templates are explicitly allowed (`design.md:115-123`). The infer prompt inventories probe outcomes and Go-carried prompt text, not unlisted inline skeletons (`design.md:190-203`). This is cheaper than creating a separate template, include action, and README entry, yet P8 is false.

2. Yes. The claim that “an inert file listed beside an inline skeleton fails” is unsupported (`design.md:119-123`). The actual check merely verifies that each listed template’s sentinel appears somewhere in some node output (`orphan-walk-and-render.sh:101-118`); a prompt can retain the inline skeleton and also render the template solely to expose its sentinel. No injected probe or recorded evidence distinguishes genuine template provenance from that duplication.

3. Yes. A correct implementation may contain an unused, syntactically valid library file such as `prompts/experimental.md`: all prompts actually used remain file-backed, all doctrine pages remain referenced, and `show` remains correct. P8 and the declared seam impose rendered-reference liveness only on doctrine pages (`declaration.md:80`), but the design rejects every unused prompt, supervisor, or pass file (`design.md:130-134`), creating a concrete false failure.

ROUTE: fail


