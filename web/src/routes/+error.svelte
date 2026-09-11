<script lang="ts">
	import { page } from '$app/state';
</script>

<h1 data-testid="title">Error {page.status}</h1>
<p data-testid="error-message">{page.error?.message}</p>
