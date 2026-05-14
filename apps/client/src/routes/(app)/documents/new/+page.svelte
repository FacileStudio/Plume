<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import Icon from '@iconify/svelte';

	let name = $state('');
	let file = $state<File | null>(null);
	let signers = $state<{ name: string; email: string }[]>([{ name: '', email: '' }]);
	let error = $state('');
	let submitting = $state(false);
	let dragging = $state(false);

	function addSigner() {
		signers = [...signers, { name: '', email: '' }];
	}

	function removeSigner(index: number) {
		signers = signers.filter((_, i) => i !== index);
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const dropped = e.dataTransfer?.files?.[0];
		if (dropped && dropped.type === 'application/pdf') {
			file = dropped;
		}
	}

	function handleFileInput(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		if (input.files?.[0]) {
			file = input.files[0];
		}
	}

	async function submit() {
		error = '';
		if (!name.trim()) {
			error = 'Document name is required';
			return;
		}
		if (!file) {
			error = 'Please upload a PDF file';
			return;
		}

		submitting = true;
		try {
			const doc = await api.documents.create(name, file);

			const validSigners = signers.filter((s) => s.name.trim() && s.email.trim());
			for (const signer of validSigners) {
				await api.signers.add(doc.id, signer.name, signer.email);
			}

			goto(`/documents/${doc.id}`);
		} catch (e: any) {
			error = e.message;
			submitting = false;
		}
	}
</script>

<svelte:head><title>New Document — Plume</title></svelte:head>

<div class="max-w-lg">
	<div class="mb-8">
		<a href="/documents" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors mb-4">
			<Icon icon="solar:arrow-left-linear" class="h-4 w-4" />
			Back to documents
		</a>
		<h1 class="text-2xl font-bold">New document</h1>
	</div>

	<form onsubmit={submit} class="space-y-6">
		<div class="space-y-2">
			<Label for="doc-name">Document name</Label>
			<Input id="doc-name" bind:value={name} placeholder="Contract, NDA, Agreement..." required />
		</div>

		<div class="space-y-2">
			<Label>PDF file</Label>
			<label
				class="flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 cursor-pointer transition-colors
					{dragging ? 'border-foreground bg-muted/50' : 'border-border hover:border-foreground/30'}"
				ondragover={(e) => { e.preventDefault(); dragging = true; }}
				ondragleave={() => (dragging = false)}
				ondrop={handleDrop}
			>
				{#if file}
					<Icon icon="solar:file-check-linear" class="h-8 w-8 text-green-600 mb-2" />
					<p class="text-sm font-medium">{file.name}</p>
					<p class="text-xs text-muted-foreground mt-1">{(file.size / 1024).toFixed(0)} KB</p>
				{:else}
					<Icon icon="solar:upload-linear" class="h-8 w-8 text-muted-foreground mb-2" />
					<p class="text-sm text-muted-foreground">Drag & drop a PDF or click to browse</p>
				{/if}
				<input type="file" accept=".pdf,application/pdf" onchange={handleFileInput} class="hidden" />
			</label>
		</div>

		<Separator />

		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<Label>Signers</Label>
				<button
					type="button"
					onclick={addSigner}
					class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
				>
					<Icon icon="mdi:plus" class="h-4 w-4" />
					Add signer
				</button>
			</div>

			{#each signers as signer, i}
				<div class="flex items-start gap-2">
					<div class="flex-1 space-y-2">
						<Input bind:value={signer.name} placeholder="Name" />
						<Input bind:value={signer.email} placeholder="Email" type="email" />
					</div>
					{#if signers.length > 1}
						<button
							type="button"
							onclick={() => removeSigner(i)}
							class="mt-2 rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
						>
							<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
						</button>
					{/if}
				</div>
			{/each}
		</div>

		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}

		<Button type="submit" disabled={submitting} class="w-full">
			{#if submitting}
				<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
				Creating...
			{:else}
				<Icon icon="mdi:file-document-plus-outline" class="h-4 w-4" />
				Create document
			{/if}
		</Button>
	</form>
</div>
