// The page's contract with Go: three remote functions, declared in
// src/routes/editor.remote.go and answered by the tractor binary. The
// editor.remote.ts beside it is generated and every body there throws, so a
// document on screen is proof that Go answered.
//
// The wire types are generated from the Go structs into
// src/lib/skgo/editor/types.ts. This module turns them into the shapes the
// page works with: the layout travels as a list of placements, because the
// generator projects no maps, and is a record keyed by node id here.

import { getDoc as getDocQuery, saveDoc, watchDoc } from '../routes/editor.remote';
import type { Diagnostic, Document, ModelResolution, Placement } from './skgo/editor/types';
import type { Layout, LayoutEntry } from './layout';

export type { Diagnostic, ModelResolution };

export type Severity = 'error' | 'warning' | 'info';

export interface ServerDoc {
	path: string;
	yaml: string;
	layout: Layout | null;
	version: string;
	diagnostics: Diagnostic[];
	parse_error: string;
	models: ModelResolution[];
}

export interface PutBody {
	yaml: string;
	layout: Layout | null;
	version: string;
}

export class ConflictError extends Error {
	current: ServerDoc;
	constructor(current: ServerDoc) {
		super('the file changed on disk');
		this.current = current;
	}
}

function fromWire(d: Document): ServerDoc {
	let layout: Layout | null = null;
	if (d.has_layout) {
		layout = {};
		for (const p of d.layout) {
			const entry: LayoutEntry = { x: p.x, y: p.y };
			if (p.w > 0) entry.w = p.w;
			if (p.h > 0) entry.h = p.h;
			layout[p.id] = entry;
		}
	}
	return {
		path: d.path,
		yaml: d.yaml,
		layout,
		version: d.version,
		diagnostics: d.diagnostics ?? [],
		parse_error: d.parse_error ?? '',
		models: d.models ?? []
	};
}

function toWire(layout: Layout): Placement[] {
	return Object.entries(layout).map(([id, e]) => ({
		id,
		x: e.x,
		y: e.y,
		w: e.w ?? 0,
		h: e.h ?? 0
	}));
}

// Read the file as it is on disk now. Kit caches a query by its argument and
// hands the cached value to the next caller, so the instance is refreshed
// before it is read: the page asks for the document only when it has reason
// to believe the file changed.
export async function getDoc(): Promise<ServerDoc> {
	const q = getDocQuery();
	await q.refresh();
	return fromWire(await q);
}

// Write the file. A save against a version other than the one on disk writes
// nothing and throws ConflictError carrying what is on disk.
export async function putDoc(body: PutBody): Promise<ServerDoc> {
	const result = await saveDoc({
		yaml: body.yaml,
		layout: body.layout ? toWire(body.layout) : [],
		write_layout: body.layout !== null,
		version: body.version
	});
	const current = fromWire(result.document);
	if (!result.saved) throw new ConflictError(current);
	return current;
}

// Subscribe to on-disk changes. The callback receives the new version; the
// first announcement is the version on disk when the stream opens. Kit keeps
// the live query connected and reconnects it with backoff if it drops.
export function subscribe(onChange: (version: string) => void): () => void {
	const iterator = watchDoc()[Symbol.asyncIterator]();
	let stopped = false;
	void (async () => {
		try {
			for (;;) {
				const { done, value } = await iterator.next();
				if (done || stopped) return;
				onChange(value.version);
			}
		} catch (e) {
			console.error('editor: the change stream ended', e);
		}
	})();
	return () => {
		stopped = true;
		void iterator.return?.();
	};
}
