// Canvas geometry: edge paths, labels, edge handles, node cards, terminals,
// and the minimap, computed from the graph and the layout.

import { boundsOf, edgeKey, isBox, type Box, type Layout, type LayoutEntry } from './layout';
import { isTerminal, linksOf, listEdges, normalizeBranch, TERMINAL, type Defaults, type Graph } from './model';
import type { ModelResolution } from './api';

export interface PathVM {
	key: string;
	d: string;
	stroke: string;
	width: number;
	dash: string;
	marker: string;
}

export interface LabelVM {
	key: string;
	maxW: number;
	x: number;
	y: number;
	text: string;
	full: string;
	color: string;
	border: string;
}

export interface HandleVM {
	key: string;
	x: number;
	y: number;
	color: string;
	dashed: boolean;
}

export interface TerminalVM {
	id: string;
	label: string;
	x: number;
	y: number;
	w: number;
	h: number;
	shadow: string;
	color: string;
	dot: string;
	handleColor: string;
}

export interface MetaVM {
	k: string;
	v: string;
	color: string;
}

export interface NodeVM {
	id: string;
	type: string;
	label: string;
	isStart: boolean;
	hasErrors: boolean;
	errorText: string;
	x: number;
	y: number;
	w: number;
	h: number;
	shadow: string;
	radius: string;
	bg: string;
	border: string;
	badgeColor: string;
	badgeBorder: string;
	meta: MetaVM[];
	handleColor: string;
}

export interface MiniVM {
	box: string;
	nodes: { id: string; x: number; y: number; w: number; h: number; rx: number; fill: string }[];
	view: { x: number; y: number; w: number; h: number; sw: number };
	bounds: Box;
	scale: number;
}

export interface Scene {
	paths: PathVM[];
	labels: LabelVM[];
	handles: HandleVM[];
	terminals: TerminalVM[];
	nodes: NodeVM[];
	mini: MiniVM;
}

export interface Viewport {
	pan: { x: number; y: number };
	zoom: number;
	vw: number;
	vh: number;
}

const FG = 'var(--foreground)';
const MUTED = 'var(--muted-foreground)';
const SUP = 'var(--chart-2)';
const IN = 'var(--chart-3)';

// The point on box B's edge that faces point P.
function anchor(b: Box, p: { x: number; y: number }): [number, number] {
	const cx = b.x + b.w / 2;
	const cy = b.y + b.h / 2;
	if (p.x < b.x) return [b.x, Math.max(b.y + 12, Math.min(b.y + b.h - 12, p.y))];
	if (p.x > b.x + b.w) return [b.x + b.w, Math.max(b.y + 12, Math.min(b.y + b.h - 12, p.y))];
	return p.y < cy ? [cx, b.y] : [cx, b.y + b.h];
}

function quad(a: [number, number], via: { x: number; y: number }, b: [number, number]): string {
	const cx = 2 * via.x - (a[0] + b[0]) / 2;
	const cy = 2 * via.y - (a[1] + b[1]) / 2;
	return `M${a[0]} ${a[1]} Q${cx} ${cy} ${b[0]} ${b[1]}`;
}

export function buildScene(
	g: Graph,
	layout: Layout,
	sel: string | null,
	errs: Record<string, string[]>,
	view: Viewport,
	showInherited: boolean,
	models: ModelResolution[]
): Scene {
	const d: Defaults = g.defaults ?? {};
	const paths: PathVM[] = [];
	const labels: LabelVM[] = [];
	const handles: HandleVM[] = [];

	g.nodes.forEach((n) => {
		const L = layout[n.id];
		if (!isBox(L)) return;
		const links = linksOf(n).filter((l) => l.to);
		links.forEach((l, i) => {
			const isOut = sel === n.id;
			const isIn = !isOut && sel === l.to;
			const selected = isOut || isIn;
			const T = layout[l.to];
			if (!isBox(T)) return;
			const key = edgeKey(n.id, i);
			const W: LayoutEntry | undefined = layout[key];
			let dpath: string;
			let mid: [number, number] | null;
			let handle: { x: number; y: number };
			if (l.dashed) {
				const P0 = W ?? { x: (L.x + L.w / 2 + T.x + T.w / 2) / 2, y: Math.min(L.y, T.y + T.h) - 40 };
				dpath = quad(anchor(L, P0), P0, anchor(T, P0));
				mid = null;
				handle = P0;
			} else if (W) {
				dpath = quad(anchor(L, W), W, anchor(T, W));
				mid = [W.x, W.y];
				handle = W;
			} else if (T.x < L.x + L.w - 20) {
				// back edge: loop underneath by default
				const P0 = { x: (L.x + L.w / 2 + T.x + T.w / 2) / 2, y: Math.max(L.y + L.h, T.y + T.h) + 70 };
				dpath = quad([L.x + L.w / 2, L.y + L.h], P0, [T.x + T.w / 2, T.y + T.h]);
				mid = [P0.x, P0.y];
				handle = P0;
			} else {
				const sy = L.y + L.h / 2 + (i - (links.length - 1) / 2) * 16;
				const sx = L.x + L.w;
				let tx = T.x;
				let ty = T.y + T.h / 2;
				if (T.x < L.x + L.w && T.x + T.w > L.x) {
					tx = T.x + T.w / 2;
					ty = T.y > L.y ? T.y : T.y + T.h;
				}
				const dx = Math.max(50, Math.abs(tx - sx) / 2);
				dpath = `M${sx} ${sy} C${sx + dx} ${sy} ${tx - dx} ${ty} ${tx} ${ty}`;
				mid = [
					(sx + 3 * (sx + dx) + 3 * (tx - dx) + tx) / 8,
					(sy + 3 * sy + 3 * ty + ty) / 8 + (i - (links.length - 1) / 2) * 18
				];
				handle = { x: mid[0], y: mid[1] };
			}
			if (l.dashed) {
				paths.push({ key, d: dpath, stroke: SUP, width: selected ? 2 : 1.5, dash: '6 4', marker: 'url(#arw-sup)' });
			} else {
				paths.push({
					key,
					d: dpath,
					stroke: isOut ? FG : isIn ? IN : MUTED,
					width: selected ? 1.75 : 1.25,
					dash: 'none',
					marker: isOut ? 'url(#arw-sel)' : isIn ? 'url(#arw-in)' : 'url(#arw)'
				});
			}
			if (mid && l.label) {
				labels.push({
					key,
					maxW: 150,
					x: mid[0],
					y: mid[1] - 13,
					text: l.label,
					full: l.label,
					color: isOut ? FG : isIn ? IN : MUTED,
					border: isIn ? IN : 'var(--border)'
				});
			}
			if (sel === n.id) handles.push({ key, x: handle.x, y: handle.y, color: l.dashed ? SUP : FG, dashed: !!l.dashed });
		});
	});

	// start arrow
	const S = g.start ? layout[g.start] : undefined;
	if (isBox(S)) {
		const y = S.y + S.h / 2;
		const x0 = S.x - 96;
		paths.push({ key: 'start', d: `M${x0} ${y} L${S.x} ${y}`, stroke: FG, width: 1.75, dash: 'none', marker: 'url(#arw-sel)' });
		labels.push({ key: 'start', maxW: 60, x: x0 + 8, y, text: 'start', full: 'Start node', color: FG, border: 'var(--border)' });
	}

	const terminals: TerminalVM[] = TERMINAL.map((t) => {
		const L = layout[t] as Box;
		const ok = t === 'success';
		const selected = sel === t;
		const ring = selected
			? '0 0 0 2px var(--foreground)'
			: ok
				? '0 0 0 1px var(--border)'
				: '0 0 0 1px color-mix(in oklch, var(--destructive) 45%, transparent)';
		return {
			id: t,
			label: t,
			x: L.x,
			y: L.y,
			w: L.w,
			h: L.h,
			shadow: ring + ', 0 1px 2px rgba(0,0,0,.05)',
			color: ok ? FG : 'var(--destructive)',
			dot: ok ? FG : 'var(--destructive)',
			handleColor: selected ? MUTED : 'transparent'
		};
	});

	const nodes: NodeVM[] = g.nodes.map((n) => {
		const L = isBox(layout[n.id]) ? (layout[n.id] as Box) : { x: 0, y: 0, w: 240, h: 120 };
		const selected = sel === n.id;
		const hasErrors = !!errs[n.id]?.length;
		const ring = hasErrors
			? '0 0 0 2px var(--destructive)'
			: selected
				? '0 0 0 2px var(--foreground)'
				: '0 0 0 1px color-mix(in oklch, var(--foreground) 10%, transparent)';
		const meta: MetaVM[] = [];
		const push = (k: string, v: unknown, inherited = false) => {
			if (v !== undefined && v !== '' && v !== null && (showInherited || !inherited)) {
				meta.push({ k, v: String(v).split('\n')[0], color: inherited ? MUTED : FG });
			}
		};
		const resolved = models.filter((model) => model.node_id === n.id);
		const pushModel = (label: string, role?: string) => {
			const model = resolved.find((candidate) => !role || candidate.role === role);
			if (model) push(label, `${model.native_model} · ${model.effective_effort}`, !model.source.startsWith('node '));
		};
		if (n.type === 'command') {
			push('cmd', n.command);
			push('timeout', n.timeout ?? d.timeout, n.timeout === undefined);
		} else if (n.type === 'loop') {
			push('checklist', n.checklist);
			pushModel('judge', 'item_judge');
			pushModel('evaluator', 'goal_evaluator');
			push('visits', n.max_visits);
		} else if (n.type === 'supervisor') {
			push('every', n.interval);
			push('watching', (n.supervises ?? []).length);
			pushModel('model', 'supervisor');
		} else {
			pushModel('model');
			push('timeout', n.timeout ?? d.timeout, n.timeout === undefined);
			push('retries', n.max_retries ?? d.max_retries, n.max_retries === undefined);
			if (n.type === 'fan_out') {
				push('branches', (n.branches ?? []).map(normalizeBranch).length);
				push('ws', n.workspace);
			}
		}
		const isSup = n.type === 'supervisor';
		return {
			id: n.id,
			type: n.type,
			label: n.label || n.id,
			isStart: g.start === n.id,
			hasErrors,
			errorText: (errs[n.id] ?? []).join('\n'),
			x: L.x,
			y: L.y,
			w: L.w,
			h: L.h,
			shadow: ring + ', 0 1px 2px rgba(0,0,0,.06)',
			radius: isSup ? '4px' : 'var(--radius-xl)',
			bg: isSup ? 'color-mix(in oklch, var(--chart-2) 7%, var(--card))' : 'var(--card)',
			border: isSup ? '1.5px dashed var(--chart-2)' : '1px solid transparent',
			badgeColor: isSup ? IN : MUTED,
			badgeBorder: isSup ? 'color-mix(in oklch, var(--chart-2) 50%, transparent)' : 'var(--border)',
			meta,
			handleColor: selected ? MUTED : 'var(--border)'
		};
	});

	// minimap
	const B = boundsOf(layout);
	const miniNodes = [...g.nodes.map((n) => n.id), ...TERMINAL]
		.map((id) => {
			const L = layout[id];
			if (!isBox(L)) return null;
			return {
				id,
				x: L.x,
				y: L.y,
				w: L.w,
				h: L.h,
				rx: isTerminal(id) ? 28 : 12,
				fill: sel === id ? FG : errs[id]?.length ? 'var(--destructive)' : isTerminal(id) ? 'var(--border)' : 'var(--ring)'
			};
		})
		.filter((x): x is NonNullable<typeof x> => x !== null);
	const { pan, zoom, vw, vh } = view;
	const mini: MiniVM = {
		box: `${B.x} ${B.y} ${B.w} ${B.h}`,
		nodes: miniNodes,
		view: { x: -pan.x / zoom, y: -pan.y / zoom, w: vw / zoom, h: vh / zoom, sw: Math.max(B.w, B.h) / 400 },
		bounds: B,
		scale: Math.max(B.w / 180, B.h / 112)
	};

	return { paths, labels, handles, terminals, nodes, mini };
}

// Count of outgoing list edges, used by the inspector to label edge rows.
export function edgeCount(g: Graph, id: string): number {
	const n = g.nodes.find((x) => x.id === id);
	return n ? listEdges(n).length : 0;
}
