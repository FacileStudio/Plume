<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space, SpaceMember } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let loading = $state(true);

	let addEmail = $state('');
	let addRole = $state('member');
	let adding = $state(false);

	let removeTarget = $state<SpaceMember | null>(null);
	let removing = $state(false);

	const spaceId = $derived(Number(page.params.id));
	const isAdminOrOwner = $derived(space?.role === 'owner' || space?.role === 'admin');

	function roleBadge(role: string): string {
		if (role === 'owner') return 'bg-amber-500/10 text-amber-600';
		if (role === 'admin') return 'bg-blue-500/10 text-blue-600';
		return 'bg-muted text-muted-foreground';
	}

	async function addMember() {
		if (!addEmail.trim()) return;
		adding = true;
		try {
			const member = await api.spaces.members.add(spaceId, { email: addEmail.trim(), role: addRole });
			members = [...members, member];
			addEmail = '';
			addRole = 'member';
			toast.success('Member added');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to add member');
		}
		adding = false;
	}

	async function updateRole(member: SpaceMember, newRole: string) {
		try {
			const updated = await api.spaces.members.updateRole(spaceId, member.id, newRole);
			members = members.map((m) => (m.id === updated.id ? updated : m));
			toast.success('Role updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update role');
		}
	}

	async function confirmRemove() {
		if (!removeTarget) return;
		removing = true;
		try {
			await api.spaces.members.remove(spaceId, removeTarget.id);
			members = members.filter((m) => m.id !== removeTarget!.id);
			removeTarget = null;
			toast.success('Member removed');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to remove member');
		}
		removing = false;
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
		<Icon icon="solar:spinner-bold-duotone" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if space}
	<div class="border-b px-4 py-4 md:px-8 md:py-5">
		<div class="flex items-center gap-3 mb-3">
			<a href="/spaces/{spaceId}" class="rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted">
				<Icon icon="solar:arrow-left-linear" class="h-5 w-5" />
			</a>
			<span class="text-sm text-muted-foreground">{space.name}</span>
		</div>
		<div class="flex items-center justify-between">
			<h1 class="text-lg font-semibold">Members</h1>
		</div>
	</div>

	<div class="p-4 md:p-8 space-y-6">
		{#if isAdminOrOwner}
			<form onsubmit={addMember} class="flex items-end gap-3">
				<div class="flex-1">
					<label for="member-email" class="mb-1.5 block text-sm font-medium">Email</label>
					<input
						id="member-email"
						type="email"
						bind:value={addEmail}
						placeholder="colleague@example.com"
						class="h-10 w-full rounded-lg border border-border bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div class="w-32">
					<label for="member-role" class="mb-1.5 block text-sm font-medium">Role</label>
					<select
						id="member-role"
						bind:value={addRole}
						class="h-10 w-full rounded-lg border border-border bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
					>
						<option value="member">Member</option>
						<option value="admin">Admin</option>
					</select>
				</div>
				<Button type="submit" disabled={adding || !addEmail.trim()} size="sm" class="h-10">
					{#if adding}
						<Icon icon="solar:spinner-bold-duotone" class="h-4 w-4 animate-spin" />
					{:else}
						<Icon icon="solar:add-circle-bold-duotone" class="h-4 w-4" />
						Add
					{/if}
				</Button>
			</form>
		{/if}

		<div>
			<h2 class="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-3">
				{members.length} member{members.length !== 1 ? 's' : ''}
			</h2>
			<div class="grid gap-3">
				{#each members as member}
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
						<div class="flex items-center gap-3 shrink-0">
							{#if isAdminOrOwner && member.role !== 'owner'}
								<select
									value={member.role}
									onchange={(e) => updateRole(member, e.currentTarget.value)}
									class="h-8 rounded-lg border border-border bg-background px-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
								>
									<option value="member">Member</option>
									<option value="admin">Admin</option>
								</select>
								<button
									onclick={() => (removeTarget = member)}
									class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
									aria-label="Remove member"
								>
									<Icon icon="solar:trash-bin-trash-bold-duotone" class="h-4 w-4" />
								</button>
							{:else}
								<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {roleBadge(member.role)}">
									{member.role}
								</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}

<AlertDialog.Root open={removeTarget !== null} onOpenChange={(open) => { if (!open) removeTarget = null; }}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Remove member</AlertDialog.Title>
			<AlertDialog.Description>
				Are you sure you want to remove <strong>{removeTarget?.name || removeTarget?.email}</strong> from this space?
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={removing}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				class="!bg-red-600 !text-white hover:!bg-red-700"
				onclick={confirmRemove}
				disabled={removing}
			>
				{#if removing}
					<Icon icon="solar:spinner-bold-duotone" class="h-4 w-4 animate-spin" />
					Removing...
				{:else}
					Remove
				{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
