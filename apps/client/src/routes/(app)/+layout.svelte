<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, currentUser } from '$lib';
	import { MobileNav, PageTransition, SideBar, SpaceSwitcher, Topbar, icons } from '@facile/muse';
	import { userStore } from '$lib/stores/user.svelte';
	import { spaceStore } from '$lib/stores/space.svelte';

	let { children } = $props();

	let collapsed = $state(false);
	let scroller: HTMLElement | null = $state(null);

	onMount(async () => {
		/* The API is the only thing that knows whether this browser has a session:
		   a local login leaves a bearer token, an SSO login leaves a cookie the
		   client cannot read. A thrown error is a bad round-trip, not a logout —
		   bouncing to /login on one would sign people out on a hiccup. */
		let me;
		try {
			me = await currentUser();
		} catch {
			return;
		}
		if (!me) {
			goto('/login');
			return;
		}
		userStore.value = me;
		try {
			spaceStore.spaces = await api.spaces.list();
		} catch {}
		/*
		 * The sync endpoint is only mounted when OIDC is configured, so firing it blind put a
		 * 404 in the log of every local-auth install on every session.
		 */
		try {
			const config = await fetch('/api/auth/config').then((response) => response.json());
			if (config?.oidc_enabled) api.auth.syncProfile().catch(() => {});
		} catch {}
	});

	const navLinks = [
		{ href: '/dashboard', label: 'Dashboard', icon: icons.dashboard },
		{ href: '/documents', label: 'Documents', icon: 'solar:document-linear' },
		{ href: '/clients', label: 'Clients', icon: 'solar:users-group-two-rounded-linear' },
		{ href: '/spaces', label: 'Spaces', icon: icons.usersGroup }
	];

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
	}

	const pathname = $derived(page.url.pathname);
	const navPages = $derived(navLinks.map((l) => ({ ...l, active: isActive(l.href) })));
	const settingsActive = $derived(isActive('/settings'));

	const user = $derived(
		userStore.value
			? {
					name: userStore.value.name?.trim() || userStore.value.email,
					avatar: userStore.value.avatar_url || undefined
				}
			: undefined
	);

	const spaces = $derived(spaceStore.spaces.map((s) => ({ id: String(s.id), name: s.name })));
	const activeSpaceId = $derived(spaceStore.activeId === null ? null : String(spaceStore.activeId));

	function selectSpace(id: string | null) {
		spaceStore.activeId = id === null ? null : Number(id);
	}

	$effect(() => {
		void pathname;
		scroller?.scrollTo({ top: 0 });
	});
</script>

<div class="flex h-[100dvh] w-full overflow-hidden bg-fc-page">
	<div class="hidden h-full shrink-0 p-3 md:block">
		<SideBar
			icon="solar:pen-new-square-bold-duotone"
			title="Plume"
			bind:collapsed
			pages={navPages}
			{spaces}
			{activeSpaceId}
			onSpaceSelect={selectSpace}
			manageSpacesHref="/spaces"
			{user}
			userHref="/settings"
			userActive={settingsActive}
			class="h-full"
		/>
	</div>

	<main bind:this={scroller} class="min-w-0 flex-1 overflow-auto overscroll-contain pb-28 md:pb-0">
		<Topbar class="md:hidden">
			<span class="text-fc-md font-semibold text-fc-fg">Plume</span>
			{#if spaces.length > 0}
				<div class="max-w-56 min-w-0 flex-1">
					<SpaceSwitcher {spaces} activeId={activeSpaceId} onSelect={selectSpace} manageHref="/spaces" />
				</div>
			{/if}
		</Topbar>

		<div class="mx-auto w-full max-w-fc-xl px-4 py-6 md:px-8 md:py-8">
			<PageTransition key={pathname}>
				{@render children()}
			</PageTransition>
		</div>
	</main>

	<MobileNav items={navPages} {user} profileHref="/settings" profileActive={settingsActive} />
</div>
