import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// A constant version keeps the committed bundle byte-identical across
			// rebuilds of unchanged source; the default is a build timestamp.
			version: { name: 'tractor' },
			adapter: adapter({
				pages: '../../internal/editor/dist',
				assets: '../../internal/editor/dist',
				fallback: 'index.html',
				precompress: false,
				strict: true
			})
		})
	],
	server: {
		proxy: {
			'/api': { target: 'http://127.0.0.1:7331', changeOrigin: false }
		}
	}
});
