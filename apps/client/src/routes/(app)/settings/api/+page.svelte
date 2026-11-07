<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Button,
		ConfirmModal,
		Drawer,
		Input,
		SecretField,
		SettingsRow,
		SettingsSection,
		StatusDot,
		Table,
		icons,
		toast
	} from '@facile/muse';
	import { api, type ApiToken } from '$lib/backend';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let tokens = $state<ApiToken[]>([]);
	let busy = $state(false);
	let createOpen = $state(false);
	let revoking = $state<ApiToken | null>(null);
	let revokeOpen = $state(false);
	let label = $state('');
	let issued = $state('');

	onMount(load);

	async function load() {
		try {
			tokens = await api.auth.apiTokens();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('api.loadFailed'));
		}
	}

	function openCreate() {
		issued = '';
		label = '';
		createOpen = true;
	}

	async function generate() {
		busy = true;
		try {
			const created = await api.auth.createApiToken(label.trim());
			issued = created.token;
			toast.success(t('api.generated'));
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('api.generateFailed'));
		} finally {
			busy = false;
		}
	}

	async function revoke() {
		if (!revoking) return;
		const target = revoking;
		busy = true;
		try {
			await api.auth.revokeApiToken(target.id);
			toast.success(t('api.revoked'));
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('api.revokeFailed'));
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{t('api.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('api.reference')} description={t('api.referenceText')}>
		{#snippet actions()}
			<Button href="/docs" target="_blank" rel="noopener" icon={icons.code}>
				{t('api.openReference')}
			</Button>
		{/snippet}

		<SettingsRow label={t('api.basePath')} description={t('api.basePathText')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
		</SettingsRow>

		<SettingsRow label={t('api.openapi')} description={t('api.openapiText')}>
			<Button variant="outline" href="/docs/openapi.json" target="_blank" rel="noopener" icon={icons.download}>
				openapi.json
			</Button>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('api.auth')} description={t('api.authText')}>
		<SettingsRow label={t('api.scheme')} description={t('api.schemeText')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">Bearer</span>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('api.tokens')} description={t('api.tokensText')}>
		{#snippet actions()}
			<Button icon={icons.plus} onclick={openCreate}>{t('api.generate')}</Button>
		{/snippet}

		{#if tokens.length}
			<Table>
				<thead>
					<tr>
						<th>{t('api.col.name')}</th>
						<th>{t('api.col.created')}</th>
						<th>{t('api.col.lastUsed')}</th>
						<th class="text-right">{t('api.col.action')}</th>
					</tr>
				</thead>
				<tbody>
					{#each tokens as token (token.id)}
						<tr>
							<td class="font-fc-mono">{token.label}</td>
							<td>{formatDate(token.created_at)}</td>
							<td>{formatDate(token.last_used_at)}</td>
							<td class="text-right">
								<Button
									variant="ghost-danger"
									size="sm"
									icon={icons.revoke}
									disabled={busy}
									onclick={() => {
										revoking = token;
										revokeOpen = true;
									}}
								>
									{t('api.revoke')}
								</Button>
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		{:else}
			<SettingsRow label={t('api.noToken')} description={t('api.noTokenText')}>
				<StatusDot tone="neutral" label={t('api.noneGenerated')} />
			</SettingsRow>
		{/if}
	</SettingsSection>
</div>

<Drawer
	bind:open={createOpen}
	title={t('api.drawer.title')}
	description={t('api.drawer.text')}
	onClose={() => (issued = '')}
>
	{#if issued}
		<div class="flex flex-col gap-4">
			<Alert tone="warning" title={t('api.copyNow')}>{t('api.copyNowText')}</Alert>
			<SecretField value={issued} label={t('api.tokenLabel')} visible autoHideMs={0} />
		</div>
	{:else}
		<div class="flex flex-col gap-2">
			<label for="token-label" class="text-fc-sm font-medium text-fc-fg">{t('api.tokenName')}</label>
			<Input id="token-label" bind:value={label} placeholder={t('api.tokenNamePlaceholder')} />
			<p class="text-fc-xs text-fc-fg-muted">{t('api.tokenWarning')}</p>
		</div>
	{/if}

	{#snippet footer()}
		{#if issued}
			<Button variant="outline" onclick={() => (createOpen = false)}>{t('common.close')}</Button>
		{:else}
			<Button variant="outline" onclick={() => (createOpen = false)}>{t('common.cancel')}</Button>
			<Button icon={icons.key} disabled={busy || !label.trim()} onclick={generate}>
				{busy ? t('api.generating') : t('api.generateShort')}
			</Button>
		{/if}
	{/snippet}
</Drawer>

<ConfirmModal
	bind:open={revokeOpen}
	title={t('api.confirm.title')}
	description={t('api.confirm.text', { name: revoking?.label ?? t('api.confirm.it') })}
	confirmLabel={t('api.revoke')}
	tone="danger"
	icon={icons.revoke}
	onConfirm={revoke}
/>
