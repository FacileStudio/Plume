<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';
	import { spaceStore } from '$lib/stores/space.svelte';

	let space = $state<Space | null>(null);
	let loading = $state(true);

	let name = $state('');
	let description = $state('');
	let saving = $state(false);

	let showDeleteConfirm = $state(false);
	let deleting = $state(false);

	const spaceId = $derived(Number(page.params.id));

	async function save() {
		if (!name.trim()) return;
		saving = true;
		try {
			const updated = await api.spaces.update(spaceId, { name: name.trim(), description: description.trim() });
			space = updated;
			spaceStore.spaces = spaceStore.spaces.map((s) => (s.id === spaceId ? updated : s));
			toast.success('Space updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update space');
		}
		saving = false;
	}

	async function confirmDelete() {
		deleting = true;
		try {
			await api.spaces.delete(spaceId);
			spaceStore.spaces = spaceStore.spaces.filter((s) => s.id !== spaceId);
			if (spaceStore.activeId === spaceId) {
				spaceStore.activeId = null;
			}
			toast.success('Space deleted');
			goto('/spaces');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete space');
		}
		deleting = false;
	}

	onMount(async () => {
		try {
			space = await api.spaces.get(spaceId);
			if (space.role !== 'owner' && space.role !== 'admin') {
				goto(`/spaces/${spaceId}`);
				return;
			}
			name = space.name;
			description = space.description;
		} catch {
			goto('/spaces');
			return;
		}
		loading = false;
	});
</script>

<svelte:head><title>Settings — {space?.name ?? 'Space'} — Plume</title></svelte:head>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if space}
	<div class="mb-8">
		<div class="flex items-center gap-2 text-sm text-muted-foreground mb-2">
			<a href="/spaces" class="hover:text-foreground">Espaces</a>
			<Icon icon="solar:alt-arrow-right-linear" class="h-3.5 w-3.5" />
			<a href="/spaces/{spaceId}" class="hover:text-foreground">{space.name}</a>
			<Icon icon="solar:alt-arrow-right-linear" class="h-3.5 w-3.5" />
			<span>Settings</span>
		</div>
		<h1 class="text-2xl font-bold">Space settings</h1>
	</div>

	<div class="space-y-8 max-w-lg">
		<div class="rounded-lg border p-6 space-y-4">
			<h2 class="text-lg font-semibold">General</h2>
			<div>
				<label for="space-name" class="block text-sm font-medium mb-1">Name</label>
				<input
					id="space-name"
					type="text"
					bind:value={name}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="space-description" class="block text-sm font-medium mb-1">Description</label>
				<textarea
					id="space-description"
					bind:value={description}
					rows="3"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
				></textarea>
			</div>
			<div class="flex gap-2 pt-1">
				<Button onclick={save} disabled={saving || !name.trim()}>
					{#if saving}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
						Saving...
					{:else}
						<Icon icon="solar:diskette-linear" class="h-4 w-4" />
						Save
					{/if}
				</Button>
			</div>
		</div>

		{#if space.role === 'owner'}
			<div class="rounded-lg border border-red-500/30 p-6 space-y-4">
				<h2 class="text-lg font-semibold text-red-500">Danger zone</h2>
				<p class="text-sm text-muted-foreground">
					Deleting a space is permanent. All members will lose access.
				</p>
				<Button
					variant="outline"
					class="border-red-500/50 text-red-500 hover:bg-red-500 hover:text-white"
					onclick={() => (showDeleteConfirm = true)}
				>
					<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
					Delete space
				</Button>
			</div>
		{/if}
	</div>

	<AlertDialog.Root open={showDeleteConfirm} onOpenChange={(open) => { if (!open) showDeleteConfirm = false; }}>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Delete space</AlertDialog.Title>
				<AlertDialog.Description>
					Are you sure you want to delete <strong>{space.name}</strong>? This action cannot be undone. All members will lose access.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel disabled={deleting}>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					class="!bg-red-600 !text-white hover:!bg-red-700"
					onclick={confirmDelete}
					disabled={deleting}
				>
					{#if deleting}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
						Deleting...
					{:else}
						<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
						Delete
					{/if}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
