<script lang="ts">
	import { page } from '$app/state';
	import { Divider, PageTransition, Tabs, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let { children } = $props();

	const sections = $derived([
		{ id: 'profile', label: t('settings.tab.profile'), icon: icons.userCircle, href: '/settings/profile' },
		{
			id: 'appearance',
			label: t('settings.tab.appearance'),
			icon: icons.palette,
			href: '/settings/appearance'
		},
		{
			id: 'notifications',
			label: t('settings.tab.notifications'),
			icon: icons.notification,
			href: '/settings/notifications'
		},
		{ id: 'api', label: t('settings.tab.api'), icon: icons.code, href: '/settings/api' },
		{
			id: 'integrations',
			label: t('settings.tab.integrations'),
			icon: icons.plug,
			href: '/settings/integrations'
		},
		{ id: 'advanced', label: t('settings.tab.advanced'), icon: icons.settings, href: '/settings/advanced' }
	]);

	const active = $derived(
		sections.find((section) => page.url.pathname.startsWith(section.href))?.id ?? 'profile'
	);
</script>

<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-10">
	<div class="flex flex-col gap-1">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('settings.heading')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('settings.subtitle')}</p>
	</div>

	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} label={t('settings.sections')} />
		<Divider />
	</div>

	<PageTransition key={active} distance={8} duration={0.25}>
		{@render children()}
	</PageTransition>
</div>
