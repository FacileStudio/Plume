<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Button,
		SettingsRow,
		SettingsSection,
		TokensSection,
		icons,
		toast,
		type TokenDraft,
		type TokenItem
	} from '@facile/muse';
	import { api, type ApiToken } from '$lib/backend';
	import { formatDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let tokens = $state<ApiToken[]>([]);

	const items = $derived<TokenItem[]>(
		tokens.map((token) => ({
			id: String(token.id),
			name: token.label,
			lastUsed: token.last_used_at ? formatDate(token.last_used_at) : undefined
		}))
	);

	onMount(load);

	async function load() {
		try {
			tokens = await api.auth.apiTokens();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('api.loadFailed'));
		}
	}

	async function generate(draft: Omit<TokenDraft, 'days'>): Promise<string> {
		const created = await api.auth.createApiToken(draft.name);
		toast.success(t('api.generated'));
		await load();
		return created.token;
	}

	async function revoke(target: TokenItem) {
		await api.auth.revokeApiToken(Number(target.id));
		toast.success(t('api.revoked'));
		await load();
	}
</script>

<svelte:head><title>{t('api.title')}</title></svelte:head>

<TokensSection tokens={items} expiry={false} onCreate={generate} onRevoke={revoke}>
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

		<SettingsRow label={t('api.scheme')} description={t('api.schemeText')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">Bearer</span>
		</SettingsRow>
	</SettingsSection>
</TokensSection>
