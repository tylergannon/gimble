friction: GitHub Actions ran only after pushes to main, so formatting failures were discovered after merge -> run the same checks on pull requests.
decision: Keep the docs site, standalone editor source, and generated embedded editor bundle in separate verification scopes; prove the editor bundle is reproducible instead of formatting generated output.
