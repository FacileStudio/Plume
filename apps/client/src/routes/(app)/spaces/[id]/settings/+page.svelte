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
		<Icon icon="solar:spinner-bold-duotone" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if space}
	<div class="border-b px-4 py-4 md:px-8 md:py-5">
		<div class="flex items-center gap-3 mb-3">
			<a href="/spaces/{spaceId}" class="rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted">
				<Icon icon="solar:arrow-left-linear" class="h-5 w-5" />
			</a>
			<span class="text-sm text-muted-foreground">{space.name}</span>
		</div>
		<h1 class="text-lg font-semibold">Settings</h1>
	</div>

	<div class="p-4 md:p-8">
		<form onsubmit={save} class="max-w-xl space-y-6">
			<div>
				<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-4">General</h2>
				<div class="space-y-4">
					<div>
						<label for="space-name" class="mb-1.5 block text-sm font-medium">Name</label>
						<input
							id="space-name"
							type="text"
							bind:value={name}
							class="h-10 w-full rounded-lg border border-border bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
						/>
					</div>
					<div>
						<label for="space-description" class="mb-1.5 block text-sm font-medium">Description</label>
						<textarea
							id="space-description"
							bind:value={description}
							rows="3"
							class="w-full rounded-lg border border-border bg-background px-3 py-2.5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
						></textarea>
					</div>
				</div>
				<div class="mt-4">
					<Button type="submit" disabled={saving || !name.trim()} size="sm">
						{#if saving}
							<Icon icon="solar:spinner-bold-duotone" class="h-4 w-4 animate-spin" />
							Saving...
						{:else}
							Save changes
						{/if}
					</Button>
				</div>
			</div>

			{#if space.role === 'owner'}
				<div class="border-t border-border pt-8">
					<h2 class="text-sm font-semibold uppercase tracking-wider text-red-500 mb-3">Danger zone</h2>
					<p class="text-sm text-muted-foreground mb-4">
						Deleting a space is permanent. All members will lose access.
					</p>
					<Button
						variant="outline"
						size="sm"
						class="border-red-500/50 text-red-500 hover:bg-red-500 hover:text-white"
						onclick={() => (showDeleteConfirm = true)}
					>
						<Icon icon="solar:trash-bin-trash-bold-duotone" class="h-4 w-4" />
						Delete space
					</Button>
				</div>
			{/if}
		</form>
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
						<Icon icon="solar:spinner-bold-duotone" class="h-4 w-4 animate-spin" />
						Deleting...
					{:else}
						Delete
					{/if}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
