<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Client, Document } from '$lib';
	import {
		Avatar,
		Badge,
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Field,
		Input,
		SettingsRow,
		SettingsSection,
		Spinner,
		Textarea,
		icons,
		toast
	} from '@facile/muse';

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

	let confirmDelete = $state(false);

	const clientId = $derived(Number(page.params.id));

	const statusTone = {
		draft: 'neutral',
		pending: 'warning',
		completed: 'success',
		declined: 'danger'
	} as const;

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

	async function saveEdit(event: SubmitEvent) {
		event.preventDefault();
		if (!client) return;
		if (!editName.trim()) {
			toast.danger('Client name is required');
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
			toast.danger(e instanceof Error ? e.message : 'Failed to update client');
		}
		saving = false;
	}

	async function runDelete() {
		if (!client) return;
		try {
			await api.clients.delete(client.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to delete client');
			throw e;
		}
		toast.success('Client deleted');
		goto('/clients');
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
		<Spinner size="lg" />
	</div>
{:else if client}
	<div class="flex flex-col gap-10">
		<div class="flex flex-col gap-4">
			<Button href="/clients" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
				Clients
			</Button>
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex min-w-0 items-center gap-3">
					<Avatar name={client.name} size="lg" />
					<div class="flex min-w-0 flex-col gap-1">
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{client.name}</h1>
						{#if client.company}
							<p class="truncate text-fc-sm text-fc-fg-muted">{client.company}</p>
						{/if}
					</div>
				</div>
				<div class="flex shrink-0 items-center gap-2">
					{#if !editing}
						<Button variant="outline" icon={icons.edit} onclick={startEdit}>Edit</Button>
					{/if}
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						onclick={() => (confirmDelete = true)}
					>
						Delete
					</Button>
				</div>
			</div>
		</div>

		{#if editing}
			<form onsubmit={saveEdit} class="flex flex-col gap-4">
				<SettingsSection title="Details" description="Contact information and notes.">
					<Field label="Name">
						<Input bind:value={editName} placeholder="Jane Doe" />
					</Field>
					<Field label="Email">
						<Input bind:value={editEmail} type="email" placeholder="jane@example.com" />
					</Field>
					<Field label="Company">
						<Input bind:value={editCompany} placeholder="Acme Inc." />
					</Field>
					<Field label="Phone">
						<Input bind:value={editPhone} type="tel" placeholder="+1 555 000 0000" />
					</Field>
					<Field label="Notes">
						<Textarea
							bind:value={editNotes}
							rows={4}
							placeholder="Anything worth remembering about this client…"
						/>
					</Field>
				</SettingsSection>
				<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
					<Button
						variant="ghost"
						disabled={saving}
						class="w-full sm:w-auto"
						onclick={() => (editing = false)}
					>
						Cancel
					</Button>
					<Button type="submit" icon={icons.check} disabled={saving} class="w-full sm:w-auto">
						{saving ? 'Saving…' : 'Save'}
					</Button>
				</div>
			</form>
		{:else}
			<SettingsSection title="Details" description="Contact information and notes.">
				<SettingsRow label="Email">
					<span class="text-fc-sm text-fc-fg">{client.email || '—'}</span>
				</SettingsRow>
				<SettingsRow label="Company">
					<span class="text-fc-sm text-fc-fg">{client.company || '—'}</span>
				</SettingsRow>
				<SettingsRow label="Phone">
					<span class="text-fc-sm text-fc-fg">{client.phone || '—'}</span>
				</SettingsRow>
				<SettingsRow label="Notes" stacked>
					<span class="whitespace-pre-wrap text-fc-sm text-fc-fg">{client.notes || '—'}</span>
				</SettingsRow>
			</SettingsSection>
		{/if}

		<section class="flex flex-col gap-4">
			<div class="flex min-w-0 flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Documents</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					{documents.length} document{documents.length === 1 ? '' : 's'} linked to this client.
				</p>
			</div>

			{#if documents.length === 0}
				<EmptyState
					icon={icons.folder}
					title="No documents linked yet"
					description="Pick this client when you create a document and it will appear here."
				/>
			{:else}
				<div class="flex flex-col gap-2">
					{#each documents as doc (doc.id)}
						<Card href="/documents/{doc.id}" class="flex items-center justify-between gap-3">
							<div class="flex min-w-0 flex-col gap-1">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{doc.name}</span>
								<span class="text-fc-xs text-fc-fg-muted">{formatDate(doc.created_at)}</span>
							</div>
							<Badge tone={statusTone[doc.status]}>{doc.status}</Badge>
						</Card>
					{/each}
				</div>
			{/if}
		</section>
	</div>
{/if}

<ConfirmModal
	bind:open={confirmDelete}
	tone="danger"
	title="Delete {client?.name ?? 'this client'}?"
	description="This cannot be undone. Documents linked to this client are kept, but unlinked."
	confirmLabel="Delete client"
	cancelLabel="Keep client"
	onConfirm={runDelete}
/>
