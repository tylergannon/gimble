<script lang="ts">
	import { RunObservation, type ObservationFrame, type RunSnapshot } from './index.js';
	import SessionTimeline from './SessionTimeline.svelte';

	let { snapshot }: { snapshot: RunSnapshot } = $props();
	const observation = $derived(new RunObservation(snapshot));
	let revision = $state(0);
	let connectionState = $state<'connecting' | 'live' | 'disconnected'>('connecting');
	const currentRun = $derived.by(() => { revision; return { ...observation.run }; });
	const views = $derived.by(() => {
		revision;
		return [...observation.invocations.values()].map((invocation) => ({
			invocation,
			session: observation.run.sessions[invocation.session],
			state: observation.state(invocation.turn)
		}));
	});

	$effect(() => {
		const generation = observation.beginConnection();
		const stream = new EventSource(`/api/runs/${encodeURIComponent(observation.run.id)}/events`);
		const names: ObservationFrame['type'][] = ['snapshot', 'event', 'lifecycle'];
		const listeners = names.map((type) => {
			const listener = (message: MessageEvent<string>) => {
				if (!observation.isCurrentConnection(generation)) return;
				const frame = { type, data: JSON.parse(message.data) } as ObservationFrame;
				if (observation.apply(frame, generation)) revision = observation.revision;
			};
			stream.addEventListener(type, listener as EventListener);
			return [type, listener] as const;
		});
		stream.onopen = () => { if (observation.isCurrentConnection(generation)) connectionState = 'live'; };
		stream.onerror = () => { if (observation.isCurrentConnection(generation)) connectionState = 'disconnected'; };
		return () => {
			observation.endConnection(generation);
			for (const [type, listener] of listeners) stream.removeEventListener(type, listener as EventListener);
			stream.close();
		};
	});
</script>

<header class="run-header">
	<div><p class="eyebrow">Run</p><h1>{currentRun.name}</h1></div>
	<div class="states"><span class={`status ${currentRun.status}`}>{currentRun.status}</span><span>{connectionState}</span></div>
</header>
{#if currentRun.error}<p class="run-error">{currentRun.error}</p>{/if}

{#each views as view (view.invocation.turn)}
	{@const invocation = view.invocation}
	{@const session = view.session}
	{@const state = view.state}
	<section class="invocation">
		<header>
			<div><h2>{session?.name ?? invocation.session}</h2><p>{session?.adapter ?? 'agent'} · {session?.model ?? 'model unavailable'}</p></div>
		</header>
		{#if state}<SessionTimeline {state} {revision} {observation} turn={invocation.turn} provenance={invocation.provenance} />{/if}
	</section>
{/each}
{#if views.length === 0}<p>No agent turns have started.</p>{/if}

<style>
	.run-header, .invocation > header, .states { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
	h1, h2, p { margin: 0; }
	.eyebrow { color: #697386; font-size: .75rem; text-transform: uppercase; letter-spacing: .08em; }
	.states { color: #697386; font-size: .8rem; }
	.status { border-radius: 999px; padding: .2rem .55rem; background: #e9edf4; }
	.status.running { background: #e3efff; color: #174d9b; }
	.status.completed { background: #e0f4e7; color: #176137; }
	.status.failed, .status.cancelled, .run-error { color: #a22525; }
	.run-error { margin-top: 1rem; }
	.invocation { margin-top: 2rem; display: grid; gap: 1rem; }
	.invocation > header p { color: #697386; font-size: .85rem; }
</style>
