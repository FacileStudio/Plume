<script lang="ts">
	import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	let {
		class: className,
		children,
		...restProps
	}: AlertDialogPrimitive.CancelProps & { class?: string; children?: Snippet } = $props();
</script>

<AlertDialogPrimitive.Cancel class={cn(buttonVariants({ variant: 'outline' }), 'mt-2 sm:mt-0', className)} {...restProps}>
	{@render children?.()}
</AlertDialogPrimitive.Cancel>
