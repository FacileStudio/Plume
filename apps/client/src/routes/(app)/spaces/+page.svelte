<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Space } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';
	import { spaceStore } from '$lib/stores/space.svelte';

	let spaces = $state<Space[]>([]);
	let loading = $state(true);
	let leaveTarget = $state<Space | null>(null);
	let leaving = $state(false);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function confirmLeave() {
		if (!leaveTarget) return;
		leaving = true;
		try {
			await api.spaces.leave(leaveTarget.id);
			spaces = spaces.filter((s) => s.id !== leaveTarget!.id);
			spaceStore.spaces = spaces;
			if (spaceStore.activeId === leaveTarget.id) {
				spaceStore.activeId = null;
			}
			leaveTarget = null;
			toast.success('You left the space');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to leave space');
		}
		leaving = false;
	}

	onMount(async () => {
		try {
			spaces = await api.spaces.list();
			spaceStore.spaces = spaces;
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>Spaces — Plume</title></svelte:head>

<div class="flex items-center justify-between mb-8">
	<h1 class="text-2xl font-bold">Espaces</h1>
	<Button href="/spaces/new">
		<Icon icon="mdi:plus" class="h-4 w-4" />
		New space
	</Button>
</div>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if spaces.length === 0}
	<div class="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
		<Icon icon="solar:users-group-rounded-linear" class="h-10 w-10 text-muted-foreground mb-3" />
		<p class="text-muted-foreground">No spaces yet. Create one to collaborate with your team.</p>
		<Button href="/spaces/new" variant="outline" class="mt-4">
			<Icon icon="mdi:plus" class="h-4 w-4" />
			New space
		</Button>
	</div>
{:else}
	<div class="space-y-2">
		{#each spaces as space}
			<div class="flex items-center justify-between rounded-lg border p-4">
				<a
					href="/spaces/{space.id}"
					class="flex items-center gap-3 min-w-0 flex-1 hover:underline"
				>
					<Icon icon="solar:users-group-rounded-linear" class="h-5 w-5 shrink-0 text-muted-foreground" />
					<div class="min-w-0">
						<p class="font-medium truncate">{space.name}</p>
						<p class="text-sm text-muted-foreground">{formatDate(space.created_at)}</p>
					</div>
				</a>
				<div class="flex items-center gap-3 shrink-0">
					<span class="rounded-full px-2.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
						{space.role}
					</span>
					{#if space.role !== 'owner'}
						<button
							onclick={() => (leaveTarget = space)}
							class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
							aria-label="Leave space"
						>
							<Icon icon="solar:logout-2-linear" class="h-4 w-4" />
						</button>
					{/if}
				</div>
			</div>
		{/each}
	</div>
{/if}

<AlertDialog.Root open={leaveTarget !== null} onOpenChange={(open) => { if (!open) leaveTarget = null; }}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Leave space</AlertDialog.Title>
			<AlertDialog.Description>
				Are you sure you want to leave <strong>{leaveTarget?.name}</strong>? You will lose access to all documents in this space.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={leaving}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				class="!bg-red-600 !text-white hover:!bg-red-700"
				onclick={confirmLeave}
				disabled={leaving}
			>
				{#if leaving}
					<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
					Leaving...
				{:else}
					Leave
				{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
