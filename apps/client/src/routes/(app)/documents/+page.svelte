<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Document } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Icon from '@iconify/svelte';
	import { spaceStore } from '$lib/stores/space.svelte';

	let documents = $state<Document[]>([]);
	let loading = $state(true);
	let deleteTarget = $state<Document | null>(null);
	let deleting = $state(false);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		deleting = true;
		try {
			await api.documents.delete(deleteTarget.id);
			documents = documents.filter((d) => d.id !== deleteTarget!.id);
			deleteTarget = null;
		} catch {}
		deleting = false;
	}

	let mounted = $state(false);

	async function load() {
		loading = true;
		try {
			documents = await api.documents.list(spaceStore.activeId);
		} catch {}
		loading = false;
	}

	onMount(async () => {
		await load();
		mounted = true;
	});

	$effect(() => {
		spaceStore.activeId;
		if (mounted) load();
	});
</script>

<svelte:head><title>Documents — Plume</title></svelte:head>

<div class="flex items-center justify-between mb-8">
	<h1 class="text-2xl font-bold">Documents</h1>
	<Button href="/documents/new">
		<Icon icon="mdi:plus" class="h-4 w-4" />
		New document
	</Button>
</div>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if documents.length === 0}
	<div class="flex flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center">
		<Icon icon="solar:document-linear" class="h-10 w-10 text-muted-foreground mb-3" />
		<p class="text-muted-foreground">No documents yet. Create your first one.</p>
		<Button href="/documents/new" variant="outline" class="mt-4">
			<Icon icon="mdi:plus" class="h-4 w-4" />
			New document
		</Button>
	</div>
{:else}
	<div class="space-y-2">
		{#each documents as doc}
			<div class="flex items-center justify-between rounded-lg border p-4">
				<a
					href="/documents/{doc.id}"
					class="flex items-center gap-3 min-w-0 flex-1 hover:underline"
				>
					<Icon icon="solar:document-text-linear" class="h-5 w-5 shrink-0 text-muted-foreground" />
					<div class="min-w-0">
						<p class="font-medium truncate">{doc.name}</p>
						<p class="text-sm text-muted-foreground">{formatDate(doc.created_at)}</p>
					</div>
				</a>
				<div class="flex items-center gap-3 shrink-0">
					{#if doc.signer_count !== undefined}
						<span class="text-sm text-muted-foreground">{doc.signer_count} signer{doc.signer_count === 1 ? '' : 's'}</span>
					{/if}
					<span class="rounded-full px-2.5 py-0.5 text-xs font-medium
						{doc.status === 'draft' ? 'bg-muted text-muted-foreground' : ''}
						{doc.status === 'pending' ? 'bg-foreground/10 text-foreground' : ''}
						{doc.status === 'completed' ? 'bg-green-500/10 text-green-700 dark:text-green-400' : ''}
						{doc.status === 'declined' ? 'bg-red-500/10 text-red-700 dark:text-red-400' : ''}
					">{doc.status}</span>
					<button
						onclick={() => (deleteTarget = doc)}
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
			<AlertDialog.Title>Delete document</AlertDialog.Title>
			<AlertDialog.Description>
				Are you sure you want to delete <strong>{deleteTarget?.name}</strong>? This action cannot be undone. The document and all associated signers and fields will be permanently removed.
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
