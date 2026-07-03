<script lang="ts">
	import type { Client } from '$lib';
	import Icon from '@iconify/svelte';

	let {
		clients,
		selectedId,
		onselect
	}: {
		clients: Client[];
		selectedId: number | null;
		onselect: (id: number | null) => void;
	} = $props();

	let open = $state(false);

	const selected = $derived(clients.find((c) => c.id === selectedId) ?? null);

	function select(id: number | null) {
		onselect(id);
		open = false;
	}

	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('.client-selector')) {
			open = false;
		}
	}

	$effect(() => {
		if (open) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});
</script>

<div class="client-selector relative">
	<button
		type="button"
		class="flex w-full items-center gap-2.5 rounded-lg border border-border/60 bg-muted/30 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/60"
		onclick={() => (open = !open)}
	>
		<Icon
			icon={selected ? 'solar:user-rounded-bold-duotone' : 'solar:user-cross-linear'}
			class="h-[18px] w-[18px] shrink-0 text-muted-foreground"
		/>
		<span class="min-w-0 flex-1 truncate {selected ? 'font-medium' : 'text-muted-foreground'}">
			{selected?.name ?? 'No client'}
		</span>
		<Icon
			icon="solar:alt-arrow-down-linear"
			class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform {open ? 'rotate-180' : ''}"
		/>
	</button>

	{#if open}
		<div class="absolute left-0 right-0 z-50 mt-1 overflow-hidden rounded-lg border border-border bg-background shadow-lg">
			<div class="max-h-64 overflow-auto p-1">
				<button
					type="button"
					class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm transition-colors {selectedId === null
						? 'bg-foreground text-background'
						: 'text-foreground hover:bg-muted'}"
					onclick={() => select(null)}
				>
					<Icon icon="solar:user-cross-linear" class="h-4 w-4 shrink-0" />
					No client
				</button>

				{#each clients as client}
					<button
						type="button"
						class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm transition-colors {selectedId === client.id
							? 'bg-foreground text-background'
							: 'text-foreground hover:bg-muted'}"
						onclick={() => select(client.id)}
					>
						<Icon icon="solar:user-rounded-bold-duotone" class="h-4 w-4 shrink-0" />
						<span class="min-w-0 flex-1 truncate">{client.name}</span>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
