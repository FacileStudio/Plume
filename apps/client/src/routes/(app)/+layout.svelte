<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, isAuthenticated, clearToken } from '$lib';
	import Icon from '@iconify/svelte';
	import { userStore } from '$lib/stores/user.svelte';
	import { spaceStore } from '$lib/stores/space.svelte';
	import SpaceSwitcher from '$lib/components/space-switcher.svelte';
	import MobileNav from '$lib/components/MobileNav.svelte';

	let { children } = $props();

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((w) => w[0])
			.filter(Boolean)
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}

	onMount(async () => {
		if (!isAuthenticated()) {
			goto('/login');
			return;
		}
		try {
			userStore.value = await api.auth.me();
		} catch {}
		try {
			spaceStore.spaces = await api.spaces.list();
		} catch {}
		api.auth.syncProfile().catch(() => {});
	});

	function logout() {
		clearToken();
		goto('/login');
	}

	const navLinks = [
		{ href: '/documents', label: 'Documents', icon: 'solar:document-linear' },
		{ href: '/spaces', label: 'Spaces', icon: 'solar:users-group-rounded-linear' },
		{ href: '/settings', label: 'Settings', icon: 'solar:settings-linear' }
	];
</script>

<div class="flex h-[100dvh] w-full overflow-hidden">
	<aside class="sticky top-0 hidden h-[100dvh] w-60 flex-col border-r bg-background md:flex">
		<div class="flex items-center gap-3 px-5 pt-8 pb-4">
			<Icon icon="solar:pen-new-square-bold-duotone" class="h-7 w-7" />
			<span class="text-2xl font-bold tracking-tight">Plume</span>
		</div>

		<SpaceSwitcher />

		<nav class="flex flex-1 flex-col gap-1 px-3">
			{#each navLinks as link}
				{@const active = page.url.pathname === link.href || page.url.pathname.startsWith(link.href + '/')}
				<a
					href={link.href}
					class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors {active
						? 'bg-foreground text-background font-medium'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				>
					<Icon icon={link.icon} class="h-4 w-4 shrink-0" />
					{link.label}
				</a>
			{/each}
		</nav>

		<div class="h-px bg-border"></div>

		<div class="flex flex-col gap-2 p-4">
			<a
				href="/profile"
				class="flex items-center gap-3 rounded-xl border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted"
			>
				{#if userStore.value?.avatar_url}
					<img src="/api{userStore.value.avatar_url}" alt="Avatar" class="h-9 w-9 shrink-0 rounded-full border border-border object-cover" />
				{:else}
					<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background">
						{userStore.value ? getInitials(userStore.value.name || userStore.value.email) : '..'}
					</div>
				{/if}
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm font-medium">{userStore.value?.name || 'Set your profile'}</p>
					<p class="truncate text-xs text-muted-foreground">{userStore.value?.email ?? ''}</p>
				</div>
			</a>
			<button
				onclick={logout}
				class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-destructive hover:bg-destructive/10"
			>
				<Icon icon="solar:logout-2-linear" class="h-4 w-4" />
				Logout
			</button>
		</div>
	</aside>

	<main class="flex-1 overflow-auto pb-24 md:pb-0">
		<div class="mx-auto w-full max-w-6xl px-4 py-6 md:px-8 md:py-8">
			{@render children()}
		</div>
	</main>
	<MobileNav user={userStore.value} />
</div>
