<script lang="ts">
	import { Minus, Moon, Plus, Sun } from '@lucide/svelte';
	import { NODE_TYPES, type NodeType } from '$lib/model';
	import type { Editor } from '$lib/store.svelte';
	import NativeSelect from './NativeSelect.svelte';

	let { editor }: { editor: Editor } = $props();

	const graphName = $derived(editor.graph.name || 'Untitled pipeline');
	const nodeCount = $derived(editor.graph.nodes.length);
	const jumpOptions = $derived([
		{ v: '', l: 'Jump to node…' },
		...editor.graph.nodes.map((n) => ({ v: n.id, l: n.label ? `${n.id} · ${n.label}` : n.id }))
	]);
	const jumpValue = $derived(editor.selNode ? editor.selNode.id : '');
	const addOptions = [{ v: '', l: 'Add node…', disabled: true }, ...NODE_TYPES.map((t) => ({ v: t, l: t }))];
	const zoomPct = $derived(Math.round(editor.zoom * 100) + '%');
</script>

<header>
	<span class="wordmark">tractor</span>
	<span class="divider"></span>
	<span class="name" title={editor.path}>{graphName}</span>
	<span class="count">{nodeCount} nodes</span>
	<span
		class="cn-badge"
		class:cn-badge-variant-destructive={editor.errorCount > 0}
		class:cn-badge-variant-outline={editor.errorCount === 0}
		title={editor.errorCount ? 'Schema and lint issues' : 'No issues'}
	>
		{editor.errorCount} {editor.errorCount === 1 ? 'issue' : 'issues'}
	</span>
	<div class="spacer"></div>
	<div class="jump">
		<NativeSelect size="sm" value={jumpValue} options={jumpOptions} onchange={(v) => v && editor.jumpTo(v)} />
	</div>
	<div class="add">
		<NativeSelect
			size="sm"
			value=""
			disabled={editor.readOnly}
			options={addOptions}
			onchange={(v, el) => {
				if (v) editor.addNode(v as NodeType);
				el.value = '';
			}}
		/>
	</div>
	<div class="zoom">
		<button class="cn-button cn-button-variant-ghost cn-button-size-icon-xs" title="Zoom out" onclick={() => editor.zoomBy(1 / 1.2)}>
			<Minus size={14} />
		</button>
		<span class="pct">{zoomPct}</span>
		<button class="cn-button cn-button-variant-ghost cn-button-size-icon-xs" title="Zoom in" onclick={() => editor.zoomBy(1.2)}>
			<Plus size={14} />
		</button>
	</div>
	<button class="cn-button cn-button-variant-outline cn-button-size-sm" onclick={() => editor.fitView()}>Fit</button>
	<button class="cn-button cn-button-variant-outline cn-button-size-sm" onclick={() => (editor.sel = null)}>Graph</button>
	<button
		class="cn-button cn-button-variant-ghost cn-button-size-icon-sm"
		title={editor.dark ? 'Light mode' : 'Dark mode'}
		onclick={() => editor.toggleDark()}
	>
		{#if editor.dark}<Sun size={16} />{:else}<Moon size={16} />{/if}
	</button>
</header>

<style>
	header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0 16px;
		border-bottom: 1px solid var(--border);
		background: var(--background);
		min-width: 0;
		overflow: hidden;
	}
	.wordmark {
		font-family: var(--font-mono);
		font-weight: 500;
		font-size: 13px;
	}
	.divider {
		width: 1px;
		height: 20px;
		background: var(--border);
	}
	.name {
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
		flex: 0 1 auto;
	}
	.count {
		color: var(--muted-foreground);
		font-size: 13px;
		white-space: nowrap;
	}
	.spacer {
		flex: 1;
	}
	.jump {
		width: 150px;
		flex: none;
	}
	.add {
		width: 120px;
		flex: none;
	}
	.zoom {
		display: flex;
		align-items: center;
		gap: 2px;
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		padding: 2px;
	}
	.pct {
		font-family: var(--font-mono);
		font-size: 12px;
		min-width: 44px;
		text-align: center;
	}
</style>
