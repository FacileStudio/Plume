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
		ConfirmModal,
		EmptyState,
		Field,
		Input,
		Select,
		SettingsSection,
		Spinner,
		icons,
		toast
	} from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let loading = $state(true);

	let addEmail = $state('');
	let addRole = $state('member');
	let adding = $state(false);

	let confirmRemove = $state(false);
	let pendingRemove = $state<SpaceMember | null>(null);

	const spaceId = $derived(Number(page.params.id));
	const canManage = $derived(
		!space?.registre_url && (space?.role === 'owner' || space?.role === 'admin')
	);

	const roleTone = { owner: 'owner', admin: 'admin', member: 'neutral' } as const;

	async function addMember(event: SubmitEvent) {
		event.preventDefault();
		if (!addEmail.trim()) return;
		adding = true;
		try {
			const member = await api.spaces.members.add(spaceId, {
				email: addEmail.trim(),
				role: addRole
			});
			members = [...members, member];
			addEmail = '';
			addRole = 'member';
			toast.success(t('members.added'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('members.addFailed'));
		}
		adding = false;
	}

	async function updateRole(member: SpaceMember, newRole: string) {
		try {
			const updated = await api.spaces.members.updateRole(spaceId, member.id, newRole);
			members = members.map((m) => (m.id === updated.id ? updated : m));
			toast.success(t('members.roleUpdated'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('members.roleFailed'));
		}
	}

	function askRemove(member: SpaceMember) {
		pendingRemove = member;
		confirmRemove = true;
	}

	async function runRemove() {
		const target = pendingRemove;
		if (!target) return;
		try {
			await api.spaces.members.remove(spaceId, target.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('members.removeFailed'));
			throw e;
		}
		members = members.filter((m) => m.id !== target.id);
		pendingRemove = null;
		toast.success(t('members.removed'));
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

<svelte:head><title>{t('members.title', { space: space?.name ?? t('space.titleFallback') })}</title></svelte:head>

{#if loading}
	<div class="flex min-h-[40dvh] items-center justify-center">
		<Spinner size="lg" />
	</div>
{:else if space}
	<div class="flex max-w-2xl flex-col gap-10">
		<div class="flex flex-col gap-2">
			<Button
				href="/spaces/{spaceId}"
				variant="ghost"
				size="sm"
				icon={icons.chevronLeft}
				class="w-fit -ml-3"
			>
				{space.name}
			</Button>
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('common.members')}</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				{t('members.count', { count: members.length })}
				{space.registre_url ? t('members.managedInRegistre') : t('members.ownersStay')}
			</p>
			{#if space.registre_url}
				<Button href={space.registre_url} variant="outline" icon={icons.space} class="w-fit">
					{t('common.manageInRegistre')}
				</Button>
			{/if}
		</div>

		{#if canManage}
			<SettingsSection title={t('members.addTitle')} description={t('members.addText')}>
				<form onsubmit={addMember} class="flex flex-col gap-3 sm:flex-row sm:items-end">
					<Field label={t('common.email')} class="min-w-0 flex-1">
						<Input
							bind:value={addEmail}
							type="email"
							placeholder={t('members.emailPlaceholder')}
							required
						/>
					</Field>
					<Field label={t('members.role')} class="sm:w-40">
						<Select bind:value={addRole}>
							<option value="member">{t('role.member')}</option>
							<option value="admin">{t('role.admin')}</option>
						</Select>
					</Field>
					<Button
						type="submit"
						icon={icons.plus}
						disabled={adding || !addEmail.trim()}
						size="lg"
						class="sm:shrink-0"
					>
						{adding ? t('common.adding') : t('common.add')}
					</Button>
				</form>
			</SettingsSection>
		{/if}

		<SettingsSection
			title={t('members.people')}
			description={canManage ? t('members.peopleManage') : t('members.peopleView')}
			bare={members.length === 0}
		>
			{#if members.length === 0}
				<EmptyState
					icon={icons.usersGroup}
					title={t('space.noMembers')}
					description={canManage ? t('members.emptyManage') : t('members.emptyView')}
				/>
			{:else}
				{#each members as member (member.id)}
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
							{#if canManage && member.role !== 'owner'}
								<Select
									value={member.role}
									aria-label={t('members.roleFor', { name: member.name || member.email })}
									class="h-9 w-36 text-fc-sm"
									onchange={(e) => updateRole(member, e.currentTarget.value)}
								>
									<option value="member">{t('role.member')}</option>
									<option value="admin">{t('role.admin')}</option>
								</Select>
								<Button
									variant="ghost-danger"
									icon={icons.remove}
									aria-label={t('members.removeAria', { name: member.name || member.email })}
									onclick={() => askRemove(member)}
								>
									{t('common.remove')}
								</Button>
							{:else}
								<Badge tone={roleTone[member.role]}>{t(`role.${member.role}`)}</Badge>
							{/if}
						</div>
					</div>
				{/each}
			{/if}
		</SettingsSection>
	</div>
{/if}

<ConfirmModal
	bind:open={confirmRemove}
	tone="danger"
	title={t('members.confirm.title', {
		name: pendingRemove?.name || pendingRemove?.email || t('members.thisMember')
	})}
	description={t('members.confirm.text')}
	confirmLabel={t('members.confirm.label')}
	cancelLabel={t('members.confirm.cancel')}
	onConfirm={runRemove}
	onCancel={() => (pendingRemove = null)}
/>
