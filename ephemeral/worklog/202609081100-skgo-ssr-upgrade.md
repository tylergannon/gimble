# skgo SSR upgrade

decision: Tyler authorized upgrading Tractor to current skgo and rebuilding its SvelteKit integration after upstream issues 105, 106, and 107 closed.
correction: The target release is skgo v0.2.5, not the v0.2.2 evaluated in the prior research; v0.2.5 includes independent Go I/O overlap during SSR.
friction: skgo v0.2.5 still fails its final goja fold for Tractor's emitted shared remote chunk when the app tsconfig excludes build; including build/.goja makes the same fold pass -> filed upstream issue #116 because adapter internals should not constrain the app tsconfig.
friction: Svelte class-field derived runes remain native private fields in the v0.2.5 SSR bundle and panic goja during the first render -> filed upstream issue #117; plain reactive getters preserve Tractor behavior and render safely, but lose derived caching.
