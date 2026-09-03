# Interview documentation sprint

friction: `TRACTOR_RUN_DIR` reached the Codex app-server but was absent from commands launched through the code-mode tool host, so the first live `tractor ask` warned instead of recording its timeline event -> propagate the run-scoped environment through tool-host command execution or document the host boundary.

correction: The first reviewer question described Tractor vs. upstream Attractor as order 4, but the file still used order 3; ask a corrected follow-up before changing the navigation order.

decision: Reviewer approved `src/content/docs/interviews.md` titled Interviews at order 3 and moving Tractor vs. upstream Attractor to order 4; recorded in interview questions 0004 and 0005.

friction: `pnpm verify` exposed pre-existing Vite+ formatting drift in `authoring-pipelines.md` and `loops.md` in addition to the edited skill -> run the repository formatter before site verification so the aggregate gate is meaningful.
