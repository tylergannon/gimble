# Gimble rename completion

decision: User confirmed that completing the rename includes a versioned public Go module release; ship v0.10.0 rather than treating the module-path change as a v1 requirement.
correction: The mascot was committed during the rename but not surfaced on the public docs landing page; make it visible in README and docs.
decision: Retain Tractor-named cache, executable, marketplace, and test values only as explicit upgrade cleanup. Remaining references in the read-only upstream submodule and a historical test comment are not product identity.
friction: A fresh release worktree has no Node dependencies; run pnpm install --frozen-lockfile before docs verification.
