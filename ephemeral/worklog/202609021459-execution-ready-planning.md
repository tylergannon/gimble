# Execution-ready planning artifacts

decision: Parse and validate recommendation.md before applying checklist cardinality and nesting rules, because an empty top-level checklist is valid only for SIMPLE while MEDIUM and LARGE require multiple items.

decision: Resolve LARGE supporting paths both lexically and through symbolic links before enforcing project-root containment and uniqueness, so generated chapter artifacts cannot escape or alias one another.

friction: The first commit attempt was intentionally stopped after the pre-commit hook applied go-modernize fixes -> rerun all required gates on the hook-modified tree before retrying the commit.
