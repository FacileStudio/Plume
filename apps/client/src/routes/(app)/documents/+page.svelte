<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Document } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';

	let documents = $state<Document[]>([]);
	let loading = $state(true);
	let deletingId = $state<number | null>(null);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function deleteDocument(id: number) {
		try {
			await api.documents.delete(id);
			documents = documents.filter((d) => d.id !== id);
		} catch {}
		deletingId = null;
	}

	onMount(async () => {
		try {
			documents = await api.documents.list();
		} catch {}
		loading = false;
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
					{#if deletingId === doc.id}
						<div class="flex items-center gap-1.5">
							<button
								onclick={() => deleteDocument(doc.id)}
								class="rounded-full bg-red-500 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-red-600"
							>Delete</button>
							<button
								onclick={() => (deletingId = null)}
								class="rounded-full bg-muted px-3 py-1 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
							>Cancel</button>
						</div>
					{:else}
						<button
							onclick={() => (deletingId = doc.id)}
							class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
						>
							<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
						</button>
					{/if}
				</div>
			</div>
		{/each}
	</div>
{/if}
