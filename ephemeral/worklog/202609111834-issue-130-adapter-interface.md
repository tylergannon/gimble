# Issue 130 adapter interface worklog

decision: This assignment is design-only until the manager settles the root callback; adapter implementation and live provider runs remain deferred.

decision: Adapters should submit exact pinned session event type/data inputs; root must allocate envelope IDs/timestamps/durable sequence, normalized session/message IDs, placement, and execution lifecycle.

discovery: Codex app-server preserves raw notification params but has no explicit per-model-call identity/boundary, monetary cost, or per-call finish reason. Its summary and raw reasoning deltas cannot both map to the single visible OpenCode reasoning stream.

discovery: The installed Claude SDK supports partial messages, but the default pre-adapter typed parser drops nested tool result identity, nested assistant message identity/model, thinking content, and full raw usage. A lossless pre-ParseMessage observation path is required.

decision: Vendored the exact installed Claude SDK source under third_party, excluding only its packaged Go build cache, and added a synchronous owned-byte observer before ParseMessage with error/close tests.

discovery: Codex thread/tokenUsage/updated is not a model-call identity. Missing usage can repeat stale last values, rate-limit updates can emit unchanged values without a completion, and compaction emits API usage plus a later local estimate. rawResponse/completed carries the native response ID and is the source-backed boundary alternative.

correction: NativeRef is one flat current sidecar using provider/sessionID/turnID/messageID/responseID/itemID/parentToolUseID, root-injected normalizedMessageID, and accounting availability; unmapped native input is written to a separate audit log and never enters AgentRecord or observation.

correction: tokenUsage presence cannot gate Codex step settlement. rawResponse/completed plus closed parts and settled tools ends the step; tokenUsage remains audit-only, including when it is late or stale.

decision: Root derives its accepted ingress registry from pinned session events plus permission.asked/replied and form.created/replied/cancelled. form.created placement is form.sessionID; unsupported global definitions remain excluded.

friction: canonical just build reached the adapter after both browser compilations but failed because another owned route exports a universal +page.ts load -> runtime/frontend owner must replace it with the Go load already planned; adapter/root changes require no workaround.
