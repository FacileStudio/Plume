<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { Document, Signer } from '$lib';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import Icon from '@iconify/svelte';
	import FieldEditor from '$lib/components/field-editor.svelte';

	let doc = $state<Document | null>(null);
	let signers = $state<Signer[]>([]);
	let loading = $state(true);
	let sending = $state(false);
	let error = $state('');
	let copiedId = $state<number | null>(null);
	let downloading = $state(false);
	let downloadingDoc = $state(false);
	let showFieldEditor = $state(false);

	function copySigningLink(signer: Signer) {
		const link = `${window.location.origin}/share/${signer.token}`;
		navigator.clipboard.writeText(link);
		copiedId = signer.id;
		setTimeout(() => (copiedId = null), 2000);
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
			const [d, s] = await Promise.all([api.documents.get(id), api.signers.list(id)]);
			doc = d;
			signers = s;
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>{doc ? `${doc.name} — Plume` : 'Plume'}</title></svelte:head>

{#if showFieldEditor && doc}
	<FieldEditor documentId={doc.id} {signers} onclose={() => (showFieldEditor = false)} />
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
			{/if}

			{#if doc.status === 'draft'}
				<Button variant="outline" onclick={() => (showFieldEditor = true)} disabled={signers.length === 0}>
					<Icon icon="solar:layers-linear" class="h-4 w-4" />
					Prepare fields
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

	<Card.Root>
		<Card.Header>
			<Card.Title>Signers</Card.Title>
			<Card.Description>{signers.length} signer{signers.length === 1 ? '' : 's'}</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if signers.length === 0}
				<p class="text-sm text-muted-foreground">No signers added yet.</p>
			{:else}
				<div class="space-y-3">
					{#each signers as signer}
						<div class="flex items-center justify-between rounded-lg border p-4">
							<div>
								<p class="font-medium">{signer.name}</p>
								<p class="text-sm text-muted-foreground">{signer.email}</p>
								{#if doc?.status === 'pending' && signer.token && signer.status === 'pending'}
									<button
										onclick={() => copySigningLink(signer)}
										class="inline-flex items-center gap-1 mt-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
									>
										<Icon icon={copiedId === signer.id ? 'solar:check-circle-linear' : 'solar:copy-linear'} class="h-3.5 w-3.5" />
										{copiedId === signer.id ? 'Copied!' : 'Copy signing link'}
									</button>
								{/if}
							</div>
							<div class="flex items-center gap-3">
								{#if signer.signed_at}
									<span class="text-xs text-muted-foreground">Signed {formatDate(signer.signed_at)}</span>
								{/if}
								<span class="rounded-full px-2.5 py-0.5 text-xs font-medium
									{signer.status === 'pending' ? 'bg-foreground/10 text-foreground' : ''}
									{signer.status === 'signed' ? 'bg-green-500/10 text-green-700 dark:text-green-400' : ''}
									{signer.status === 'declined' ? 'bg-red-500/10 text-red-700 dark:text-red-400' : ''}
								">{signer.status}</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
{/if}

{/if}
