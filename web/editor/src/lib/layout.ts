// Node placement. Saved entries come from the layout sidecar; nodes without
// one get a layered position: columns by depth from `start`, rows by order,
// supervisors in a row beneath, terminals to the right.

import { isTerminal, linksOf, type Graph } from './model';

export interface LayoutEntry {
	x: number;
	y: number;
	w?: number;
	h?: number;
}

export type Layout = Record<string, LayoutEntry>;

export interface Box {
	x: number;
	y: number;
	w: number;
	h: number;
}

export const NODE_W = 240;
export const NODE_H = 120;
export const TERM_W = 150;
export const TERM_H = 56;

const X0 = 60;
const Y0 = 80;
const COL = 340;
const ROW = 170;

export function isBox(l: LayoutEntry | undefined): l is Box {
	return !!l && typeof l.w === 'number' && typeof l.h === 'number';
}

export function edgeKey(from: string, index: number): string {
	return `e:${from}>${index}`;
}

export function autoLayout(g: Graph, saved: Layout): Layout {
	const out: Layout = { ...saved };
	const nodes = g.nodes.filter((n) => n.id);
	const byId = new Map(nodes.map((n) => [n.id, n]));
	const walkers = nodes.filter((n) => n.type !== 'supervisor');

	// Depth by breadth-first walk along every outgoing route from start.
	const depth = new Map<string, number>();
	if (g.start && byId.has(g.start) && byId.get(g.start)!.type !== 'supervisor') {
		depth.set(g.start, 0);
		const queue = [g.start];
		while (queue.length) {
			const id = queue.shift()!;
			const d = depth.get(id)!;
			for (const l of linksOf(byId.get(id)!)) {
				if (!l.to || isTerminal(l.to) || !byId.has(l.to) || depth.has(l.to)) continue;
				if (byId.get(l.to)!.type === 'supervisor') continue;
				depth.set(l.to, d + 1);
				queue.push(l.to);
			}
		}
	}
	let maxDepth = -1;
	for (const d of depth.values()) maxDepth = Math.max(maxDepth, d);
	const unreached = walkers.filter((n) => !depth.has(n.id));
	unreached.forEach((n) => depth.set(n.id, maxDepth + 1));

	const rows = new Map<number, number>();
	let bottom = Y0;
	for (const n of walkers) {
		const col = depth.get(n.id) ?? 0;
		const row = rows.get(col) ?? 0;
		rows.set(col, row + 1);
		const y = Y0 + row * ROW;
		bottom = Math.max(bottom, y + NODE_H);
		if (!isBox(out[n.id])) out[n.id] = { x: X0 + col * COL, y, w: NODE_W, h: NODE_H };
	}

	// Supervisors sit in a row beneath the walk.
	let supIndex = 0;
	for (const n of nodes) {
		if (n.type !== 'supervisor') continue;
		if (!isBox(out[n.id])) {
			out[n.id] = { x: X0 + supIndex * COL, y: bottom + ROW - NODE_H, w: NODE_W, h: NODE_H };
		}
		supIndex++;
	}

	// Terminals: to the right of everything placed so far.
	if (!isBox(out.success) || !isBox(out.failure)) {
		const boxes = nodes.map((n) => out[n.id]).filter(isBox);
		const mx = boxes.length ? Math.max(...boxes.map((l) => l.x + l.w)) : X0;
		const my = boxes.length ? Math.min(...boxes.map((l) => l.y)) : Y0;
		if (!isBox(out.success)) out.success = { x: mx + 160, y: my + 36, w: TERM_W, h: TERM_H };
		if (!isBox(out.failure)) out.failure = { x: mx + 160, y: my + 136, w: TERM_W, h: TERM_H };
	}
	return out;
}

export function boundsOf(layout: Layout): Box {
	const ls = Object.values(layout).filter(isBox);
	if (!ls.length) return { x: 0, y: 0, w: 1000, h: 600 };
	const x0 = Math.min(...ls.map((l) => l.x)) - 200;
	const y0 = Math.min(...ls.map((l) => l.y)) - 200;
	return {
		x: x0,
		y: y0,
		w: Math.max(...ls.map((l) => l.x + l.w)) + 200 - x0,
		h: Math.max(...ls.map((l) => l.y + l.h)) + 200 - y0
	};
}
