<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Client, Document } from '$lib';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	let client = $state<Client | null>(null);
	let documents = $state<Document[]>([]);
	let loading = $state(true);

	let editing = $state(false);
	let saving = $state(false);
	let editName = $state('');
	let editEmail = $state('');
	let editCompany = $state('');
	let editPhone = $state('');
	let editNotes = $state('');

	let deleteOpen = $state(false);
	let deleting = $state(false);

	const clientId = $derived(Number(page.params.id));

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function startEdit() {
		if (!client) return;
		editName = client.name;
		editEmail = client.email;
		editCompany = client.company;
		editPhone = client.phone;
		editNotes = client.notes;
		editing = true;
	}

	function cancelEdit() {
		editing = false;
	}

	async function saveEdit() {
		if (!client) return;
		if (!editName.trim()) {
			toast.error('Client name is required');
			return;
		}
		saving = true;
		try {
			client = await api.clients.update(client.id, {
				name: editName.trim(),
				email: editEmail.trim(),
				company: editCompany.trim(),
				phone: editPhone.trim(),
				notes: editNotes.trim()
			});
			editing = false;
			toast.success('Client updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update client');
		}
		saving = false;
	}

	async function confirmDelete() {
		if (!client) return;
		deleting = true;
		try {
			await api.clients.delete(client.id);
			toast.success('Client deleted');
			goto('/clients');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete client');
			deleting = false;
		}
	}

	onMount(async () => {
		try {
			const [c, d] = await Promise.all([
				api.clients.get(clientId),
				api.clients.documents(clientId)
			]);
			client = c;
			documents = d;
		} catch {
			goto('/clients');
			return;
		}
		loading = false;
	});
</script>

<svelte:head><title>{client?.name ?? 'Client'} — Plume</title></svelte:head>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if client}
	<div class="mb-6 border-b pb-5">
		<div class="flex items-center gap-3 mb-3">
			<a href="/clients" class="rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted">
				<Icon icon="solar:arrow-left-linear" class="h-5 w-5" />
			</a>
			<span class="text-sm text-muted-foreground">Clients</span>
		</div>
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3 min-w-0">
				<div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full border border-border bg-muted text-sm font-semibold">
					{client.name.charAt(0).toUpperCase()}
				</div>
				<div class="min-w-0">
					<h1 class="text-2xl font-bold truncate">{client.name}</h1>
					{#if client.company}
						<p class="text-sm text-muted-foreground truncate">{client.company}</p>
					{/if}
				</div>
			</div>
			<div class="flex items-center gap-2 shrink-0">
				{#if !editing}
					<Button variant="outline" size="sm" onclick={startEdit}>
						<Icon icon="solar:pen-linear" class="h-4 w-4" />
						Edit
					</Button>
				{/if}
				<Button variant="outline" size="sm" onclick={() => (deleteOpen = true)}>
					<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
					Delete
				</Button>
			</div>
		</div>
	</div>

	<div class="space-y-6">
		<Card.Root>
			<Card.Header>
				<div class="flex items-start justify-between gap-4">
					<div>
						<Card.Title>Details</Card.Title>
						<Card.Description>Contact information and notes.</Card.Description>
					</div>
					{#if editing}
						<div class="flex items-center gap-2">
							<Button size="sm" onclick={saveEdit} disabled={saving}>
								{#if saving}
									<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
								{:else}
									<Icon icon="solar:check-circle-linear" class="h-4 w-4" />
								{/if}
								Save
							</Button>
							<Button variant="ghost" size="sm" onclick={cancelEdit} disabled={saving}>
								Cancel
							</Button>
						</div>
					{/if}
				</div>
			</Card.Header>
			<Card.Content>
				{#if editing}
					<div class="space-y-4">
						<div class="space-y-2">
							<Label for="edit-name">Name</Label>
							<Input id="edit-name" bind:value={editName} placeholder="Jane Doe" />
						</div>
						<div class="space-y-2">
							<Label for="edit-email">Email</Label>
							<Input id="edit-email" bind:value={editEmail} placeholder="jane@example.com" type="email" />
						</div>
						<div class="space-y-2">
							<Label for="edit-company">Company</Label>
							<Input id="edit-company" bind:value={editCompany} placeholder="Acme Inc." />
						</div>
						<div class="space-y-2">
							<Label for="edit-phone">Phone</Label>
							<Input id="edit-phone" bind:value={editPhone} placeholder="+1 555 000 0000" type="tel" />
						</div>
						<div class="space-y-2">
							<Label for="edit-notes">Notes</Label>
							<textarea
								id="edit-notes"
								bind:value={editNotes}
								placeholder="Anything worth remembering about this client..."
								rows="4"
								class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 min-h-16 w-full min-w-0 rounded-lg border bg-transparent px-2.5 py-1.5 text-base outline-none transition-colors placeholder:text-muted-foreground focus-visible:ring-3 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
							></textarea>
						</div>
					</div>
				{:else}
					<div class="space-y-2.5 text-sm">
						<div class="flex justify-between gap-4">
							<span class="text-muted-foreground">Email</span>
							<span class="text-right">{client.email || '—'}</span>
						</div>
						<div class="flex justify-between gap-4">
							<span class="text-muted-foreground">Company</span>
							<span class="text-right">{client.company || '—'}</span>
						</div>
						<div class="flex justify-between gap-4">
							<span class="text-muted-foreground">Phone</span>
							<span class="text-right">{client.phone || '—'}</span>
						</div>
						<div class="flex flex-col gap-1 pt-1">
							<span class="text-muted-foreground">Notes</span>
							<span class="whitespace-pre-wrap">{client.notes || '—'}</span>
						</div>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Documents</Card.Title>
				<Card.Description>
					{documents.length} document{documents.length === 1 ? '' : 's'} linked to this client
				</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if documents.length === 0}
					<p class="text-sm text-muted-foreground">No documents linked yet.</p>
				{:else}
					<div class="space-y-2">
						{#each documents as doc}
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
								<span class="rounded-full px-2.5 py-0.5 text-xs font-medium shrink-0
									{doc.status === 'draft' ? 'bg-muted text-muted-foreground' : ''}
									{doc.status === 'pending' ? 'bg-foreground/10 text-foreground' : ''}
									{doc.status === 'completed' ? 'bg-green-500/10 text-green-700 dark:text-green-400' : ''}
									{doc.status === 'declined' ? 'bg-red-500/10 text-red-700 dark:text-red-400' : ''}
								">{doc.status}</span>
							</a>
						{/each}
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>
{/if}

<AlertDialog.Root open={deleteOpen} onOpenChange={(open) => (deleteOpen = open)}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete client</AlertDialog.Title>
			<AlertDialog.Description>
				Are you sure you want to delete <strong>{client?.name}</strong>? This action cannot be undone. Documents linked to this client will be unlinked.
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
