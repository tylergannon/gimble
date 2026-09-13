import { SessionProjection, type JSONObject, type JSONValue, type ProjectionState, type Snapshot } from '../sessionstate/index.js'
import type { Usage } from '../skgo/observation/types.js'

export type { Usage }

export type RunStatus = 'running' | 'completed' | 'failed' | 'cancelled'
export type RunSession = { name: string; adapter: string; model: string; scope: string; parent?: string }
export type InvocationSnapshot = { scope: string; session: string; turn: string; snapshot: Snapshot; provenance: Record<string, unknown> }
export type RunSnapshot = {
	run: { id: string; name: string; status: RunStatus; error?: string; sessions: Record<string, RunSession>; usage?: Record<string, Usage> }
	invocations: Record<string, InvocationSnapshot>
}
export type LifecycleRecord = { seq: number; time: string; scope: string; session?: string; turn?: string; event: { kind: string; [key: string]: unknown } }
export type ObservationFrame =
	| { type: 'snapshot'; data: RunSnapshot }
	| { type: 'event'; data: { scope: string; session: string; turn: string; event: JSONObject; nativeRef?: JSONValue } }
	| { type: 'lifecycle'; data: LifecycleRecord }
export type Invocation = Omit<InvocationSnapshot, 'snapshot'> & { projection: SessionProjection }

const emptySnapshot = (): Snapshot => ({ state: { info: {}, family: {}, active: {}, message: {}, pending: {}, permission: {}, form: {} } })
const clone = <T>(value: T): T => structuredClone(value)

/** Live run state. Revision invalidates renderers without cloning transcript data per frame. */
export class RunObservation {
	run: RunSnapshot['run']
	readonly invocations = new Map<string, Invocation>()
	revision = 0
	private messageRevisions = new Map<string, number>()
	private snapshotRevision = 0
	private connectionGeneration = 0

	constructor(snapshot: RunSnapshot) { this.run = clone(snapshot.run); this.replace(snapshot) }
	beginConnection(): number { return ++this.connectionGeneration }
	endConnection(generation: number) { if (generation === this.connectionGeneration) this.connectionGeneration++ }
	isCurrentConnection(generation: number) { return generation === this.connectionGeneration }

	replace(snapshot: RunSnapshot, generation?: number): boolean {
		if (generation !== undefined && !this.isCurrentConnection(generation)) return false
		this.run = clone(snapshot.run)
		this.messageRevisions.clear()
		this.snapshotRevision = this.revision + 1
		this.invocations.clear()
		for (const [turn, value] of Object.entries(snapshot.invocations)) this.invocations.set(turn, {
			scope: value.scope, session: value.session, turn: value.turn,
			projection: SessionProjection.restore(value.snapshot), provenance: clone(value.provenance)
		})
		this.revision++
		return true
	}

	apply(frame: ObservationFrame, generation: number): boolean {
		if (!this.isCurrentConnection(generation)) return false
		switch (frame.type) {
			case 'snapshot': return this.replace(frame.data, generation)
			case 'event': {
				const value = frame.data
				const messageID = canonicalMessageID(value.event)
				if (messageID) this.messageRevisions.set(`${value.turn}\0${messageID}`, this.revision + 1)
				let invocation = this.invocations.get(value.turn)
				if (!invocation) {
					invocation = { scope: value.scope, session: value.session, turn: value.turn, projection: SessionProjection.restore(emptySnapshot()), provenance: {} }
					this.invocations.set(value.turn, invocation)
				}
				invocation.projection.apply(value.event)
				foldProvenance(invocation.provenance, value.event, value.nativeRef)
				break
			}
			case 'lifecycle': foldLifecycle(this.run, frame.data); break
		}
		this.revision++
		return true
	}

	state(turn: string): Readonly<ProjectionState> | undefined { return this.invocations.get(turn)?.projection.viewState() }
	messageRevision(turn: string, message: string): number { return this.messageRevisions.get(`${turn}\0${message}`) ?? this.snapshotRevision }
	snapshot(): RunSnapshot {
		return { run: clone(this.run), invocations: Object.fromEntries([...this.invocations].map(([turn, value]) => [turn, {
			scope: value.scope, session: value.session, turn: value.turn,
			snapshot: value.projection.snapshot(), provenance: clone(value.provenance)
		}])) }
	}
}

export function foldLifecycle(run: RunSnapshot['run'], record: LifecycleRecord) {
	const event = record.event
	switch (event.kind) {
		case 'run_started': run.name = String(event.name); run.status = 'running'; delete run.error; break
		case 'run_ended': run.status = event.error ? 'failed' : 'completed'; if (event.error) run.error = String(event.error); else delete run.error; break
		case 'run_cancelled': run.status = 'cancelled'; if (event.error) run.error = String(event.error); else delete run.error; break
		case 'session_created':
			if (record.session) run.sessions[record.session] = { name: String(event.name), adapter: String(event.adapter), model: String(event.model), scope: record.scope, ...(event.parent ? { parent: String(event.parent) } : {}) }
			break
	}
}

export function canonicalMessageID(event: JSONObject): string | undefined {
	const data = event.data ?? {}
	if (typeof data.assistantMessageID === 'string') return data.assistantMessageID
	if (typeof data.messageID === 'string') return data.messageID
	if (typeof data.message?.id === 'string') return data.message.id
	if (typeof data.inboxID === 'string') return data.inboxID
	return undefined
}

export function foldProvenance(target: Record<string, unknown>, event: JSONObject, nativeRef?: JSONValue) {
	if (nativeRef === undefined) return
	const normalized = typeof nativeRef === 'object' && nativeRef !== null && !Array.isArray(nativeRef) ? nativeRef.normalizedMessageID : undefined
	const messageID = typeof normalized === 'string' ? normalized : canonicalMessageID(event)
	if (messageID) target[messageID] = clone(nativeRef)
}

/** The five token counts and the cost of one message or session, zero-filled.
 * A count a harness did not report is 0: zero is a number, and nothing here
 * says whether it was measured. */
export function usageOf(value: JSONObject | undefined): Usage {
	const tokens = (value?.tokens ?? {}) as JSONObject
	const cache = (tokens.cache ?? {}) as JSONObject
	return {
		cost: numberOf(value?.cost),
		tokens: {
			input: numberOf(tokens.input),
			output: numberOf(tokens.output),
			reasoning: numberOf(tokens.reasoning),
			cache: { read: numberOf(cache.read), write: numberOf(cache.write) }
		}
	}
}

/** One line of usage, for a message, a session total or a scope. */
export const usageText = (usage: Usage): string =>
	`${usage.tokens.input} in · ${usage.tokens.output} out · ${usage.tokens.reasoning} reasoning · ` +
	`${usage.tokens.cache.read} cache read · ${usage.tokens.cache.write} cache write · $${usage.cost}`

const numberOf = (value: unknown): number => (typeof value === 'number' && Number.isFinite(value) ? value : 0)
