# Pass 4: proof quality

Question: taking all ten designs under `validation/` together, and the
`command` and `infer` fields in chapter 4's ledger, is there any
validator a coding agent could satisfy while its promise is false, any
check that is trivially true or depends on evidence the design does not
capture, or any design stricter than its promise? Look especially for
patterns a single review would miss: the same lazy mechanism reused
across promises, evidence that is asserted but never captured, a
scripted answerer that makes every scenario pass by construction.
Any such finding is a fail; owning node `design`, naming the promise.
