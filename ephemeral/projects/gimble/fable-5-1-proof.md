# Fable 5.1 alias proof

On 2026-09-01, Claude Code 2.1.252 still resolved its own unversioned `fable`
alias to `claude-fable-5`. A no-persistence probe using the new concrete ID
completed successfully and reported `claude-fable-5-1` as its canonical model.

A fresh build of Tractor's `cmd/agent` was then invoked with a Codex caller
identity, leaving its Fable version implicit. The real CLI turn produced:

```text
[claude-code:unrecognized_model] {"model":"claude-fable-5-1","query_source":"sdk"}
```

The turn called Claude's `Write` tool and created `model-proof.txt`. An
independent byte dump of the file was:

```text
46 41 42 4c 45 5f 35 5f 31 5f 52 45 41 44 59 0a
 F  A  B  L  E  _  5  _  1  _  R  E  A  D  Y \n
```

The CLI returned a native session ID and the structured outcome:

```json
{"next":"done","notes":"Wrote model-proof.txt containing \"FABLE_5_1_READY\\n\" in the workspace directory."}
```

This demonstrates both halves of the claim at the public helper entry point:
the implicit Fable selection was concretized to 5.1 before Claude Code saw it,
and that concrete model completed observable workspace work.
