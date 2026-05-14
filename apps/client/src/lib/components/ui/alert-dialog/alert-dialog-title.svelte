<script lang="ts">
	import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	let {
		class: className,
		children,
		...restProps
	}: AlertDialogPrimitive.TitleProps & { class?: string; children?: Snippet } = $props();
</script>

<AlertDialogPrimitive.Title class={cn('text-lg font-semibold', className)} {...restProps}>
	{@render children?.()}
</AlertDialogPrimitive.Title>
