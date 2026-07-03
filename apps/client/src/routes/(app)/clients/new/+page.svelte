<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	let name = $state('');
	let email = $state('');
	let company = $state('');
	let phone = $state('');
	let notes = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function submit() {
		error = '';
		if (!name.trim()) {
			error = 'Client name is required';
			return;
		}
		submitting = true;
		try {
			const c = await api.clients.create({
				name: name.trim(),
				email: email.trim(),
				company: company.trim(),
				phone: phone.trim(),
				notes: notes.trim()
			});
			goto(`/clients/${c.id}`);
		} catch (e: any) {
			error = e.message;
			toast.error(e instanceof Error ? e.message : 'Failed to create client');
			submitting = false;
		}
	}
</script>

<svelte:head><title>New Client — Plume</title></svelte:head>

<div class="max-w-lg">
	<div class="mb-8">
		<a href="/clients" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors mb-4">
			<Icon icon="solar:arrow-left-linear" class="h-4 w-4" />
			Back to clients
		</a>
		<h1 class="text-2xl font-bold">New client</h1>
	</div>

	<form onsubmit={submit} class="space-y-6">
		<div class="space-y-2">
			<Label for="client-name">Name</Label>
			<Input id="client-name" bind:value={name} placeholder="Jane Doe" required />
		</div>

		<div class="space-y-2">
			<Label for="client-email">Email</Label>
			<Input id="client-email" bind:value={email} placeholder="jane@example.com" type="email" />
		</div>

		<div class="space-y-2">
			<Label for="client-company">Company</Label>
			<Input id="client-company" bind:value={company} placeholder="Acme Inc." />
		</div>

		<div class="space-y-2">
			<Label for="client-phone">Phone</Label>
			<Input id="client-phone" bind:value={phone} placeholder="+1 555 000 0000" type="tel" />
		</div>

		<div class="space-y-2">
			<Label for="client-notes">Notes</Label>
			<textarea
				id="client-notes"
				bind:value={notes}
				placeholder="Anything worth remembering about this client..."
				rows="4"
				class="dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 min-h-16 w-full min-w-0 rounded-lg border bg-transparent px-2.5 py-1.5 text-base outline-none transition-colors placeholder:text-muted-foreground focus-visible:ring-3 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm"
			></textarea>
		</div>

		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}

		<Button type="submit" disabled={submitting} class="w-full">
			{#if submitting}
				<Icon icon="solar:spinner-linear" class="h-4 w-4 animate-spin" />
				Creating...
			{:else}
				<Icon icon="solar:user-plus-linear" class="h-4 w-4" />
				Create client
			{/if}
		</Button>
	</form>
</div>
