<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Client } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	let clients = $state<Client[]>([]);
	let loading = $state(true);
	let deleteTarget = $state<Client | null>(null);
	let deleting = $state(false);

	async function confirmDelete() {
		if (!deleteTarget) return;
		deleting = true;
		try {
			await api.clients.delete(deleteTarget.id);
			clients = clients.filter((c) => c.id !== deleteTarget!.id);
			deleteTarget = null;
			toast.success('Client deleted');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete client');
		}
		deleting = false;
	}

	onMount(async () => {
		try {
			clients = await api.clients.list();
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>Clients — Plume</title></svelte:head>

<div class="mb-6 flex items-center justify-between border-b pb-5">
	<h1 class="text-2xl font-bold">Clients</h1>
	<Button href="/clients/new">
		<Icon icon="mdi:plus" class="h-4 w-4" />
		New client
	</Button>
</div>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if clients.length === 0}
	<div class="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
		<Icon icon="solar:users-group-two-rounded-linear" class="h-10 w-10 text-muted-foreground mb-3" />
		<p class="text-muted-foreground">No clients yet. Add your first one.</p>
		<Button href="/clients/new" variant="outline" class="mt-4">
			<Icon icon="mdi:plus" class="h-4 w-4" />
			New client
		</Button>
	</div>
{:else}
	<div class="space-y-2">
		{#each clients as c}
			<div class="flex items-center justify-between rounded-lg border p-4">
				<a
					href="/clients/{c.id}"
					class="flex items-center gap-3 min-w-0 flex-1 hover:underline"
				>
					<Icon icon="solar:user-rounded-linear" class="h-5 w-5 shrink-0 text-muted-foreground" />
					<div class="min-w-0">
						<p class="font-medium truncate">{c.name}</p>
						<p class="text-sm text-muted-foreground truncate">
							{c.email || c.company || 'No contact details'}
						</p>
					</div>
				</a>
				<div class="flex items-center gap-3 shrink-0">
					<button
						onclick={() => (deleteTarget = c)}
						class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
					>
						<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
					</button>
				</div>
			</div>
		{/each}
	</div>
{/if}

<AlertDialog.Root open={deleteTarget !== null} onOpenChange={(open) => { if (!open) deleteTarget = null; }}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete client</AlertDialog.Title>
			<AlertDialog.Description>
				Are you sure you want to delete <strong>{deleteTarget?.name}</strong>? This action cannot be undone. Documents linked to this client will be unlinked.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={deleting}>
				<Icon icon="solar:close-circle-linear" class="h-4 w-4" />
				Cancel
			</AlertDialog.Cancel>
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
