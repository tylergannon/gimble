<script lang="ts">
	import NativeSelect, { type Option } from './NativeSelect.svelte';

	export interface FieldVM {
		key: string;
		label: string;
		value: string;
		kind: 'text' | 'textarea' | 'enum';
		options: Option[];
		rows: number;
		placeholder: string;
		required: boolean;
		inherited: string;
		error: string;
		mono: boolean;
		// input fires on every keystroke; change fires on commit
		commitOnly?: boolean;
	}

	let {
		field,
		showRequired = true,
		onvalue
	}: { field: FieldVM; showRequired?: boolean; onvalue: (value: string) => void } = $props();
</script>

<div class="field">
	<div class="head">
		<label class="cn-label" for={`field-${field.key}`}>{field.label}</label>
		{#if showRequired && field.required}<span class="hint">required</span>{/if}
		{#if field.inherited}<span class="hint inherited">default: {field.inherited}</span>{/if}
	</div>
	{#if field.kind === 'text'}
		<input
			id={`field-${field.key}`}
			class="cn-input"
			class:mono={field.mono}
			value={field.value}
			placeholder={field.placeholder}
			oninput={(e) => !field.commitOnly && onvalue(e.currentTarget.value)}
			onchange={(e) => field.commitOnly && onvalue(e.currentTarget.value)}
		/>
	{:else if field.kind === 'textarea'}
		<textarea
			id={`field-${field.key}`}
			class="cn-textarea"
			class:mono={field.mono}
			rows={field.rows}
			value={field.value}
			placeholder={field.placeholder}
			oninput={(e) => onvalue(e.currentTarget.value)}
		></textarea>
	{:else}
		<NativeSelect value={field.value} options={field.options} onchange={(v) => onvalue(v)} />
	{/if}
	{#if field.error}<div class="field-error">{field.error}</div>{/if}
</div>

<style>
	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.head {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.hint {
		font-size: 11px;
		color: var(--muted-foreground);
	}
	.inherited {
		margin-left: auto;
	}
	textarea.cn-textarea {
		resize: vertical;
		line-height: 1.45;
	}
</style>
