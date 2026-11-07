<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space, SpaceMember } from '$lib';
	import {
		Avatar,
		Badge,
		Button,
		EmptyState,
		SettingsSection,
		Spinner,
		StatCard,
		icons
	} from '@facile/muse';
	import { formatShortDate } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let loading = $state(true);

	const spaceId = $derived(Number(page.params.id));
	const isAdminOrOwner = $derived(space?.role === 'owner' || space?.role === 'admin');
	const preview = $derived(members.slice(0, 5));

	const roleTone = { owner: 'owner', admin: 'admin', member: 'neutral' } as const;

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

<svelte:head><title>{space?.name ?? t('space.titleFallback')} — Plume</title></svelte:head>

{#snippet membersAction()}
	{#if members.length > 5}
		<Button variant="secondary" size="sm" href="/spaces/{spaceId}/members">
			{t('space.viewAllMembers')}
		</Button>
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
				{t('spaces.heading')}
			</Button>
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex min-w-0 flex-col gap-2">
					<div class="flex min-w-0 flex-wrap items-center gap-3">
						<h1 class="truncate text-fc-2xl font-semibold text-fc-fg">{space.name}</h1>
						<Badge tone={roleTone[space.role]}>{t(`role.${space.role}`)}</Badge>
					</div>
					<p class="text-fc-sm text-fc-fg-muted">
						{space.description || t('space.noDescription')}
					</p>
				</div>
				<div class="flex shrink-0 items-center gap-2">
					<Button variant="secondary" icon={icons.usersGroup} href="/spaces/{spaceId}/members">
						{t('space.membersCount', { count: members.length })}
					</Button>
					{#if space.registre_url}
						<Button variant="secondary" icon={icons.space} href={space.registre_url}>
							{t('common.manageInRegistre')}
						</Button>
					{:else if isAdminOrOwner}
						<Button variant="secondary" icon={icons.settings} href="/spaces/{spaceId}/settings">
							{t('space.settings')}
						</Button>
					{/if}
				</div>
			</div>
		</div>

		<section class="grid gap-4 sm:grid-cols-2">
			<StatCard label={t('common.members')} value={members.length} />
			<StatCard label={t('common.created')} value={formatShortDate(space.created_at)} />
		</section>

		<SettingsSection
			title={t('common.members')}
			description={t('space.membersText')}
			bare={members.length === 0}
			actions={membersAction}
		>
			{#if members.length === 0}
				<EmptyState
					icon={icons.usersGroup}
					title={t('space.noMembers')}
					description={t('space.noMembersText')}
				/>
			{:else}
				{#each preview as member (member.id)}
					<div
						class="flex flex-col gap-3 py-4 first:pt-0 last:pb-0 sm:flex-row sm:items-center sm:justify-between sm:gap-6"
					>
						<div class="flex min-w-0 items-center gap-3">
							<Avatar src={member.avatar_url || undefined} name={member.name || member.email} size="sm" />
							<div class="flex min-w-0 flex-col gap-1">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{member.name || member.email}</span>
								<p class="truncate text-fc-xs text-fc-fg-muted">{member.email}</p>
							</div>
						</div>
						<div class="flex items-center gap-2 sm:shrink-0 sm:justify-end">
							<Badge tone={roleTone[member.role]}>{t(`role.${member.role}`)}</Badge>
						</div>
					</div>
				{/each}
			{/if}
		</SettingsSection>
	</div>
{/if}
