<script lang="ts">
	import { TriangleAlert, X } from '@lucide/svelte';
	import { isBox } from '$lib/layout';
	import { isTerminal } from '$lib/model';
	import { buildScene } from '$lib/scene';
	import type { Editor } from '$lib/store.svelte';

	let { editor }: { editor: Editor } = $props();

	let viewport: HTMLDivElement | undefined = $state();

	const scene = $derived(
		buildScene(
			editor.graph,
			editor.layout,
			editor.sel,
			editor.errors,
			{ pan: editor.pan, zoom: editor.zoom, vw: editor.vw, vh: editor.vh },
			editor.showInherited
		)
	);

	const xform = $derived(`translate(${editor.pan.x}px, ${editor.pan.y}px) scale(${editor.zoom})`);
	const gridSize = $derived(`${24 * editor.zoom}px ${24 * editor.zoom}px`);
	const gridRepeat = $derived(editor.zoom < 0.4 ? 'no-repeat' : 'repeat');
	const gridPos = $derived(`${editor.pan.x}px ${editor.pan.y}px`);
	const bgCursor = $derived(editor.drag?.type === 'pan' ? 'grabbing' : 'grab');

	// Non-passive wheel listener and viewport size tracking.
	$effect(() => {
		const vp = viewport;
		if (!vp) return;
		const onWheel = (e: WheelEvent) => {
			e.preventDefault();
			if (e.ctrlKey || e.metaKey) {
				const r = vp.getBoundingClientRect();
				editor.zoomAt(e.clientX - r.left, e.clientY - r.top, editor.zoom * Math.exp(-e.deltaY * 0.0015));
			} else {
				editor.pan = { x: editor.pan.x - e.deltaX, y: editor.pan.y - e.deltaY };
			}
		};
		vp.addEventListener('wheel', onWheel, { passive: false });
		const ro = new ResizeObserver(() => {
			editor.vw = vp.clientWidth;
			editor.vh = vp.clientHeight;
		});
		ro.observe(vp);
		editor.vw = vp.clientWidth;
		editor.vh = vp.clientHeight;
		return () => {
			vp.removeEventListener('wheel', onWheel);
			ro.disconnect();
		};
	});

	// Fit the view once the first document has loaded and the viewport is measured.
	$effect(() => {
		if (editor.needsFit && !editor.loading && editor.vw > 0) {
			editor.fitView();
			editor.needsFit = false;
		}
	});

	function onBgDown(e: MouseEvent) {
		if (e.button !== 0) return;
		editor.sel = null;
		editor.drag = { type: 'pan', sx: e.clientX, sy: e.clientY, ox: editor.pan.x, oy: editor.pan.y };
	}

	function downMove(id: string) {
		return (e: MouseEvent) => {
			if (e.button !== 0) return;
			e.stopPropagation();
			const L = editor.layout[id];
			editor.sel = id;
			editor.drag = { type: 'move', id, sx: e.clientX, sy: e.clientY, ox: L?.x ?? 0, oy: L?.y ?? 0 };
		};
	}

	function downResize(id: string) {
		return (e: MouseEvent) => {
			if (e.button !== 0) return;
			e.stopPropagation();
			const L = editor.layout[id];
			editor.sel = id;
			editor.drag = {
				type: 'resize',
				id,
				sx: e.clientX,
				sy: e.clientY,
				ow: isBox(L) ? L.w : 240,
				oh: isBox(L) ? L.h : 120
			};
		};
	}

	function downHandle(key: string, x: number, y: number) {
		return (e: MouseEvent) => {
			if (e.button !== 0) return;
			e.stopPropagation();
			editor.drag = { type: 'edge', key, sx: e.clientX, sy: e.clientY, ox: x, oy: y };
		};
	}

	function onMove(e: MouseEvent) {
		const d = editor.drag;
		if (!d) return;
		const zoom = editor.zoom;
		if (d.type === 'pan') {
			editor.pan = { x: d.ox + e.clientX - d.sx, y: d.oy + e.clientY - d.sy };
			return;
		}
		if (d.type === 'edge') {
			editor.setLayout(d.key, { x: d.ox + (e.clientX - d.sx) / zoom, y: d.oy + (e.clientY - d.sy) / zoom });
			return;
		}
		const cur = editor.layout[d.id];
		const base = isBox(cur) ? cur : { x: 0, y: 0, w: isTerminal(d.id) ? 150 : 240, h: isTerminal(d.id) ? 56 : 120 };
		if (d.type === 'move') {
			const snap = editor.snap ? (v: number) => Math.round(v / 24) * 24 : (v: number) => v;
			editor.setLayout(d.id, {
				...base,
				x: snap(d.ox + (e.clientX - d.sx) / zoom),
				y: snap(d.oy + (e.clientY - d.sy) / zoom)
			});
		} else if (d.type === 'resize') {
			editor.setLayout(d.id, {
				...base,
				w: Math.max(isTerminal(d.id) ? 96 : 160, d.ow + (e.clientX - d.sx) / zoom),
				h: Math.max(isTerminal(d.id) ? 36 : 84, d.oh + (e.clientY - d.sy) / zoom)
			});
		}
	}

	function onUp() {
		editor.drag = null;
	}

	function onMiniDown(e: MouseEvent) {
		e.stopPropagation();
		const r = e.currentTarget as HTMLElement;
		const rect = r.getBoundingClientRect();
		const B = scene.mini.bounds;
		const s = scene.mini.scale;
		const offX = (180 - B.w / s) / 2;
		const offY = (112 - B.h / s) / 2;
		const gx = B.x + (e.clientX - rect.left - offX) * s;
		const gy = B.y + (e.clientY - rect.top - offY) * s;
		editor.pan = { x: editor.vw / 2 - gx * editor.zoom, y: editor.vh / 2 - gy * editor.zoom };
	}
</script>

<svelte:window onmousemove={onMove} onmouseup={onUp} />

<div
	class="viewport"
	bind:this={viewport}
	role="presentation"
	onmousedown={onBgDown}
	style:background-size={gridSize}
	style:background-repeat={gridRepeat}
	style:background-position={gridPos}
	style:cursor={bgCursor}
>
	<div class="world" style:transform={xform}>
		<svg width="1" height="1" class="edges">
			<defs>
				<marker id="arw" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
					<path d="M0 0L10 5L0 10z" style="fill:var(--muted-foreground)"></path>
				</marker>
				<marker id="arw-sup" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
					<path d="M0 0L10 5L0 10z" style="fill:var(--chart-2)"></path>
				</marker>
				<marker id="arw-in" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
					<path d="M0 0L10 5L0 10z" style="fill:var(--chart-3)"></path>
				</marker>
				<marker id="arw-sel" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
					<path d="M0 0L10 5L0 10z" style="fill:var(--foreground)"></path>
				</marker>
			</defs>
			{#each scene.paths as p (p.key)}
				<path d={p.d} fill="none" stroke={p.stroke} stroke-width={p.width} stroke-dasharray={p.dash} marker-end={p.marker}></path>
			{/each}
		</svg>

		{#each scene.terminals as t (t.id)}
			<div
				class="terminal"
				role="presentation"
				onmousedown={downMove(t.id)}
				style:left="{t.x}px"
				style:top="{t.y}px"
				style:width="{t.w}px"
				style:height="{t.h}px"
				style:box-shadow={t.shadow}
				style:color={t.color}
			>
				<span class="dot" style:background={t.dot}></span>{t.label}
				<div
					class="resize terminal-resize"
					role="presentation"
					onmousedown={downResize(t.id)}
					style:border-color={t.handleColor}
				></div>
			</div>
		{/each}

		{#each scene.nodes as n (n.id)}
			<div
				class="node"
				role="presentation"
				onmousedown={downMove(n.id)}
				style:left="{n.x}px"
				style:top="{n.y}px"
				style:width="{n.w}px"
				style:height="{n.h}px"
				style:background={n.bg}
				style:border-radius={n.radius}
				style:border={n.border}
				style:box-shadow={n.shadow}
			>
				<div class="node-head">
					<span class="badge" style:border-color={n.badgeBorder} style:color={n.badgeColor}>{n.type}</span>
					<div class="spacer"></div>
					{#if n.hasErrors}
						<span class="warn tip" data-tip={n.errorText}>
							<TriangleAlert size={15} />
						</span>
					{/if}
				</div>
				<div class="node-body">
					<div class="label">{n.label}</div>
					<div class="id">{n.id}</div>
				</div>
				<div class="meta">
					{#each n.meta as m (m.k)}
						<span class="chip"><span class="k">{m.k}</span><span class="v" style:color={m.color}>{m.v}</span></span>
					{/each}
				</div>
				<div class="resize node-resize" role="presentation" onmousedown={downResize(n.id)} style:border-color={n.handleColor}></div>
			</div>
		{/each}

		{#each scene.handles as h (h.key)}
			<div
				class="handle"
				role="presentation"
				title="Drag to reshape edge"
				onmousedown={downHandle(h.key, h.x, h.y)}
				style:left="{h.x}px"
				style:top="{h.y}px"
				style:border-color={h.color}
			></div>
		{/each}

		{#each scene.labels as l (l.key)}
			<div
				class="edge-label"
				title={l.full}
				style:left="{l.x}px"
				style:top="{l.y}px"
				style:max-width="{l.maxW}px"
				style:border-color={l.border}
				style:color={l.color}
			>
				{l.text}
			</div>
		{/each}
	</div>

	{#if editor.banner}
		<div class="banner cn-alert cn-alert-variant-default" role="status">
			<span>{editor.banner}</span>
			<button class="cn-button cn-button-variant-ghost cn-button-size-icon-xs" title="Dismiss" onclick={() => (editor.banner = '')}>
				<X size={14} />
			</button>
		</div>
	{/if}

	<div class="minimap" role="presentation" onmousedown={onMiniDown}>
		<svg width="180" height="112" viewBox={scene.mini.box} preserveAspectRatio="xMidYMid meet">
			{#each scene.mini.nodes as r (r.id)}
				<rect x={r.x} y={r.y} width={r.w} height={r.h} rx={r.rx} style:fill={r.fill}></rect>
			{/each}
			<rect
				class="mini-view"
				x={scene.mini.view.x}
				y={scene.mini.view.y}
				width={scene.mini.view.w}
				height={scene.mini.view.h}
				style:stroke-width={scene.mini.view.sw}
			></rect>
		</svg>
	</div>

	<div class="hint">
		<span>Drag canvas to pan · scroll to pan · ⌘/Ctrl+scroll to zoom</span>
		<label class="snap">
			<button
				type="button"
				class="cn-checkbox"
				role="checkbox"
				aria-checked={editor.snap}
				data-checked={editor.snap ? '' : undefined}
				onmousedown={(e) => e.stopPropagation()}
				onclick={() => editor.setSnap(!editor.snap)}
			>
				{#if editor.snap}<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>{/if}
			</button>
			<span>Snap to grid</span>
		</label>
	</div>
</div>

<style>
	.viewport {
		position: relative;
		overflow: hidden;
		background-color: var(--surface);
		background-image: radial-gradient(color-mix(in oklch, var(--foreground) 14%, transparent) 1px, transparent 1px);
		user-select: none;
		min-width: 0;
		min-height: 0;
	}
	.world {
		position: absolute;
		left: 0;
		top: 0;
		transform-origin: 0 0;
	}
	.edges {
		position: absolute;
		left: 0;
		top: 0;
		overflow: visible;
		pointer-events: none;
	}
	.terminal {
		position: absolute;
		border-radius: 999px;
		font-family: var(--font-mono);
		font-size: 12px;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		background: var(--background);
		cursor: move;
		overflow: hidden;
	}
	.dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
	}
	.resize {
		position: absolute;
		right: 0;
		bottom: 0;
		cursor: nwse-resize;
		border-right: 2px solid;
		border-bottom: 2px solid;
	}
	.terminal-resize {
		right: 10px;
		width: 12px;
		height: 12px;
		border-radius: 0 0 6px 0;
		margin: 4px;
	}
	.node-resize {
		width: 14px;
		height: 14px;
		border-radius: 0 0 10px 0;
		margin: 3px;
	}
	.node {
		position: absolute;
		color: var(--card-foreground);
		cursor: move;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}
	.node-head {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px 0 12px;
	}
	.badge {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 500;
		padding: 1px 7px;
		border-radius: 999px;
		border: 1px solid;
		white-space: nowrap;
	}
	.spacer {
		flex: 1;
	}
	.warn {
		display: inline-flex;
		color: var(--destructive);
		position: relative;
	}
	.tip::after {
		content: attr(data-tip);
		position: absolute;
		bottom: calc(100% + 6px);
		right: -6px;
		display: none;
		white-space: pre-line;
		max-width: 280px;
		width: max-content;
		border-radius: var(--radius-md);
		padding: 6px 10px;
		font-family: var(--font-sans);
		font-size: var(--text-xs);
		line-height: 1.4;
		background: var(--primary);
		color: var(--primary-foreground);
		box-shadow: var(--shadow-md);
		z-index: 5;
		text-align: left;
	}
	.tip:hover::after {
		display: block;
	}
	.node-body {
		padding: 8px 12px 0 12px;
		min-height: 0;
	}
	.label {
		font-weight: 500;
		font-size: 14px;
		line-height: 1.3;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.id {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--muted-foreground);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 10px;
		padding: 8px 12px 10px 12px;
		margin-top: auto;
		font-size: 11px;
		overflow: hidden;
	}
	.chip {
		display: inline-flex;
		gap: 4px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 100%;
	}
	.chip .k {
		color: var(--muted-foreground);
	}
	.chip .v {
		font-family: var(--font-mono);
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.handle {
		position: absolute;
		width: 12px;
		height: 12px;
		margin: -6px 0 0 -6px;
		border-radius: 50%;
		background: var(--background);
		border: 2px solid;
		cursor: grab;
		z-index: 2;
	}
	.edge-label {
		position: absolute;
		transform: translate(-50%, -50%);
		padding: 1px 7px;
		border-radius: 999px;
		background: var(--background);
		border: 1px solid;
		font-size: 11px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		pointer-events: auto;
		cursor: default;
	}
	.banner {
		position: absolute;
		top: 12px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px 6px 12px;
		box-shadow: var(--shadow-md);
		max-width: calc(100% - 32px);
		z-index: 6;
	}
	.minimap {
		position: absolute;
		right: 16px;
		bottom: 16px;
		width: 180px;
		height: 112px;
		background: var(--background);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
		overflow: hidden;
		cursor: crosshair;
	}
	.minimap svg {
		display: block;
	}
	.mini-view {
		fill: color-mix(in oklch, var(--foreground) 6%, transparent);
		stroke: var(--foreground);
	}
	.hint {
		position: absolute;
		left: 16px;
		bottom: 16px;
		display: flex;
		align-items: center;
		gap: 14px;
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.hint > span {
		pointer-events: none;
	}
	.snap {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		cursor: pointer;
	}
</style>
