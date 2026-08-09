<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import type { Space } from '$lib';
	import {
		Button,
		ConfirmModal,
		Field,
		Input,
		SettingsRow,
		SettingsSection,
		Spinner,
		Textarea,
		icons,
		toast
	} from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';

	let space = $state<Space | null>(null);
	let loading = $state(true);

	let name = $state('');
	let description = $state('');
	let saving = $state(false);

	let confirmDelete = $state(false);

	const spaceId = $derived(Number(page.params.id));

	async function save(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;
		saving = true;
		try {
			const updated = await api.spaces.update(spaceId, {
				name: name.trim(),
				description: description.trim()
			});
			space = updated;
			spaceStore.spaces = spaceStore.spaces.map((s) => (s.id === spaceId ? updated : s));
			toast.success('Space updated');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to update space');
		}
		saving = false;
	}

	async function runDelete() {
		try {
			await api.spaces.delete(spaceId);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to delete space');
			throw e;
		}
		spaceStore.spaces = spaceStore.spaces.filter((s) => s.id !== spaceId);
		if (spaceStore.activeId === spaceId) {
			spaceStore.activeId = null;
		}
		toast.success('Space deleted');
		goto('/spaces');
	}

	onMount(async () => {
		try {
			space = await api.spaces.get(spaceId);
			if (space.role !== 'owner' && space.role !== 'admin') {
				goto(`/spaces/${spaceId}`);
				return;
			}
			name = space.name;
			description = space.description;
		} catch {
			goto('/spaces');
			return;
		}
		loading = false;
	});
</script>

<svelte:head><title>Settings — {space?.name ?? 'Space'} — Plume</title></svelte:head>

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
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Space settings</h1>
		</div>

		<form onsubmit={save} class="flex flex-col gap-4">
			<SettingsSection title="General" description="How this space is named and described.">
				<Field label="Name">
					<Input bind:value={name} placeholder="My team" />
				</Field>
				<Field label="Description" helper="Optional — what the space is for.">
					<Textarea bind:value={description} rows={3} placeholder="Client contracts and NDAs." />
				</Field>
			</SettingsSection>
			<div class="flex justify-end">
				<Button type="submit" icon={icons.check} disabled={saving || !name.trim()}>
					{saving ? 'Saving…' : 'Save changes'}
				</Button>
			</div>
		</form>

		{#if space.role === 'owner'}
			<SettingsSection title="Danger zone" description="Irreversible, and nobody can undo it for you.">
				<SettingsRow
					label="Delete this space"
					description="Every member loses access. Documents in the space go with it."
				>
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						onclick={() => (confirmDelete = true)}
					>
						Delete space
					</Button>
				</SettingsRow>
			</SettingsSection>
		{/if}
	</div>

	<ConfirmModal
		bind:open={confirmDelete}
		tone="danger"
		title="Delete {space.name}?"
		description="This cannot be undone. All members lose access immediately."
		confirmLabel="Delete space"
		cancelLabel="Keep space"
		onConfirm={runDelete}
	/>
{/if}
