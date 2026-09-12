<script lang="ts">
	import RunViewer from '#lib/observation/RunViewer.svelte';
	import type { RunSnapshot } from '#lib/observation/index.js';

	let { data }: { data: { snapshot: RunSnapshot } } = $props();
</script>

<RunViewer snapshot={data.snapshot} />
