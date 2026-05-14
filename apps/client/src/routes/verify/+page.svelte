<script lang="ts">
	import { api, isAuthenticated } from '$lib';
	import type { VerifyResponse } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import Icon from '@iconify/svelte';

	let file = $state<File | null>(null);
	let dragging = $state(false);
	let loading = $state(false);
	let error = $state('');
	let result = $state<VerifyResponse | null>(null);

	function reset() {
		file = null;
		result = null;
		error = '';
	}

	function pickFile(input: File | undefined) {
		if (!input) return;
		if (input.type && input.type !== 'application/pdf') {
			error = 'Only PDF files can be verified';
			return;
		}
		error = '';
		file = input;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		pickFile(e.dataTransfer?.files?.[0]);
	}

	function handleInput(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		pickFile(input.files?.[0]);
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

	function statusTone(status: string): string {
		switch (status) {
			case 'completed':
				return 'bg-green-500/10 text-green-700 dark:text-green-400';
			case 'declined':
				return 'bg-red-500/10 text-red-700 dark:text-red-400';
			case 'pending':
			case 'signed':
				return 'bg-foreground/10 text-foreground';
			default:
				return 'bg-muted text-muted-foreground';
		}
	}
</script>

<svelte:head>
	<title>Verify a document — Plume</title>
	<meta name="description" content="Check whether a PDF was issued and signed through Plume." />
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<header class="border-b border-border">
		<div class="mx-auto flex max-w-3xl items-center justify-between px-6 py-4">
			<a href="/" class="flex h-14 items-center gap-3">
				<Icon icon="solar:document-add-bold-duotone" class="w-7 h-7" />
				<span class="text-2xl font-bold tracking-tight">Plume</span>
			</a>
			<div class="flex items-center gap-2">
				{#if isAuthenticated()}
					<Button variant="ghost" href="/dashboard">Dashboard</Button>
				{:else}
					<Button variant="ghost" href="/login">Log in</Button>
				{/if}
			</div>
		</div>
	</header>

	<main class="mx-auto max-w-3xl px-6 py-16">
		<div class="mb-10 text-center">
			<div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-border bg-background">
				<Icon icon="solar:shield-check-linear" class="h-6 w-6" />
			</div>
			<h1 class="text-3xl font-bold tracking-tight">Verify a document</h1>
			<p class="mx-auto mt-3 max-w-lg text-muted-foreground">
				Drop a PDF below. Plume re-computes its SHA-256 fingerprint and checks whether it matches a document we have on file. Nothing is stored.
			</p>
		</div>

		<Card.Root>
			<Card.Content class="p-6">
				<label
					class="flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-10 cursor-pointer transition-colors
						{dragging ? 'border-foreground bg-muted/50' : 'border-border hover:border-foreground/30'}"
					ondragover={(e) => { e.preventDefault(); dragging = true; }}
					ondragleave={() => (dragging = false)}
					ondrop={handleDrop}
				>
					{#if file}
						<Icon icon="solar:file-check-linear" class="h-10 w-10 text-green-600 mb-3" />
						<p class="text-sm font-medium">{file.name}</p>
						<p class="text-xs text-muted-foreground mt-1">{(file.size / 1024).toFixed(0)} KB</p>
					{:else}
						<Icon icon="solar:upload-linear" class="h-10 w-10 text-muted-foreground mb-3" />
						<p class="text-sm text-muted-foreground">Drag & drop a PDF or click to browse</p>
						<p class="text-xs text-muted-foreground mt-1">Max 50 MB</p>
					{/if}
					<input type="file" accept=".pdf,application/pdf" onchange={handleInput} class="hidden" />
				</label>

				{#if error}
					<p class="mt-4 text-sm text-destructive">{error}</p>
				{/if}

				<div class="mt-6 flex items-center gap-2">
					<Button onclick={verify} disabled={!file || loading} class="flex-1">
						{#if loading}
							<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
							Verifying...
						{:else}
							<Icon icon="solar:shield-check-linear" class="h-4 w-4" />
							Verify document
						{/if}
					</Button>
					{#if file || result || error}
						<Button variant="outline" onclick={reset} disabled={loading}>
							<Icon icon="solar:refresh-linear" class="h-4 w-4" />
							Reset
						</Button>
					{/if}
				</div>
			</Card.Content>
		</Card.Root>

		{#if result}
			<div class="mt-6">
				{#if result.match && result.document}
					<Card.Root class="border-green-500/40 bg-green-500/5">
						<Card.Header>
							<div class="flex items-start gap-3">
								<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-green-500/15">
									<Icon icon="solar:check-circle-bold" class="h-6 w-6 text-green-600 dark:text-green-400" />
								</div>
								<div class="flex-1">
									<Card.Title class="text-lg">Authentic — issued by Plume</Card.Title>
									<Card.Description>
										This document matches our records. You uploaded the
										<span class="font-medium text-foreground">
											{result.variant === 'signed' ? 'signed' : 'original'}
										</span>
										version.
									</Card.Description>
								</div>
							</div>
						</Card.Header>
						<Card.Content class="space-y-5">
							<div>
								<h3 class="mb-2 text-sm font-semibold">Document</h3>
								<div class="rounded-md border bg-background p-4 space-y-2 text-sm">
									<div class="flex justify-between gap-4">
										<span class="text-muted-foreground">Name</span>
										<span class="font-medium text-right">{result.document.name}</span>
									</div>
									<div class="flex justify-between gap-4">
										<span class="text-muted-foreground">File</span>
										<span class="font-medium text-right break-all">{result.document.file_name}</span>
									</div>
									<div class="flex justify-between gap-4">
										<span class="text-muted-foreground">Status</span>
										<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {statusTone(result.document.status)}">
											{result.document.status}
										</span>
									</div>
									<div class="flex justify-between gap-4">
										<span class="text-muted-foreground">Uploaded</span>
										<span>{formatDate(result.document.created_at)}</span>
									</div>
									{#if result.document.completed_at}
										<div class="flex justify-between gap-4">
											<span class="text-muted-foreground">Completed</span>
											<span>{formatDate(result.document.completed_at)}</span>
										</div>
									{/if}
								</div>
							</div>

							{#if result.signers && result.signers.length > 0}
								<div>
									<h3 class="mb-2 text-sm font-semibold">Signers</h3>
									<div class="space-y-2">
										{#each result.signers as signer}
											<div class="flex items-center justify-between rounded-md border bg-background p-3">
												<div class="min-w-0">
													<p class="text-sm font-medium truncate">{signer.name}</p>
													<p class="text-xs text-muted-foreground truncate">{signer.email}</p>
												</div>
												<div class="flex items-center gap-3 shrink-0">
													{#if signer.signed_at}
														<span class="text-xs text-muted-foreground">{formatDate(signer.signed_at)}</span>
													{/if}
													<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {statusTone(signer.status)}">
														{signer.status}
													</span>
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}

							<div>
								<h3 class="mb-2 text-sm font-semibold">SHA-256 fingerprint</h3>
								<p class="break-all rounded-md border bg-background p-3 font-mono text-xs">{result.hash}</p>
							</div>
						</Card.Content>
					</Card.Root>
				{:else}
					<Card.Root class="border-red-500/40 bg-red-500/5">
						<Card.Header>
							<div class="flex items-start gap-3">
								<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-red-500/15">
									<Icon icon="solar:close-circle-bold" class="h-6 w-6 text-red-600 dark:text-red-400" />
								</div>
								<div class="flex-1">
									<Card.Title class="text-lg">Not recognized</Card.Title>
									<Card.Description>
										Plume has no document matching this fingerprint. The file may have been altered after signing, was not issued through this server, or belongs to a different Plume instance.
									</Card.Description>
								</div>
							</div>
						</Card.Header>
						<Card.Content>
							<h3 class="mb-2 text-sm font-semibold">SHA-256 fingerprint computed</h3>
							<p class="break-all rounded-md border bg-background p-3 font-mono text-xs">{result.hash}</p>
						</Card.Content>
					</Card.Root>
				{/if}
			</div>
		{/if}

		<Separator class="my-12" />

		<div class="text-center text-sm text-muted-foreground">
			<p>
				Verification compares the SHA-256 hash of the uploaded file against hashes Plume recorded at upload and at completion. Any change to a single byte produces a different fingerprint.
			</p>
		</div>
	</main>

	<footer class="border-t border-border text-center text-muted">
		<div class="mx-auto max-w-3xl px-6 py-6 text-sm text-muted-foreground">
			© {new Date().getFullYear()} Plume by
			<a href="https://facile.studio" class="underline hover:cursor-pointer font-semibold">Facile.</a>
		</div>
	</footer>
</div>
