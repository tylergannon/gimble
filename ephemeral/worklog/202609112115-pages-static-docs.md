correction: GitHub Pages cannot host the Go SSR runtime. Restore the prior Astro layout as a separate static documentation build, while retaining the current plain-language Go-library guide copy.
decision: Keep the Go SvelteKit runtime for the future live run UI, but publish only the static Astro output to GitHub Pages.
friction: Astro's root type-check discovered unrelated Go-runtime Playwright files after the docs site returned to the repository root -> exclude e2e and web from the static site's TypeScript project.
friction: main gained a vendored OpenCode oracle while this change was in progress -> exclude third_party from the static docs TypeScript project so Pages validates only its own sources.
