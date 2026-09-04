<script lang="ts">
	import { Plus, Trash2, TriangleAlert, X } from '@lucide/svelte';
	import {
		DEFAULT_KEYS,
		DUR_RE,
		FIELD_META,
		FIELD_SECTIONS,
		REQUIRED,
		TYPE_DESC,
		getPath,
		listEdges,
		normalizeBranch,
		type Defaults,
		type GraphNode
	} from '$lib/model';
	import type { Editor } from '$lib/store.svelte';
	import Field, { type FieldVM } from './Field.svelte';
	import NativeSelect from './NativeSelect.svelte';

	let { editor }: { editor: Editor } = $props();

	const graph = $derived(editor.graph);
	const defaults: Defaults = $derived(graph.defaults ?? {});
	const ids = $derived(graph.nodes.map((n) => n.id));
	const sel: GraphNode | null = $derived(editor.selNode);
	const selErrors = $derived(sel ? (editor.errors[sel.id] ?? []) : []);
	const selIdError = $derived(selErrors.find((e) => /^id |^duplicate/.test(e)) ?? '');

	const known = (x: string) => !x || ids.includes(x) || x === 'success' || x === 'failure';

	function parseValue(kind: string, v: string): string | number | undefined {
		if (v === '') return undefined;
		if (kind === 'int') {
			const n = parseInt(v, 10);
			return Number.isNaN(n) ? undefined : n;
		}
		return v;
	}

	function field(key: string, raw: unknown, required: boolean, inheritFrom: Defaults | null): FieldVM {
		const m = FIELD_META[key] ?? { label: key };
		const kind = m.kind ?? 'text';
		const value = raw === undefined || raw === null ? '' : String(raw);
		const inherited =
			m.inherit && inheritFrom && inheritFrom[key] !== undefined && raw === undefined ? String(inheritFrom[key]) : '';
		let error = '';
		if (kind === 'duration' && value && !DUR_RE.test(value)) error = 'Integer followed by ms, s, m, h, or d';
		const options =
			kind === 'enum'
				? [{ v: '', l: inherited ? `inherit (${inherited})` : '—' }, ...(m.options ?? []).map((o) => ({ v: o, l: o }))]
				: [];
		return {
			key,
			label: m.label,
			value,
			kind: kind === 'textarea' ? 'textarea' : kind === 'enum' ? 'enum' : 'text',
			options,
			rows: m.rows ?? 3,
			placeholder: inherited || m.placeholder || '',
			required,
			inherited,
			error,
			mono: !!(m.mono || kind === 'duration' || kind === 'int')
		};
	}

	// ---- node inspector ----

	const sections = $derived.by(() => {
		if (!sel) return [];
		const req = REQUIRED[sel.type] ?? [];
		return (FIELD_SECTIONS[sel.type] ?? []).map(([title, keys]) => ({
			title,
			fields: keys.map((k) => ({ vm: field(k, sel[k], req.includes(k), defaults), kind: FIELD_META[k]?.kind ?? 'text' }))
		}));
	});

	const targetOptions = $derived([
		{ v: '', l: '— choose target —' },
		...graph.nodes
			.filter((n) => n.type !== 'supervisor')
			.map((n) => ({ v: n.id, l: `${n.id}${n.label ? ' · ' + n.label : ''}` })),
		{ v: 'success', l: 'success' },
		{ v: 'failure', l: 'failure' }
	]);

	interface Transition {
		key: string;
		label: string;
		to: string;
		required: boolean;
		error: string;
	}

	const transitions: Transition[] = $derived.by(() => {
		if (!sel || (sel.type !== 'command' && sel.type !== 'loop')) return [];
		const keys: [string, string, boolean][] =
			sel.type === 'command'
				? [
						['edges.success', 'success', true],
						['edges.error', 'error', false]
					]
				: [
						['edges.loop', 'loop', true],
						['edges.exit', 'exit', true]
					];
		return keys.map(([key, label, required]) => {
			const raw = getPath(sel, key);
			const to = raw === undefined || raw === null ? '' : String(raw);
			return { key, label, to, required, error: !known(to) ? 'Unknown target' : required && !to ? 'Required' : '' };
		});
	});

	const hasTransitions = $derived(!!sel && (sel.type === 'command' || sel.type === 'loop'));
	const hasEdgeList = $derived(!!sel && !hasTransitions && sel.type !== 'supervisor');
	const linksHint = $derived(
		!sel
			? ''
			: sel.type === 'command'
				? 'Exit code decides the edge; no conditions needed.'
				: sel.type === 'loop'
					? 'Loop is the entry node of one lap; exit is where the evaluator sends the walk.'
					: sel.type === 'fan_out'
						? 'Branch edges leave the fan-out once every branch has finished. Leave the condition blank for an unconditional edge.'
						: 'The agent picks one edge by its condition. Leave the condition blank for an unconditional edge.'
	);
	const edges = $derived(sel && hasEdgeList ? listEdges(sel) : []);

	const superviseOptions = $derived(
		sel && sel.type === 'supervisor'
			? graph.nodes
					.filter((n) => n.id !== sel.id)
					.map((n) => ({ id: n.id, label: n.label ?? '', checked: (sel.supervises ?? []).includes(n.id) }))
			: []
	);

	const branches = $derived(sel && sel.type === 'fan_out' ? (sel.branches ?? []).map(normalizeBranch) : []);

	function toggleSupervise(id: string, on: boolean) {
		if (!sel) return;
		const cur = sel.supervises ?? [];
		editor.setSupervises(sel.id, on ? [...cur.filter((x) => x !== id), id] : cur.filter((x) => x !== id));
	}

	// ---- graph settings ----

	const startOptions = $derived([
		{ v: '', l: '—' },
		...graph.nodes.filter((n) => n.type !== 'supervisor').map((n) => ({ v: n.id, l: n.id }))
	]);
	const startError = $derived(
		graph.start && !ids.includes(graph.start) ? 'Start node does not exist' : !graph.start ? 'Start is required' : ''
	);
	const defaultFields = $derived(
		DEFAULT_KEYS.map((k) => {
			const f = field(k, defaults[k], false, null);
			if (f.kind === 'enum') f.options[0] = { v: '', l: '—' };
			return { vm: f, kind: FIELD_META[k]?.kind ?? 'text' };
		})
	);
	const nameField: FieldVM = $derived({
		key: 'name',
		label: 'Name',
		value: graph.name ?? '',
		kind: 'text',
		options: [],
		rows: 1,
		placeholder: '',
		required: false,
		inherited: '',
		error: '',
		mono: false
	});
	const goalField: FieldVM = $derived({
		key: 'goal',
		label: 'Goal',
		value: graph.goal ?? '',
		kind: 'textarea',
		options: [],
		rows: 4,
		placeholder: 'Objective exposed to prompt expansion',
		required: false,
		inherited: '',
		error: '',
		mono: false
	});
</script>

<aside>
	{#if !sel}
		<div class="panel">
			<div>
				<div class="title">Graph settings</div>
				<div class="subtitle">Pipeline metadata and file-level defaults.</div>
			</div>
			{#if editor.graphIssues.length}
				<div class="cn-alert cn-alert-variant-destructive">
					<TriangleAlert />
					<div class="cn-alert-title">Pipeline issues</div>
					<div class="cn-alert-description">
						{#each editor.graphIssues as e, i (i)}<div>{e}</div>{/each}
					</div>
				</div>
			{/if}
			<div class="section">
				<div class="section-title">Pipeline</div>
				<Field field={nameField} onvalue={(v) => editor.setGraphField('name', v)} />
				<Field field={goalField} onvalue={(v) => editor.setGraphField('goal', v)} />
				<div class="field">
					<label class="cn-label" for="field-start">Start node</label>
					<NativeSelect value={graph.start ?? ''} options={startOptions} onchange={(v) => editor.setGraphField('start', v)} />
					{#if startError}<div class="field-error">{startError}</div>{/if}
				</div>
			</div>
			<div class="section">
				<div class="section-title">Defaults</div>
				{#each defaultFields as f (f.vm.key)}
					<Field field={f.vm} showRequired={false} onvalue={(v) => editor.setDefault(f.vm.key, parseValue(f.kind, v))} />
				{/each}
			</div>
		</div>
	{:else}
		{@const node = sel}
		<div class="panel">
			<div class="head">
				<div class="head-main">
					<div class="type-row">
						<span class="type-badge">{node.type}</span>
						<span class="type-desc">{TYPE_DESC[node.type] ?? ''}</span>
					</div>
					<div class="field id-field">
						<label class="cn-label" for="field-id">id</label>
						<input
							id="field-id"
							class="cn-input mono"
							value={node.id}
							onchange={(e) => editor.renameNode(node.id, e.currentTarget.value)}
						/>
						{#if selIdError}<div class="field-error">{selIdError}</div>{/if}
					</div>
				</div>
				<button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm" title="Delete node" onclick={() => editor.deleteNode(node.id)}>
					<Trash2 size={16} />
				</button>
			</div>

			{#if selErrors.length}
				<div class="cn-alert cn-alert-variant-destructive">
					<TriangleAlert />
					<div class="cn-alert-title">Schema issues</div>
					<div class="cn-alert-description">
						{#each selErrors as e, i (i)}<div>{e}</div>{/each}
					</div>
				</div>
			{/if}

			<div class="start-row">
				<button
					type="button"
					class="cn-switch"
					role="switch"
					aria-label="Start node"
					aria-checked={graph.start === node.id}
					data-checked={graph.start === node.id ? '' : undefined}
					onclick={() => editor.setStart(node.id, graph.start !== node.id)}
				>
					<span class="cn-switch-thumb"></span>
				</button>
				<span>Start node</span>
			</div>

			{#each sections as s (s.title)}
				<div class="section">
					<div class="section-title">{s.title}</div>
					{#each s.fields as f (f.vm.key)}
						<Field field={f.vm} onvalue={(v) => editor.setNodeField(node.id, f.vm.key, parseValue(f.kind, v))} />
					{/each}
				</div>
			{/each}

			{#if hasTransitions}
				<div class="section">
					<div class="section-head"><span class="section-title">Transitions</span></div>
					<div class="links-hint">{linksHint}</div>
					{#each transitions as t (t.key)}
						<div class="link-card">
							<div class="link-main">
								<div class="link-row">
									<span class="link-label">{t.label}</span>
									<NativeSelect
										size="sm"
										value={t.to}
										options={targetOptions}
										onchange={(v) => editor.setNodeField(node.id, t.key, v || (t.required ? '' : undefined))}
									/>
								</div>
								{#if t.error}<div class="field-error">{t.error}</div>{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}

			{#if hasEdgeList}
				<div class="section">
					<div class="section-head">
						<span class="section-title">Outgoing edges</span>
						<button class="cn-button cn-button-variant-outline cn-button-size-xs" onclick={() => editor.addEdge(node.id)}>
							<Plus size={12} />Add edge
						</button>
					</div>
					<div class="links-hint">{linksHint}</div>
					{#each edges as e, i (i)}
						<div class="link-card">
							<div class="link-main">
								<div class="link-row">
									<span class="link-label">edge {i + 1}</span>
									<NativeSelect size="sm" value={e.to ?? ''} options={targetOptions} onchange={(v) => editor.setEdge(node.id, i, { to: v })} />
								</div>
								<input
									class="cn-input cn-input-size-sm"
									value={e.condition ?? ''}
									placeholder="Condition — why the agent should take this edge"
									oninput={(ev) => editor.setEdge(node.id, i, { condition: ev.currentTarget.value })}
								/>
								{#if !known(e.to ?? '')}<div class="field-error">Unknown target</div>{/if}
							</div>
							<button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm" title="Remove edge" onclick={() => editor.removeEdge(node.id, i)}>
								<X size={14} />
							</button>
						</div>
					{/each}
				</div>
			{/if}

			{#if node.type === 'supervisor'}
				<div class="section supervises">
					<div class="section-title">Supervises</div>
					{#each superviseOptions as o (o.id)}
						<label class="check-row">
							<button
								type="button"
								class="cn-checkbox"
								role="checkbox"
								aria-label={`Supervise ${o.id}`}
								aria-checked={o.checked}
								data-checked={o.checked ? '' : undefined}
								onclick={() => toggleSupervise(o.id, !o.checked)}
							>
								{#if o.checked}<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>{/if}
							</button>
							<span class="check-id">{o.id}</span><span class="check-label">{o.label}</span>
						</label>
					{/each}
				</div>
			{/if}

			{#if node.type === 'fan_out'}
				<div class="section">
					<div class="section-head">
						<span class="section-title">Branches</span>
						<button class="cn-button cn-button-variant-outline cn-button-size-xs" onclick={() => editor.addBranch(node.id)}>
							<Plus size={12} />Add branch
						</button>
					</div>
					{#each branches as b, i (i)}
						<div class="link-card">
							<div class="link-main">
								<input
									class="cn-input cn-input-size-sm mono-sm"
									value={b.id ?? ''}
									placeholder="branch id"
									onchange={(e) => editor.setBranch(node.id, i, { id: e.currentTarget.value })}
								/>
								<input
									class="cn-input cn-input-size-sm mono-sm"
									value={(b.artifacts ?? []).join(', ')}
									placeholder="artifacts, comma separated"
									onchange={(e) =>
										editor.setBranch(node.id, i, {
											artifacts: e.currentTarget.value
												.split(',')
												.map((s) => s.trim())
												.filter(Boolean)
										})}
								/>
								<textarea
									class="cn-textarea branch-prompt"
									rows="3"
									value={b.agent?.prompt ?? ''}
									placeholder="Prompt override (optional)"
									oninput={(e) => editor.setBranch(node.id, i, { prompt: e.currentTarget.value })}
								></textarea>
							</div>
							<button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm" title="Remove branch" onclick={() => editor.removeBranch(node.id, i)}>
								<X size={14} />
							</button>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</aside>

<style>
	aside {
		border-left: 1px solid var(--border);
		overflow: auto;
		background: var(--background);
		min-height: 0;
	}
	.panel {
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	.title {
		font-weight: 600;
		font-size: 16px;
		letter-spacing: -0.015em;
	}
	.subtitle {
		color: var(--muted-foreground);
		font-size: 13px;
		margin-top: 2px;
	}
	.section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.section.supervises {
		gap: 10px;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.head {
		display: flex;
		align-items: flex-start;
		gap: 10px;
	}
	.head-main {
		flex: 1;
		min-width: 0;
	}
	.type-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.type-desc {
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.id-field {
		margin-top: 8px;
	}
	.start-row {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
	}
	.start-row span:last-child {
		white-space: nowrap;
	}
	.links-hint {
		font-size: 12px;
		color: var(--muted-foreground);
		margin-top: -6px;
	}
	.link-card {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 8px;
		padding: 10px;
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		background: var(--card);
	}
	.link-main {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
	}
	.link-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.link-label {
		font-size: 12px;
		color: var(--muted-foreground);
		min-width: 52px;
	}
	.check-row {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13px;
		cursor: pointer;
	}
	.check-id {
		font-family: var(--font-mono);
		font-size: 12px;
	}
	.check-label {
		color: var(--muted-foreground);
	}
	.mono-sm {
		font-family: var(--font-mono);
		font-size: 12px;
	}
	.branch-prompt {
		font-size: 13px;
		resize: vertical;
	}
	.cn-alert-description div + div {
		margin-top: 2px;
	}
</style>
