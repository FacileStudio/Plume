<script lang="ts">
	import { spaceStore } from '$lib/stores/space.svelte';
	import Icon from '@iconify/svelte';

	let open = $state(false);

	function select(id: number | null) {
		spaceStore.activeId = id;
		open = false;
	}

	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('.space-switcher')) {
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

{#if spaceStore.spaces.length > 0}
	<div class="space-switcher relative px-3 pb-2">
		<button
			class="flex w-full items-center gap-2.5 rounded-lg border border-border/60 bg-muted/30 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/60"
			onclick={() => (open = !open)}
		>
			<Icon
				icon={spaceStore.active ? 'solar:users-group-rounded-bold-duotone' : 'solar:user-circle-bold-duotone'}
				class="h-[18px] w-[18px] shrink-0 text-muted-foreground"
			/>
			<span class="min-w-0 flex-1 truncate font-medium">
				{spaceStore.active?.name ?? 'Personal'}
			</span>
			<Icon
				icon="solar:alt-arrow-down-linear"
				class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform {open ? 'rotate-180' : ''}"
			/>
		</button>

		{#if open}
			<div class="absolute left-3 right-3 z-50 mt-1 overflow-hidden rounded-lg border border-border bg-background shadow-lg">
				<div class="max-h-64 overflow-auto p-1">
					<button
						class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm transition-colors {spaceStore.activeId === null
							? 'bg-foreground text-background'
							: 'text-foreground hover:bg-muted'}"
						onclick={() => select(null)}
					>
						<Icon icon="solar:user-circle-bold-duotone" class="h-4 w-4 shrink-0" />
						Personal
					</button>

					{#each spaceStore.spaces as space}
						<button
							class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm transition-colors {spaceStore.activeId === space.id
								? 'bg-foreground text-background'
								: 'text-foreground hover:bg-muted'}"
							onclick={() => select(space.id)}
						>
							<Icon icon="solar:users-group-rounded-bold-duotone" class="h-4 w-4 shrink-0" />
							<span class="min-w-0 flex-1 truncate">{space.name}</span>
						</button>
					{/each}
				</div>

				<div class="border-t border-border p-1">
					<a
						href="/spaces"
						class="flex w-full items-center gap-2.5 rounded-md px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
						onclick={() => (open = false)}
					>
						<Icon icon="solar:settings-linear" class="h-4 w-4 shrink-0" />
						Manage spaces
					</a>
				</div>
			</div>
		{/if}
	</div>
{/if}
