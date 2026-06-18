<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';
	import { spaceStore } from '$lib/stores/space.svelte';

	let name = $state('');
	let description = $state('');
	let saving = $state(false);

	async function create() {
		if (!name.trim()) return;
		saving = true;
		try {
			const space = await api.spaces.create({ name: name.trim(), description: description.trim() });
			spaceStore.spaces = [...spaceStore.spaces, space];
			spaceStore.activeId = space.id;
			toast.success('Space created');
			goto(`/spaces/${space.id}`);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to create space');
		}
		saving = false;
	}
</script>

<svelte:head><title>New Space — Plume</title></svelte:head>

<div class="max-w-lg">
	<h1 class="text-2xl font-bold mb-6">New space</h1>

	<div class="rounded-lg border p-6 space-y-4">
		<div>
			<label for="space-name" class="block text-sm font-medium mb-1">Name</label>
			<input
				id="space-name"
				type="text"
				bind:value={name}
				placeholder="My team"
				class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
			/>
		</div>
		<div>
			<label for="space-description" class="block text-sm font-medium mb-1">Description</label>
			<textarea
				id="space-description"
				bind:value={description}
				placeholder="What is this space for?"
				rows="3"
				class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
			></textarea>
		</div>
		<div class="flex gap-2 pt-1">
			<Button onclick={create} disabled={saving || !name.trim()}>
				{#if saving}
					<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
					Creating...
				{:else}
					<Icon icon="mdi:plus" class="h-4 w-4" />
					Create space
				{/if}
			</Button>
			<Button variant="outline" href="/spaces">Cancel</Button>
		</div>
	</div>
</div>
