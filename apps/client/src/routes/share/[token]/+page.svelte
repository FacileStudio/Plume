<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib';
	import type { SigningPayload, Field } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import Icon from '@iconify/svelte';
	import PdfViewer from '$lib/components/pdf-viewer.svelte';

	let payload = $state<SigningPayload | null>(null);
	let loading = $state(true);
	let notFound = $state(false);
	let submitting = $state(false);
	let signed = $state(false);
	let declined = $state(false);
	let error = $state('');
	let fieldValues = $state<Record<string, string>>({});

	function initFields(fields: Field[]) {
		const values: Record<string, string> = {};
		for (const f of fields) {
			values[String(f.id)] = f.value ?? '';
		}
		fieldValues = values;
	}

	function fieldLabel(f: Field): string {
		switch (f.field_type) {
			case 'signature': return 'Signature';
			case 'text': return 'Text';
			case 'date': return 'Date';
			case 'checkbox': return 'Checkbox';
			default: return 'Field';
		}
	}

	async function signDocument() {
		const token = (page.params as Record<string, string>).token;
		submitting = true;
		error = '';
		try {
			await api.signing.sign(token, fieldValues);
			signed = true;
		} catch (e: any) {
			error = e.message;
		}
		submitting = false;
	}

	async function declineDocument() {
		const token = (page.params as Record<string, string>).token;
		submitting = true;
		error = '';
		try {
			await api.signing.decline(token);
			declined = true;
		} catch (e: any) {
			error = e.message;
		}
		submitting = false;
	}

	onMount(async () => {
		const token = (page.params as Record<string, string>).token;
		try {
			payload = await api.signing.get(token);
			if (payload.signer.status === 'signed') {
				signed = true;
			} else if (payload.signer.status === 'declined') {
				declined = true;
			} else {
				initFields(payload.fields);
			}
		} catch {
			notFound = true;
		}
		loading = false;
	});
</script>

<svelte:head><title>{payload ? `Sign — ${payload.document.name}` : 'Sign Document'} — Plume</title></svelte:head>

<div class="flex min-h-[100dvh] flex-col">
	<header class="flex items-center gap-3 border-b px-6 py-4">
		<Icon icon="solar:document-add-bold-duotone" class="h-6 w-6" />
		<span class="text-lg font-bold tracking-tight">Plume</span>
	</header>

	<main class="flex flex-1 items-start justify-center p-6">
		{#if loading}
			<div class="flex flex-col items-center gap-3">
				<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
			</div>
		{:else if notFound}
			<div class="flex flex-col items-center gap-4 text-center">
				<Icon icon="solar:eye-closed-linear" class="h-12 w-12 text-muted-foreground" />
				<h1 class="text-xl font-semibold">Link not found</h1>
				<p class="text-muted-foreground">This signing link may be invalid or expired.</p>
			</div>
		{:else if signed}
			<div class="flex flex-col items-center gap-4 text-center">
				<Icon icon="solar:check-circle-bold-duotone" class="h-14 w-14 text-green-600" />
				<h1 class="text-xl font-semibold">Document signed successfully</h1>
				<p class="text-muted-foreground">You can close this page.</p>
			</div>
		{:else if declined}
			<div class="flex flex-col items-center gap-4 text-center">
				<Icon icon="solar:close-circle-bold-duotone" class="h-14 w-14 text-red-500" />
				<h1 class="text-xl font-semibold">Document declined</h1>
				<p class="text-muted-foreground">You have declined to sign this document.</p>
			</div>
		{:else if payload}
			<div class="flex w-full max-w-6xl gap-8 flex-col lg:flex-row">
				<div class="flex-1 min-w-0 max-h-[calc(100dvh-10rem)] overflow-y-auto rounded-lg border bg-muted/30 p-4">
					<PdfViewer url={api.signing.fileUrl((page.params as Record<string, string>).token)} />
				</div>

				<div class="w-full lg:w-80 shrink-0 space-y-6">
					<div class="text-center">
						<h1 class="text-xl font-semibold mb-1">{payload.document.name}</h1>
						<p class="text-sm text-muted-foreground">
							Signing as <span class="font-medium text-foreground">{payload.signer.name}</span>
						</p>
					</div>

					<Separator />

					{#if payload.fields.length > 0}
						<div class="space-y-4">
							{#each payload.fields as field}
								<div class="space-y-2">
									<Label for="field-{field.id}">
										{fieldLabel(field)}
										{#if field.required}
											<span class="text-destructive">*</span>
										{/if}
									</Label>
									{#if field.field_type === 'signature'}
										<Input
											id="field-{field.id}"
											bind:value={fieldValues[String(field.id)]}
											placeholder="Type your full name as signature"
											class="font-serif italic text-lg"
										/>
									{:else if field.field_type === 'date'}
										<Input
											id="field-{field.id}"
											type="date"
											bind:value={fieldValues[String(field.id)]}
										/>
									{:else if field.field_type === 'checkbox'}
										<label class="flex items-center gap-2 cursor-pointer">
											<input
												type="checkbox"
												checked={fieldValues[String(field.id)] === 'true'}
												onchange={(e) => {
													fieldValues[String(field.id)] = (e.currentTarget as HTMLInputElement).checked ? 'true' : 'false';
												}}
												class="h-4 w-4 rounded border-border"
											/>
											<span class="text-sm">I agree</span>
										</label>
									{:else}
										<Input
											id="field-{field.id}"
											bind:value={fieldValues[String(field.id)]}
											placeholder="Enter text"
										/>
									{/if}
								</div>
							{/each}
						</div>

						<Separator />
					{/if}

					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}

					<div class="flex gap-3">
						<Button onclick={signDocument} disabled={submitting} class="flex-1">
							{#if submitting}
								<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
							{:else}
								<Icon icon="solar:pen-new-square-linear" class="h-4 w-4" />
							{/if}
							Sign & complete
						</Button>
						<Button onclick={declineDocument} disabled={submitting} variant="outline">
							Decline
						</Button>
					</div>
				</div>
			</div>
		{/if}
	</main>

	<footer class="flex items-center justify-center border-t px-6 py-4 text-xs text-muted-foreground">
		<Icon icon="solar:document-add-bold-duotone" class="mr-1.5 h-3.5 w-3.5" />
		Powered by Plume
	</footer>
</div>
