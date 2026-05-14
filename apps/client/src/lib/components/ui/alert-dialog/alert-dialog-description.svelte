<script lang="ts">
	import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	let {
		class: className,
		children,
		...restProps
	}: AlertDialogPrimitive.DescriptionProps & { class?: string; children?: Snippet } = $props();
</script>

<AlertDialogPrimitive.Description class={cn('text-sm text-muted-foreground', className)} {...restProps}>
	{@render children?.()}
</AlertDialogPrimitive.Description>
