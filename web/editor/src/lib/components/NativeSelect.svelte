<script lang="ts">
	import { ChevronDown } from '@lucide/svelte';

	export interface Option {
		v: string;
		l: string;
		disabled?: boolean;
	}

	let {
		value = '',
		options,
		size = 'default',
		title = '',
		disabled = false,
		onchange
	}: {
		value?: string;
		options: Option[];
		size?: 'default' | 'sm';
		title?: string;
		disabled?: boolean;
		onchange: (value: string, el: HTMLSelectElement) => void;
	} = $props();
</script>

<span class="cn-native-select-wrapper">
	<select
		class="cn-native-select"
		class:cn-native-select-size-sm={size === 'sm'}
		{title}
		{value}
		{disabled}
		onchange={(e) => onchange(e.currentTarget.value, e.currentTarget)}
	>
		{#each options as o (o.v + o.l)}
			<option value={o.v} disabled={o.disabled ?? false}>{o.l}</option>
		{/each}
	</select>
	<ChevronDown class="cn-native-select-icon" />
</span>
