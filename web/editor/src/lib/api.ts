// HTTP contract with the Go server: GET/PUT /api/doc and SSE on /api/events.

export type Severity = 'error' | 'warning' | 'info';

export interface Diagnostic {
	rule: string;
	severity: Severity;
	message: string;
	node_id?: string;
	edge?: [string, string];
	fix?: string;
}

export interface ServerDoc {
	path: string;
	yaml: string;
	layout: Record<string, unknown> | null;
	version: string;
	diagnostics: Diagnostic[];
	parse_error: string;
}

export interface PutBody {
	yaml: string;
	layout: Record<string, unknown> | null;
	version: string;
}

export class ConflictError extends Error {
	current: ServerDoc;
	constructor(current: ServerDoc) {
		super('the file changed on disk');
		this.current = current;
	}
}

async function readError(res: Response): Promise<string> {
	try {
		const body = (await res.json()) as { error?: string };
		if (body && typeof body.error === 'string') return body.error;
	} catch {
		// fall through to the status line
	}
	return `${res.status} ${res.statusText}`;
}

export async function getDoc(): Promise<ServerDoc> {
	const res = await fetch('/api/doc', { headers: { Accept: 'application/json' } });
	if (!res.ok) throw new Error(await readError(res));
	return (await res.json()) as ServerDoc;
}

export async function putDoc(body: PutBody): Promise<ServerDoc> {
	const res = await fetch('/api/doc', {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
		body: JSON.stringify(body)
	});
	if (res.status === 409) throw new ConflictError((await res.json()) as ServerDoc);
	if (!res.ok) throw new Error(await readError(res));
	return (await res.json()) as ServerDoc;
}

// Subscribe to on-disk changes. The callback receives the new version.
export function subscribe(onChange: (version: string) => void): () => void {
	const source = new EventSource('/api/events');
	source.addEventListener('change', (event) => {
		try {
			const data = JSON.parse((event as MessageEvent).data) as { version?: string };
			if (data && typeof data.version === 'string') onChange(data.version);
		} catch {
			// ignore malformed events
		}
	});
	return () => source.close();
}
