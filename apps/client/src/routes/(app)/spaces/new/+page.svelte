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

<div class="mb-6 border-b pb-5">
	<div class="flex items-center gap-3">
		<a href="/spaces" class="rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted">
			<Icon icon="solar:arrow-left-linear" class="h-5 w-5" />
		</a>
		<h1 class="text-lg font-semibold">New space</h1>
	</div>
</div>

<div>
	<form onsubmit={create} class="max-w-xl space-y-6">
		<div>
			<label for="space-name" class="mb-1.5 block text-sm font-medium">Name</label>
			<input
				id="space-name"
				type="text"
				bind:value={name}
				placeholder="My team"
				class="h-10 w-full rounded-lg border border-border bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
			/>
		</div>
		<div>
			<label for="space-description" class="mb-1.5 block text-sm font-medium">Description</label>
			<textarea
				id="space-description"
				bind:value={description}
				placeholder="What is this space for?"
				rows="3"
				class="w-full rounded-lg border border-border bg-background px-3 py-2.5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
			></textarea>
		</div>
		<div class="flex items-center gap-3">
			<Button type="submit" disabled={saving || !name.trim()}>
				{#if saving}
					<Icon icon="solar:spinner-bold-duotone" class="h-4 w-4 animate-spin" />
					Creating...
				{:else}
					Create space
				{/if}
			</Button>
			<Button variant="outline" href="/spaces">Cancel</Button>
		</div>
	</form>
</div>
