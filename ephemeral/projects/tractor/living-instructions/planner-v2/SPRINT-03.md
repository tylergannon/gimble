# Sprint 3: run it once

Run the new workflow on a small seed and make it work:

```
tractor workflow run plan --workdir <scratch> --logs <scratch>
```

Use a seed of a few sentences (a feature that is plainly three or four
sprints). Answer its questions with `tractor answer`. Expect the first run
to break; fix the prompts or the graph, not the engine, and run again.
Stop after three runs and ask if it still fails.

Record the result under `planner-v2/proof/README.md`: the seed, the run id,
the questions asked, the reviewer's findings, and the final checklist.
