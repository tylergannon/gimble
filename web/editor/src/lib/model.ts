// Schema tables for the real pipeline schema (see `gimble print-schema`)
// plus the client-side checks the page runs before the server lints.

export const ID_RE = /^[A-Za-z_][A-Za-z0-9_]*$/;
export const DUR_RE = /^[0-9]+(ms|s|m|h|d)$/;
export const EFFORT = ['low', 'medium', 'high'] as const;
export const FIDELITY = ['full', 'compacted', 'none'] as const;
export const WORKSPACE = ['isolated', 'shared'] as const;
export const TERMINAL = ['success', 'failure'] as const;
export const NODE_TYPES = ['agent', 'fan_out', 'fan_in', 'command', 'supervisor', 'loop'] as const;

export type NodeType = (typeof NODE_TYPES)[number];
export type Terminal = (typeof TERMINAL)[number];

export interface Edge {
	to: string;
	condition?: string;
}

export interface RouteEdges {
	success?: string;
	error?: string;
	loop?: string;
	exit?: string;
}

export interface ModelSelection {
	name?: string;
	version?: string;
	effort?: string;
}

export interface LoopRole {
	model?: ModelSelection;
}

export interface AgentOverride {
	label?: string;
	prompt?: string;
	model?: ModelSelection;
	fidelity?: string;
	timeout?: string;
	max_retries?: number;
	max_visits?: number;
	thread_id?: string;
}

export interface Branch {
	id: string;
	artifacts?: string[];
	agent?: AgentOverride;
}

export interface GraphNode {
	id: string;
	type: NodeType;
	label?: string;
	prompt?: string;
	edges?: Edge[] | RouteEdges;
	branch_edges?: Edge[];
	branches?: (Branch | string)[];
	supervises?: string[];
	command?: string;
	checklist?: string;
	workspace?: string;
	max_parallel?: number;
	interval?: string;
	timeout?: string;
	max_retries?: number;
	max_visits?: number;
	thread_id?: string;
	model?: ModelSelection;
	fidelity?: string;
	item_judge?: LoopRole;
	goal_evaluator?: LoopRole;
	[key: string]: unknown;
}

export interface Defaults {
	model?: ModelSelection;
	fidelity?: string;
	timeout?: string;
	max_retries?: number;
	[key: string]: unknown;
}

export interface Graph {
	name?: string;
	goal?: string;
	start?: string;
	defaults?: Defaults;
	nodes: GraphNode[];
}

export const TYPE_DESC: Record<NodeType, string> = {
	agent: 'Runs an LLM task.',
	fan_out: 'Concurrently walks each branch.',
	fan_in: 'Evaluates branch evidence with an LLM turn.',
	command: 'Executes one shell command.',
	supervisor: 'Observes nodes and coaches them outside the walk.',
	loop: 'Iterates a checklist file until done.'
};

const MODEL = ['model.name', 'model.version', 'model.effort'];
const LLM = [...MODEL, 'fidelity'];
const LIMITS = ['timeout', 'max_retries', 'max_visits', 'thread_id'];

export type Section = [title: string, keys: string[]];

export const FIELD_SECTIONS: Record<NodeType, Section[]> = {
	agent: [
		['Task', ['label', 'prompt']],
		['Model', LLM],
		['Limits', LIMITS]
	],
	fan_in: [
		['Task', ['label', 'prompt']],
		['Model', LLM],
		['Limits', LIMITS]
	],
	fan_out: [
		['Task', ['label', 'prompt', 'workspace', 'max_parallel']],
		['Model', LLM],
		['Limits', LIMITS]
	],
	command: [
		['Command', ['label', 'command']],
		['Limits', ['timeout', 'max_visits']]
	],
	supervisor: [
		['Coaching', ['label', 'prompt', 'interval']],
		['Model', MODEL],
		['Limits', ['timeout']]
	],
	loop: [
		['Loop', ['label', 'checklist']],
		['Item judge', ['item_judge.model.name', 'item_judge.model.version', 'item_judge.model.effort']],
		['Goal evaluator', ['goal_evaluator.model.name', 'goal_evaluator.model.version', 'goal_evaluator.model.effort']],
		['Limits', ['timeout', 'max_visits']]
	]
};

export type FieldKind = 'text' | 'textarea' | 'enum' | 'int' | 'duration';

export interface FieldMeta {
	label: string;
	kind?: FieldKind;
	options?: readonly string[];
	rows?: number;
	mono?: boolean;
	inherit?: boolean;
	placeholder?: string;
}

export const FIELD_META: Record<string, FieldMeta> = {
	label: { label: 'Label' },
	prompt: { label: 'Prompt', kind: 'textarea', rows: 8 },
	command: { label: 'Command', kind: 'textarea', rows: 3, mono: true, placeholder: 'go test ./...' },
	checklist: { label: 'Checklist path', mono: true, placeholder: 'docs/CHECKLIST.md' },
	workspace: { label: 'Workspace', kind: 'enum', options: WORKSPACE },
	max_parallel: { label: 'Max parallel', kind: 'int' },
	'model.name': { label: 'Model name', inherit: true, mono: true },
	'model.version': { label: 'Version (optional)', inherit: true, mono: true },
	'model.effort': { label: 'Reasoning effort', kind: 'enum', options: EFFORT, inherit: true },
	fidelity: { label: 'Fidelity', kind: 'enum', options: FIDELITY, inherit: true },
	'item_judge.model.name': { label: 'Model name', mono: true, placeholder: 'flash' },
	'item_judge.model.version': { label: 'Version (optional)', mono: true },
	'item_judge.model.effort': { label: 'Reasoning effort', kind: 'enum', options: EFFORT },
	'goal_evaluator.model.name': { label: 'Model name', inherit: true, mono: true },
	'goal_evaluator.model.version': { label: 'Version (optional)', inherit: true, mono: true },
	'goal_evaluator.model.effort': { label: 'Reasoning effort', kind: 'enum', options: EFFORT, inherit: true },
	timeout: { label: 'Timeout', kind: 'duration', inherit: true, placeholder: '30m' },
	interval: { label: 'Interval', kind: 'duration', placeholder: '5m' },
	max_retries: { label: 'Max retries', kind: 'int', inherit: true },
	max_visits: { label: 'Max visits', kind: 'int' },
	thread_id: { label: 'Thread id', mono: true }
};

// Required keys per type; dotted keys reach into nested maps.
export const REQUIRED: Partial<Record<NodeType, string[]>> = {
	command: ['command', 'edges.success'],
	supervisor: ['prompt', 'supervises'],
	loop: ['edges.loop', 'edges.exit'],
	fan_out: ['branches']
};

export const DEFAULT_KEYS = [
	'model.name',
	'model.version',
	'model.effort',
	'fidelity',
	'timeout',
	'max_retries'
] as const;

export function isTerminal(id: string | null | undefined): id is Terminal {
	return id === 'success' || id === 'failure';
}

export function getPath(obj: unknown, key: string): unknown {
	let cur: unknown = obj;
	for (const part of key.split('.')) {
		if (cur === null || typeof cur !== 'object') return undefined;
		cur = (cur as Record<string, unknown>)[part];
	}
	return cur;
}

export function routeEdges(n: GraphNode): RouteEdges {
	const e = n.edges;
	return e && !Array.isArray(e) && typeof e === 'object' ? e : {};
}

export function listEdges(n: GraphNode): Edge[] {
	const e = n.type === 'fan_out' ? n.branch_edges : n.edges;
	return Array.isArray(e) ? e.filter((x) => x && typeof x === 'object') : [];
}

// The key that holds a node's outgoing edge list.
export function edgesKey(n: GraphNode): 'edges' | 'branch_edges' {
	return n.type === 'fan_out' ? 'branch_edges' : 'edges';
}

export interface Link {
	to: string;
	label: string;
	dashed?: boolean;
}

export function linksOf(n: GraphNode): Link[] {
	switch (n.type) {
		case 'command': {
			const e = routeEdges(n);
			return [
				{ to: e.success ?? '', label: 'success' },
				{ to: e.error ?? '', label: 'error' }
			];
		}
		case 'loop': {
			const e = routeEdges(n);
			return [
				{ to: e.loop ?? '', label: 'loop' },
				{ to: e.exit ?? '', label: 'exit' }
			];
		}
		case 'supervisor':
			return (n.supervises ?? []).map((to) => ({ to: String(to), label: '', dashed: true }));
		default:
			return listEdges(n).map((e) => ({ to: String(e.to ?? ''), label: e.condition ?? '' }));
	}
}

export function normalizeBranch(b: Branch | string): Branch {
	return typeof b === 'string' ? { id: b, artifacts: [] } : b;
}

export function validate(g: Graph): Record<string, string[]> {
	const ids = g.nodes.map((n) => n.id);
	const errs: Record<string, string[]> = {};
	const push = (id: string, m: string) => (errs[id] = errs[id] || []).push(m);
	const known = (x: string) => ids.includes(x) || isTerminal(x);
	g.nodes.forEach((n) => {
		const id = String(n.id ?? '');
		if (!ID_RE.test(id)) push(id, 'id must match ^[A-Za-z_][A-Za-z0-9_]*$');
		if (isTerminal(id)) push(id, 'id may not be success or failure');
		if (ids.filter((x) => x === id).length > 1) push(id, 'duplicate id');
		if (!NODE_TYPES.includes(n.type)) push(id, `unknown type "${n.type}"`);
		for (const k of ['timeout', 'interval'] as const) {
			const v = n[k];
			if (v !== undefined && v !== '' && !DUR_RE.test(String(v)))
				push(id, `${k} must be an integer followed by ms, s, m, h, or d`);
		}
		for (const k of REQUIRED[n.type] ?? []) {
			const v = getPath(n, k);
			if (v === undefined || v === null || v === '' || (Array.isArray(v) && !v.length))
				push(id, `${k} is required`);
		}
		linksOf(n).forEach((l) => {
			if (l.to && !known(l.to)) push(id, `target "${l.to}" does not exist`);
		});
		if (n.type === 'supervisor') {
			(n.supervises ?? []).forEach((x) => {
				if (x === id) push(id, 'cannot supervise itself');
			});
		}
		if (n.type === 'fan_out') {
			(n.branches ?? []).forEach((b) => {
				if (typeof b === 'object' && (!b.id || !(b.artifacts ?? []).length))
					push(id, `branch ${b.id || '(unnamed)'} needs an id and at least one artifact`);
			});
		}
		const selections = [n.model, n.item_judge?.model, n.goal_evaluator?.model];
		for (const selection of selections) {
			if (!selection) continue;
			if (!selection.name?.trim()) push(id, 'model.name is required when a model selection is present');
			if (selection.version !== undefined && !selection.version.trim()) push(id, 'model.version must be nonblank when present');
			if (selection.effort !== undefined && !EFFORT.includes(selection.effort as (typeof EFFORT)[number]))
				push(id, 'model.effort must be low, medium, or high');
		}
	});
	return errs;
}
