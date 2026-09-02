# nlspec Pseudocode Notation

Language-neutral notation for the technical artifacts of an nlspec. Exact
examples of a specified language, schema, or wire format are allowed when the
public syntax must be shown exactly.

## Record

Describes a data type definition.

```
RECORD Session:
    id                : String                  -- UUID, assigned at creation
    provider_profile  : ProviderProfile         -- tools + system prompt for the active model
    execution_env     : ExecutionEnvironment    -- where tools run
    history           : List<Turn>              -- ordered conversation turns
    event_emitter     : EventEmitter            -- delivers events to host application
    config            : SessionConfig           -- limits, timeouts, settings
    state             : SessionState            -- current lifecycle state
    llm_client        : Client                  -- from the Unified LLM SDK
    steering_queue    : Queue<String>           -- messages to inject between tool rounds
    followup_queue    : Queue<String>           -- messages to process after current input completes
    subagents         : Map<String, SubAgent>   -- active child agents
```

## Enum

```
ENUM SessionState:
    IDLE              -- waiting for user input
    PROCESSING        -- running the agentic loop
    AWAITING_INPUT    -- model asked the user a question
    CLOSED            -- session terminated (normal or error)
```

## Interface

Preferred way to define seams. Should include type definitions for
user-defined types.

```
RECORD ToolDefinition:
    name        : String            -- unique identifier
    description : String            -- for the LLM
    parameters  : Dict              -- JSON Schema (root must be "object")

RECORD RegisteredTool:
    definition  : ToolDefinition
    executor    : Function          -- (arguments, execution_env) -> String

RECORD ToolRegistry:
    _tools      : Map<String, RegisteredTool>

    register(tool)                  -- add or replace a tool
    unregister(name)                -- remove a tool
    get(name) -> RegisteredTool | None
    definitions() -> List<ToolDefinition>
    names() -> List<String>

INTERFACE ProviderProfile:
    id              : String            -- "openai", "anthropic", "gemini"
    model           : String            -- model identifier (e.g., "gpt-5.2-codex")
    tool_registry   : ToolRegistry      -- all tools available to this profile

    FUNCTION build_system_prompt(environment, project_docs) -> String
    FUNCTION tools() -> List<ToolDefinition>
    FUNCTION provider_options() -> Map | None

    -- Capability flags
    supports_reasoning           : Boolean
    supports_streaming           : Boolean
    supports_parallel_tool_calls : Boolean
    context_window_size          : Integer
```

## State Machine

Coupled with a state enum, to define the allowed transitions.

```
IDLE -> PROCESSING          -- on submit()
PROCESSING -> PROCESSING    -- tool loop continues
PROCESSING -> AWAITING_INPUT -- model asks user a question (no tool calls, open-ended)
PROCESSING -> IDLE          -- natural completion or turn limit
PROCESSING -> CLOSED        -- unrecoverable error
IDLE -> CLOSED              -- explicit close()
any -> CLOSED               -- abort signal (after graceful shutdown cleanup)
AWAITING_INPUT -> PROCESSING -- user provides answer
```

## Function

A pseudocode algorithm definition. Use only for core behavior whose ordering,
precedence, or edge cases would be ambiguous in prose. It defines required
semantics, not a prescribed implementation.

```
FUNCTION check_context_usage(session):
    approx_tokens = total_chars_in_history(session.history) / 4
    threshold = session.provider_profile.context_window_size * 0.8
    IF approx_tokens > threshold:
        session.emit(WARNING, message = "Context usage at ~"
            + ROUND(approx_tokens / session.provider_profile.context_window_size * 100)
            + "% of context window")
```

## Service Definition

Rare; only when the algorithms shown are non-obvious but simple and/or a prose
description would be more verbose.

```
HandlerRegistry:
    handlers        : Map<String, Handler>   -- type string -> handler instance
    default_handler : Handler                -- fallback handler (typically codergen)

    FUNCTION register(type_string, handler):
        handlers[type_string] = handler
        -- Registering for an already-registered type replaces the previous handler

    FUNCTION resolve(node) -> Handler:
        -- 1. Explicit type attribute
        IF node.type is not empty AND node.type IN handlers:
            RETURN handlers[node.type]

        -- 2. Shape-based resolution
        handler_type = SHAPE_TO_TYPE[node.shape]
        IF handler_type IN handlers:
            RETURN handlers[handler_type]

        -- 3. Default
        RETURN default_handler
```

## Slice-design artifacts

Used in ephemeral slice specs (see slice-design), not normally in the nlspec
itself.

**Call-stack tree** — for orchestration or control-flow changes; diff syntax
when the interesting part is what changes:

```diff
 entrypoint
   runCommand
+    handleCreateResource
+      ResourceClient.create(input)
+        POST /resources
+      renderResult
-    legacyCreateFlow
```

**File-tree diff** — the projection of the seam model onto disk:

```diff
 src
 └── resource
+    ├── resource-client.ts      # NEW - wraps API contract calls
+    ├── resource-client.test.ts # NEW - covers request/response mapping
~    └── resource-route.ts       # MODIFIED - wires create action into UI
```
