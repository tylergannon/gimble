# danielmiessler/fabric

## Purpose
Go CLI whose entire product is a library of markdown prompts ("patterns") with a small `{{var}}` template language, a raw-print command, and a `--dry-run` that prints what would be sent. Looked at because it is a Go prompt library with a show path, even though patterns are not embedded.

## Pinned
- Repo: https://github.com/danielmiessler/fabric
- Commit: `a8afe606083a1f5321eb1580d574537cef805770` (main)
- License: MIT

## Key concepts
- Layout: `data/patterns/<name>/system.md` (255 patterns at this commit), optional `user.md`. Not embedded: patterns are copied to `~/.config/fabric/patterns` and read from disk — https://raw.githubusercontent.com/danielmiessler/fabric/a8afe606083a1f5321eb1580d574537cef805770/internal/plugins/db/fsdb/db.go L21-25; a custom directory can shadow builtins via `CUSTOM_PATTERNS_DIRECTORY` L59-67.
- Template language is regex-based, not `text/template`: `{{input}}`, `{{name}}` variables, `{{plugin:datetime:now}}`, `{{ext:name:op}}` — `internal/plugins/template/template.go` L34-35; `{{input}}` is protected with a sentinel before variable expansion (`fsdb/patterns.go` L102-125).
- Raw print: `PrintPattern` writes the unrendered pattern (`patterns.go` L174-181).
- Materialized print: `--dry-run` swaps in a fake vendor that formats the final message list (`System:\n...`) instead of sending — `internal/cli/flags.go` L78, `internal/plugins/ai/dryrun/dryrun.go` L50-56. Same message list the real vendor would receive.
- No render-all test, no orphan test; CI only zips `data/patterns/**` into an artifact on change (`.github/workflows/patterns.yaml`).

## Bounded comparison
Like ours in printing the materialized text by substituting the sender, but only for disk-resident patterns with a bespoke placeholder syntax, unlike our embedded `text/template` set.

## Gotchas
- `{{...}}` collides with Go `text/template` delimiters; fabric avoids `text/template` entirely. If our doctrine pages ever contain literal `{{`, `text/template` needs `Delims` or escaping.
- Patterns on disk means version skew between binary and prompt library; embedding avoids this but loses fabric's "edit in place" property.

## Recipe
- To see a dry-run sender that prints the exact request, start at `internal/plugins/ai/dryrun/dryrun.go` L50-110.
