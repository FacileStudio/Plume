<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space, SpaceMember } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let loading = $state(true);

	const spaceId = $derived(Number(page.params.id));
	const isAdminOrOwner = $derived(space?.role === 'owner' || space?.role === 'admin');

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	onMount(async () => {
		try {
			const [s, m] = await Promise.all([
				api.spaces.get(spaceId),
				api.spaces.members.list(spaceId)
			]);
			space = s;
			members = m;
		} catch {
			goto('/spaces');
			return;
		}
		loading = false;
	});
</script>

<svelte:head><title>{space?.name ?? 'Space'} — Plume</title></svelte:head>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if space}
	<div class="mb-8">
		<div class="flex items-center gap-2 text-sm text-muted-foreground mb-2">
			<a href="/spaces" class="hover:text-foreground">Espaces</a>
			<Icon icon="solar:alt-arrow-right-linear" class="h-3.5 w-3.5" />
			<span>{space.name}</span>
		</div>
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold">{space.name}</h1>
				{#if space.description}
					<p class="text-muted-foreground mt-1">{space.description}</p>
				{/if}
			</div>
			{#if isAdminOrOwner}
				<div class="flex items-center gap-2">
					<Button variant="outline" href="/spaces/{spaceId}/members">
						<Icon icon="solar:users-group-rounded-linear" class="h-4 w-4" />
						Members ({members.length})
					</Button>
					<Button variant="outline" href="/spaces/{spaceId}/settings">
						<Icon icon="solar:settings-linear" class="h-4 w-4" />
						Settings
					</Button>
				</div>
			{:else}
				<Button variant="outline" href="/spaces/{spaceId}/members">
					<Icon icon="solar:users-group-rounded-linear" class="h-4 w-4" />
					Members ({members.length})
				</Button>
			{/if}
		</div>
	</div>

	<div class="rounded-lg border p-6">
		<h2 class="text-lg font-semibold mb-4">Members</h2>
		<div class="space-y-3">
			{#each members.slice(0, 5) as member}
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3 min-w-0">
						<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-muted text-xs font-semibold">
							{(member.name || member.email).charAt(0).toUpperCase()}
						</div>
						<div class="min-w-0">
							<p class="text-sm font-medium truncate">{member.name || member.email}</p>
							<p class="text-xs text-muted-foreground truncate">{member.email}</p>
						</div>
					</div>
					<span class="rounded-full px-2.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
						{member.role}
					</span>
				</div>
			{/each}
		</div>
		{#if members.length > 5}
			<div class="mt-4 text-center">
				<Button href="/spaces/{spaceId}/members" variant="outline" size="sm">View all members</Button>
			</div>
		{/if}
	</div>

	<div class="mt-6 rounded-lg border p-6">
		<h2 class="text-lg font-semibold mb-2">Details</h2>
		<div class="space-y-2 text-sm">
			<div class="flex justify-between">
				<span class="text-muted-foreground">Your role</span>
				<span class="font-medium">{space.role}</span>
			</div>
			<div class="flex justify-between">
				<span class="text-muted-foreground">Created</span>
				<span>{formatDate(space.created_at)}</span>
			</div>
			<div class="flex justify-between">
				<span class="text-muted-foreground">Members</span>
				<span>{members.length}</span>
			</div>
		</div>
	</div>
{/if}
