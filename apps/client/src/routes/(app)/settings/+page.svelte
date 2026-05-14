<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Webhook } from '$lib';
	import Icon from '@iconify/svelte';

	let webhooks = $state<Webhook[]>([]);
	let showWebhookForm = $state(false);
	let webhookUrl = $state('');
	let webhookSecret = $state('');
	let editingWebhookId = $state<number | null>(null);
	let webhookSaving = $state(false);
	let deletingWebhookId = $state<number | null>(null);

	async function loadWebhooks() {
		try {
			webhooks = await api.webhooks.list();
		} catch {}
	}

	function resetWebhookForm() {
		webhookUrl = '';
		webhookSecret = '';
		editingWebhookId = null;
		showWebhookForm = false;
	}

	function startEditWebhook(wh: Webhook) {
		webhookUrl = wh.url;
		webhookSecret = '';
		editingWebhookId = wh.id;
		showWebhookForm = true;
	}

	async function saveWebhook() {
		webhookSaving = true;
		try {
			if (editingWebhookId) {
				const existing = webhooks.find((w) => w.id === editingWebhookId);
				await api.webhooks.update(editingWebhookId, {
					url: webhookUrl,
					secret: webhookSecret,
					enabled: existing?.enabled ?? true
				});
			} else {
				await api.webhooks.create({
					url: webhookUrl,
					secret: webhookSecret
				});
			}
			resetWebhookForm();
			await loadWebhooks();
		} catch {}
		webhookSaving = false;
	}

	async function toggleWebhookEnabled(wh: Webhook) {
		try {
			await api.webhooks.update(wh.id, {
				url: wh.url,
				secret: '',
				enabled: !wh.enabled
			});
			await loadWebhooks();
		} catch {}
	}

	async function deleteWebhook(id: number) {
		try {
			await api.webhooks.delete(id);
			deletingWebhookId = null;
			await loadWebhooks();
		} catch {}
	}

	onMount(async () => {
		await loadWebhooks();
	});
</script>

<svelte:head><title>Settings — Plume</title></svelte:head>

<h1 class="text-2xl font-bold mb-6">Settings</h1>

<div class="space-y-8 max-w-lg">
	<div class="rounded-lg border p-6">
		<div class="flex items-center justify-between mb-1">
			<h2 class="text-lg font-semibold">Webhooks</h2>
			{#if !showWebhookForm && webhooks.length > 0}
				<button
					onclick={() => { resetWebhookForm(); showWebhookForm = true; }}
					class="flex items-center gap-1.5 rounded-full bg-foreground px-3 py-1 text-sm font-medium text-background transition-colors hover:bg-foreground/90"
				>
					<Icon icon="mdi:plus-circle-outline" class="h-4 w-4" />
					Add
				</button>
			{/if}
		</div>
		<p class="text-sm text-muted-foreground mb-4">Receive notifications when documents are signed, completed, or declined</p>

		{#if webhooks.length === 0 && !showWebhookForm}
			<button
				onclick={() => (showWebhookForm = true)}
				class="flex items-center gap-2 rounded-lg border border-dashed px-4 py-3 text-sm text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground w-full justify-center"
			>
				<Icon icon="mdi:plus-circle-outline" class="h-4 w-4" />
				Add Webhook
			</button>
		{/if}

		{#if webhooks.length > 0}
			<div class="space-y-3 mb-4">
				{#each webhooks as wh}
					<div class="rounded-lg border p-4">
						{#if deletingWebhookId === wh.id}
							<div class="flex items-center justify-between">
								<p class="text-sm">Delete this webhook?</p>
								<div class="flex gap-2">
									<button
										onclick={() => deleteWebhook(wh.id)}
										class="rounded-full bg-red-500 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-red-600"
									>
										Delete
									</button>
									<button
										onclick={() => (deletingWebhookId = null)}
										class="rounded-full bg-muted px-3 py-1 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
									>
										Cancel
									</button>
								</div>
							</div>
						{:else}
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium">{wh.url}</p>
									<div class="mt-1.5 flex flex-wrap items-center gap-2">
										<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {wh.enabled ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground'}">
											{wh.enabled ? 'Active' : 'Disabled'}
										</span>
									</div>
								</div>
								<div class="flex items-center gap-1.5 shrink-0">
									<button
										onclick={() => toggleWebhookEnabled(wh)}
										class="relative h-6 w-10 rounded-full transition-colors {wh.enabled ? 'bg-green-500' : 'bg-muted'}"
										aria-label="{wh.enabled ? 'Disable' : 'Enable'} webhook"
									>
										<span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform {wh.enabled ? 'left-[18px]' : 'left-0.5'}"></span>
									</button>
									<button
										onclick={() => startEditWebhook(wh)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted"
										aria-label="Edit webhook"
									>
										<Icon icon="solar:pen-linear" class="h-3.5 w-3.5" />
									</button>
									<button
										onclick={() => (deletingWebhookId = wh.id)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
										aria-label="Delete webhook"
									>
										<Icon icon="solar:trash-bin-trash-linear" class="h-3.5 w-3.5" />
									</button>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		{#if showWebhookForm}
			<div class="rounded-lg border p-4 space-y-4">
				<div>
					<label for="webhook-url" class="block text-sm font-medium mb-1">URL</label>
					<input
						id="webhook-url"
						type="text"
						bind:value={webhookUrl}
						placeholder="https://nook.example.com/webhook/plume"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div>
					<label for="webhook-secret" class="block text-sm font-medium mb-1">Secret</label>
					<input
						id="webhook-secret"
						type="password"
						bind:value={webhookSecret}
						placeholder="Shared HMAC secret"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div class="flex gap-2 pt-1">
					<button
						onclick={saveWebhook}
						disabled={webhookSaving || !webhookUrl}
						class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<Icon icon="solar:diskette-linear" class="h-4 w-4" />
						{webhookSaving ? 'Saving…' : editingWebhookId ? 'Update' : 'Save'}
					</button>
					<button
						onclick={resetWebhookForm}
						class="flex items-center gap-1.5 rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
					>
						<Icon icon="solar:close-circle-linear" class="h-4 w-4" />
						Cancel
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
