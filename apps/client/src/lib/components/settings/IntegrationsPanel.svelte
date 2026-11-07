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
	import { formatDateTime } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

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
			toast.success(t('smtp.saved'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('smtp.saveFailed'));
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
			toast.success(t('smtp.removed'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('smtp.removeFailed'));
		}
		smtpDeleting = false;
	}

	async function testSmtp() {
		const email = userStore.value?.email;
		if (!email) {
			toast.danger(t('smtp.noAddress'));
			return;
		}
		smtpTesting = true;
		try {
			await api.smtp.test(email);
			toast.success(t('smtp.testSent', { email }));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('smtp.testFailed'));
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
			toast.success(t('webhooks.saved'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.saveFailed'));
		}
		webhookSaving = false;
	}

	async function toggleWebhookEnabled(wh: Webhook) {
		try {
			await api.webhooks.update(wh.id, { url: wh.url, secret: '', enabled: !wh.enabled });
			await loadWebhooks();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.updateFailed'));
		}
	}

	async function deleteWebhook() {
		const target = pendingWebhook;
		if (!target) return;
		try {
			await api.webhooks.delete(target.id);
			await loadWebhooks();
			toast.success(t('webhooks.deleted'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.deleteFailed'));
		}
		pendingWebhook = null;
	}

	async function testWebhook(id: number) {
		if (testingWebhookId !== null) return;
		testingWebhookId = id;
		try {
			await api.webhooks.test(id);
			toast.success(t('webhooks.delivered'));
			await loadWebhooks();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('webhooks.deliveryFailed'));
		}
		testingWebhookId = null;
	}

	function lastDelivery(wh: Webhook) {
		return wh.last_sent_at
			? t('webhooks.lastDelivered', { date: formatDateTime(wh.last_sent_at) })
			: t('webhooks.neverDelivered');
	}
</script>

<svelte:head><title>{t('integrations.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('smtp.title')} description={t('smtp.text')}>
		{#snippet actions()}
			{#if smtpConfigured}
				<Button variant="outline" icon={icons.mail} disabled={smtpTesting} onclick={testSmtp}>
					{smtpTesting ? t('common.sending') : t('smtp.sendTest')}
				</Button>
			{/if}
		{/snippet}

		<SettingsRow label={t('common.status')} description={t('smtp.statusText')}>
			<StatusDot
				tone={smtpConfigured ? 'success' : 'neutral'}
				label={smtpConfigured ? t('smtp.configured', { host: smtpHost }) : t('smtp.notConfigured')}
			/>
		</SettingsRow>

		<SettingsRow stacked label={t('smtp.host')} description={t('smtp.hostText')}>
			<Field>
				<Input bind:value={smtpHost} placeholder="smtp.example.com" aria-label={t('smtp.hostAria')} />
			</Field>
		</SettingsRow>

		<SettingsRow stacked label={t('smtp.port')} description={t('smtp.portText')}>
			<Field>
				<Input type="number" bind:value={smtpPort} placeholder="587" aria-label={t('smtp.portAria')} />
			</Field>
		</SettingsRow>

		<SettingsRow stacked label={t('smtp.username')} description={t('smtp.usernameText')}>
			<Field>
				<Input
					bind:value={smtpUsername}
					placeholder="user@example.com"
					aria-label={t('smtp.usernameAria')}
				/>
			</Field>
		</SettingsRow>

		<SettingsRow stacked>
			<SecretField
				bind:value={smtpPassword}
				editable
				label={t('smtp.password')}
				helper={smtpConfigured ? t('smtp.passwordKeep') : t('smtp.passwordNew')}
				class="w-full"
			/>
		</SettingsRow>

		<SettingsRow stacked label={t('smtp.fromEmail')} description={t('smtp.fromEmailText')}>
			<Field>
				<Input
					type="email"
					bind:value={smtpFromEmail}
					placeholder="noreply@example.com"
					aria-label={t('smtp.fromEmail')}
				/>
			</Field>
		</SettingsRow>

		<SettingsRow stacked label={t('smtp.fromName')} description={t('smtp.fromNameText')}>
			<Field>
				<Input bind:value={smtpFromName} placeholder="Plume" aria-label={t('smtp.fromName')} />
			</Field>
		</SettingsRow>

		<div class="flex flex-wrap gap-2 pt-1">
			<Button icon={icons.check} disabled={smtpSaving || !smtpHost} onclick={saveSmtp}>
				{smtpSaving ? t('common.saving') : t('common.save')}
			</Button>
			{#if smtpConfigured}
				<Button
					variant="ghost-danger"
					icon={icons.remove}
					disabled={smtpDeleting}
					onclick={() => (confirmSmtpDelete = true)}
				>
					{smtpDeleting ? t('common.deleting') : t('smtp.deleteConfig')}
				</Button>
			{/if}
		</div>
	</SettingsSection>

	<SettingsSection title={t('webhooks.title')} description={t('webhooks.text')}>
		{#snippet actions()}
			<Button variant="outline" icon={icons.plus} onclick={() => openWebhookDrawer()}>
				{t('webhooks.add')}
			</Button>
		{/snippet}

		{#if webhooks.length === 0}
			<EmptyState
				bare
				icon={icons.plug}
				title={t('webhooks.emptyTitle')}
				description={t('webhooks.emptyText')}
			/>
		{:else}
			{#each webhooks as wh (wh.id)}
				<SettingsRow label={wh.url} description={lastDelivery(wh)}>
					<div class="flex flex-wrap items-center gap-2">
						<StatusDot
							tone={wh.enabled ? 'success' : 'neutral'}
							label={wh.enabled ? t('webhooks.active') : t('webhooks.disabled')}
						/>
						<Switch
							checked={wh.enabled}
							onchange={() => toggleWebhookEnabled(wh)}
							aria-label={wh.enabled
								? t('webhooks.disable', { url: wh.url })
								: t('webhooks.enable', { url: wh.url })}
						/>
						<Button
							variant="ghost"
							size="sm"
							icon={icons.bolt}
							disabled={testingWebhookId === wh.id}
							onclick={() => testWebhook(wh.id)}
						>
							{testingWebhookId === wh.id ? t('common.sending') : t('webhooks.test')}
						</Button>
						<Button
							variant="ghost"
							size="sm"
							icon={icons.edit}
							onclick={() => openWebhookDrawer(wh)}
						>
							{t('common.edit')}
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
							{t('common.delete')}
						</Button>
					</div>
				</SettingsRow>
			{/each}
		{/if}
	</SettingsSection>

	<SettingsSection title={t('webhooks.contract')} description={t('webhooks.contractText')}>
		<SettingsRow stacked label={t('webhooks.eventTypes')} description={t('webhooks.eventTypesText')}>
			<div class="flex flex-wrap gap-1.5">
				{#each webhookEventTypes as eventType (eventType)}
					<span class="rounded-fc-sm bg-fc-surface px-1.5 py-0.5 font-fc-mono text-fc-xs text-fc-fg-muted">
						{eventType}
					</span>
				{/each}
			</div>
		</SettingsRow>

		<SettingsRow label={t('webhooks.signatureHeader')} description={t('webhooks.signatureHeaderText')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">x-plume-signature-256</span>
		</SettingsRow>
	</SettingsSection>
</div>

<Drawer
	bind:open={webhookDrawer}
	title={editingWebhookId ? t('webhooks.drawer.edit') : t('webhooks.add')}
	description={t('webhooks.drawer.text')}
>
	<div class="flex flex-col gap-4">
		<Field label={t('webhooks.url')}>
			<Input bind:value={webhookUrl} placeholder="https://nook.example.com/webhook/plume" />
		</Field>

		<SecretField
			bind:value={webhookSecret}
			editable
			label={t('webhooks.secret')}
			helper={editingWebhookId ? t('webhooks.secretKeep') : t('webhooks.secretNew')}
			class="w-full"
		/>
	</div>

	{#snippet footer()}
		<Button icon={icons.check} disabled={webhookSaving || !webhookUrl} onclick={saveWebhook}>
			{webhookSaving ? t('common.saving') : editingWebhookId ? t('common.update') : t('common.create')}
		</Button>
		<Button variant="outline" onclick={() => (webhookDrawer = false)}>{t('common.cancel')}</Button>
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={confirmSmtpDelete}
	tone="danger"
	title={t('smtp.confirm.title')}
	description={t('smtp.confirm.text')}
	confirmLabel={t('smtp.deleteConfig')}
	onConfirm={deleteSmtp}
/>

<ConfirmModal
	bind:open={confirmWebhookDelete}
	tone="danger"
	title={t('webhooks.confirm.title')}
	description={pendingWebhook ? t('webhooks.confirm.text', { url: pendingWebhook.url }) : ''}
	confirmLabel={t('webhooks.confirm.label')}
	onConfirm={deleteWebhook}
	onCancel={() => (pendingWebhook = null)}
/>
