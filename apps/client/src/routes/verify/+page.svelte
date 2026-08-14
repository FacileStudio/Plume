<script lang="ts">
	import { onMount } from 'svelte';
	import { api, currentUser } from '$lib';
	import type { VerifyResponse } from '$lib';
	import { Alert, Badge, Button, Card, Divider, Dropzone, Spinner, icons } from '@facile/muse';

	const MAX_SIZE = 50 * 1024 * 1024;

	let files = $state<File[]>([]);
	let loading = $state(false);
	let error = $state('');
	let result = $state<VerifyResponse | null>(null);
	let signedIn = $state(false);

	onMount(async () => {
		signedIn = !!(await currentUser().catch(() => null));
	});

	const file = $derived(files[0] ?? null);

	function reset() {
		files = [];
		result = null;
		error = '';
	}

	function onFiles() {
		error = '';
		result = null;
	}

	function onReject(rejections: { file: File; reason: 'type' | 'size' | 'count' }[]) {
		const reason = rejections[0]?.reason;
		error =
			reason === 'size'
				? 'That file is larger than 50 MB.'
				: reason === 'count'
					? 'Verify one document at a time.'
					: 'Only PDF files can be verified.';
	}

	async function verify() {
		if (!file) return;
		loading = true;
		error = '';
		result = null;
		try {
			result = await api.verify.check(file);
		} catch (e: any) {
			error = e?.message ?? 'Failed to verify document';
		} finally {
			loading = false;
		}
	}

	function formatDate(iso?: string | null): string {
		if (!iso) return '';
		return new Date(iso).toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function statusTone(status: string): 'neutral' | 'info' | 'success' | 'danger' {
		switch (status) {
			case 'completed':
				return 'success';
			case 'declined':
				return 'danger';
			case 'pending':
			case 'signed':
				return 'info';
			default:
				return 'neutral';
		}
	}
</script>

<svelte:head>
	<title>Verify a document — Plume</title>
	<meta name="description" content="Check whether a PDF was issued and signed through Plume." />
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="flex min-h-[100dvh] flex-col bg-fc-page text-fc-fg">
	<header class="border-b border-fc-border">
		<div class="mx-auto flex w-full max-w-3xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
			<a
				href="/"
				class="flex items-center gap-2 rounded-fc-md focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
			>
				<iconify-icon icon="solar:pen-new-square-bold-duotone" width="24" height="24" class="block"
				></iconify-icon>
				<span class="text-fc-lg font-semibold tracking-tight">Plume</span>
			</a>
			{#if signedIn}
				<Button variant="ghost" href="/dashboard">Dashboard</Button>
			{:else}
				<Button variant="ghost" href="/login">Log in</Button>
			{/if}
		</div>
	</header>

	<main class="mx-auto w-full max-w-3xl flex-1 px-4 py-10 sm:px-6 sm:py-16">
		<div class="flex flex-col gap-10">
			<div class="flex flex-col items-center gap-4 text-center">
				<span
					class="flex size-12 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-fg-muted"
				>
					<iconify-icon icon={icons.shield} width="24" height="24" class="block"></iconify-icon>
				</span>
				<div class="flex flex-col gap-1">
					<h1 class="text-fc-2xl font-semibold text-fc-fg">Verify a document</h1>
					<p class="mx-auto max-w-lg text-fc-sm text-fc-fg-muted">
						Drop a PDF below. Plume re-computes its SHA-256 fingerprint and checks whether it matches
						a document we have on file. Nothing is stored.
					</p>
				</div>
			</div>

			<section class="flex flex-col gap-4">
				<Dropzone
					bind:files
					accept=".pdf,application/pdf"
					maxSize={MAX_SIZE}
					disabled={loading}
					label={file ? file.name : 'Drop a PDF here'}
					hint={file ? `${(file.size / 1024).toFixed(0)} KB` : 'Max 50 MB'}
					{onFiles}
					{onReject}
				/>

				{#if error}
					<Alert tone="danger" title="Could not verify">{error}</Alert>
				{/if}

				<div class="flex flex-col gap-2 sm:flex-row">
					<Button
						size="lg"
						icon={loading ? undefined : icons.shield}
						onclick={verify}
						disabled={!file || loading}
						class="flex-1"
					>
						{#if loading}
							<Spinner size="sm" class="border-current/30 border-t-current" />
							Verifying…
						{:else}
							Verify document
						{/if}
					</Button>
					{#if file || result || error}
						<Button variant="outline" size="lg" icon={icons.refresh} onclick={reset} disabled={loading}>
							Reset
						</Button>
					{/if}
				</div>
			</section>

			{#if result}
				<section class="flex flex-col gap-4">
					{#if result.match && result.document}
						<Alert tone="success" title="Authentic — issued by Plume">
							This document matches our records. You uploaded the
							{result.variant === 'signed' ? 'signed' : 'original'} version.
						</Alert>

						<Card class="flex flex-col gap-4">
							<p class="text-fc-sm font-medium text-fc-fg">Document</p>
							<dl class="flex flex-col gap-2 text-fc-sm">
								<div class="flex items-start justify-between gap-4">
									<dt class="text-fc-fg-muted">Name</dt>
									<dd class="min-w-0 text-right font-medium text-fc-fg">{result.document.name}</dd>
								</div>
								<div class="flex items-start justify-between gap-4">
									<dt class="text-fc-fg-muted">File</dt>
									<dd class="min-w-0 break-all text-right font-medium text-fc-fg">
										{result.document.file_name}
									</dd>
								</div>
								<div class="flex items-center justify-between gap-4">
									<dt class="text-fc-fg-muted">Status</dt>
									<dd>
										<Badge tone={statusTone(result.document.status)}>{result.document.status}</Badge>
									</dd>
								</div>
								<div class="flex items-start justify-between gap-4">
									<dt class="text-fc-fg-muted">Uploaded</dt>
									<dd class="text-right text-fc-fg">{formatDate(result.document.created_at)}</dd>
								</div>
								{#if result.document.completed_at}
									<div class="flex items-start justify-between gap-4">
										<dt class="text-fc-fg-muted">Completed</dt>
										<dd class="text-right text-fc-fg">{formatDate(result.document.completed_at)}</dd>
									</div>
								{/if}
							</dl>
						</Card>

						{#if result.signers && result.signers.length > 0}
							<Card class="flex flex-col gap-4">
								<p class="text-fc-sm font-medium text-fc-fg">Signers</p>
								<ul class="flex flex-col">
									{#each result.signers as signer (signer.email + signer.name)}
										<li
											class="flex flex-wrap items-center justify-between gap-3 border-t border-fc-border py-3 first:border-t-0 first:pt-0 last:pb-0"
										>
											<div class="min-w-0">
												<p class="truncate text-fc-sm font-medium text-fc-fg">{signer.name}</p>
												<p class="truncate text-fc-xs text-fc-fg-muted">{signer.email}</p>
											</div>
											<div class="flex shrink-0 items-center gap-3">
												{#if signer.signed_at}
													<span class="text-fc-xs text-fc-fg-muted">{formatDate(signer.signed_at)}</span>
												{/if}
												<Badge tone={statusTone(signer.status)}>{signer.status}</Badge>
											</div>
										</li>
									{/each}
								</ul>
							</Card>
						{/if}
					{:else}
						<Alert tone="danger" title="Not recognized">
							Plume has no document matching this fingerprint. The file may have been altered after
							signing, was not issued through this server, or belongs to a different Plume instance.
						</Alert>
					{/if}

					<Card class="flex flex-col gap-4">
						<p class="text-fc-sm font-medium text-fc-fg">SHA-256 fingerprint</p>
						<p class="break-all rounded-fc-md bg-fc-surface p-3 font-mono text-fc-xs text-fc-fg">
							{result.hash}
						</p>
					</Card>
				</section>
			{/if}

			<div class="flex flex-col gap-4">
				<Divider class="my-0" />
				<p class="text-center text-fc-sm text-fc-fg-muted">
					Verification compares the SHA-256 hash of the uploaded file against hashes Plume recorded at
					upload and at completion. Any change to a single byte produces a different fingerprint.
				</p>
			</div>
		</div>
	</main>

	<footer class="border-t border-fc-border">
		<div class="mx-auto w-full max-w-3xl px-4 py-6 text-center text-fc-sm text-fc-fg-muted sm:px-6">
			© {new Date().getFullYear()} Plume by
			<a
				href="https://facile.studio"
				class="font-medium text-fc-fg underline underline-offset-2 hover:opacity-80"
			>
				Facile.
			</a>
		</div>
	</footer>
</div>
