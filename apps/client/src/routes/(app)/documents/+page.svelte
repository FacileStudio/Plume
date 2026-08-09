<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Document } from '$lib';
	import {
		Badge,
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Skeleton,
		Table,
		icons,
		toast
	} from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';

	const statusTone = {
		draft: 'neutral',
		pending: 'info',
		completed: 'success',
		declined: 'danger'
	} as const;

	let documents = $state<Document[]>([]);
	let loading = $state(true);
	let mounted = $state(false);
	let confirmDelete = $state(false);
	let pendingDelete = $state<Document | null>(null);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function signerLabel(count: number): string {
		return `${count} signer${count === 1 ? '' : 's'}`;
	}

	function askDelete(doc: Document) {
		pendingDelete = doc;
		confirmDelete = true;
	}

	async function runDelete() {
		const target = pendingDelete;
		if (!target) return;
		try {
			await api.documents.delete(target.id);
			documents = documents.filter((d) => d.id !== target.id);
			pendingDelete = null;
			toast.success('Document deleted');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to delete the document');
			throw e;
		}
	}

	async function load() {
		loading = true;
		try {
			documents = await api.documents.list(spaceStore.activeId);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to load documents');
		}
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

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-1">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Documents</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				Everything you have sent out for signature, and everything still in draft.
			</p>
		</div>
		<Button href="/documents/new" icon={icons.plus}>New document</Button>
	</div>

	<section class="flex flex-col gap-4">
		{#if loading}
			<div class="flex flex-col gap-2">
				{#each [0, 1, 2, 3] as row (row)}
					<Skeleton class="h-14 w-full" />
				{/each}
			</div>
		{:else if documents.length === 0}
			<EmptyState
				icon="solar:document-text-linear"
				title="No documents yet"
				description="Upload a PDF, place the fields, and send it out for signature."
			>
				<Button href="/documents/new" icon={icons.plus}>New document</Button>
			</EmptyState>
		{:else}
			<div class="hidden md:block">
				<Table>
					<thead>
						<tr>
							<th scope="col">Document</th>
							<th scope="col">Created</th>
							<th scope="col">Signers</th>
							<th scope="col">Status</th>
							<th aria-label="Actions"></th>
						</tr>
					</thead>
					<tbody>
						{#each documents as doc (doc.id)}
							<tr>
								<td>
									<a
										href="/documents/{doc.id}"
										class="flex min-w-0 items-center gap-2.5 text-fc-fg hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
									>
										<iconify-icon
											icon="solar:document-text-linear"
											width="16"
											height="16"
											class="block shrink-0 text-fc-fg-muted"
										></iconify-icon>
										<span class="truncate font-medium">{doc.name}</span>
									</a>
								</td>
								<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(doc.created_at)}</td>
								<td class="whitespace-nowrap text-fc-fg-muted">
									{doc.signer_count === undefined ? '—' : signerLabel(doc.signer_count)}
								</td>
								<td><Badge tone={statusTone[doc.status]}>{doc.status}</Badge></td>
								<td>
									<div class="flex justify-end">
										<Button
											variant="ghost-danger"
											size="sm"
											icon={icons.remove}
											aria-label="Delete {doc.name}"
											onclick={() => askDelete(doc)}
										>
											Delete
										</Button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</Table>
			</div>

			<div class="flex flex-col gap-2 md:hidden">
				{#each documents as doc (doc.id)}
					<Card class="flex flex-col gap-3">
						<div class="flex min-w-0 items-start justify-between gap-3">
							<a href="/documents/{doc.id}" class="flex min-w-0 flex-col gap-1">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{doc.name}</span>
								<span class="text-fc-xs text-fc-fg-muted">
									{formatDate(doc.created_at)}
									{#if doc.signer_count !== undefined}
										· {signerLabel(doc.signer_count)}
									{/if}
								</span>
							</a>
							<Badge tone={statusTone[doc.status]}>{doc.status}</Badge>
						</div>
						<div class="flex justify-end">
							<Button
								variant="ghost-danger"
								size="lg"
								icon={icons.remove}
								aria-label="Delete {doc.name}"
								onclick={() => askDelete(doc)}
							>
								Delete
							</Button>
						</div>
					</Card>
				{/each}
			</div>
		{/if}
	</section>
</div>

<ConfirmModal
	bind:open={confirmDelete}
	tone="danger"
	title="Delete {pendingDelete?.name ?? 'this document'}?"
	description="The document, its signers and every field placed on it are removed permanently. This cannot be undone."
	confirmLabel="Delete document"
	onConfirm={runDelete}
	onCancel={() => (pendingDelete = null)}
/>
