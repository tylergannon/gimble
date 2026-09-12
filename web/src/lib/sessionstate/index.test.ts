import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { describe, test } from 'node:test'
import { SessionProjection, type ProjectionState } from './index.ts'

const initial = (fixture: any): ProjectionState => {
	const info = Object.fromEntries(fixture.seed.info.map((item: any) => [item.id, item]))
	const family = Object.fromEntries(fixture.seed.info.filter((item: any) => !item.parentID).map((item: any) => [item.id, [item.id]]))
	return { info, family, active: {}, message: structuredClone(fixture.seed.messages ?? {}), pending: structuredClone(fixture.seed.pending ?? {}), permission: {}, form: {} }
}

describe('session event projection', () => {
	for (const name of ['stream', 'native-backend', 'overlapping-reasoning']) {
		test(`${name} restores at every event cut`, async () => {
			const fixture = JSON.parse(await readFile(new URL(`./fixtures/${name}.json`, import.meta.url), 'utf8'))
			const uninterrupted = new SessionProjection(initial(fixture))
			for (const event of fixture.events) uninterrupted.apply(event)
			for (let cut = 0; cut <= fixture.events.length; cut++) {
				const before = new SessionProjection(initial(fixture))
				for (const event of fixture.events.slice(0, cut)) before.apply(event)
				const restored = SessionProjection.restore(before.snapshot())
				for (const event of fixture.events.slice(cut)) restored.apply(event)
				assert.deepEqual(restored.snapshot().state, uninterrupted.snapshot().state, `cut ${cut}`)
			}
		})
	}
})
