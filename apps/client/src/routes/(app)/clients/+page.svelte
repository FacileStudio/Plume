<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Client } from '$lib';
	import {
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Spinner,
		Table,
		icons,
		toast
	} from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let clients = $state<Client[]>([]);
	let loading = $state(true);
	let confirmDelete = $state(false);
	let pendingDelete = $state<Client | null>(null);

	function askDelete(client: Client) {
		pendingDelete = client;
		confirmDelete = true;
	}

	async function runDelete() {
		const target = pendingDelete;
		if (!target) return;
		try {
			await api.clients.delete(target.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('clients.deleteFailed'));
			throw e;
		}
		clients = clients.filter((c) => c.id !== target.id);
		pendingDelete = null;
		toast.success(t('clients.deleted'));
	}

	onMount(async () => {
		try {
			clients = await api.clients.list();
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>{t('clients.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('clients.heading')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">{t('clients.subtitle')}</p>
		</div>
		<Button href="/clients/new" icon={icons.plus}>{t('clients.new')}</Button>
	</div>

	{#if loading}
		<div class="flex min-h-[40dvh] items-center justify-center">
			<Spinner size="lg" />
		</div>
	{:else if clients.length === 0}
		<EmptyState
			icon={icons.usersGroup}
			title={t('clients.empty.title')}
			description={t('clients.empty.text')}
		>
			<Button href="/clients/new" variant="secondary" icon={icons.plus}>{t('clients.new')}</Button>
		</EmptyState>
	{:else}
		<div class="hidden md:block">
			<Table>
				<thead>
					<tr>
						<th scope="col">{t('clients.col.client')}</th>
						<th scope="col">{t('common.company')}</th>
						<th scope="col">{t('common.phone')}</th>
						<th aria-label={t('common.actions')}></th>
					</tr>
				</thead>
				<tbody>
					{#each clients as client (client.id)}
						<tr>
							<td>
								<a
									href="/clients/{client.id}"
									class="flex min-w-0 flex-col rounded-fc-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
								>
									<span class="truncate font-medium text-fc-fg">{client.name}</span>
									<span class="truncate text-fc-xs text-fc-fg-muted">
										{client.email || t('clients.noEmail')}
									</span>
								</a>
							</td>
							<td class="text-fc-fg-muted">{client.company || '—'}</td>
							<td class="text-fc-fg-muted">{client.phone || '—'}</td>
							<td class="text-right">
								<Button
									variant="ghost-danger"
									size="sm"
									icon={icons.remove}
									aria-label={t('documents.deleteAria', { name: client.name })}
									onclick={() => askDelete(client)}
								>
									{t('common.delete')}
								</Button>
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		</div>

		<div class="flex flex-col gap-2 md:hidden">
			{#each clients as client (client.id)}
				<Card class="flex flex-col gap-3">
					<a
						href="/clients/{client.id}"
						class="flex min-w-0 flex-col rounded-fc-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
					>
						<span class="truncate text-fc-sm font-medium text-fc-fg">{client.name}</span>
						<span class="truncate text-fc-xs text-fc-fg-muted">
							{client.email || client.company || t('clients.noContact')}
						</span>
					</a>
					<div class="flex items-center justify-end">
						<Button
							variant="ghost-danger"
							size="lg"
							icon={icons.remove}
							aria-label={t('documents.deleteAria', { name: client.name })}
							onclick={() => askDelete(client)}
						>
							{t('common.delete')}
						</Button>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmDelete}
	tone="danger"
	title={t('clients.confirm.title', { name: pendingDelete?.name ?? t('clients.confirm.thisClient') })}
	description={t('clients.confirm.text')}
	confirmLabel={t('clients.confirm.label')}
	cancelLabel={t('clients.confirm.cancel')}
	onConfirm={runDelete}
	onCancel={() => (pendingDelete = null)}
/>
