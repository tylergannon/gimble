import { readFile } from 'node:fs/promises'
import { createRoot } from 'solid-js'
import { Schema } from 'effect'
import { EventManifest } from '@opencode/schema/event-manifest'
import { mock } from 'bun:test'
import * as realStore from 'solid-js/store'
import type { OpenCodeEvent, SessionInfo } from './upstream/packages/client/src/promise/index.ts'

let capturedStore: any
const originalCreateStore = realStore.createStore
mock.module('solid-js/store', () => ({
	...realStore,
	createStore(initial: unknown, ...rest: unknown[]) {
		const result = (originalCreateStore as any)(initial, ...rest)
		capturedStore = result[0]
		return result
	}
}))
const { createData } = await import('./upstream/packages/client/src/solid/data.ts')
type CreateDataInput = Parameters<typeof createData>[0]

type Fixture = {
	seed: { info: SessionInfo[]; messages?: Record<string, any[]>; pending?: Record<string, any[]> }
	events: unknown[]
}

const path = process.argv[2]
if (!path) throw new Error('usage: run.sh FIXTURE.json')
const fixture = JSON.parse(await readFile(path, 'utf8')) as Fixture

async function prefix(count: number) {
	const listeners = new Set<Parameters<CreateDataInput['event']['listen']>[0]>()
	const take = async (value: unknown) => structuredClone(value)
	const seedInfo = structuredClone(fixture.seed.info)
	const seedMessages = structuredClone(fixture.seed.messages ?? {})
	const seedPending = structuredClone(fixture.seed.pending ?? {})
	const byID = new Map(seedInfo.map(info => [info.id, info]))
	const api = {
		session: {
			get: (args: any) => take(byID.get(args.sessionID)),
			list: () => take({ data: [] }),
			message: (args: any) => take(seedMessages[args.sessionID]?.find((x:any) => x.id === args.messageID)),
			inbox: { list: (args: any) => take(seedPending[args.sessionID] ?? []) }
		},
		message: { list: (args: any) => take({ data: [...(seedMessages[args.sessionID] ?? [])].reverse(), cursor: {} }) }
	}
	const bus: CreateDataInput['event'] = { on: () => () => {}, listen(handler) { listeners.add(handler); return () => listeners.delete(handler) } }
	const setup = createRoot(dispose => ({ data: createData({ api: () => api as never, directory: '/oracle', event: bus, connection: { status: () => 'connected' }, onError: () => {} }), dispose }))
	for (const info of seedInfo) setup.data.session.remember(info)
	for (const sid of Object.keys(seedMessages)) { setup.data.session.message.invalidate(sid); await setup.data.session.message.sync(sid) }
	for (const sid of Object.keys(seedPending)) { setup.data.session.pending.invalidate(sid); await setup.data.session.pending.sync(sid) }
	for (const raw of fixture.events.slice(0, count)) {
		const type = (raw as any).type
		const definition = EventManifest.ServerDefinitions.find(item => item.type === type)
		if (!definition) throw new Error(`unknown event type: ${type}`)
		const event = Schema.decodeUnknownSync(definition)(raw) as OpenCodeEvent
		listeners.forEach(handler => handler({ name: event.type, details: event }))
	}
	const immediate = observe(capturedStore)
	setup.dispose()
	return { prefix: count, immediate }
}

function observe(store: any) {
	return structuredClone({
		info: store.session.info,
		family: store.session.family,
		active: store.session.active,
		message: store.session.message,
		pending: store.session.pending,
		permission: store.session.permission,
		form: store.session.form
	})
}

for (let cut = 0; cut <= fixture.events.length; cut++) console.log(JSON.stringify(await prefix(cut)))
