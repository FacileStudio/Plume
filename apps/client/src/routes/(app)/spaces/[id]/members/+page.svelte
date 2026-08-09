<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space, SpaceMember } from '$lib';
	import {
		Badge,
		Button,
		ConfirmModal,
		EmptyState,
		Field,
		Input,
		Select,
		SettingsRow,
		SettingsSection,
		Spinner,
		icons,
		toast
	} from '@facile/muse';

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let loading = $state(true);

	let addEmail = $state('');
	let addRole = $state('member');
	let adding = $state(false);

	let confirmRemove = $state(false);
	let pendingRemove = $state<SpaceMember | null>(null);

	const spaceId = $derived(Number(page.params.id));
	const isAdminOrOwner = $derived(space?.role === 'owner' || space?.role === 'admin');

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
			toast.success('Member added');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to add member');
		}
		adding = false;
	}

	async function updateRole(member: SpaceMember, newRole: string) {
		try {
			const updated = await api.spaces.members.updateRole(spaceId, member.id, newRole);
			members = members.map((m) => (m.id === updated.id ? updated : m));
			toast.success('Role updated');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to update role');
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
			toast.danger(e instanceof Error ? e.message : 'Failed to remove member');
			throw e;
		}
		members = members.filter((m) => m.id !== target.id);
		pendingRemove = null;
		toast.success('Member removed');
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

<svelte:head><title>Members — {space?.name ?? 'Space'} — Plume</title></svelte:head>

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
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Members</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				{members.length} member{members.length === 1 ? '' : 's'} in this space. Owners cannot be
				removed.
			</p>
		</div>

		{#if isAdminOrOwner}
			<SettingsSection
				title="Add a member"
				description="They get access to every document in this space straight away."
			>
				<form onsubmit={addMember} class="flex flex-col gap-3 sm:flex-row sm:items-end">
					<Field label="Email" class="min-w-0 flex-1">
						<Input
							bind:value={addEmail}
							type="email"
							placeholder="colleague@example.com"
							required
						/>
					</Field>
					<Field label="Role" class="sm:w-40">
						<Select bind:value={addRole}>
							<option value="member">Member</option>
							<option value="admin">Admin</option>
						</Select>
					</Field>
					<Button
						type="submit"
						icon={icons.plus}
						disabled={adding || !addEmail.trim()}
						size="lg"
						class="sm:shrink-0"
					>
						{adding ? 'Adding…' : 'Add'}
					</Button>
				</form>
			</SettingsSection>
		{/if}

		<SettingsSection
			title="People"
			description="Change a role or revoke access."
			bare={members.length === 0}
		>
			{#if members.length === 0}
				<EmptyState
					icon={icons.usersGroup}
					title="No one here yet"
					description="Add a teammate above and they will show up in this list."
				/>
			{:else}
				{#each members as member (member.id)}
					<SettingsRow label={member.name || member.email} description={member.email}>
						{#if isAdminOrOwner && member.role !== 'owner'}
							<Select
								value={member.role}
								aria-label="Role for {member.name || member.email}"
								class="h-9 w-36 text-fc-sm"
								onchange={(e) => updateRole(member, e.currentTarget.value)}
							>
								<option value="member">Member</option>
								<option value="admin">Admin</option>
							</Select>
							<Button
								variant="ghost-danger"
								icon={icons.remove}
								aria-label="Remove {member.name || member.email}"
								onclick={() => askRemove(member)}
							>
								Remove
							</Button>
						{:else}
							<Badge tone={roleTone[member.role]}>{member.role}</Badge>
						{/if}
					</SettingsRow>
				{/each}
			{/if}
		</SettingsSection>
	</div>
{/if}

<ConfirmModal
	bind:open={confirmRemove}
	tone="danger"
	title="Remove {pendingRemove?.name || pendingRemove?.email || 'this member'}?"
	description="They lose access to every document in this space immediately."
	confirmLabel="Remove member"
	cancelLabel="Keep access"
	onConfirm={runRemove}
	onCancel={() => (pendingRemove = null)}
/>
