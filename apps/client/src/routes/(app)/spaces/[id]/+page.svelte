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

	function roleBadge(role: string): string {
		if (role === 'owner') return 'bg-amber-500/10 text-amber-600';
		if (role === 'admin') return 'bg-blue-500/10 text-blue-600';
		return 'bg-muted text-muted-foreground';
	}

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
		<Icon icon="solar:spinner-bold-duotone" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if space}
	<div class="mb-6 border-b pb-5">
		<div class="flex items-center gap-3 mb-3">
			<a href="/spaces" class="rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted">
				<Icon icon="solar:arrow-left-linear" class="h-5 w-5" />
			</a>
			<span class="text-sm text-muted-foreground">{space.name}</span>
		</div>
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-lg font-semibold">{space.name}</h1>
				{#if space.description}
					<p class="mt-1 text-sm text-muted-foreground">{space.description}</p>
				{/if}
			</div>
			<div class="flex items-center gap-2">
				<Button variant="outline" size="sm" href="/spaces/{spaceId}/members">
					<Icon icon="solar:users-group-rounded-bold-duotone" class="h-4 w-4" />
					Members ({members.length})
				</Button>
				{#if isAdminOrOwner}
					<Button variant="outline" size="sm" href="/spaces/{spaceId}/settings">
						<Icon icon="solar:settings-bold-duotone" class="h-4 w-4" />
						Settings
					</Button>
				{/if}
			</div>
		</div>
	</div>

	<div class="space-y-6">
		<div>
			<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-3">Members</h2>
			<div class="grid gap-3">
				{#each members.slice(0, 5) as member}
					<div class="flex items-center justify-between rounded-lg border border-border p-4 transition-colors hover:bg-muted/50">
						<div class="flex items-center gap-3 min-w-0">
							<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border bg-muted text-xs font-semibold">
								{(member.name || member.email).charAt(0).toUpperCase()}
							</div>
							<div class="min-w-0">
								<p class="text-sm font-medium truncate">{member.name || member.email}</p>
								<p class="text-xs text-muted-foreground truncate">{member.email}</p>
							</div>
						</div>
						<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {roleBadge(member.role)}">
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

		<div class="rounded-lg border border-border bg-muted/20 p-4">
			<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-3">Details</h2>
			<div class="space-y-2.5 text-sm">
				<div class="flex justify-between">
					<span class="text-muted-foreground">Your role</span>
					<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {roleBadge(space.role)}">{space.role}</span>
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
	</div>
{/if}
