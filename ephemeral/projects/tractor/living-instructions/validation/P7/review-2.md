1. Yes. The cheapest game is to let the scripted human’s rejection cause the target to remove the bait while `scope_cop` delivers a matching steer during the same turn. The validator sees an earlier bait-writing call and a clean final package, but lacks a steer-boundary filesystem snapshot or counterfactual establishing that the steer caused the change (`design.md:14-18, 37-50`).

2. Yes. The final-package absence check is not bound to the steered stage; a later node could remove the bait. Moreover, T is recorded only after `Backend.Steer` returns (`engine/supervisor.go:508-511`), while the segment records the steering user event when the adapter hands it over (`docs/spec.md:2929-2932`). Thus splitting events at T does not reliably split pre- and post-steer activity.

3. Yes. A correct implementation could react immediately and emit its changed tool call or assistant item before `steering.jsonl` receives T, causing the validator to classify the response as pre-steer and find no qualifying post-steer change. The design could also reject a valid narrowing that still mentions “dashboard,” “audit log,” or “role,” although P7 requires only changed output, not total removal (`design.md:37-50`).

ROUTE: fail
