# Codex desktop queue delivery

Use this only when the user asks the current local Codex desktop task to keep
watch over a Gimble run.

Read `CODEX_THREAD_ID` and `CODEX_SESSION_ID` from this task's shell. Those
values belong to the caller; environment inherited by the shared Gimble MCP
server does not identify which task called it. Pass them to `start_run`:

```json
{
  "parent": {
    "thread_id": "<CODEX_THREAD_ID>",
    "host_id": "local",
    "session_id": "<CODEX_SESSION_ID>"
  }
}
```

That explicit attachment opts the run into direct Codex delivery. Gimble's
detached wake service calls:

```text
codex queue --thread <CODEX_THREAD_ID> --message <bounded run digest>
```

Do not create an automation, heartbeat, shell timer, or parent-side polling
loop. Do not infer the parent from a title, cwd, or recent activity.

A loaded idle task starts queued input as a new turn. A busy task retains it
until the current turn becomes idle. An unloaded task retains it until resume;
do not claim enqueue alone reopened a cold task. Gimble records rejected queue
requests as `HostWake` failures and retries unread news without failing the
pipeline.
