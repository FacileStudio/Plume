<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { PreferencesSection, ProfileSection, SettingsLayout, icons, settingsRoute } from '@facile/muse';
	import { userStore } from '$lib/stores/user.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { theme } from '$lib/theme.svelte';
	import AdvancedPanel from '$lib/components/settings/AdvancedPanel.svelte';
	import IntegrationsPanel from '$lib/components/settings/IntegrationsPanel.svelte';
	import NotificationsPanel from '$lib/components/settings/NotificationsPanel.svelte';
	import TokensPanel from '$lib/components/settings/TokensPanel.svelte';

	const LEGACY: Record<string, string> = { appearance: 'preferences', api: 'tokens' };

	const SECTIONS = [
		{ id: 'profile', icon: icons.userCircle },
		{ id: 'preferences', icon: icons.palette },
		{ id: 'notifications', icon: icons.notification },
		{ id: 'integrations', icon: icons.plug },
		{ id: 'tokens', icon: icons.code },
		{ id: 'advanced', icon: icons.settings }
	] as const;

	const sections = $derived(
		SECTIONS.map((s) => ({ ...s, label: t(`settings.tab.${s.id}`), href: `/settings/${s.id}` }))
	);

	const requested = $derived(page.params.section);
	const user = $derived(userStore.value);

	$effect(() => {
		const legacy = requested ? LEGACY[requested] : undefined;
		if (legacy) void goto(`/settings/${legacy}`, { replaceState: true });
		else if (requested && !settingsRoute(requested, sections).known) void goto('/settings', { replaceState: true });
	});
</script>

<svelte:head><title>{t('profile.title')}</title></svelte:head>

<SettingsLayout
	class="mx-auto w-full max-w-fc-xl"
	{sections}
	active={requested}
	backHref="/settings"
	title={t('settings.heading')}
	description={t('settings.subtitle')}
	navLabel={t('settings.sections')}
>
	{#snippet children(id)}
		{#if id === 'profile'}
			<ProfileSection
				name={user?.name?.trim() || user?.email || t('profile.account')}
				email={user?.email}
				avatar={user?.avatar_url || undefined}
			/>
		{:else if id === 'preferences'}
			<PreferencesSection theme={theme.preference} onThemeChange={(next) => theme.set(next)} onLocaleChange={rememberLocale} />
		{:else if id === 'notifications'}
			<NotificationsPanel />
		{:else if id === 'integrations'}
			<IntegrationsPanel />
		{:else if id === 'tokens'}
			<TokensPanel />
		{:else}
			<AdvancedPanel />
		{/if}
	{/snippet}
</SettingsLayout>
