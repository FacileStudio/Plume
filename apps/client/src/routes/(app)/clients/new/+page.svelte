<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Alert, Button, Card, Field, Input, Textarea, icons, toast } from '@facile/muse';

	let name = $state('');
	let email = $state('');
	let company = $state('');
	let phone = $state('');
	let notes = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
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
			toast.success('Client created');
			goto(`/clients/${c.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create client';
			toast.danger(error);
			submitting = false;
		}
	}
</script>

<svelte:head><title>New Client — Plume</title></svelte:head>

<div class="flex max-w-xl flex-col gap-10">
	<div class="flex flex-col gap-2">
		<Button href="/clients" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
			Back to clients
		</Button>
		<h1 class="text-fc-2xl font-semibold text-fc-fg">New client</h1>
		<p class="text-fc-sm text-fc-fg-muted">
			Only the name is required — everything else can be filled in later.
		</p>
	</div>

	<form onsubmit={submit} class="flex flex-col gap-4">
		<Card class="flex flex-col gap-4">
			<Field label="Name">
				<Input bind:value={name} placeholder="Jane Doe" required />
			</Field>
			<Field label="Email">
				<Input bind:value={email} type="email" placeholder="jane@example.com" />
			</Field>
			<Field label="Company">
				<Input bind:value={company} placeholder="Acme Inc." />
			</Field>
			<Field label="Phone">
				<Input bind:value={phone} type="tel" placeholder="+1 555 000 0000" />
			</Field>
			<Field label="Notes" helper="Optional — anything worth remembering about this client.">
				<Textarea bind:value={notes} rows={4} placeholder="Prefers PDFs, invoices monthly…" />
			</Field>
		</Card>

		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}

		<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
			<Button href="/clients" variant="ghost" class="w-full sm:w-auto">Cancel</Button>
			<Button type="submit" icon={icons.plus} disabled={submitting} class="w-full sm:w-auto">
				{submitting ? 'Creating…' : 'Create client'}
			</Button>
		</div>
	</form>
</div>
