<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Document, Signer, Field } from '$lib';
	import { spaceStore } from '$lib/stores/space.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Separator } from '$lib/components/ui/separator';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';
	import FieldEditor from '$lib/components/field-editor.svelte';

	// The space this document was opened under. If the user switches context via
	// the space switcher, this document belongs to the previous context, so leave
	// the detail view and let the documents list load the new context.
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
	let fields = $state<Field[]>([]);
	let loading = $state(true);

	const fieldsBySigner = $derived(
		fields.reduce<Map<number, number>>((m, f) => m.set(f.signer_id, (m.get(f.signer_id) ?? 0) + 1), new Map())
	);

	async function refreshFields() {
		if (!doc) return;
		try {
			fields = await api.fields.list(doc.id);
		} catch {}
	}

	function closeFieldEditor() {
		showFieldEditor = false;
		refreshFields();
	}
	let sending = $state(false);
	let error = $state('');
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
	let addingSigner = $state(false);

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

	function isWaitingSigner(signer: Signer): boolean {
		if (!doc || doc.status !== 'pending' || !doc.sequential) return false;
		if (signer.status !== 'pending') return false;
		if (signer.role !== 'signer' && signer.role !== 'approver') return false;
		return activeOrderNum !== null && signer.order_num > activeOrderNum;
	}

	async function addSignerToDoc() {
		if (!doc || doc.status !== 'draft') return;
		const name = newSignerName.trim();
		const email = newSignerEmail.trim();
		if (!name || !email) {
			error = 'Name and email are required';
			return;
		}
		addingSigner = true;
		error = '';
		try {
			const created = await api.signers.add(doc.id, name, email);
			signers = [...signers, created];
			newSignerName = '';
			newSignerEmail = '';
			showAddSigner = false;
		} catch (e: any) {
			error = e.message;
		}
		addingSigner = false;
	}

	async function toggleSequential() {
		if (!doc || doc.status !== 'draft') return;
		togglingSequential = true;
		error = '';
		try {
			doc = await api.documents.update(doc.id, { sequential: !doc.sequential });
		} catch (e: any) {
			error = e.message;
		}
		togglingSequential = false;
	}

	function copySigningLink(signer: Signer) {
		const link = `${window.location.origin}/share/${signer.token}`;
		navigator.clipboard.writeText(link);
		copiedId = signer.id;
		setTimeout(() => (copiedId = null), 2000);
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
			toast.error(e instanceof Error ? e.message : 'Failed to send reminder');
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
				headers: { Authorization: `Bearer ${localStorage.getItem('token')}` }
			});
			if (!res.ok) throw new Error('Download failed');
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			a.click();
			URL.revokeObjectURL(url);
		} catch (e: any) {
			error = e.message;
		}
		setLoading(false);
	}

	function downloadCertificate() {
		downloadFile(api.documents.certificateUrl(doc!.id), `${doc!.name}_certificate.pdf`, (v) => (downloading = v));
	}

	function downloadDocument() {
		downloadFile(api.documents.fileUrl(doc!.id), `${doc!.name}.pdf`, (v) => (downloadingDoc = v));
	}

	function downloadAuditTrail() {
		downloadFile(api.documents.auditTrailUrl(doc!.id), `${doc!.name}_audit_trail.pdf`, (v) => (downloadingAudit = v));
	}

	async function sendForSigning() {
		if (!doc) return;
		sending = true;
		error = '';
		try {
			doc = await api.documents.send(doc.id);
			signers = await api.signers.list(doc.id);
		} catch (e: any) {
			error = e.message;
		}
		sending = false;
	}

	onMount(async () => {
		const id = Number(page.params.id);
		try {
			const [d, s, f] = await Promise.all([
				api.documents.get(id),
				api.signers.list(id),
				api.fields.list(id).catch(() => [])
			]);
			doc = d;
			signers = s;
			fields = f;
		} catch {}
		loading = false;
		pageMounted = true;
	});
</script>

<svelte:head><title>{doc ? `${doc.name} — Plume` : 'Plume'}</title></svelte:head>

{#if showFieldEditor && doc}
	<FieldEditor documentId={doc.id} {signers} onclose={closeFieldEditor} />
{:else}

<a href="/documents" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors mb-6">
	<Icon icon="solar:arrow-left-linear" class="h-4 w-4" />
	Back to documents
</a>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if doc}
	<div class="flex items-start justify-between mb-8">
		<div>
			<div class="flex items-center gap-3 mb-1">
				<h1 class="text-2xl font-bold">{doc.name}</h1>
				<span class="rounded-full px-2.5 py-0.5 text-xs font-medium
					{doc.status === 'draft' ? 'bg-muted text-muted-foreground' : ''}
					{doc.status === 'pending' ? 'bg-foreground/10 text-foreground' : ''}
					{doc.status === 'completed' ? 'bg-green-500/10 text-green-700 dark:text-green-400' : ''}
					{doc.status === 'declined' ? 'bg-red-500/10 text-red-700 dark:text-red-400' : ''}
				">{doc.status}</span>
			</div>
			<p class="text-sm text-muted-foreground">Created {formatDate(doc.created_at)}</p>
			{#if doc.file_name}
				<p class="text-sm text-muted-foreground mt-1">
					<Icon icon="solar:file-linear" class="inline h-3.5 w-3.5" />
					{doc.file_name}
				</p>
			{/if}
			{#if doc.status === 'draft'}
				<p class="mt-2.5 inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium
					{fields.length > 0 ? 'bg-green-500/10 text-green-700 dark:text-green-400' : 'bg-muted text-muted-foreground'}">
					<Icon icon={fields.length > 0 ? 'solar:check-circle-bold' : 'solar:layers-minimalistic-linear'} class="h-3.5 w-3.5" />
					{#if fields.length > 0}
						{fields.length} field{fields.length === 1 ? '' : 's'} prepared
					{:else}
						No fields prepared yet
					{/if}
				</p>
			{/if}
		</div>

		<div class="flex items-center gap-2">
			{#if doc.status === 'completed'}
				<Button variant="outline" onclick={downloadDocument} disabled={downloadingDoc}>
					{#if downloadingDoc}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
					{:else}
						<Icon icon="solar:download-minimalistic-linear" class="h-4 w-4" />
					{/if}
					Download document
				</Button>
				<Button variant="outline" onclick={downloadCertificate} disabled={downloading}>
					{#if downloading}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
					{:else}
						<Icon icon="solar:document-linear" class="h-4 w-4" />
					{/if}
					Download certificate
				</Button>
				<Button variant="outline" onclick={downloadAuditTrail} disabled={downloadingAudit}>
					{#if downloadingAudit}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
					{:else}
						<Icon icon="solar:shield-check-linear" class="h-4 w-4" />
					{/if}
					Download audit
				</Button>
			{/if}

			{#if doc.status === 'draft'}
				<Button variant="outline" onclick={() => (showFieldEditor = true)} disabled={signers.length === 0}>
					<Icon icon="solar:layers-linear" class="h-4 w-4" />
					{fields.length > 0 ? `Edit fields (${fields.length})` : 'Prepare fields'}
				</Button>
				<Button onclick={sendForSigning} disabled={sending || signers.length === 0}>
					{#if sending}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
						Sending...
					{:else}
						<Icon icon="solar:plain-linear" class="h-4 w-4" />
						Send for signing
					{/if}
				</Button>
			{/if}
		</div>
	</div>

	{#if error}
		<p class="text-sm text-destructive mb-4">{error}</p>
	{/if}

	<Card.Root class="mb-6">
		<Card.Header>
			<Card.Title>Signing order</Card.Title>
			<Card.Description>
				{#if doc.status === 'draft'}
					Choose how signers are invited to sign this document.
				{:else}
					Sequential signing: {doc.sequential ? 'On' : 'Off'}
				{/if}
			</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if doc.status === 'draft'}
				<button
					type="button"
					onclick={toggleSequential}
					disabled={togglingSequential}
					class="flex w-full items-start justify-between gap-4 rounded-lg border p-4 text-left transition-colors hover:bg-accent disabled:opacity-60"
				>
					<div>
						<p class="font-medium">Sequential signing</p>
						<p class="text-sm text-muted-foreground">
							Signers are invited one at a time in order. Signer N+1 is invited only after signer N completes.
						</p>
					</div>
					<div
						class="relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors
							{doc.sequential ? 'bg-foreground' : 'bg-muted'}"
					>
						<span
							class="inline-block h-5 w-5 transform rounded-full bg-background shadow transition-transform
								{doc.sequential ? 'translate-x-5' : 'translate-x-0.5'}"
						></span>
					</div>
				</button>
			{:else}
				<div class="flex items-center gap-2 text-sm text-muted-foreground">
					<Icon
						icon={doc.sequential ? 'solar:list-check-linear' : 'solar:users-group-rounded-linear'}
						class="h-4 w-4"
					/>
					{#if doc.sequential}
						Signers are invited one at a time in order.
					{:else}
						All signers were invited simultaneously.
					{/if}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<div class="flex items-start justify-between gap-4">
				<div>
					<Card.Title>Signers</Card.Title>
					<Card.Description>{signers.length} signer{signers.length === 1 ? '' : 's'}</Card.Description>
				</div>
				{#if doc.status !== 'completed' && doc.status !== 'declined'}
					<Button
						variant="outline"
						size="sm"
						onclick={() => (showAddSigner = !showAddSigner)}
						disabled={doc.status !== 'draft'}
						title={doc.status !== 'draft' ? 'Signers can only be added before signing starts' : undefined}
					>
						<Icon icon="mdi:plus" class="h-4 w-4" />
						Add signer
					</Button>
				{/if}
			</div>
		</Card.Header>
		<Card.Content>
			{#if doc.status === 'draft' && showAddSigner}
				<div class="mb-4 rounded-lg border p-4 space-y-3">
					<div class="grid gap-2 sm:grid-cols-2">
						<Input bind:value={newSignerName} placeholder="Name" />
						<Input bind:value={newSignerEmail} placeholder="Email" type="email" />
					</div>
					<div class="flex items-center gap-2">
						<Button size="sm" onclick={addSignerToDoc} disabled={addingSigner}>
							{#if addingSigner}
								<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
							{:else}
								<Icon icon="mdi:plus" class="h-4 w-4" />
							{/if}
							Add signer
						</Button>
						<Button variant="ghost" size="sm" onclick={() => (showAddSigner = false)} disabled={addingSigner}>
							Cancel
						</Button>
					</div>
				</div>
			{/if}
			{#if signers.length === 0}
				<p class="text-sm text-muted-foreground">No signers added yet.</p>
			{:else}
				<div class="space-y-3">
					{#each sortedSigners as signer}
						{@const waiting = isWaitingSigner(signer)}
						<div class="flex items-center justify-between rounded-lg border p-4">
							<div>
								<p class="font-medium">{signer.name}</p>
								<p class="text-sm text-muted-foreground">{signer.email}</p>
								{#if doc?.status === 'draft'}
									{@const fc = fieldsBySigner.get(signer.id) ?? 0}
									<p class="mt-1 inline-flex items-center gap-1 text-xs {fc > 0 ? 'text-green-700 dark:text-green-400' : 'text-muted-foreground'}">
										<Icon icon={fc > 0 ? 'solar:check-circle-linear' : 'solar:layers-minimalistic-linear'} class="h-3.5 w-3.5" />
										{fc > 0 ? `${fc} field${fc === 1 ? '' : 's'} assigned` : 'No fields assigned'}
									</p>
								{:else if doc?.status === 'pending' && signer.token && signer.status === 'pending' && !waiting}
									<div class="mt-2 flex flex-wrap items-center gap-2">
										<Button variant="outline" size="sm" onclick={() => copySigningLink(signer)}>
											<Icon icon={copiedId === signer.id ? 'solar:check-circle-linear' : 'solar:copy-linear'} class="h-3.5 w-3.5" />
											{copiedId === signer.id ? 'Copied!' : 'Copy link'}
										</Button>
										<Button variant="outline" size="sm" onclick={() => remindSigner(signer)} disabled={remindingId === signer.id}>
											<Icon icon="solar:plain-linear" class="h-3.5 w-3.5 {remindingId === signer.id ? 'animate-spin' : ''}" />
											{remindingId === signer.id ? 'Sending…' : 'Resend email'}
										</Button>
										{#if signer.last_reminded_at}
											<span class="text-xs text-muted-foreground">Sent {formatRelative(signer.last_reminded_at)}</span>
										{/if}
									</div>
								{:else if waiting}
									<p class="inline-flex items-center gap-1 mt-1.5 text-xs text-muted-foreground">
										<Icon icon="solar:clock-circle-linear" class="h-3.5 w-3.5" />
										Waiting for previous signer
									</p>
								{/if}
							</div>
							<div class="flex items-center gap-3">
								{#if signer.signed_at}
									<span class="text-xs text-muted-foreground">Signed {formatDate(signer.signed_at)}</span>
								{/if}
								<span class="rounded-full px-2.5 py-0.5 text-xs font-medium
									{signer.status === 'pending' && !waiting ? 'bg-foreground/10 text-foreground' : ''}
									{signer.status === 'pending' && waiting ? 'bg-muted text-muted-foreground' : ''}
									{signer.status === 'signed' ? 'bg-green-500/10 text-green-700 dark:text-green-400' : ''}
									{signer.status === 'declined' ? 'bg-red-500/10 text-red-700 dark:text-red-400' : ''}
								">{waiting ? 'waiting' : signer.status}</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
{/if}

{/if}
