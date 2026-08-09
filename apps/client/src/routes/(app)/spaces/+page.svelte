<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { Space } from '$lib';
	import {
		Badge,
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Spinner,
		Table,
		icons,
		toast
	} from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';

	let spaces = $state<Space[]>([]);
	let loading = $state(true);
	let confirmLeave = $state(false);
	let pendingLeave = $state<Space | null>(null);

	const roleTone = { owner: 'owner', admin: 'admin', member: 'neutral' } as const;

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function askLeave(space: Space) {
		pendingLeave = space;
		confirmLeave = true;
	}

	async function runLeave() {
		const target = pendingLeave;
		if (!target) return;
		try {
			await api.spaces.leave(target.id);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to leave space');
			throw e;
		}
		spaces = spaces.filter((s) => s.id !== target.id);
		spaceStore.spaces = spaces;
		if (spaceStore.activeId === target.id) {
			spaceStore.activeId = null;
		}
		pendingLeave = null;
		toast.success('You left the space');
	}

	onMount(async () => {
		try {
			spaces = await api.spaces.list();
			spaceStore.spaces = spaces;
		} catch {}
		loading = false;
	});
</script>

<svelte:head><title>Spaces — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex min-w-0 flex-col gap-2">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Spaces</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				Shared workspaces. Everyone in a space sees its documents and clients.
			</p>
		</div>
		<Button href="/spaces/new" icon={icons.plus}>New space</Button>
	</div>

	{#if loading}
		<div class="flex min-h-[40dvh] items-center justify-center">
			<Spinner size="lg" />
		</div>
	{:else if spaces.length === 0}
		<EmptyState
			icon={icons.usersGroup}
			title="No spaces yet"
			description="Create one to collaborate with your team on the same documents."
		>
			<Button href="/spaces/new" variant="outline" icon={icons.plus}>New space</Button>
		</EmptyState>
	{:else}
		<div class="hidden md:block">
			<Table>
				<thead>
					<tr>
						<th scope="col">Space</th>
						<th scope="col">Your role</th>
						<th scope="col">Created</th>
						<th aria-label="Actions"></th>
					</tr>
				</thead>
				<tbody>
					{#each spaces as space (space.id)}
						<tr>
							<td>
								<a
									href="/spaces/{space.id}"
									class="flex min-w-0 flex-col rounded-fc-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
								>
									<span class="truncate font-medium text-fc-fg">{space.name}</span>
									{#if space.description}
										<span class="truncate text-fc-xs text-fc-fg-muted">{space.description}</span>
									{/if}
								</a>
							</td>
							<td><Badge tone={roleTone[space.role]}>{space.role}</Badge></td>
							<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(space.created_at)}</td>
							<td class="text-right">
								{#if space.role !== 'owner'}
									<Button
										variant="ghost-danger"
										size="sm"
										icon={icons.logout}
										aria-label="Leave {space.name}"
										onclick={() => askLeave(space)}
									>
										Leave
									</Button>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		</div>

		<div class="flex flex-col gap-2 md:hidden">
			{#each spaces as space (space.id)}
				<Card class="flex flex-col gap-3">
					<div class="flex min-w-0 items-start justify-between gap-3">
						<a
							href="/spaces/{space.id}"
							class="flex min-w-0 flex-col rounded-fc-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						>
							<span class="truncate text-fc-sm font-medium text-fc-fg">{space.name}</span>
							<span class="text-fc-xs text-fc-fg-muted">{formatDate(space.created_at)}</span>
						</a>
						<Badge tone={roleTone[space.role]}>{space.role}</Badge>
					</div>
					{#if space.role !== 'owner'}
						<div class="flex items-center justify-end">
							<Button
								variant="ghost-danger"
								size="lg"
								icon={icons.logout}
								aria-label="Leave {space.name}"
								onclick={() => askLeave(space)}
							>
								Leave
							</Button>
						</div>
					{/if}
				</Card>
			{/each}
		</div>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmLeave}
	tone="danger"
	title="Leave {pendingLeave?.name ?? 'this space'}?"
	description="You lose access to every document in this space. An owner can invite you back."
	confirmLabel="Leave space"
	cancelLabel="Stay"
	onConfirm={runLeave}
	onCancel={() => (pendingLeave = null)}
/>
