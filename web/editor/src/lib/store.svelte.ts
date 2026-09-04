// Editor state: the YAML document, the layout, selection, viewport, the
// debounced save, and the merge of server and client diagnostics.

import { ConflictError, getDoc, putDoc, subscribe, type Diagnostic, type ServerDoc } from './api';
import { PipelineDoc } from './doc';
import { autoLayout, isBox, NODE_H, NODE_W, type Layout, type LayoutEntry } from './layout';
import {
	edgesKey,
	isTerminal,
	listEdges,
	normalizeBranch,
	routeEdges,
	TERMINAL,
	validate,
	type Branch,
	type Edge,
	type Graph,
	type GraphNode,
	type NodeType
} from './model';

const SAVE_DELAY = 400;
export const DARK_KEY = 'tractor-editor-dark';
export const SNAP_KEY = 'tractor-editor-snap';

export type Drag =
	| { type: 'pan'; sx: number; sy: number; ox: number; oy: number }
	| { type: 'move'; id: string; sx: number; sy: number; ox: number; oy: number }
	| { type: 'resize'; id: string; sx: number; sy: number; ow: number; oh: number }
	| { type: 'edge'; key: string; sx: number; sy: number; ox: number; oy: number };

const ROUTE_KEYS = ['success', 'error', 'loop', 'exit'] as const;

function readFlag(key: string): boolean {
	try {
		return localStorage.getItem(key) === '1';
	} catch {
		return false;
	}
}

function writeFlag(key: string, value: boolean): void {
	try {
		localStorage.setItem(key, value ? '1' : '0');
	} catch {
		// storage unavailable
	}
}

export class Editor {
	// document
	path = $state('');
	version = $state('');
	serverDiagnostics = $state<Diagnostic[]>([]);
	parseError = $state('');
	banner = $state('');
	loading = $state(true);
	saving = $state(false);
	needsFit = $state(false);

	private doc: PipelineDoc = PipelineDoc.parse('');
	private rev = $state(0);
	graph: Graph = $derived.by(() => {
		void this.rev;
		return this.doc.toGraph();
	});
	yamlErrors: string[] = $derived.by(() => {
		void this.rev;
		return this.doc.errors;
	});

	// layout
	savedLayout = $state<Layout>({});
	private layoutFromDisk = false;
	private layoutTouched = false;
	layout: Layout = $derived(autoLayout(this.graph, this.savedLayout));

	// view
	sel = $state<string | null>(null);
	pan = $state({ x: 40, y: 60 });
	zoom = $state(0.85);
	vw = $state(1000);
	vh = $state(700);
	drag = $state<Drag | null>(null);
	dark = $state(readFlag(DARK_KEY));
	snap = $state(readFlag(SNAP_KEY));
	showInherited = $state(true);

	// diagnostics
	clientErrors: Record<string, string[]> = $derived(validate(this.graph));
	errors: Record<string, string[]> = $derived.by(() => {
		const out: Record<string, string[]> = {};
		for (const [id, list] of Object.entries(this.clientErrors)) out[id] = [...list];
		for (const d of this.serverDiagnostics) {
			const id = d.node_id || d.edge?.[0];
			if (!id) continue;
			const msg = d.severity === 'error' ? d.message : `${d.severity}: ${d.message}`;
			const list = (out[id] ??= []);
			if (!list.includes(msg)) list.push(msg);
		}
		return out;
	});
	graphIssues: string[] = $derived.by(() => {
		const list: string[] = [];
		if (this.parseError) list.push(this.parseError);
		for (const e of this.yamlErrors) list.push(e);
		const g = this.graph;
		const ids = g.nodes.map((n) => n.id);
		if (!g.start) list.push('Start is required');
		else if (!ids.includes(g.start)) list.push('Start node does not exist');
		for (const d of this.serverDiagnostics) {
			if (d.node_id || d.edge) continue;
			const msg = d.severity === 'error' ? d.message : `${d.severity}: ${d.message}`;
			if (!list.includes(msg)) list.push(msg);
		}
		return list;
	});
	errorCount: number = $derived(
		Object.values(this.errors).reduce((a, b) => a + b.length, 0) + this.graphIssues.length
	);

	selNode: GraphNode | null = $derived(
		this.sel ? (this.graph.nodes.find((n) => n.id === this.sel) ?? null) : null
	);

	// saving
	private timer: ReturnType<typeof setTimeout> | null = null;
	private inflight = false;
	private dirty = false;

	get pending(): boolean {
		return this.timer !== null || this.inflight || this.dirty;
	}

	// ---- lifecycle ----

	start(): () => void {
		void this.load();
		const stop = subscribe((version) => {
			if (version !== this.version && !this.pending) void this.load();
		});
		return () => {
			stop();
			if (this.timer) clearTimeout(this.timer);
		};
	}

	async load(): Promise<void> {
		try {
			const d = await getDoc();
			const first = this.loading;
			this.adopt(d);
			if (first) this.needsFit = true;
		} catch (e) {
			this.banner = `Could not load the document: ${e instanceof Error ? e.message : String(e)}`;
			this.loading = false;
		}
	}

	private adopt(d: ServerDoc): void {
		this.doc = PipelineDoc.parse(d.yaml);
		this.rev++;
		this.savedLayout = (d.layout ?? {}) as Layout;
		this.layoutFromDisk = d.layout !== null;
		this.layoutTouched = false;
		this.path = d.path;
		this.version = d.version;
		this.serverDiagnostics = d.diagnostics ?? [];
		this.parseError = d.parse_error ?? '';
		if (this.sel && !isTerminal(this.sel) && !this.graph.nodes.some((n) => n.id === this.sel)) {
			this.sel = null;
		}
		this.loading = false;
	}

	private touch(): void {
		this.rev++;
		this.scheduleSave();
	}

	private scheduleSave(): void {
		if (this.timer) clearTimeout(this.timer);
		this.timer = setTimeout(() => {
			this.timer = null;
			void this.flush();
		}, SAVE_DELAY);
	}

	private async flush(): Promise<void> {
		if (this.inflight) {
			this.dirty = true;
			return;
		}
		this.inflight = true;
		this.saving = true;
		const layout = this.layoutFromDisk || this.layoutTouched ? $state.snapshot(this.savedLayout) : null;
		try {
			const d = await putDoc({ yaml: this.doc.toString(), layout, version: this.version });
			this.path = d.path;
			this.version = d.version;
			this.serverDiagnostics = d.diagnostics ?? [];
			this.parseError = d.parse_error ?? '';
			if (d.layout !== null) this.layoutFromDisk = true;
		} catch (e) {
			if (e instanceof ConflictError) {
				if (this.timer) clearTimeout(this.timer);
				this.timer = null;
				this.dirty = false;
				this.adopt(e.current);
				this.banner = 'The file changed on disk; your unsaved edits were dropped.';
			} else {
				this.banner = `Save failed: ${e instanceof Error ? e.message : String(e)}`;
			}
		} finally {
			this.inflight = false;
			this.saving = false;
			if (this.dirty) {
				this.dirty = false;
				this.scheduleSave();
			}
		}
	}

	// ---- graph edits ----

	private index(id: string): number {
		return this.graph.nodes.findIndex((n) => n.id === id);
	}

	setGraphField(key: 'name' | 'goal' | 'start', value: string): void {
		if (key === 'start') this.doc.set(['start'], value);
		else this.doc.set([key], value === '' ? undefined : value);
		this.touch();
	}

	setDefault(key: string, value: string | number | undefined): void {
		if (value === undefined) this.doc.deleteAndPrune(['defaults', key]);
		else this.doc.set(['defaults', key], value);
		this.touch();
	}

	// `key` may be dotted to reach into nested maps (edges.success).
	setNodeField(id: string, key: string, value: unknown): void {
		const i = this.index(id);
		if (i < 0) return;
		this.doc.set(['nodes', i, ...key.split('.')], value);
		this.touch();
	}

	setEdge(id: string, index: number, patch: Partial<Edge>): void {
		const i = this.index(id);
		const n = this.graph.nodes[i];
		if (!n) return;
		const key = edgesKey(n);
		if ('to' in patch) this.doc.set(['nodes', i, key, index, 'to'], patch.to ?? '');
		if ('condition' in patch) {
			this.doc.set(['nodes', i, key, index, 'condition'], patch.condition || undefined);
		}
		this.touch();
	}

	addEdge(id: string): void {
		const i = this.index(id);
		const n = this.graph.nodes[i];
		if (!n) return;
		this.doc.push(['nodes', i, edgesKey(n)], { to: '' });
		this.touch();
	}

	removeEdge(id: string, index: number): void {
		const i = this.index(id);
		const n = this.graph.nodes[i];
		if (!n) return;
		this.doc.delete(['nodes', i, edgesKey(n), index]);
		this.touch();
	}

	setSupervises(id: string, list: string[]): void {
		this.setNodeField(id, 'supervises', list);
	}

	branchesOf(id: string): Branch[] {
		const n = this.graph.nodes.find((x) => x.id === id);
		return (n?.branches ?? []).map(normalizeBranch);
	}

	setBranch(id: string, index: number, patch: { id?: string; artifacts?: string[]; prompt?: string }): void {
		const i = this.index(id);
		const n = this.graph.nodes[i];
		if (!n) return;
		const branches = n.branches ?? [];
		const base: (string | number)[] = ['nodes', i, 'branches', index];
		if (typeof branches[index] === 'string') {
			// A legacy branch-root reference becomes an object branch.
			this.doc.set(base, { id: branches[index], artifacts: [] });
		}
		if (patch.id !== undefined) this.doc.set([...base, 'id'], patch.id);
		if (patch.artifacts !== undefined) this.doc.set([...base, 'artifacts'], patch.artifacts);
		if (patch.prompt !== undefined) {
			if (patch.prompt === '') this.doc.deleteAndPrune([...base, 'agent', 'prompt']);
			else this.doc.set([...base, 'agent', 'prompt'], patch.prompt);
		}
		this.touch();
	}

	addBranch(id: string): void {
		const i = this.index(id);
		if (i < 0) return;
		this.doc.push(['nodes', i, 'branches'], { id: '', artifacts: [] });
		this.touch();
	}

	removeBranch(id: string, index: number): void {
		const i = this.index(id);
		if (i < 0) return;
		this.doc.delete(['nodes', i, 'branches', index]);
		this.touch();
	}

	setStart(id: string, on: boolean): void {
		const g = this.graph;
		if (on) this.doc.set(['start'], id);
		else if (g.start === id) this.doc.set(['start'], '');
		this.touch();
	}

	renameNode(oldId: string, newId: string): void {
		newId = newId.trim();
		if (!newId || newId === oldId) return;
		const g = this.graph;
		const i = this.index(oldId);
		if (i < 0) return;
		this.doc.set(['nodes', i, 'id'], newId);
		g.nodes.forEach((n, j) => {
			if (n.type === 'command' || n.type === 'loop') {
				const e = routeEdges(n);
				for (const k of ROUTE_KEYS) if (e[k] === oldId) this.doc.set(['nodes', j, 'edges', k], newId);
			} else if (n.type !== 'supervisor') {
				listEdges(n).forEach((e, k) => {
					if (e.to === oldId) this.doc.set(['nodes', j, edgesKey(n), k, 'to'], newId);
				});
			}
			if (n.type === 'supervisor') {
				(n.supervises ?? []).forEach((s, k) => {
					if (s === oldId) this.doc.set(['nodes', j, 'supervises', k], newId);
				});
			}
		});
		if (g.start === oldId) this.doc.set(['start'], newId);
		const layout: Layout = {};
		const prefix = `e:${oldId}>`;
		for (const [k, v] of Object.entries(this.savedLayout)) {
			const nk = k === oldId ? newId : k.startsWith(prefix) ? `e:${newId}>${k.slice(prefix.length)}` : k;
			layout[nk] = v;
		}
		this.savedLayout = layout;
		this.layoutTouched = true;
		if (this.sel === oldId) this.sel = newId;
		this.touch();
	}

	deleteNode(id: string): void {
		if (!id || isTerminal(id)) return;
		const g = this.graph;
		const i = this.index(id);
		if (i < 0) return;
		g.nodes.forEach((n, j) => {
			if (j === i) return;
			if (n.type === 'command' || n.type === 'loop') {
				const e = routeEdges(n);
				for (const k of ROUTE_KEYS) {
					if (e[k] === id) this.doc.set(['nodes', j, 'edges', k], k === 'error' ? undefined : '');
				}
			} else if (n.type !== 'supervisor') {
				const edges = listEdges(n);
				if (edges.some((e) => e.to === id)) {
					this.doc.set(['nodes', j, edgesKey(n)], edges.filter((e) => e.to !== id));
				}
			}
			if (n.type === 'supervisor' && (n.supervises ?? []).includes(id)) {
				this.doc.set(['nodes', j, 'supervises'], (n.supervises ?? []).filter((x) => x !== id));
			}
		});
		if (g.start === id) this.doc.set(['start'], '');
		this.doc.delete(['nodes', i]);
		const layout: Layout = {};
		const prefix = `e:${id}>`;
		for (const [k, v] of Object.entries(this.savedLayout)) {
			if (k !== id && !k.startsWith(prefix)) layout[k] = v;
		}
		this.savedLayout = layout;
		this.layoutTouched = true;
		if (this.sel === id) this.sel = null;
		this.touch();
	}

	addNode(type: NodeType): void {
		const ids = this.graph.nodes.map((n) => n.id);
		let i = 1;
		let id: string = type;
		while (ids.includes(id)) id = `${type}_${++i}`;
		const n: Record<string, unknown> = { id, type };
		switch (type) {
			case 'command':
				Object.assign(n, { command: '', edges: { success: '' } });
				break;
			case 'supervisor':
				Object.assign(n, { prompt: '', supervises: [] });
				break;
			case 'loop':
				Object.assign(n, { edges: { loop: '', exit: '' } });
				break;
			case 'fan_out':
				Object.assign(n, { branches: [], branch_edges: [] });
				break;
			default:
				Object.assign(n, { edges: [] });
		}
		this.doc.push(['nodes'], n);
		let cx = (this.vw / 2 - this.pan.x) / this.zoom - NODE_W / 2;
		let cy = (this.vh / 2 - this.pan.y) / this.zoom - NODE_H / 2;
		// Step aside when another card already sits at the viewport centre.
		const taken = Object.values(this.layout).filter(isBox);
		while (taken.some((l) => Math.abs(l.x - cx) < 24 && Math.abs(l.y - cy) < 24)) {
			cx += 32;
			cy += 32;
		}
		this.savedLayout[id] = { x: cx, y: cy, w: NODE_W, h: NODE_H };
		this.layoutTouched = true;
		this.sel = id;
		this.touch();
	}

	// ---- layout ----

	setLayout(key: string, entry: LayoutEntry): void {
		this.savedLayout[key] = entry;
		this.layoutTouched = true;
		this.scheduleSave();
	}

	// ---- view ----

	fitView(): void {
		const ls = Object.values(this.layout).filter(isBox);
		if (!ls.length) return;
		const x0 = Math.min(...ls.map((l) => l.x)) - 140;
		const y0 = Math.min(...ls.map((l) => l.y)) - 40;
		const x1 = Math.max(...ls.map((l) => l.x + l.w)) + 140;
		const y1 = Math.max(...ls.map((l) => l.y + l.h)) + 40;
		const zoom = Math.min(1.5, this.vw / (x1 - x0), this.vh / (y1 - y0));
		this.zoom = zoom;
		this.pan = {
			x: (this.vw - (x1 - x0) * zoom) / 2 - x0 * zoom,
			y: (this.vh - (y1 - y0) * zoom) / 2 - y0 * zoom
		};
	}

	zoomBy(f: number): void {
		const nz = Math.min(2.5, Math.max(0.2, this.zoom * f));
		this.pan = {
			x: this.vw / 2 - ((this.vw / 2 - this.pan.x) * nz) / this.zoom,
			y: this.vh / 2 - ((this.vh / 2 - this.pan.y) * nz) / this.zoom
		};
		this.zoom = nz;
	}

	zoomAt(mx: number, my: number, nz: number): void {
		nz = Math.min(2.5, Math.max(0.2, nz));
		this.pan = {
			x: mx - ((mx - this.pan.x) * nz) / this.zoom,
			y: my - ((my - this.pan.y) * nz) / this.zoom
		};
		this.zoom = nz;
	}

	jumpTo(id: string): void {
		const L = this.layout[id];
		if (!isBox(L)) return;
		this.sel = id;
		this.pan = {
			x: this.vw / 2 - (L.x + L.w / 2) * this.zoom,
			y: this.vh / 2 - (L.y + L.h / 2) * this.zoom
		};
	}

	toggleDark(): void {
		this.dark = !this.dark;
		writeFlag(DARK_KEY, this.dark);
	}

	setSnap(on: boolean): void {
		this.snap = on;
		writeFlag(SNAP_KEY, on);
	}

	terminalIds(): readonly string[] {
		return TERMINAL;
	}
}
