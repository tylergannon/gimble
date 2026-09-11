<script lang="ts">
	import { greet, status } from './hello.remote';

	let name = $state('world');
	let busy = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy) return;
		busy = true;
		try {
			// A command is a POST, so this is also the check that the app's
			// origin and the binary's agree. Single flight: the command's
			// response carries the refreshed query.
			await greet(name).updates(status());
		} finally {
			busy = false;
		}
	}
</script>

<!--
	No `pending` snippet: with one, Svelte renders only the snippet on the server
	and the awaited content on the client. Without it the render waits for Go's
	answer, so the document arrives with the values in it.
-->
<svelte:boundary>
	{@const s = await status()}
	<h1 data-testid="title">{s.name}</h1>
	<p data-testid="answered-by">Served by {s.goVersion}.</p>
	<p>Greetings so far: <strong data-testid="greetings">{s.greetings}</strong></p>
	<p data-testid="last-greeting">Last greeting: {s.lastGreeting || '(none yet)'}</p>
	{#snippet failed(error)}
		<p data-testid="status-failed">{(error as Error).message}</p>
	{/snippet}
</svelte:boundary>

<form onsubmit={submit}>
	<input data-testid="name" bind:value={name} />
	<button data-testid="greet" type="submit" disabled={busy}>Greet</button>
</form>
