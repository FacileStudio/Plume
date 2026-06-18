<script lang="ts">
	import { spaceStore } from '$lib/stores/space.svelte';
	import Icon from '@iconify/svelte';

	let open = $state(false);

	function select(id: number | null) {
		spaceStore.activeId = id;
		open = false;
	}
</script>

<div class="relative">
	<button
		onclick={() => (open = !open)}
		class="flex w-full items-center gap-2 rounded-md border border-border/70 bg-muted/40 px-3 py-2 text-sm transition-colors hover:bg-muted"
	>
		<Icon icon="solar:users-group-rounded-linear" class="h-4 w-4 shrink-0 text-muted-foreground" />
		<span class="flex-1 truncate text-left font-medium">
			{spaceStore.active?.name ?? 'Personal'}
		</span>
		<Icon icon="solar:alt-arrow-down-linear" class="h-3.5 w-3.5 text-muted-foreground" />
	</button>

	{#if open}
		<button class="fixed inset-0 z-40" onclick={() => (open = false)} aria-label="Close menu"></button>
		<div class="absolute left-0 right-0 z-50 mt-1 rounded-md border bg-background shadow-lg">
			<div class="p-1">
				<button
					onclick={() => select(null)}
					class="flex w-full items-center gap-2 rounded-sm px-2.5 py-2 text-sm transition-colors hover:bg-muted {spaceStore.activeId === null ? 'bg-muted font-medium' : ''}"
				>
					<Icon icon="solar:user-rounded-linear" class="h-4 w-4 shrink-0 text-muted-foreground" />
					Personal
				</button>

				{#each spaceStore.spaces as space}
					<button
						onclick={() => select(space.id)}
						class="flex w-full items-center gap-2 rounded-sm px-2.5 py-2 text-sm transition-colors hover:bg-muted {spaceStore.activeId === space.id ? 'bg-muted font-medium' : ''}"
					>
						<Icon icon="solar:users-group-rounded-linear" class="h-4 w-4 shrink-0 text-muted-foreground" />
						<span class="flex-1 truncate text-left">{space.name}</span>
						<span class="text-xs text-muted-foreground">{space.role}</span>
					</button>
				{/each}
			</div>
			<div class="border-t p-1">
				<a
					href="/spaces"
					onclick={() => (open = false)}
					class="flex w-full items-center gap-2 rounded-sm px-2.5 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
				>
					<Icon icon="solar:settings-linear" class="h-4 w-4" />
					Manage spaces
				</a>
			</div>
		</div>
	{/if}
</div>
