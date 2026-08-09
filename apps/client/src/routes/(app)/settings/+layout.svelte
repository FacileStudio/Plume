<script lang="ts">
	import { page } from '$app/state';
	import { Divider, PageTransition, Tabs, icons } from '@facile/muse';

	let { children } = $props();

	const sections = [
		{ id: 'profile', label: 'Profile', icon: icons.userCircle, href: '/settings/profile' },
		{ id: 'appearance', label: 'Appearance', icon: icons.palette, href: '/settings/appearance' },
		{
			id: 'notifications',
			label: 'Notifications',
			icon: icons.notification,
			href: '/settings/notifications'
		},
		{ id: 'api', label: 'API', icon: icons.code, href: '/settings/api' },
		{ id: 'integrations', label: 'Integrations', icon: icons.plug, href: '/settings/integrations' },
		{ id: 'advanced', label: 'Advanced', icon: icons.settings, href: '/settings/advanced' }
	];

	const active = $derived(
		sections.find((section) => page.url.pathname.startsWith(section.href))?.id ?? 'profile'
	);
</script>

<div class="mx-auto flex w-full max-w-fc-lg flex-col gap-10">
	<div class="flex flex-col gap-1">
		<h1 class="text-fc-2xl font-semibold text-fc-fg">Settings</h1>
		<p class="text-fc-sm text-fc-fg-muted">
			Your account, this browser, and how Plume reaches the outside world.
		</p>
	</div>

	<div class="flex flex-col gap-4">
		<Tabs items={sections} value={active} label="Settings sections" />
		<Divider class="my-0" />
	</div>

	<PageTransition key={active} distance={8} duration={0.25}>
		{@render children()}
	</PageTransition>
</div>
