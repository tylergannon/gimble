import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite-plus';
import skgo from './skgo-adapter.js';

export default defineConfig({
	plugins: [
		sveltekit({
			// The build is committed and embedded, so it is not compressed and
			// carries a constant version: rebuilding unchanged source yields
			// byte-identical files, and the tree shows no spurious diff.
			adapter: skgo({ precompress: false }),
			version: { name: 'tractor' },
			experimental: { remoteFunctions: true },
			compilerOptions: {
				experimental: { async: true },
				// Runes mode for the app's own files; libraries keep their default.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			}
		})
	]
});
