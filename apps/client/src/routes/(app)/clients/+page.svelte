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
			toast.danger(e instanceof Error ? e.message : 'Failed to delete client');
			throw e;
		}
		clients = clients.filter((c) => c.id !== target.id);
		pendingDelete = null;
		toast.success('Client deleted');
	}

	onMount(async () => {
		try {
			clients = await api.clients.list();
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>Clients — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Clients</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				The people and companies you send documents to.
			</p>
		</div>
		<Button href="/clients/new" icon={icons.plus}>New client</Button>
	</div>

	{#if loading}
		<div class="flex min-h-[40dvh] items-center justify-center">
			<Spinner size="lg" />
		</div>
	{:else if clients.length === 0}
		<EmptyState
			icon={icons.usersGroup}
			title="No clients yet"
			description="Add your first one and it will show up here, along with every document linked to it."
		>
			<Button href="/clients/new" variant="outline" icon={icons.plus}>New client</Button>
		</EmptyState>
	{:else}
		<div class="hidden md:block">
			<Table>
				<thead>
					<tr>
						<th scope="col">Client</th>
						<th scope="col">Company</th>
						<th scope="col">Phone</th>
						<th aria-label="Actions"></th>
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
										{client.email || 'No email'}
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
									aria-label="Delete {client.name}"
									onclick={() => askDelete(client)}
								>
									Delete
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
							{client.email || client.company || 'No contact details'}
						</span>
					</a>
					<div class="flex items-center justify-end">
						<Button
							variant="ghost-danger"
							size="lg"
							icon={icons.remove}
							aria-label="Delete {client.name}"
							onclick={() => askDelete(client)}
						>
							Delete
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
	title="Delete {pendingDelete?.name ?? 'this client'}?"
	description="This cannot be undone. Documents linked to this client are kept, but unlinked."
	confirmLabel="Delete client"
	cancelLabel="Keep client"
	onConfirm={runDelete}
	onCancel={() => (pendingDelete = null)}
/>
