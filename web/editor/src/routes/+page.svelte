<script lang="ts">
	import Canvas from '#lib/components/Canvas.svelte';
	import Header from '#lib/components/Header.svelte';
	import Inspector from '#lib/components/Inspector.svelte';
	import { fromWire } from '#lib/api.ts';
	import { Editor } from '#lib/store.svelte.ts';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	// Route data may change after navigation, but this one-page editor adopts
	// later file changes through its live query. The server value is its seed.
	const initialDocument = () => fromWire(data);
	const editor = new Editor(initialDocument());

	$effect(() => editor.start());

	$effect(() => {
		document.documentElement.classList.toggle('dark', editor.dark);
	});

	function onKey(e: KeyboardEvent) {
		const t = e.target as HTMLElement | null;
		if (t && (/^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName) || t.isContentEditable)) return;
		if ((e.key === 'Delete' || e.key === 'Backspace') && editor.sel) {
			e.preventDefault();
			editor.deleteNode(editor.sel);
		}
	}
</script>

<svelte:window onkeydown={onKey} />

<div class="root">
	<Header {editor} />
	<div class="body">
		<Canvas {editor} />
		<Inspector {editor} />
	</div>
</div>

<style>
	.root {
		position: fixed;
		inset: 0;
		display: grid;
		grid-template-rows: 48px minmax(0, 1fr);
		grid-template-columns: minmax(0, 1fr);
		background: var(--background);
		color: var(--foreground);
		font-family: var(--font-sans);
		font-size: 14px;
		overflow: hidden;
	}
	.body {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 380px;
		min-height: 0;
		min-width: 0;
	}
</style>
