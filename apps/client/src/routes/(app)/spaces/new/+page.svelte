<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib';
	import { Button, Card, Field, Input, Textarea, icons, toast } from '@facile/muse';
	import { spaceStore } from '$lib/stores/space.svelte';
	import { t } from '$lib/i18n/index.svelte';

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
			toast.success(t('spaceNew.created'));
			goto(`/spaces/${space.id}`);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('spaceNew.failed'));
		}
		saving = false;
	}
</script>

<svelte:head><title>{t('spaceNew.title')}</title></svelte:head>

<div class="flex max-w-xl flex-col gap-10">
	<div class="flex flex-col gap-2">
		<Button href="/spaces" variant="ghost" size="sm" icon={icons.chevronLeft} class="w-fit -ml-3">
			{t('spaceNew.back')}
		</Button>
		<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('spaces.new')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('spaceNew.subtitle')}</p>
	</div>

	<form onsubmit={create} class="flex flex-col gap-4">
		<Card class="flex flex-col gap-4">
			<Field label={t('common.name')}>
				<Input bind:value={name} placeholder={t('spaceNew.namePlaceholder')} required />
			</Field>
			<Field label={t('common.description')} helper={t('spaceNew.descriptionHelp')}>
				<Textarea bind:value={description} rows={3} placeholder={t('spaceNew.descriptionPlaceholder')} />
			</Field>
		</Card>

		<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
			<Button href="/spaces" variant="ghost" class="w-full sm:w-auto">{t('common.cancel')}</Button>
			<Button
				type="submit"
				icon={icons.plus}
				disabled={saving || !name.trim()}
				class="w-full sm:w-auto"
			>
				{saving ? t('common.creating') : t('spaceNew.submit')}
			</Button>
		</div>
	</form>
</div>
