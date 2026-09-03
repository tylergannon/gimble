# opencode-ai/opencode (Go, archived)

## Purpose
The archived Go predecessor of crush; looked at as the "before" state: prompts as Go string constants. Useful as the counterexample the crush migration moved away from.

## Pinned
- Repo: https://github.com/opencode-ai/opencode (archived)
- Commit: `73ee493265acf15fcd8caab2bc8cd3bd375b63cb` (main)
- License: MIT

## Key concepts
- Layout: `internal/llm/prompt/{prompt,coder,task,title,summarizer}.go`; no `.md` files, no `embed`.
- Prompt bodies are `const` Go strings selected by provider: https://raw.githubusercontent.com/opencode-ai/opencode/73ee493265acf15fcd8caab2bc8cd3bd375b63cb/internal/llm/prompt/coder.go L16-24 (`CoderPrompt` picks `baseAnthropicCoderPrompt` or `baseOpenAICoderPrompt`, then `fmt.Sprintf` joins env info and LSP info).
- Dispatch by agent name with a fallback literal `"You are a helpful assistant"`: `internal/llm/prompt/prompt.go` L15-39.
- Project context files are appended by string concatenation, cached in a `sync.Once`: `prompt.go` L41-58.
- Only test covers context-file gathering, not prompt text: `internal/llm/prompt/prompt_test.go` L14-40.
- No CLI command prints the prompt.

## Bounded comparison
Unlike ours because every prompt is a Go string with `fmt.Sprintf` composition; it is the exact shape our first promise forbids.

## Gotchas
- Archived; do not cite as current practice. Its successor (crush) moved the same bodies into `.md.tpl` files without changing the composition model (data in, one string out).

## Recipe
- To see what "prompt in a Go string" looks like at scale, start at `internal/llm/prompt/coder.go` L26-170 and compare with crush's `internal/agent/templates/coder.md.tpl`.
