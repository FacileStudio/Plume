<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Document, DocumentStats } from '$lib';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';
	import { spaceStore } from '$lib/stores/space.svelte';

	let documents = $state<Document[]>([]);
	let stats = $state<DocumentStats>({ total: 0, pending: 0, completed: 0 });
	let loading = $state(true);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	let mounted = $state(false);

	async function load() {
		loading = true;
		try {
			const sid = spaceStore.activeId;
			const [docs, s] = await Promise.all([api.documents.list(sid), api.documents.stats(sid)]);
			documents = docs;
			stats = s;
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

<svelte:head><title>Dashboard — Plume</title></svelte:head>

<div class="mb-6 flex items-center justify-between border-b pb-5">
	<h1 class="text-2xl font-bold">Dashboard</h1>
	<div class="flex items-center gap-2">
		<Button variant="outline" href="/verify">
			<Icon icon="solar:shield-check-linear" class="h-4 w-4" />
			Verify a document
		</Button>
		<Button href="/documents/new">
			<Icon icon="mdi:plus" class="h-4 w-4" />
			New document
		</Button>
	</div>
</div>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else}
	<div class="grid gap-4 md:grid-cols-3 mb-8">
		<Card.Root>
			<Card.Header>
				<Card.Description>Total Documents</Card.Description>
				<Card.Title class="text-3xl">{stats.total}</Card.Title>
			</Card.Header>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Description>Pending Signatures</Card.Description>
				<Card.Title class="text-3xl">{stats.pending}</Card.Title>
			</Card.Header>
		</Card.Root>
		<Card.Root>
			<Card.Header>
				<Card.Description>Completed</Card.Description>
				<Card.Title class="text-3xl">{stats.completed}</Card.Title>
			</Card.Header>
		</Card.Root>
	</div>

	<h2 class="text-lg font-semibold mb-4">Recent documents</h2>

	{#if documents.length === 0}
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
			{#each documents.slice(0, 10) as doc}
				<a
					href="/documents/{doc.id}"
					class="flex items-center justify-between rounded-lg border p-4 transition-colors hover:bg-muted/50"
				>
					<div class="flex items-center gap-3 min-w-0">
						<Icon icon="solar:document-text-linear" class="h-5 w-5 shrink-0 text-muted-foreground" />
						<div class="min-w-0">
							<p class="font-medium truncate">{doc.name}</p>
							<p class="text-sm text-muted-foreground">{formatDate(doc.created_at)}</p>
						</div>
					</div>
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
					</div>
				</a>
			{/each}
		</div>

		{#if documents.length > 10}
			<div class="mt-4 text-center">
				<Button href="/documents" variant="outline" size="sm">View all documents</Button>
			</div>
		{/if}
	{/if}
{/if}
