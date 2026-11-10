<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, currentUser, logout } from '$lib';
	import { MobileNav, PageTransition, SideBar, SpaceSwitcher, Topbar, icons } from '@facile/muse';
	import { userStore } from '$lib/stores/user.svelte';
	import { rememberLocale, setSessionLocale, t } from '$lib/i18n/index.svelte';
	import { theme, type ThemePreference } from '$lib/theme.svelte';
	import { spaceStore } from '$lib/stores/space.svelte';

	let { children } = $props();

	let collapsed = $state(false);
	let invitations = $state<{ count: number; href: string } | null>(null);
	let scroller: HTMLElement | null = $state(null);

	onMount(async () => {
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
		setSessionLocale(me.locale);
		fetch('/api/me/invitations', { credentials: 'same-origin' })
			.then((r) => (r.ok ? r.json() : null))
			.then((b) => (invitations = b?.href ? { count: b.count, href: b.href } : null))
			.catch(() => (invitations = null));
		try {
			spaceStore.spaces = await api.spaces.list();
		} catch {}
		api.auth
			.syncProfile()
			.then(() => api.auth.me())
			.then((fresh) => {
				userStore.value = fresh;
				setSessionLocale(fresh.locale);
			})
			.catch(() => {});
	});

	const navLinks = $derived([
		{ href: '/dashboard', label: t('nav.dashboard'), icon: icons.dashboard },
		{ href: '/documents', label: t('nav.documents'), icon: 'solar:document-linear' },
		{ href: '/clients', label: t('nav.clients'), icon: 'solar:users-group-two-rounded-linear' },
		{ href: '/spaces', label: t('nav.spaces'), icon: icons.space }
	]);

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
	}

	const pathname = $derived(page.url.pathname);
	const navPages = $derived(navLinks.map((l) => ({ ...l, active: isActive(l.href) })));
	/**
	 * Ends the session on both transports, forgets the cached identity and returns to the login page.
	 */
	async function signOut() {
		await logout();
		userStore.value = null;
		setSessionLocale(null);
		void goto('/login');
	}

	const userMenu = $derived({
		subscription: { href: 'https://sso.facile.studio/spaces' },
		helpHref: `${page.url.origin}/docs`,
		onSignOut: signOut,
		theme: theme.preference,
		onThemeChange: (next: ThemePreference) => theme.set(next),
		onLocaleChange: rememberLocale,
		invitations
	});

	const settingsActive = $derived(isActive('/settings'));

	const user = $derived(
		userStore.value
			? {
					name: userStore.value.name?.trim() || userStore.value.email,
					email: userStore.value.email,
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
			{userMenu}
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

	<MobileNav items={navPages} {user} profileHref="/settings" profileActive={settingsActive} {userMenu} />
</div>
