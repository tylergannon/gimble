// The browser owns the YAML document. Edits go back into the parsed
// `yaml` Document at the changed path so comments and key order survive.
// Scalars that were not touched are written back from their source text,
// so untouched prose keeps its original wrapping.

import {
	isCollection,
	isScalar,
	isSeq,
	parseDocument,
	visit,
	type Document,
	type Scalar,
	type ToStringOptions
} from 'yaml';
import type { Graph, GraphNode } from './model';

export type Path = (string | number)[];

const STR_TAG = 'tag:yaml.org,2002:str';
const ORIGINAL = Symbol('original value');

type SourceScalar = Scalar & { [ORIGINAL]?: unknown };

interface StringifyCtx {
	indent: string;
	implicitKey?: boolean;
	inFlow?: boolean;
}

type StringifyFn = (
	item: unknown,
	ctx: unknown,
	onComment?: () => void,
	onChompKeep?: () => void
) => string;

interface BlockScalarToken {
	type: 'block-scalar';
	props: { type: string; source: string }[];
	source: string;
}

interface FlowScalarToken {
	type: 'scalar' | 'single-quoted-scalar' | 'double-quoted-scalar';
	source: string;
}

// Original source for an unchanged scalar, re-indented to the current
// indent, or null when the default stringifier should run.
function sourceOf(item: SourceScalar, ctx: StringifyCtx): string | null {
	const tok = item.srcToken as BlockScalarToken | FlowScalarToken | undefined;
	if (!tok || item[ORIGINAL] !== item.value || item.comment || ctx.implicitKey) return null;
	if (tok.type === 'block-scalar') {
		if (tok.props.some((p) => p.type !== 'block-scalar-header' && p.type !== 'newline')) return null;
		const header = tok.props.find((p) => p.type === 'block-scalar-header')?.source ?? '';
		if (!header || /[0-9+]/.test(header)) return null;
		const lines = tok.source.replace(/\n+$/, '').split('\n');
		const indents = lines.filter((l) => l.trim()).map((l) => (l.match(/^ */) as RegExpMatchArray)[0].length);
		if (!indents.length) return null;
		const min = Math.min(...indents);
		const body = lines.map((l) => (l.trim() ? ctx.indent + l.slice(min) : '')).join('\n');
		return header + '\n' + body;
	}
	if (tok.type === 'scalar' || tok.type === 'single-quoted-scalar' || tok.type === 'double-quoted-scalar') {
		if (tok.source.includes('\n')) return null;
		if (tok.type === 'scalar' && item.type !== 'PLAIN') return null;
		return tok.source;
	}
	return null;
}

function preserveSources(doc: Document): void {
	visit(doc, {
		Scalar(_, node) {
			(node as SourceScalar)[ORIGINAL] = node.value;
		}
	});
	doc.schema.tags = doc.schema.tags.map((tag) => {
		if (tag.tag !== STR_TAG || !tag.stringify) return tag;
		const base = tag.stringify as StringifyFn;
		return {
			...tag,
			stringify: (item: unknown, ctx: unknown, onComment?: () => void, onChompKeep?: () => void) =>
				(isScalar(item) ? sourceOf(item as SourceScalar, ctx as StringifyCtx) : null) ??
				base(item, ctx, onComment, onChompKeep)
		};
	});
}

function detectIndent(text: string): number {
	const m = text.match(/^( +)\S/m);
	return m ? m[1].length : 2;
}

export class PipelineDoc {
	readonly doc: Document;
	private readonly options: ToStringOptions;

	private constructor(doc: Document, options: ToStringOptions) {
		this.doc = doc;
		this.options = options;
	}

	static parse(text: string): PipelineDoc {
		const doc = parseDocument(text, { keepSourceTokens: true });
		preserveSources(doc);
		return new PipelineDoc(doc, { flowCollectionPadding: false, indent: detectIndent(text) });
	}

	get errors(): string[] {
		return this.doc.errors.map((e) => e.message.split('\n')[0]);
	}

	toString(): string {
		if (this.doc.contents === null) return '';
		return this.doc.toString(this.options);
	}

	toGraph(): Graph {
		const js = this.doc.toJS() as unknown;
		const root = js && typeof js === 'object' && !Array.isArray(js) ? (js as Record<string, unknown>) : {};
		const rawNodes = Array.isArray(root.nodes) ? (root.nodes as unknown[]) : [];
		const nodes = rawNodes
			.filter((n): n is GraphNode => !!n && typeof n === 'object' && !Array.isArray(n))
			.map((n) => ({ ...n, id: n.id === undefined || n.id === null ? '' : String(n.id) }));
		const defaults =
			root.defaults && typeof root.defaults === 'object' && !Array.isArray(root.defaults)
				? (root.defaults as Graph['defaults'])
				: undefined;
		const str = (v: unknown) => (v === undefined || v === null ? undefined : String(v));
		return { name: str(root.name), goal: str(root.goal), start: str(root.start), defaults, nodes };
	}

	get(path: Path): unknown {
		return this.doc.getIn(path);
	}

	has(path: Path): boolean {
		return this.doc.hasIn(path);
	}

	// Set a value; `undefined` removes the key. When the existing value is a
	// scalar of the same JS type, its style (block, folded, quoted) is kept;
	// a replaced flow sequence stays a flow sequence.
	set(path: Path, value: unknown): void {
		if (value === undefined) {
			this.doc.deleteIn(path);
			return;
		}
		const existing = this.doc.getIn(path, true);
		if (
			isScalar(existing) &&
			(typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') &&
			typeof existing.value === typeof value
		) {
			existing.value = value;
			return;
		}
		const flow = isSeq(existing) && existing.flow;
		this.doc.setIn(path, this.doc.createNode(value));
		if (flow && Array.isArray(value)) {
			const created = this.doc.getIn(path, true);
			if (isSeq(created)) created.flow = true;
		}
	}

	delete(path: Path): void {
		this.doc.deleteIn(path);
	}

	// Append to the sequence at `path`, creating it when missing.
	push(path: Path, value: unknown): void {
		const existing = this.doc.getIn(path, true);
		if (isCollection(existing)) {
			this.doc.addIn(path, this.doc.createNode(value));
		} else {
			this.doc.setIn(path, this.doc.createNode([value]));
		}
	}

	// Remove a key and then its parent when the parent collection is now empty.
	deleteAndPrune(path: Path): void {
		this.doc.deleteIn(path);
		const parent = path.slice(0, -1);
		if (!parent.length) return;
		const node = this.doc.getIn(parent, true);
		if (isCollection(node) && node.items.length === 0) this.doc.deleteIn(parent);
	}
}
