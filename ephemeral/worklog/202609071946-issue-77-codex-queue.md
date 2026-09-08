# Issue 77 Codex queue delivery

decision: Codex desktop delivery was initially modeled as a host-scheduled pull by the launching parent, while Claude Code remained an engine-scheduled socket push.
friction: The initial design depended on automation_update even though that tool was absent from this task's callable catalog -> this was a symptom of choosing the wrong transport, not a reason to block Tractor delivery.
decision: A Codex parent is identified explicitly by CODEX_THREAD_ID and host identity supplied by the launching task, never inferred from the shared MCP server environment, task title, or recent activity.
correction: agent-protocol forbids writes under docs/ without express permission; reverted the unauthorized docs/spec.md edit before the issue checkpoint was made merge-ready.
correction: a green implementation suite is not the issue's live same-parent proof; do not declare the PR merge-ready before live delivery is observed.
correction: Tyler rejected the heartbeat/pull architecture as needless indirection; the real requirement is external delivery into a waiting Codex agent, and Codex already exposes `codex queue --thread <id> --message <text>` for that purpose.
decision: issue #77 now uses Tractor's existing news-triggered wake service with a Codex queue channel; `turn/steer` remains active-turn input, while `thread/queue/add` starts a loaded idle thread or waits for a busy one.
decision: cold/unloaded threads retain queued input until resume; treat that as a bounded lifecycle limitation rather than replacing direct messaging with a polling automation.
doc_bug: issue #77's prior “native in-chat heartbeat” decision conflated direct delivery, busy deferral, and cold-thread recovery -> issue body and title were rewritten on 2026-09-07 to specify direct queue delivery.
proof: live detached run 030580088758f6ef8c92e1a40f43c302 queued marker TRACTOR_CODEX_QUEUE_77_DIRECT_LIVE_1 to parent 01a07eab-99d5-7923-84ac-f41374506751; `ephemeral/issue-77-proof/direct-queue-live-1/run/timeline.jsonl` records six source events followed by delivered HostWake through codex_desktop.
proof: after the launching turn completed, the same parent task received TRACTOR_CODEX_QUEUE_77_DIRECT_LIVE_1 as a new queue-triggered user turn without human input; the message contained the exact run directory, all six events, and the full timeline artifact path once, demonstrating external delivery plus active-turn-to-idle deferral.
