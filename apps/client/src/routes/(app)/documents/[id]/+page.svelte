<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, getToken } from '$lib';
	import type { Document, Signer, Field as DocumentField, Client } from '$lib';
	import { spaceStore } from '$lib/stores/space.svelte';
	import {
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

	const statusTone = {
		draft: 'neutral',
		pending: 'info',
		completed: 'success',
		declined: 'danger'
	} as const;

	type Stage = { label: string; tone: 'neutral' | 'accent' | 'info' | 'success' | 'warning' | 'danger' };

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
			toast.danger(e instanceof Error ? e.message : 'Failed to reload the fields');
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
		if (signer.status === 'signed') return { label: 'signed', tone: 'success' };
		if (signer.status === 'declined') return { label: 'declined', tone: 'danger' };
		if (doc?.status === 'draft') return { label: 'not sent', tone: 'neutral' };
		if (waiting) return { label: 'waiting', tone: 'neutral' };
		if (signer.viewed_at) return { label: 'opened document', tone: 'info' };
		if (signer.email_opened_at) return { label: 'opened email', tone: 'warning' };
		return { label: 'sent', tone: 'accent' };
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
			newSignerError = 'Name and email are required';
			return;
		}
		addingSigner = true;
		newSignerError = '';
		try {
			const created = await api.signers.add(doc.id, name, email);
			signers = [...signers, created];
			showAddSigner = false;
			toast.success('Signer added');
		} catch (e) {
			newSignerError = e instanceof Error ? e.message : 'Failed to add the signer';
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
			toast.danger(e instanceof Error ? e.message : 'Failed to change the signing order');
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
			toast.danger(e instanceof Error ? e.message : 'Failed to update client');
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
			toast.danger('Copy failed');
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
			toast.success('Reminder sent');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to send reminder');
		}
		setTimeout(() => {
			if (remindingId === signer.id) remindingId = null;
		}, 1500);
	}

	function formatRelative(iso: string): string {
		const then = new Date(iso).getTime();
		const diff = Date.now() - then;
		if (diff < 60_000) return 'just now';
		const minutes = Math.floor(diff / 60_000);
		if (minutes < 60) return `${minutes}m ago`;
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		return `${days}d ago`;
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function downloadFile(urlPath: string, filename: string, setLoading: (v: boolean) => void) {
		if (!doc) return;
		setLoading(true);
		try {
			const res = await fetch(urlPath, {
				headers: { Authorization: `Bearer ${getToken()}` }
			});
			if (!res.ok) throw new Error('Download failed');
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Download failed');
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
			toast.success('Document sent');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to send the document');
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
			toast.danger(e instanceof Error ? e.message : 'Failed to load the document');
		}
		loading = false;
		pageMounted = true;
	});
</script>

<svelte:head><title>{doc ? `${doc.name} — Plume` : 'Plume'}</title></svelte:head>

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
			Back to documents
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
						<Badge tone={statusTone[doc.status]}>{doc.status}</Badge>
						{#if doc.status === 'draft'}
							<Badge tone={fields.length > 0 ? 'success' : 'neutral'}>
								{fields.length > 0
									? `${fields.length} field${fields.length === 1 ? '' : 's'} prepared`
									: 'No fields prepared yet'}
							</Badge>
						{/if}
					</div>

					<div class="flex flex-col gap-1 text-fc-sm text-fc-fg-muted">
						<span>Created {formatDate(doc.created_at)}</span>
						{#if doc.file_name}
							<span class="truncate">{doc.file_name}</span>
						{/if}
					</div>

					<div class="flex flex-wrap items-center gap-2">
						<span class="text-fc-sm text-fc-fg-muted">Client</span>
						{#if doc.status === 'draft'}
							<Select
								value={clientChoice}
								disabled={updatingClient}
								class="w-full sm:w-56"
								aria-label="Linked client"
								onchange={(e) => selectClient(e.currentTarget.value)}
							>
								<option value="">No client</option>
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
							<span class="text-fc-sm text-fc-fg-muted">No client</span>
						{/if}
					</div>
				</div>

				<div class="flex shrink-0 flex-wrap items-center gap-2">
					{#if doc.status === 'completed'}
						<Button
							variant="outline"
							icon={icons.download}
							onclick={downloadDocument}
							disabled={downloadingDoc}
						>
							{downloadingDoc ? 'Downloading…' : 'Download document'}
						</Button>
						<Button
							variant="outline"
							icon="solar:document-linear"
							onclick={downloadCertificate}
							disabled={downloading}
						>
							{downloading ? 'Downloading…' : 'Download certificate'}
						</Button>
						<Button
							variant="outline"
							icon={icons.shield}
							onclick={downloadAuditTrail}
							disabled={downloadingAudit}
						>
							{downloadingAudit ? 'Downloading…' : 'Download audit'}
						</Button>
					{/if}

					{#if doc.status === 'draft'}
						<Button
							variant="outline"
							icon="solar:layers-linear"
							onclick={() => (showFieldEditor = true)}
							disabled={signers.length === 0}
						>
							{fields.length > 0 ? `Edit fields (${fields.length})` : 'Prepare fields'}
						</Button>
						<Button
							icon="solar:plain-linear"
							onclick={sendForSigning}
							disabled={sending || signers.length === 0}
						>
							{sending ? 'Sending…' : 'Send for signing'}
						</Button>
					{/if}
				</div>
			</div>

			<SettingsSection
				title="Signing order"
				description={doc.status === 'draft'
					? 'Choose how signers are invited to sign this document.'
					: doc.sequential
						? 'Signers are invited one at a time in order.'
						: 'All signers were invited simultaneously.'}
			>
				<SettingsRow
					label="Sequential signing"
					description="Signers are invited one at a time in order. Signer N+1 is invited only after signer N completes."
				>
					{#if doc.status === 'draft'}
						<Switch
							checked={sequential}
							disabled={togglingSequential}
							aria-label="Sequential signing"
							onchange={(e) => toggleSequential(e.currentTarget.checked)}
						/>
					{:else}
						<Badge tone={doc.sequential ? 'accent' : 'neutral'}>
							{doc.sequential ? 'On' : 'Off'}
						</Badge>
					{/if}
				</SettingsRow>
			</SettingsSection>

			<SettingsSection
				title="Signers"
				description="{signers.length} signer{signers.length === 1 ? '' : 's'} on this document."
			>
				{#snippet actions()}
					{#if doc && doc.status !== 'completed' && doc.status !== 'declined'}
						<Button
							variant="outline"
							size="sm"
							icon={icons.plus}
							onclick={openAddSigner}
							disabled={doc.status !== 'draft'}
							title={doc.status !== 'draft'
								? 'Signers can only be added before signing starts'
								: undefined}
						>
							Add signer
						</Button>
					{/if}
				{/snippet}

				{#if signers.length === 0}
					<EmptyState
						bare
						icon={icons.usersGroup}
						title="No signers yet"
						description="Add at least one signer before this document can be sent."
					/>
				{:else}
					{#each sortedSigners as signer (signer.id)}
						{@const waiting = isWaitingSigner(signer)}
						{@const stage = signerStage(signer, waiting)}
						{@const assigned = fieldsBySigner.get(signer.id) ?? 0}
						<div
							class="flex flex-col gap-3 border-t border-fc-border pt-4 first:border-t-0 first:pt-0 sm:flex-row sm:items-start sm:justify-between"
						>
							<div class="flex min-w-0 flex-col gap-2">
								<div class="min-w-0">
									<p class="text-fc-sm font-medium break-words text-fc-fg">{signer.name}</p>
									<p class="truncate text-fc-sm text-fc-fg-muted">{signer.email}</p>
								</div>

								{#if doc?.status === 'draft'}
									<p class="text-fc-xs {assigned > 0 ? 'text-fc-success' : 'text-fc-fg-muted'}">
										{assigned > 0
											? `${assigned} field${assigned === 1 ? '' : 's'} assigned`
											: 'No fields assigned'}
									</p>
								{:else if doc?.status === 'pending' && signer.token && signer.status === 'pending' && !waiting}
									<div class="flex flex-wrap items-center gap-2">
										<Button
											variant="outline"
											size="sm"
											icon={copiedId === signer.id ? icons.check : icons.copy}
											onclick={() => copySigningLink(signer)}
										>
											{copiedId === signer.id ? 'Copied!' : 'Copy link'}
										</Button>
										<Button
											variant="outline"
											size="sm"
											icon="solar:plain-linear"
											onclick={() => remindSigner(signer)}
											disabled={remindingId === signer.id}
										>
											{remindingId === signer.id ? 'Sending…' : 'Resend email'}
										</Button>
										{#if signer.last_reminded_at}
											<span class="text-fc-xs text-fc-fg-muted">
												Sent {formatRelative(signer.last_reminded_at)}
											</span>
										{/if}
									</div>
									{#if signer.viewed_at}
										<p class="text-fc-xs text-fc-fg-muted">
											Opened the document {formatRelative(signer.viewed_at)}
										</p>
									{:else if signer.email_opened_at}
										<p class="text-fc-xs text-fc-fg-muted">
											Opened the email {formatRelative(signer.email_opened_at)}
										</p>
									{/if}
								{:else if waiting}
									<p class="text-fc-xs text-fc-fg-muted">Waiting for previous signer</p>
								{/if}
							</div>

							<div class="flex shrink-0 flex-wrap items-center gap-3">
								{#if signer.signed_at}
									<span class="text-fc-xs text-fc-fg-muted">
										Signed {formatDate(signer.signed_at)}
									</span>
								{/if}
								<Badge tone={stage.tone}>{stage.label}</Badge>
							</div>
						</div>
					{/each}
				{/if}
			</SettingsSection>
		{:else}
			<EmptyState
				icon={icons.error}
				title="Document not found"
				description="It may have been deleted, or it belongs to another space."
			>
				<Button href="/documents" variant="outline" icon={icons.chevronLeft}>
					Back to documents
				</Button>
			</EmptyState>
		{/if}
	</div>
{/if}

<Modal bind:open={showAddSigner} title="Add signer" showClose>
	<div class="flex flex-col gap-4">
		{#if clients.length > 0}
			<Field label="Fill from client" helper="Copies the client's name and email into the fields below.">
				<Select value={fillChoice} onchange={(e) => fillSignerFromClient(e.currentTarget.value)}>
					<option value="">Pick a client…</option>
					{#each clients as client (client.id)}
						<option value={String(client.id)}>{client.name}</option>
					{/each}
				</Select>
			</Field>
		{/if}
		<Field label="Name">
			<Input bind:value={newSignerName} placeholder="Jane Doe" />
		</Field>
		<Field label="Email" error={newSignerError}>
			<Input bind:value={newSignerEmail} type="email" placeholder="jane@example.com" />
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
				Cancel
			</Button>
			<Button
				icon={icons.plus}
				class="w-full sm:w-auto"
				disabled={addingSigner}
				onclick={addSignerToDoc}
			>
				{addingSigner ? 'Adding…' : 'Add signer'}
			</Button>
		</div>
	{/snippet}
</Modal>
