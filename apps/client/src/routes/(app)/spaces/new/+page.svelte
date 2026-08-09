<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Button, Card, Field, Input, Textarea, icons, toast } from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';

	let name = $state('');
	let description = $state('');
	let saving = $state(false);

	async function create(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) return;
		saving = true;
		try {
			const space = await api.spaces.create({
				name: name.trim(),
				description: description.trim()
			});
			spaceStore.spaces = [...spaceStore.spaces, space];
			spaceStore.activeId = space.id;
			toast.success('Space created');
			goto(`/spaces/${space.id}`);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to create space');
		}
		saving = false;
	}
</script>

<svelte:head><title>New Space — Plume</title></svelte:head>

<div class="flex max-w-xl flex-col gap-10">
	<div class="flex flex-col gap-2">
		<Button href="/spaces" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
			Back to spaces
		</Button>
		<h1 class="text-fc-2xl font-semibold text-fc-fg">New space</h1>
		<p class="text-fc-sm text-fc-fg-muted">
			A space is a shared workspace. You start as its owner and can invite people afterwards.
		</p>
	</div>

	<form onsubmit={create} class="flex flex-col gap-4">
		<Card class="flex flex-col gap-4">
			<Field label="Name">
				<Input bind:value={name} placeholder="My team" required />
			</Field>
			<Field label="Description" helper="Optional — what the space is for.">
				<Textarea bind:value={description} rows={3} placeholder="Client contracts and NDAs." />
			</Field>
		</Card>

		<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
			<Button href="/spaces" variant="ghost" class="w-full sm:w-auto">Cancel</Button>
			<Button
				type="submit"
				icon={icons.plus}
				disabled={saving || !name.trim()}
				class="w-full sm:w-auto"
			>
				{saving ? 'Creating…' : 'Create space'}
			</Button>
		</div>
	</form>
</div>
