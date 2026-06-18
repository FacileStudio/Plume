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

	let showAddForm = $state(false);
	let addEmail = $state('');
	let addRole = $state('member');
	let adding = $state(false);

	let removeTarget = $state<SpaceMember | null>(null);
	let removing = $state(false);

	const spaceId = $derived(Number(page.params.id));
	const isAdminOrOwner = $derived(space?.role === 'owner' || space?.role === 'admin');

	async function addMember() {
		if (!addEmail.trim()) return;
		adding = true;
		try {
			const member = await api.spaces.members.add(spaceId, { email: addEmail.trim(), role: addRole });
			members = [...members, member];
			addEmail = '';
			addRole = 'member';
			showAddForm = false;
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
		<Icon icon="solar:spinner-linear" class="h-8 w-8 animate-spin text-muted-foreground" />
	</div>
{:else if space}
	<div class="mb-8">
		<div class="flex items-center gap-2 text-sm text-muted-foreground mb-2">
			<a href="/spaces" class="hover:text-foreground">Espaces</a>
			<Icon icon="solar:alt-arrow-right-linear" class="h-3.5 w-3.5" />
			<a href="/spaces/{spaceId}" class="hover:text-foreground">{space.name}</a>
			<Icon icon="solar:alt-arrow-right-linear" class="h-3.5 w-3.5" />
			<span>Members</span>
		</div>
		<div class="flex items-center justify-between">
			<h1 class="text-2xl font-bold">Members</h1>
			{#if isAdminOrOwner && !showAddForm}
				<Button onclick={() => (showAddForm = true)}>
					<Icon icon="mdi:plus" class="h-4 w-4" />
					Add member
				</Button>
			{/if}
		</div>
	</div>

	{#if showAddForm && isAdminOrOwner}
		<div class="rounded-lg border p-6 mb-6 space-y-4 max-w-lg">
			<h2 class="text-lg font-semibold">Add member</h2>
			<div>
				<label for="member-email" class="block text-sm font-medium mb-1">Email</label>
				<input
					id="member-email"
					type="email"
					bind:value={addEmail}
					placeholder="colleague@example.com"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="member-role" class="block text-sm font-medium mb-1">Role</label>
				<select
					id="member-role"
					bind:value={addRole}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				>
					<option value="member">Member</option>
					<option value="admin">Admin</option>
				</select>
			</div>
			<div class="flex gap-2 pt-1">
				<Button onclick={addMember} disabled={adding || !addEmail.trim()}>
					{#if adding}
						<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
						Adding...
					{:else}
						<Icon icon="mdi:plus" class="h-4 w-4" />
						Add
					{/if}
				</Button>
				<Button variant="outline" onclick={() => { showAddForm = false; addEmail = ''; addRole = 'member'; }}>
					Cancel
				</Button>
			</div>
		</div>
	{/if}

	<div class="space-y-2">
		{#each members as member}
			<div class="flex items-center justify-between rounded-lg border p-4">
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
							class="rounded-md border bg-background px-2 py-1 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
						>
							<option value="member">Member</option>
							<option value="admin">Admin</option>
						</select>
						<button
							onclick={() => (removeTarget = member)}
							class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
							aria-label="Remove member"
						>
							<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
						</button>
					{:else}
						<span class="rounded-full px-2.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
							{member.role}
						</span>
					{/if}
				</div>
			</div>
		{/each}
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
					<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
					Removing...
				{:else}
					Remove
				{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
