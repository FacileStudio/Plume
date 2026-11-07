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
	import { formatShortDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

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
			toast.danger(t('client.nameRequired'));
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
			toast.success(t('client.updated'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('client.updateFailed'));
		}
		saving = false;
	}

	async function runDelete() {
		if (!client) return;
		try {
			await api.clients.delete(client.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('clients.deleteFailed'));
			throw e;
		}
		toast.success(t('clients.deleted'));
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

<svelte:head><title>{client?.name ?? t('client.titleFallback')} — Plume</title></svelte:head>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Spinner size="lg" />
	</div>
{:else if client}
	<div class="flex flex-col gap-10">
		<div class="flex flex-col gap-4">
			<Button href="/clients" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
				{t('client.back')}
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
						<Button variant="secondary" icon={icons.edit} onclick={startEdit}>{t('common.edit')}</Button>
					{/if}
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						onclick={() => (confirmDelete = true)}
					>
						{t('common.delete')}
					</Button>
				</div>
			</div>
		</div>

		{#if editing}
			<form onsubmit={saveEdit} class="flex flex-col gap-4">
				<SettingsSection title={t('client.details')} description={t('client.detailsText')}>
					<Field label={t('common.name')}>
						<Input bind:value={editName} placeholder={t('clientNew.namePlaceholder')} />
					</Field>
					<Field label={t('common.email')}>
						<Input bind:value={editEmail} type="email" placeholder={t('clientNew.emailPlaceholder')} />
					</Field>
					<Field label={t('common.company')}>
						<Input bind:value={editCompany} placeholder={t('clientNew.companyPlaceholder')} />
					</Field>
					<Field label={t('common.phone')}>
						<Input bind:value={editPhone} type="tel" placeholder={t('clientNew.phonePlaceholder')} />
					</Field>
					<Field label={t('common.notes')}>
						<Textarea
							bind:value={editNotes}
							rows={4}
							placeholder={t('client.notesPlaceholder')}
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
						{t('common.cancel')}
					</Button>
					<Button type="submit" icon={icons.check} disabled={saving} class="w-full sm:w-auto">
						{saving ? t('common.saving') : t('common.save')}
					</Button>
				</div>
			</form>
		{:else}
			<SettingsSection title={t('client.details')} description={t('client.detailsText')}>
				<SettingsRow label={t('common.email')}>
					<span class="text-fc-sm text-fc-fg">{client.email || '—'}</span>
				</SettingsRow>
				<SettingsRow label={t('common.company')}>
					<span class="text-fc-sm text-fc-fg">{client.company || '—'}</span>
				</SettingsRow>
				<SettingsRow label={t('common.phone')}>
					<span class="text-fc-sm text-fc-fg">{client.phone || '—'}</span>
				</SettingsRow>
				<SettingsRow label={t('common.notes')} stacked>
					<span class="whitespace-pre-wrap text-fc-sm text-fc-fg">{client.notes || '—'}</span>
				</SettingsRow>
			</SettingsSection>
		{/if}

		<section class="flex flex-col gap-4">
			<div class="flex min-w-0 flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">{t('client.documents')}</h2>
				<p class="text-fc-sm text-fc-fg-muted">{t('client.documentCount', { count: documents.length })}</p>
			</div>

			{#if documents.length === 0}
				<EmptyState
					icon={icons.folder}
					title={t('client.noDocuments')}
					description={t('client.noDocumentsText')}
				/>
			{:else}
				<div class="flex flex-col gap-2">
					{#each documents as doc (doc.id)}
						<Card href="/documents/{doc.id}" class="flex items-center justify-between gap-3">
							<div class="flex min-w-0 flex-col gap-1">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{doc.name}</span>
								<span class="text-fc-xs text-fc-fg-muted">{formatShortDate(doc.created_at)}</span>
							</div>
							<Badge tone={statusTone[doc.status]}>{t(`status.${doc.status}`)}</Badge>
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
	title={t('clients.confirm.title', { name: client?.name ?? t('clients.confirm.thisClient') })}
	description={t('clients.confirm.text')}
	confirmLabel={t('clients.confirm.label')}
	cancelLabel={t('clients.confirm.cancel')}
	onConfirm={runDelete}
/>
