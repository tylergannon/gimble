# Unified model selection worklog

decision: Issue 53 is the accepted contract; model selection is one atomic object and authored provider fields are rejected everywhere they were previously accepted.
decision: Preserve authored model provenance through parsing; resolve effective selections through one shared resolver before runner or harness construction.

friction: The editor production build resolves TypeScript config through the repository root even though `web/editor/.svelte-kit/tsconfig.json` is local; a package-local install was sufficient for `pnpm check` but `pnpm build` failed with `Tsconfig not found astro/tsconfigs/strict` until the root lockfile was provisioned.
lesson: Validate model declarations before ordinary lint so an invalid selection cannot collapse to an empty provider/model during shared-thread analysis; execution, CLI, MCP, and editor all consume the same resolver result.
lesson: Model-role observability needs a dedicated event and role stamped on every event because item judge and goal evaluator share the loop node ID.
