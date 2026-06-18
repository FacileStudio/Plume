<script lang="ts">
	import { spaceStore } from '$lib/stores/space.svelte';
	import Icon from '@iconify/svelte';

	let open = $state(false);

	function select(id: number | null) {
		spaceStore.activeId = id;
		open = false;
	}
</script>

<div class="relative px-3 pb-2">
	<button
		onclick={() => (open = !open)}
		class="flex w-full items-center gap-2.5 rounded-lg border border-border/60 bg-muted/30 px-3 py-2 text-sm transition-colors hover:bg-muted"
	>
		{#if spaceStore.active}
			<Icon icon="solar:users-group-rounded-bold-duotone" class="h-4 w-4 shrink-0 text-muted-foreground" />
		{:else}
			<Icon icon="solar:user-circle-bold-duotone" class="h-4 w-4 shrink-0 text-muted-foreground" />
		{/if}
		<span class="flex-1 truncate text-left font-medium">
			{spaceStore.active?.name ?? 'Personal'}
		</span>
		<Icon
			icon="solar:alt-arrow-down-linear"
			class="h-3.5 w-3.5 text-muted-foreground transition-transform {open ? 'rotate-180' : ''}"
		/>
	</button>

	{#if open}
		<button class="fixed inset-0 z-40" onclick={() => (open = false)} aria-label="Close menu"></button>
		<div class="absolute left-3 right-3 z-50 mt-1 max-h-64 overflow-auto rounded-lg border border-border bg-background p-1 shadow-lg">
			<button
				onclick={() => select(null)}
				class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors {spaceStore.activeId === null
					? 'bg-foreground text-background font-medium'
					: 'hover:bg-muted'}"
			>
				<Icon icon="solar:user-circle-bold-duotone" class="h-4 w-4 shrink-0" />
				Personal
			</button>

			{#each spaceStore.spaces as space}
				<button
					onclick={() => select(space.id)}
					class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition-colors {spaceStore.activeId === space.id
						? 'bg-foreground text-background font-medium'
						: 'hover:bg-muted'}"
				>
					<Icon icon="solar:users-group-rounded-bold-duotone" class="h-4 w-4 shrink-0" />
					<span class="flex-1 truncate text-left">{space.name}</span>
				</button>
			{/each}

			<div class="border-t border-border mt-1 pt-1">
				<a
					href="/spaces"
					onclick={() => (open = false)}
					class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
				>
					<Icon icon="solar:settings-bold-duotone" class="h-4 w-4" />
					Manage spaces
				</a>
			</div>
		</div>
	{/if}
</div>
