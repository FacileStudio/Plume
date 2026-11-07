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
	import { formatShortDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

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
			toast.success(t('documents.deleted'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('documents.deleteFailed'));
			throw e;
		}
	}

	async function load() {
		loading = true;
		try {
			documents = await api.documents.list(spaceStore.activeId);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('documents.loadFailed'));
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

<svelte:head><title>{t('documents.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-1">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('documents.heading')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">{t('documents.subtitle')}</p>
		</div>
		<Button href="/documents/new" icon={icons.plus}>{t('dashboard.newDocument')}</Button>
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
				title={t('documents.empty.title')}
				description={t('documents.empty.text')}
			>
				<Button href="/documents/new" icon={icons.plus}>{t('dashboard.newDocument')}</Button>
			</EmptyState>
		{:else}
			<div class="hidden md:block">
				<Table>
					<thead>
						<tr>
							<th scope="col">{t('documents.col.document')}</th>
							<th scope="col">{t('common.created')}</th>
							<th scope="col">{t('documents.col.signers')}</th>
							<th scope="col">{t('common.status')}</th>
							<th aria-label={t('common.actions')}></th>
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
								<td class="whitespace-nowrap text-fc-fg-muted">{formatShortDate(doc.created_at)}</td>
								<td class="whitespace-nowrap text-fc-fg-muted">
									{doc.signer_count === undefined ? '—' : t('documents.signerCount', { count: doc.signer_count })}
								</td>
								<td><Badge tone={statusTone[doc.status]}>{t(`status.${doc.status}`)}</Badge></td>
								<td>
									<div class="flex justify-end">
										<Button
											variant="ghost-danger"
											size="sm"
											icon={icons.remove}
											aria-label={t('documents.deleteAria', { name: doc.name })}
											onclick={() => askDelete(doc)}
										>
											{t('common.delete')}
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
									{formatShortDate(doc.created_at)}
									{#if doc.signer_count !== undefined}
										· {t('documents.signerCount', { count: doc.signer_count })}
									{/if}
								</span>
							</a>
							<Badge tone={statusTone[doc.status]}>{t(`status.${doc.status}`)}</Badge>
						</div>
						<div class="flex justify-end">
							<Button
								variant="ghost-danger"
								size="lg"
								icon={icons.remove}
								aria-label={t('documents.deleteAria', { name: doc.name })}
								onclick={() => askDelete(doc)}
							>
								{t('common.delete')}
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
	title={t('documents.confirm.title', { name: pendingDelete?.name ?? t('documents.confirm.thisDocument') })}
	description={t('documents.confirm.text')}
	confirmLabel={t('documents.confirm.label')}
	onConfirm={runDelete}
	onCancel={() => (pendingDelete = null)}
/>
