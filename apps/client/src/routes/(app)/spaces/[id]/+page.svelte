<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space, SpaceMember } from '$lib';
	import {
		Badge,
		Button,
		EmptyState,
		SettingsRow,
		SettingsSection,
		Spinner,
		StatCard,
		icons
	} from '@facile/muse';

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let loading = $state(true);

	const spaceId = $derived(Number(page.params.id));
	const isAdminOrOwner = $derived(space?.role === 'owner' || space?.role === 'admin');
	const preview = $derived(members.slice(0, 5));

	const roleTone = { owner: 'owner', admin: 'admin', member: 'neutral' } as const;

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

{#snippet membersAction()}
	{#if members.length > 5}
		<Button variant="outline" size="sm" href="/spaces/{spaceId}/members">View all members</Button>
	{/if}
{/snippet}

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Spinner size="lg" />
	</div>
{:else if space}
	<div class="flex flex-col gap-10">
		<div class="flex flex-col gap-4">
			<Button href="/spaces" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
				Spaces
			</Button>
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex min-w-0 flex-col gap-2">
					<div class="flex min-w-0 flex-wrap items-center gap-3">
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{space.name}</h1>
						<Badge tone={roleTone[space.role]}>{space.role}</Badge>
					</div>
					<p class="text-fc-sm text-fc-fg-muted">
						{space.description || 'No description yet.'}
					</p>
				</div>
				<div class="flex shrink-0 items-center gap-2">
					<Button variant="outline" icon={icons.usersGroup} href="/spaces/{spaceId}/members">
						Members ({members.length})
					</Button>
					{#if isAdminOrOwner}
						<Button variant="outline" icon={icons.settings} href="/spaces/{spaceId}/settings">
							Settings
						</Button>
					{/if}
				</div>
			</div>
		</div>

		<section class="grid gap-4 sm:grid-cols-2">
			<StatCard label="Members" value={members.length} />
			<StatCard label="Created" value={formatDate(space.created_at)} />
		</section>

		<SettingsSection
			title="Members"
			description="The people who can see everything in this space."
			bare={members.length === 0}
			actions={membersAction}
		>
			{#if members.length === 0}
				<EmptyState
					icon={icons.usersGroup}
					title="No one here yet"
					description="Invite a teammate and they will see this space as soon as they accept."
				/>
			{:else}
				{#each preview as member (member.id)}
					<SettingsRow label={member.name || member.email} description={member.email}>
						<Badge tone={roleTone[member.role]}>{member.role}</Badge>
					</SettingsRow>
				{/each}
			{/if}
		</SettingsSection>
	</div>
{/if}
