import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import { RunObservation, accounting, foldProvenance, type LifecycleRecord, type RunSnapshot } from './index.ts'
import type { Snapshot } from '../sessionstate/index.ts'

const projection = (): Snapshot => ({
	state: { info: { ses: { id: 'ses', title: 'start' } }, family: { ses: ['ses'] }, active: {}, message: { ses: [] }, pending: {}, permission: {}, form: {} }
})
const snapshot = (title = 'start'): RunSnapshot => {
	const value = projection(); value.state.info.ses.title = title
	return { run: { id: 'run', name: 'Run', status: 'running', sessions: { ses: { name: 'agent', adapter: 'codex', model: 'm', scope: 'lap' } } }, scopes: { 'loop.1': { name: 'loop.1', status: 'running' } }, invocations: { turn: { scope: 'lap', session: 'ses', turn: 'turn', snapshot: value, provenance: {} } } }
}

test('replacement snapshot resets machines and stale connection callbacks cannot mutate them', () => {
	const observation = new RunObservation(snapshot('first'))
	const oldConnection = observation.beginConnection()
	observation.apply({ type: 'snapshot', data: snapshot('replacement') }, oldConnection)
	const currentConnection = observation.beginConnection()
	assert.equal(observation.apply({ type: 'event', data: { scope: 'lap', session: 'ses', turn: 'turn', event: { id: 'old', created: 1, type: 'session.renamed', data: { sessionID: 'ses', title: 'stale' } } } }, oldConnection), false)
	observation.apply({ type: 'event', data: { scope: 'lap', session: 'ses', turn: 'turn', event: { id: 'new', created: 2, type: 'session.permissions.updated', data: { sessionID: 'ses', permissions: ['read'] } } } }, currentConnection)
	assert.equal(observation.state('turn')?.info.ses.title, 'replacement')
	assert.deepEqual(observation.state('turn')?.info.ses.permissions, ['read'])
})

test('lifecycle only controls workflow terminal status and relationships', () => {
	const observation = new RunObservation(snapshot())
	const connection = observation.beginConnection()
	observation.apply({ type: 'lifecycle', data: { seq: 1, time: '2026-01-01T00:00:00Z', scope: 'lap.2', session: 'reviewer', event: { kind: 'session_created', name: 'reviewer', adapter: 'claude', model: 'opus', workdir: '/tmp', parent: 'ses' } } }, connection)
	observation.apply({ type: 'event', data: { scope: 'lap', session: 'ses', turn: 'turn', event: { id: 'ended', created: 3, type: 'session.step.ended', data: { sessionID: 'ses', assistantMessageID: 'none', finish: 'stop' } } } }, connection)
	assert.equal(observation.run.status, 'running')
	assert.equal(observation.run.sessions.reviewer.parent, 'ses')
	observation.apply({ type: 'lifecycle', data: { seq: 2, time: '2026-01-01T00:00:01Z', scope: '', event: { kind: 'run_ended', name: 'Run', error: '' } } }, connection)
	assert.equal(observation.run.status, 'completed')
})

test('message revision survives a following lifecycle frame', () => {
	const observation = new RunObservation(snapshot())
	const connection = observation.beginConnection()
	observation.apply({ type: 'event', data: { scope: 'lap', session: 'ses', turn: 'turn', event: { id: 'delta', created: 3, type: 'session.text.delta', data: { sessionID: 'ses', assistantMessageID: 'message', delta: 'done' } } } }, connection)
	const changed = observation.messageRevision('turn', 'message')
	observation.apply({ type: 'lifecycle', data: { seq: 2, time: '2026-01-01T00:00:01Z', scope: '', event: { kind: 'run_ended', name: 'Run', error: '' } } }, connection)
	assert.equal(observation.messageRevision('turn', 'message'), changed)
})

test('accounting distinguishes unavailable from measured zero', () => {
	assert.equal(accounting({}, undefined), 'unavailable')
	assert.equal(accounting({ cost: 0, tokens: { input: 0 } }, undefined), 'unavailable')
	assert.equal(accounting({ cost: 0, tokens: { input: 0 } }, { accounting: { costAvailable: true, tokensAvailable: true, costSource: 'provider' } }), 'cost 0; tokens {"input":0}')
})

test('provenance keys the latest native sidecar by normalized message ID', () => {
	const provenance: Record<string, unknown> = {}
	foldProvenance(provenance, { data: { assistantMessageID: 'source-id' } }, { normalizedMessageID: 'msg_canonical', messageID: 'provider-id', accounting: { costAvailable: false, tokensAvailable: false, costSource: 'unavailable' } })
	assert.deepEqual(Object.keys(provenance), ['msg_canonical'])
	assert.equal((provenance.msg_canonical as { messageID: string }).messageID, 'provider-id')
})

test('a cancelled run stays cancelled, as the Go store decides it', () => {
	const observation = new RunObservation(snapshot())
	const connection = observation.beginConnection()
	const fixture = new URL('../../../../internal/observation/testdata/cancelled-run.jsonl', import.meta.url)
	for (const line of readFileSync(fixture, 'utf8').trim().split('\n')) {
		observation.apply({ type: 'lifecycle', data: JSON.parse(line) as LifecycleRecord }, connection)
	}
	assert.equal(observation.run.status, 'cancelled')
	assert.equal(observation.run.error, 'context canceled')
})

test('the snapshot carries the scope tree and workflow lifecycle folds into it', () => {
	const observation = new RunObservation(snapshot())
	const connection = observation.beginConnection()
	const frame = (seq: number, scope: string, event: Record<string, unknown>) =>
		observation.apply({ type: 'lifecycle', data: { seq, time: '2026-01-01T00:00:00Z', scope, event } as LifecycleRecord }, connection)
	frame(1, 'loop.1', { kind: 'planner_decision', task: { name: 'write a.txt' } })
	frame(2, 'loop.1/task.2', { kind: 'scope_began', name: 'task.2', task: { name: 'write a.txt' } })
	frame(3, 'loop.1/task.2', { kind: 'value_set', key: 'worker result', value: '"done"' })
	frame(4, 'loop.1/task.2', { kind: 'scope_ended', error: '' })
	frame(5, 'loop.1', { kind: 'planner_decision' })
	// An ended scope with no error is ended, never succeeded.
	assert.deepEqual(observation.scopes['loop.1/task.2'], { name: 'task.2', status: 'ended', task: { name: 'write a.txt' }, values: { 'worker result': 'done' } })
	assert.deepEqual(observation.scopes['loop.1'].decisions, [{ task: { name: 'write a.txt' } }, {}])
	assert.deepEqual(observation.snapshot().scopes, observation.scopes)
})
