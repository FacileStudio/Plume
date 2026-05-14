<script lang="ts">
	import { AlertDialog as AlertDialogPrimitive } from 'bits-ui';
	import { buttonVariants, type ButtonVariant } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	let {
		class: className,
		variant = 'default' as ButtonVariant,
		children,
		...restProps
	}: AlertDialogPrimitive.ActionProps & { class?: string; variant?: ButtonVariant; children?: Snippet } = $props();
</script>

<AlertDialogPrimitive.Action class={cn(buttonVariants({ variant }), className)} {...restProps}>
	{@render children?.()}
</AlertDialogPrimitive.Action>
