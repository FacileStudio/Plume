<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Alert, Button, Card, Field, Input, Textarea, icons, toast } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

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
			error = t('clientNew.nameRequired');
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
			toast.success(t('clientNew.created'));
			goto(`/clients/${c.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : t('clientNew.failed');
			toast.danger(error);
			submitting = false;
		}
	}
</script>

<svelte:head><title>{t('clientNew.title')}</title></svelte:head>

<div class="flex max-w-xl flex-col gap-10">
	<div class="flex flex-col gap-2">
		<Button href="/clients" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
			{t('clientNew.back')}
		</Button>
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('clients.new')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('clientNew.subtitle')}</p>
	</div>

	<form onsubmit={submit} class="flex flex-col gap-4">
		<Card class="flex flex-col gap-4">
			<Field label={t('common.name')}>
				<Input bind:value={name} placeholder={t('clientNew.namePlaceholder')} required />
			</Field>
			<Field label={t('common.email')}>
				<Input bind:value={email} type="email" placeholder={t('clientNew.emailPlaceholder')} />
			</Field>
			<Field label={t('common.company')}>
				<Input bind:value={company} placeholder={t('clientNew.companyPlaceholder')} />
			</Field>
			<Field label={t('common.phone')}>
				<Input bind:value={phone} type="tel" placeholder={t('clientNew.phonePlaceholder')} />
			</Field>
			<Field label={t('common.notes')} helper={t('clientNew.notesHelp')}>
				<Textarea bind:value={notes} rows={4} placeholder={t('clientNew.notesPlaceholder')} />
			</Field>
		</Card>

		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}

		<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
			<Button href="/clients" variant="ghost" class="w-full sm:w-auto">{t('common.cancel')}</Button>
			<Button type="submit" icon={icons.plus} disabled={submitting} class="w-full sm:w-auto">
				{submitting ? t('common.creating') : t('clientNew.submit')}
			</Button>
		</div>
	</form>
</div>
