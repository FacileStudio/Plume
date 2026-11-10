<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Document, Signer, Field as DocumentField, Client } from '$lib';
	import { spaceStore } from '$lib/stores/space.svelte';
	import {
		Avatar,
		Badge,
		Button,
		EmptyState,
		Field,
		Input,
		Modal,
		Select,
		SettingsRow,
		SettingsSection,
		Spinner,
		Switch,
		icons,
		toast
	} from '@facile/muse';
	import FieldEditor from '$lib/components/field-editor.svelte';
	import { formatRelative, formatShortDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	const statusTone = {
		draft: 'neutral',
		pending: 'info',
		completed: 'success',
		declined: 'danger'
	} as const;

	type Stage = { label: () => string; tone: 'neutral' | 'accent' | 'info' | 'success' | 'warning' | 'danger' };

	const openedSpaceId = spaceStore.activeId;
	let pageMounted = $state(false);

	$effect(() => {
		const current = spaceStore.activeId;
		if (pageMounted && current !== openedSpaceId) {
			goto('/documents');
		}
	});

	let doc = $state<Document | null>(null);
	let signers = $state<Signer[]>([]);
	let fields = $state<DocumentField[]>([]);
	let clients = $state<Client[]>([]);
	let loading = $state(true);
	let updatingClient = $state(false);
	let clientChoice = $state('');
	let sequential = $state(false);
	let sending = $state(false);
	let copiedId = $state<number | null>(null);
	let downloading = $state(false);
	let downloadingDoc = $state(false);
	let downloadingAudit = $state(false);
	let showFieldEditor = $state(false);
	let remindingId = $state<number | null>(null);
	let togglingSequential = $state(false);
	let showAddSigner = $state(false);
	let newSignerName = $state('');
	let newSignerEmail = $state('');
	let newSignerError = $state('');
	let fillChoice = $state('');
	let addingSigner = $state(false);

	const linkedClient = $derived(
		doc?.client_id ? (clients.find((c) => c.id === doc!.client_id) ?? null) : null
	);

	const fieldsBySigner = $derived(
		fields.reduce<Map<number, number>>(
			(m, f) => m.set(f.signer_id, (m.get(f.signer_id) ?? 0) + 1),
			new Map()
		)
	);

	const sortedSigners = $derived(
		[...signers].sort((a, b) => a.order_num - b.order_num || a.id - b.id)
	);

	const activeOrderNum = $derived.by(() => {
		if (!doc || doc.status !== 'pending' || !doc.sequential) return null;
		const pending = sortedSigners.filter(
			(s) => s.status === 'pending' && (s.role === 'signer' || s.role === 'approver')
		);
		if (pending.length === 0) return null;
		return pending[0].order_num;
	});

	async function refreshFields() {
		if (!doc) return;
		try {
			fields = await api.fields.list(doc.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('doc.fieldsReloadFailed'));
		}
	}

	function closeFieldEditor() {
		showFieldEditor = false;
		refreshFields();
	}

	function isWaitingSigner(signer: Signer): boolean {
		if (!doc || doc.status !== 'pending' || !doc.sequential) return false;
		if (signer.status !== 'pending') return false;
		if (signer.role !== 'signer' && signer.role !== 'approver') return false;
		return activeOrderNum !== null && signer.order_num > activeOrderNum;
	}

	function signerStage(signer: Signer, waiting: boolean): Stage {
		if (signer.status === 'signed') return { label: () => t('doc.stage.signed'), tone: 'success' };
		if (signer.status === 'declined') return { label: () => t('doc.stage.declined'), tone: 'danger' };
		if (doc?.status === 'draft') return { label: () => t('doc.stage.notSent'), tone: 'neutral' };
		if (waiting) return { label: () => t('doc.stage.waiting'), tone: 'neutral' };
		if (signer.viewed_at) return { label: () => t('doc.stage.openedDocument'), tone: 'info' };
		if (signer.email_opened_at) return { label: () => t('doc.stage.openedEmail'), tone: 'warning' };
		return { label: () => t('doc.stage.sent'), tone: 'accent' };
	}

	function openAddSigner() {
		newSignerName = '';
		newSignerEmail = '';
		newSignerError = '';
		fillChoice = '';
		showAddSigner = true;
	}

	async function addSignerToDoc() {
		if (!doc || doc.status !== 'draft') return;
		const name = newSignerName.trim();
		const email = newSignerEmail.trim();
		if (!name || !email) {
			newSignerError = t('doc.addSigner.required');
			return;
		}
		addingSigner = true;
		newSignerError = '';
		try {
			const created = await api.signers.add(doc.id, name, email);
			signers = [...signers, created];
			showAddSigner = false;
			toast.success(t('doc.addSigner.added'));
		} catch (e) {
			newSignerError = e instanceof Error ? e.message : t('doc.addSigner.failed');
		}
		addingSigner = false;
	}

	async function toggleSequential(next: boolean) {
		if (!doc || doc.status !== 'draft') return;
		const previous = sequential;
		sequential = next;
		togglingSequential = true;
		try {
			doc = await api.documents.update(doc.id, { sequential: next });
			sequential = doc.sequential;
		} catch (e) {
			sequential = previous;
			toast.danger(e instanceof Error ? e.message : t('doc.order.failed'));
		}
		togglingSequential = false;
	}

	async function selectClient(value: string) {
		if (!doc || doc.status !== 'draft') return;
		const previous = clientChoice;
		clientChoice = value;
		updatingClient = true;
		try {
			doc = await api.documents.update(doc.id, { client_id: value ? Number(value) : 0 });
			clientChoice = doc.client_id ? String(doc.client_id) : '';
		} catch (e) {
			clientChoice = previous;
			toast.danger(e instanceof Error ? e.message : t('doc.clientFailed'));
		}
		updatingClient = false;
	}

	function fillSignerFromClient(value: string) {
		fillChoice = value;
		const client = clients.find((c) => String(c.id) === value);
		if (!client) return;
		newSignerName = client.name;
		newSignerEmail = client.email;
	}

	async function copySigningLink(signer: Signer) {
		const link = `${window.location.origin}/share/${signer.token}`;
		try {
			await navigator.clipboard.writeText(link);
			copiedId = signer.id;
			setTimeout(() => (copiedId = null), 2000);
		} catch {
			toast.danger(t('doc.signers.copyFailed'));
		}
	}

	async function remindSigner(signer: Signer) {
		if (remindingId === signer.id) return;
		remindingId = signer.id;
		try {
			const res = await api.signers.remind(signer.id);
			signers = signers.map((s) =>
				s.id === signer.id ? { ...s, last_reminded_at: res.reminded_at } : s
			);
			toast.success(t('doc.signers.reminded'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('doc.signers.remindFailed'));
		}
		setTimeout(() => {
			if (remindingId === signer.id) remindingId = null;
		}, 1500);
	}

	async function downloadFile(urlPath: string, filename: string, setLoading: (v: boolean) => void) {
		if (!doc) return;
		setLoading(true);
		try {
			const res = await fetch(urlPath, { credentials: 'same-origin' });
			if (!res.ok) throw new Error(t('doc.downloadFailed'));
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('doc.downloadFailed'));
		}
		setLoading(false);
	}

	function downloadCertificate() {
		downloadFile(
			api.documents.certificateUrl(doc!.id),
			`${doc!.name}_certificate.pdf`,
			(v) => (downloading = v)
		);
	}

	function downloadDocument() {
		downloadFile(api.documents.fileUrl(doc!.id), `${doc!.name}.pdf`, (v) => (downloadingDoc = v));
	}

	function downloadAuditTrail() {
		downloadFile(
			api.documents.auditTrailUrl(doc!.id),
			`${doc!.name}_audit_trail.pdf`,
			(v) => (downloadingAudit = v)
		);
	}

	async function sendForSigning() {
		if (!doc) return;
		sending = true;
		try {
			doc = await api.documents.send(doc.id);
			signers = await api.signers.list(doc.id);
			toast.success(t('doc.sent'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('doc.sendFailed'));
		}
		sending = false;
	}

	onMount(async () => {
		const id = Number(page.params.id);
		try {
			const [d, s, f, c] = await Promise.all([
				api.documents.get(id),
				api.signers.list(id),
				api.fields.list(id).catch(() => []),
				api.clients.list().catch(() => [])
			]);
			doc = d;
			signers = s;
			fields = f;
			clients = c;
			sequential = d.sequential;
			clientChoice = d.client_id ? String(d.client_id) : '';
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('doc.loadFailed'));
		}
		loading = false;
		pageMounted = true;
	});
</script>

<svelte:head><title>{doc ? `${doc.name} — Plume` : t('doc.titleFallback')}</title></svelte:head>

{#if showFieldEditor && doc}
	<FieldEditor documentId={doc.id} {signers} onclose={closeFieldEditor} />
{:else}
	<div class="flex flex-col gap-10">
		<Button
			href="/documents"
			variant="ghost"
			size="sm"
			icon={icons.chevronLeft}
			class="self-start px-2"
		>
			{t('common.backToDocuments')}
		</Button>

		{#if loading}
			<div class="flex min-h-[40dvh] items-center justify-center">
				<Spinner size="lg" />
			</div>
		{:else if doc}
			<div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
				<div class="flex min-w-0 flex-col gap-3">
					<div class="flex flex-wrap items-center gap-x-3 gap-y-2">
						<h1 class="min-w-0 text-fc-2xl font-semibold break-words text-fc-fg">{doc.name}</h1>
						<Badge tone={statusTone[doc.status]}>{t(`status.${doc.status}`)}</Badge>
						{#if doc.status === 'draft'}
							<Badge tone={fields.length > 0 ? 'success' : 'neutral'}>
								{fields.length > 0
									? t('doc.fieldsPrepared', { count: fields.length })
									: t('doc.noFieldsPrepared')}
							</Badge>
						{/if}
					</div>

					<div class="flex flex-col gap-1 text-fc-sm text-fc-fg-muted">
						<span>{t('doc.createdOn', { date: formatShortDate(doc.created_at) })}</span>
						{#if doc.file_name}
							<span class="truncate">{doc.file_name}</span>
						{/if}
					</div>

					<div class="flex flex-wrap items-center gap-2">
						<span class="text-fc-sm text-fc-fg-muted">{t('doc.client')}</span>
						{#if doc.status === 'draft'}
							<Select
								value={clientChoice}
								disabled={updatingClient}
								class="w-full sm:w-56"
								aria-label={t('doc.linkedClient')}
								onchange={(e) => selectClient(e.currentTarget.value)}
							>
								<option value="">{t('doc.noClient')}</option>
								{#each clients as client (client.id)}
									<option value={String(client.id)}>{client.name}</option>
								{/each}
							</Select>
						{:else if linkedClient}
							<a
								href="/clients/{linkedClient.id}"
								class="text-fc-sm text-fc-fg hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
							>
								{linkedClient.name}
							</a>
						{:else}
							<span class="text-fc-sm text-fc-fg-muted">{t('doc.noClient')}</span>
						{/if}
					</div>
				</div>

				<div class="flex shrink-0 flex-wrap items-center gap-2">
					{#if doc.status === 'completed'}
						<Button
							variant="secondary"
							icon={icons.download}
							onclick={downloadDocument}
							disabled={downloadingDoc}
						>
							{downloadingDoc ? t('common.downloading') : t('common.download')}
						</Button>
						<Button
							variant="secondary"
							icon="solar:document-linear"
							onclick={downloadCertificate}
							disabled={downloading}
						>
							{downloading ? t('common.downloading') : t('doc.downloadCertificate')}
						</Button>
						<Button
							variant="secondary"
							icon={icons.shield}
							onclick={downloadAuditTrail}
							disabled={downloadingAudit}
						>
							{downloadingAudit ? t('common.downloading') : t('doc.downloadAudit')}
						</Button>
					{/if}

					{#if doc.status === 'draft'}
						<Button
							variant="secondary"
							icon="solar:layers-linear"
							onclick={() => (showFieldEditor = true)}
							disabled={signers.length === 0}
						>
							{fields.length > 0 ? t('doc.editFields', { count: fields.length }) : t('doc.prepareFields')}
						</Button>
						<Button
							icon="solar:plain-linear"
							onclick={sendForSigning}
							disabled={sending || signers.length === 0}
						>
							{sending ? t('common.sending') : t('doc.send')}
						</Button>
					{/if}
				</div>
			</div>

			<SettingsSection
				title={t('doc.order.title')}
				description={doc.status === 'draft'
					? t('doc.order.draft')
					: doc.sequential
						? t('doc.order.sequentialOn')
						: t('doc.order.sequentialOff')}
			>
				<SettingsRow label={t('doc.order.label')} description={t('doc.order.labelText')}>
					{#if doc.status === 'draft'}
						<Switch
							checked={sequential}
							disabled={togglingSequential}
							aria-label={t('doc.order.label')}
							onchange={(e) => toggleSequential(e.currentTarget.checked)}
						/>
					{:else}
						<Badge tone={doc.sequential ? 'accent' : 'neutral'}>
							{doc.sequential ? t('common.on') : t('common.off')}
						</Badge>
					{/if}
				</SettingsRow>
			</SettingsSection>

			<SettingsSection
				title={t('doc.signers.title')}
				description={t('doc.signers.count', { count: signers.length })}
			>
				{#snippet actions()}
					{#if doc && doc.status !== 'completed' && doc.status !== 'declined'}
						<Button
							variant="secondary"
							size="sm"
							icon={icons.plus}
							onclick={openAddSigner}
							disabled={doc.status !== 'draft'}
							title={doc.status !== 'draft' ? t('doc.signers.addDisabled') : undefined}
						>
							{t('docNew.addSigner')}
						</Button>
					{/if}
				{/snippet}

				{#if signers.length === 0}
					<EmptyState
						bare
						icon={icons.usersGroup}
						title={t('doc.signers.emptyTitle')}
						description={t('doc.signers.emptyText')}
					/>
				{:else}
					{#each sortedSigners as signer (signer.id)}
						{@const waiting = isWaitingSigner(signer)}
						{@const stage = signerStage(signer, waiting)}
						{@const assigned = fieldsBySigner.get(signer.id) ?? 0}
						<div class="flex flex-col gap-3 pt-4 first:pt-0 sm:flex-row sm:items-start sm:justify-between">
							<div class="flex min-w-0 flex-col gap-2">
								<div class="flex min-w-0 items-center gap-3">
									<Avatar src={signer.avatar_url || undefined} name={signer.name || signer.email} size="sm" />
									<div class="min-w-0">
										<p class="text-fc-sm font-medium break-words text-fc-fg">{signer.name}</p>
										<p class="truncate text-fc-sm text-fc-fg-muted">{signer.email}</p>
									</div>
								</div>

								{#if doc?.status === 'draft'}
									<p class="text-fc-xs {assigned > 0 ? 'text-fc-success' : 'text-fc-fg-muted'}">
										{assigned > 0
											? t('doc.signers.fieldsAssigned', { count: assigned })
											: t('doc.signers.noFieldsAssigned')}
									</p>
								{:else if doc?.status === 'pending' && signer.token && signer.status === 'pending' && !waiting}
									<div class="flex flex-wrap items-center gap-2">
										<Button
											variant="secondary"
											size="sm"
											icon={copiedId === signer.id ? icons.check : icons.copy}
											onclick={() => copySigningLink(signer)}
										>
											{copiedId === signer.id ? t('doc.signers.copied') : t('doc.signers.copyLink')}
										</Button>
										<Button
											variant="secondary"
											size="sm"
											icon="solar:plain-linear"
											onclick={() => remindSigner(signer)}
											disabled={remindingId === signer.id}
										>
											{remindingId === signer.id ? t('common.sending') : t('doc.signers.resend')}
										</Button>
										{#if signer.last_reminded_at}
											<span class="text-fc-xs text-fc-fg-muted">
												{t('doc.signers.sentAgo', { when: formatRelative(signer.last_reminded_at) })}
											</span>
										{/if}
									</div>
									{#if signer.viewed_at}
										<p class="text-fc-xs text-fc-fg-muted">
											{t('doc.signers.openedDocument', { when: formatRelative(signer.viewed_at) })}
										</p>
									{:else if signer.email_opened_at}
										<p class="text-fc-xs text-fc-fg-muted">
											{t('doc.signers.openedEmail', { when: formatRelative(signer.email_opened_at) })}
										</p>
									{/if}
								{:else if waiting}
									<p class="text-fc-xs text-fc-fg-muted">{t('doc.signers.waiting')}</p>
								{/if}
							</div>

							<div class="flex shrink-0 flex-wrap items-center gap-3">
								{#if signer.signed_at}
									<span class="text-fc-xs text-fc-fg-muted">
										{t('doc.signers.signedOn', { date: formatShortDate(signer.signed_at) })}
									</span>
								{/if}
								<Badge tone={stage.tone}>{stage.label()}</Badge>
							</div>
						</div>
					{/each}
				{/if}
			</SettingsSection>
		{:else}
			<EmptyState
				icon={icons.error}
				title={t('doc.notFound.title')}
				description={t('doc.notFound.text')}
			>
				<Button href="/documents" variant="secondary" icon={icons.chevronLeft}>
					{t('common.backToDocuments')}
				</Button>
			</EmptyState>
		{/if}
	</div>
{/if}

<Modal bind:open={showAddSigner} title={t('doc.addSigner.title')} showClose>
	<div class="flex flex-col gap-4">
		{#if clients.length > 0}
			<Field label={t('doc.addSigner.fill')} helper={t('doc.addSigner.fillHelp')}>
				<Select value={fillChoice} onchange={(e) => fillSignerFromClient(e.currentTarget.value)}>
					<option value="">{t('doc.addSigner.pick')}</option>
					{#each clients as client (client.id)}
						<option value={String(client.id)}>{client.name}</option>
					{/each}
				</Select>
			</Field>
		{/if}
		<Field label={t('common.name')}>
			<Input bind:value={newSignerName} placeholder={t('docNew.signerNamePlaceholder')} />
		</Field>
		<Field label={t('common.email')} error={newSignerError}>
			<Input bind:value={newSignerEmail} type="email" placeholder={t('docNew.signerEmailPlaceholder')} />
		</Field>
	</div>
	{#snippet footer()}
		<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
			<Button
				variant="ghost"
				class="w-full sm:w-auto"
				disabled={addingSigner}
				onclick={() => (showAddSigner = false)}
			>
				{t('common.cancel')}
			</Button>
			<Button
				icon={icons.plus}
				class="w-full sm:w-auto"
				disabled={addingSigner}
				onclick={addSignerToDoc}
			>
				{addingSigner ? t('common.adding') : t('docNew.addSigner')}
			</Button>
		</div>
	{/snippet}
</Modal>
