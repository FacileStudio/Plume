<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Button,
		ConfirmModal,
		Drawer,
		EmptyState,
		Field,
		Input,
		SecretField,
		SettingsRow,
		SettingsSection,
		StatusDot,
		Switch,
		icons,
		toast
	} from '@facile/muse';
	import { api } from '$lib';
	import type { Webhook } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';

	let smtpHost = $state('');
	let smtpPort = $state(587);
	let smtpUsername = $state('');
	let smtpPassword = $state('');
	let smtpFromEmail = $state('');
	let smtpFromName = $state('');
	let smtpConfigured = $state(false);
	let smtpSaving = $state(false);
	let smtpTesting = $state(false);
	let smtpDeleting = $state(false);
	let confirmSmtpDelete = $state(false);

	let webhooks = $state<Webhook[]>([]);
	let webhookDrawer = $state(false);
	let webhookUrl = $state('');
	let webhookSecret = $state('');
	let editingWebhookId = $state<number | null>(null);
	let webhookSaving = $state(false);
	let testingWebhookId = $state<number | null>(null);
	let confirmWebhookDelete = $state(false);
	let pendingWebhook = $state<Webhook | null>(null);

	const webhookEventTypes = [
		'document.created',
		'document.sent',
		'document.completed',
		'document.declined',
		'document.deleted',
		'signer.added',
		'signer.email_opened',
		'signer.viewed',
		'signer.signed',
		'signer.declined',
		'signer.reminded'
	];

	onMount(() => {
		loadSmtp();
		loadWebhooks();
	});

	async function loadSmtp() {
		try {
			const config = await api.smtp.get();
			if (!config) {
				smtpConfigured = false;
				return;
			}
			smtpHost = config.host;
			smtpPort = config.port;
			smtpUsername = config.username;
			smtpPassword = '';
			smtpFromEmail = config.from_email;
			smtpFromName = config.from_name;
			smtpConfigured = true;
		} catch {
			smtpConfigured = false;
		}
	}

	async function saveSmtp() {
		smtpSaving = true;
		try {
			await api.smtp.save({
				host: smtpHost,
				port: smtpPort,
				username: smtpUsername,
				password: smtpPassword,
				from_email: smtpFromEmail,
				from_name: smtpFromName
			});
			smtpPassword = '';
			smtpConfigured = true;
			toast.success('SMTP configuration saved');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to save SMTP configuration');
		}
		smtpSaving = false;
	}

	async function deleteSmtp() {
		smtpDeleting = true;
		try {
			await api.smtp.delete();
			smtpHost = '';
			smtpPort = 587;
			smtpUsername = '';
			smtpPassword = '';
			smtpFromEmail = '';
			smtpFromName = '';
			smtpConfigured = false;
			toast.success('SMTP configuration removed');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to delete SMTP configuration');
		}
		smtpDeleting = false;
	}

	async function testSmtp() {
		const email = userStore.value?.email;
		if (!email) {
			toast.danger('No email address found for current user');
			return;
		}
		smtpTesting = true;
		try {
			await api.smtp.test(email);
			toast.success(`Test email sent to ${email}`);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to send test email');
		}
		smtpTesting = false;
	}

	async function loadWebhooks() {
		try {
			webhooks = await api.webhooks.list();
		} catch {}
	}

	function openWebhookDrawer(wh?: Webhook) {
		editingWebhookId = wh?.id ?? null;
		webhookUrl = wh?.url ?? '';
		webhookSecret = '';
		webhookDrawer = true;
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
				await api.webhooks.create({ url: webhookUrl, secret: webhookSecret });
			}
			webhookDrawer = false;
			editingWebhookId = null;
			webhookUrl = '';
			webhookSecret = '';
			await loadWebhooks();
			toast.success('Webhook saved');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to save webhook');
		}
		webhookSaving = false;
	}

	async function toggleWebhookEnabled(wh: Webhook) {
		try {
			await api.webhooks.update(wh.id, { url: wh.url, secret: '', enabled: !wh.enabled });
			await loadWebhooks();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to update webhook');
		}
	}

	async function deleteWebhook() {
		const target = pendingWebhook;
		if (!target) return;
		try {
			await api.webhooks.delete(target.id);
			await loadWebhooks();
			toast.success('Webhook deleted');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to delete webhook');
		}
		pendingWebhook = null;
	}

	async function testWebhook(id: number) {
		if (testingWebhookId !== null) return;
		testingWebhookId = id;
		try {
			await api.webhooks.test(id);
			toast.success('Test event delivered');
			await loadWebhooks();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Test delivery failed');
		}
		testingWebhookId = null;
	}

	function lastDelivery(wh: Webhook) {
		return wh.last_sent_at
			? `Last delivered ${new Date(wh.last_sent_at).toLocaleString()}`
			: 'Never delivered';
	}
</script>

<svelte:head><title>Integrations — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Email (SMTP)"
		description="The relay Plume sends signing invitations, reminders and completion notices through."
	>
		{#snippet actions()}
			{#if smtpConfigured}
				<Button variant="outline" icon={icons.mail} disabled={smtpTesting} onclick={testSmtp}>
					{smtpTesting ? 'Sending…' : 'Send a test'}
				</Button>
			{/if}
		{/snippet}

		<SettingsRow label="Status" description="Saved credentials, not a live connection check.">
			<StatusDot
				tone={smtpConfigured ? 'success' : 'neutral'}
				label={smtpConfigured ? `Configured — ${smtpHost}` : 'Not configured'}
			/>
		</SettingsRow>

		<SettingsRow stacked label="Host" description="The relay's hostname, without a scheme.">
			<Field>
				<Input bind:value={smtpHost} placeholder="smtp.example.com" aria-label="SMTP host" />
			</Field>
		</SettingsRow>

		<SettingsRow stacked label="Port" description="587 for STARTTLS, 465 for implicit TLS.">
			<Field>
				<Input type="number" bind:value={smtpPort} placeholder="587" aria-label="SMTP port" />
			</Field>
		</SettingsRow>

		<SettingsRow stacked label="Username" description="Leave empty for a relay that takes no auth.">
			<Field>
				<Input bind:value={smtpUsername} placeholder="user@example.com" aria-label="SMTP username" />
			</Field>
		</SettingsRow>

		<SettingsRow stacked>
			<SecretField
				bind:value={smtpPassword}
				editable
				label="Password"
				helper={smtpConfigured
					? 'Plume never sends a stored password back to the browser. Leave this empty to keep the one on file, or type a new one to replace it.'
					: 'Stored server-side and never returned once saved.'}
				class="w-full"
			/>
		</SettingsRow>

		<SettingsRow stacked label="From email" description="The envelope sender every Plume email uses.">
			<Field>
				<Input
					type="email"
					bind:value={smtpFromEmail}
					placeholder="noreply@example.com"
					aria-label="From email"
				/>
			</Field>
		</SettingsRow>

		<SettingsRow stacked label="From name" description="The display name recipients see.">
			<Field>
				<Input bind:value={smtpFromName} placeholder="Plume" aria-label="From name" />
			</Field>
		</SettingsRow>

		<div class="flex flex-wrap gap-2 pt-1">
			<Button icon={icons.check} disabled={smtpSaving || !smtpHost} onclick={saveSmtp}>
				{smtpSaving ? 'Saving…' : 'Save'}
			</Button>
			{#if smtpConfigured}
				<Button
					variant="ghost-danger"
					icon={icons.remove}
					disabled={smtpDeleting}
					onclick={() => (confirmSmtpDelete = true)}
				>
					{smtpDeleting ? 'Deleting…' : 'Delete configuration'}
				</Button>
			{/if}
		</div>
	</SettingsSection>

	<SettingsSection
		title="Webhooks"
		description="Every document and signer event is POSTed to each enabled endpoint."
	>
		{#snippet actions()}
			<Button variant="outline" icon={icons.plus} onclick={() => openWebhookDrawer()}>
				Add webhook
			</Button>
		{/snippet}

		{#if webhooks.length === 0}
			<EmptyState
				bare
				icon={icons.plug}
				title="No endpoints yet"
				description="Add one and Plume starts delivering signed events to it immediately."
			/>
		{:else}
			{#each webhooks as wh (wh.id)}
				<SettingsRow label={wh.url} description={lastDelivery(wh)}>
					<div class="flex flex-wrap items-center gap-2">
						<StatusDot
							tone={wh.enabled ? 'success' : 'neutral'}
							label={wh.enabled ? 'Active' : 'Disabled'}
						/>
						<Switch
							checked={wh.enabled}
							onchange={() => toggleWebhookEnabled(wh)}
							aria-label={wh.enabled ? `Disable ${wh.url}` : `Enable ${wh.url}`}
						/>
						<Button
							variant="ghost"
							size="sm"
							icon={icons.bolt}
							disabled={testingWebhookId === wh.id}
							onclick={() => testWebhook(wh.id)}
						>
							{testingWebhookId === wh.id ? 'Sending…' : 'Test'}
						</Button>
						<Button
							variant="ghost"
							size="sm"
							icon={icons.edit}
							onclick={() => openWebhookDrawer(wh)}
						>
							Edit
						</Button>
						<Button
							variant="ghost-danger"
							size="sm"
							icon={icons.remove}
							onclick={() => {
								pendingWebhook = wh;
								confirmWebhookDelete = true;
							}}
						>
							Delete
						</Button>
					</div>
				</SettingsRow>
			{/each}
		{/if}
	</SettingsSection>

	<SettingsSection
		title="Event contract"
		description="What a receiver has to be ready for. Failed deliveries retry up to three times with backoff."
	>
		<SettingsRow
			stacked
			label="Event types"
			description="Every endpoint receives all of them — filtering happens on your side."
		>
			<div class="flex flex-wrap gap-1.5">
				{#each webhookEventTypes as eventType (eventType)}
					<span class="rounded-fc-sm bg-fc-surface px-1.5 py-0.5 font-fc-mono text-fc-xs text-fc-fg-muted">
						{eventType}
					</span>
				{/each}
			</div>
		</SettingsRow>

		<SettingsRow
			label="Signature header"
			description="Each delivery is signed HMAC-SHA256 with the endpoint's secret."
		>
			<span class="font-fc-mono text-fc-sm text-fc-fg">x-plume-signature-256</span>
		</SettingsRow>
	</SettingsSection>
</div>

<Drawer
	bind:open={webhookDrawer}
	title={editingWebhookId ? 'Edit webhook' : 'Add webhook'}
	description="Plume POSTs a JSON event to this URL and signs the body with the shared secret."
>
	<div class="flex flex-col gap-4">
		<Field label="URL">
			<Input bind:value={webhookUrl} placeholder="https://antenne.example.com/webhook/plume" />
		</Field>

		<SecretField
			bind:value={webhookSecret}
			editable
			label="Signing secret"
			helper={editingWebhookId
				? 'The stored secret is never sent back. Leave this empty to keep it, or type a new one to replace it.'
				: 'Used to compute the HMAC-SHA256 signature. Store the same value on the receiving end.'}
			class="w-full"
		/>
	</div>

	{#snippet footer()}
		<Button icon={icons.check} disabled={webhookSaving || !webhookUrl} onclick={saveWebhook}>
			{webhookSaving ? 'Saving…' : editingWebhookId ? 'Update' : 'Create'}
		</Button>
		<Button variant="outline" onclick={() => (webhookDrawer = false)}>Cancel</Button>
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={confirmSmtpDelete}
	tone="danger"
	title="Delete the SMTP configuration?"
	description="Plume stops sending signing invitations, reminders and completion notices until a relay is configured again. The stored password is destroyed and cannot be recovered."
	confirmLabel="Delete configuration"
	onConfirm={deleteSmtp}
/>

<ConfirmModal
	bind:open={confirmWebhookDelete}
	tone="danger"
	title="Delete this webhook?"
	description={pendingWebhook
		? `${pendingWebhook.url} stops receiving events immediately and its signing secret is destroyed. Anything downstream that relies on these deliveries goes quiet.`
		: ''}
	confirmLabel="Delete webhook"
	onConfirm={deleteWebhook}
	onCancel={() => (pendingWebhook = null)}
/>
