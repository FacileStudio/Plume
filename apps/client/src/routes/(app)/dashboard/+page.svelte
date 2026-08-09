<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Document, DocumentStats } from '$lib';
	import {
		Badge,
		BarChart,
		Button,
		Card,
		DonutChart,
		EmptyState,
		Skeleton,
		StatCard,
		Table,
		icons
	} from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';

	let documents = $state<Document[]>([]);
	let stats = $state<DocumentStats>({ total: 0, pending: 0, completed: 0 });
	let loading = $state(true);

	const statusTone = {
		draft: 'neutral',
		pending: 'info',
		completed: 'success',
		declined: 'danger'
	} as const;

	const statusLabel = {
		draft: 'Draft',
		pending: 'Pending',
		completed: 'Completed',
		declined: 'Declined'
	} as const;

	const STATUS_ORDER = ['draft', 'pending', 'completed', 'declined'] as const;

	const recent = $derived(documents.slice(0, 10));

	const byStatus = $derived(
		STATUS_ORDER.map((status) => ({
			label: statusLabel[status],
			value: documents.filter((d) => d.status === status).length
		})).filter((slice) => slice.value > 0)
	);

	const months = $derived.by(() => {
		const now = new Date();
		return Array.from({ length: 6 }, (_, i) => {
			const d = new Date(now.getFullYear(), now.getMonth() - (5 - i), 1);
			return { key: `${d.getFullYear()}-${d.getMonth()}`, label: d.toLocaleDateString('en-US', { month: 'short' }) };
		});
	});

	const perMonth = $derived(
		months.map(
			(m) =>
				documents.filter((doc) => {
					const d = new Date(doc.created_at);
					return `${d.getFullYear()}-${d.getMonth()}` === m.key;
				}).length
		)
	);

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

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-1">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Dashboard</h1>
			<p class="text-fc-sm text-fc-fg-muted">Everything waiting on a signature, at a glance.</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Button variant="outline" href="/verify" icon={icons.shield}>Verify a document</Button>
			<Button href="/documents/new" icon={icons.plus}>New document</Button>
		</div>
	</div>

	{#if loading}
		<div class="flex flex-col gap-10">
			<div class="grid gap-4 sm:grid-cols-3">
				{#each [0, 1, 2] as tile (tile)}
					<Skeleton class="h-24 w-full" />
				{/each}
			</div>
			<div class="flex flex-col gap-2">
				{#each [0, 1, 2, 3] as row (row)}
					<Skeleton class="h-14 w-full" />
				{/each}
			</div>
		</div>
	{:else}
		<section class="grid gap-4 sm:grid-cols-3">
			<StatCard label="Total documents" value={stats.total} />
			<StatCard label="Pending signatures" value={stats.pending} />
			<StatCard label="Completed" value={stats.completed} />
		</section>

		{#if documents.length > 0}
			<section class="grid gap-4 lg:grid-cols-2">
				<Card class="flex flex-col gap-4">
					<p class="text-fc-sm font-medium text-fc-fg">Documents created</p>
					<BarChart
						series={[{ name: 'Documents', data: perMonth }]}
						labels={months.map((m) => m.label)}
						height={220}
						yFormat={(n) => `${n}`}
					/>
				</Card>
				<Card class="flex flex-col gap-4">
					<p class="text-fc-sm font-medium text-fc-fg">By status</p>
					<DonutChart
						data={byStatus}
						centerLabel="documents"
						centerValue={documents.length}
						valueFormat={(n) => `${n}`}
						class="flex-1"
					/>
				</Card>
			</section>
		{/if}

		<section class="flex flex-col gap-4">
			<div class="flex flex-wrap items-end justify-between gap-4">
				<div class="flex min-w-0 flex-col gap-1">
					<h2 class="text-fc-lg font-semibold text-fc-fg">Recent documents</h2>
					<p class="text-fc-sm text-fc-fg-muted">The ten most recent in this space.</p>
				</div>
				{#if documents.length > recent.length}
					<Button variant="outline" href="/documents" iconRight={icons.arrow}>
						View all documents
					</Button>
				{/if}
			</div>

			{#if recent.length === 0}
				<EmptyState
					icon={icons.folder}
					title="No documents yet"
					description="Upload a PDF, drop a few fields on it, and send it out for signature."
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
							</tr>
						</thead>
						<tbody>
							{#each recent as doc (doc.id)}
								<tr>
									<td>
										<a
											href="/documents/{doc.id}"
											class="block truncate font-medium text-fc-fg hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
										>
											{doc.name}
										</a>
									</td>
									<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(doc.created_at)}</td>
									<td class="tabular-nums text-fc-fg-muted">
										{doc.signer_count ?? 0}
									</td>
									<td><Badge tone={statusTone[doc.status]}>{statusLabel[doc.status]}</Badge></td>
								</tr>
							{/each}
						</tbody>
					</Table>
				</div>

				<div class="flex flex-col gap-2 md:hidden">
					{#each recent as doc (doc.id)}
						<Card href="/documents/{doc.id}" class="flex items-center justify-between gap-3">
							<div class="min-w-0">
								<p class="truncate text-fc-sm font-medium text-fc-fg">{doc.name}</p>
								<p class="truncate text-fc-xs text-fc-fg-muted">
									{formatDate(doc.created_at)}
									{#if doc.signer_count !== undefined}
										· {doc.signer_count} signer{doc.signer_count === 1 ? '' : 's'}
									{/if}
								</p>
							</div>
							<Badge tone={statusTone[doc.status]}>{statusLabel[doc.status]}</Badge>
						</Card>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</div>
