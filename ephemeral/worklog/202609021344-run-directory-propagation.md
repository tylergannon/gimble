# Run directory propagation

discovery: `runPipeline` constructs Codex and agy adapters before `Runner.Run`, while both default adapter configs snapshot `os.Environ()` at construction. Since the engine owns `TRACTOR_RUN_DIR` setup at run start, default CLI process configs must inherit the environment at process spawn instead of adapter construction.

decision: Reviewer answered “No. Environment only. Add a fallback when a backend proves it scrubs the environment, not before.” Do not add a run-directory preamble field or `tractor ask --run` fallback in this sprint.
