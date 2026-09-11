import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite-plus';
import skgo from '@skgo/sveltekit-adapter';

export default defineConfig({
	plugins: [
		sveltekit({
			adapter: skgo({ out: '../internal/webembed/build' }),
			experimental: { remoteFunctions: true },
			compilerOptions: { experimental: { async: true } }
		})
	]
});
