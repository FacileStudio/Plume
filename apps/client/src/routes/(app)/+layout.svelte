<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, isAuthenticated, clearToken } from '$lib';
	import Icon from '@iconify/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { userStore } from '$lib/stores/user.svelte';
	import { spaceStore } from '$lib/stores/space.svelte';
	import SpaceSwitcher from '$lib/components/space-switcher.svelte';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import FileText from '@lucide/svelte/icons/file-text';
	import Settings from '@lucide/svelte/icons/settings';
	import LogOut from '@lucide/svelte/icons/log-out';
	import ShieldCheck from '@lucide/svelte/icons/shield-check';

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
		{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/documents', label: 'Documents', icon: FileText },
		{ href: '/verify', label: 'Verify', icon: ShieldCheck },
		{ href: '/settings', label: 'Settings', icon: Settings }
	];
</script>

<div class="flex h-screen w-full overflow-hidden">
	<aside class="sticky top-0 flex h-screen w-60 flex-col border-r bg-background">
		<div class="flex items-center gap-3 px-5 pt-8 pb-4">
			<Icon icon="solar:document-add-bold-duotone" class="w-7 h-7" />
			<span class="text-2xl font-bold tracking-tight">Plume</span>
		</div>

		<div class="px-3 pb-4">
			<SpaceSwitcher />
		</div>

		<nav class="flex flex-1 flex-col gap-1 px-3">
			{#each navLinks as link}
				{@const active = page.url.pathname === link.href || page.url.pathname.startsWith(link.href + '/')}
				<a
					href={link.href}
					class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors {active
						? 'bg-foreground text-background font-medium'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				>
					<link.icon class="h-4 w-4 shrink-0" />
					{link.label}
				</a>
			{/each}
		</nav>

		<Separator />

		<div class="flex flex-col gap-2 p-4">
			<a
				href="/profile"
				class="flex items-center gap-3 rounded-lg border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted"
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
			<Button
				variant="ghost"
				size="sm"
				class="w-full justify-start gap-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
				onclick={logout}
			>
				<LogOut class="h-4 w-4" />
				Logout
			</Button>
		</div>
	</aside>

	<main class="flex-1 overflow-auto p-8">
		{@render children()}
	</main>
</div>
