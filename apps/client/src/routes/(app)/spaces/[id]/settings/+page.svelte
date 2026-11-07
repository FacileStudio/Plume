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
	import { t } from '$lib/i18n/index.svelte';

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
			toast.success(t('spaceSettings.updated'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('spaceSettings.updateFailed'));
		}
		saving = false;
	}

	async function runDelete() {
		try {
			await api.spaces.delete(spaceId);
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('spaceSettings.deleteFailed'));
			throw e;
		}
		spaceStore.spaces = spaceStore.spaces.filter((s) => s.id !== spaceId);
		if (spaceStore.activeId === spaceId) {
			spaceStore.activeId = null;
		}
		toast.success(t('spaceSettings.deleted'));
		goto('/spaces');
	}

	onMount(async () => {
		try {
			space = await api.spaces.get(spaceId);
			if (space.registre_url || (space.role !== 'owner' && space.role !== 'admin')) {
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

<svelte:head><title>{t('spaceSettings.title', { space: space?.name ?? t('space.titleFallback') })}</title></svelte:head>

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
			<h1 class="text-fc-2xl font-semibold text-fc-fg">{t('spaceSettings.heading')}</h1>
		</div>

		<form onsubmit={save} class="flex flex-col gap-4">
			<SettingsSection title={t('spaceSettings.general')} description={t('spaceSettings.generalText')}>
				<Field label={t('common.name')}>
					<Input bind:value={name} placeholder={t('spaceNew.namePlaceholder')} />
				</Field>
				<Field label={t('common.description')} helper={t('spaceNew.descriptionHelp')}>
					<Textarea bind:value={description} rows={3} placeholder={t('spaceNew.descriptionPlaceholder')} />
				</Field>
			</SettingsSection>
			<div class="flex justify-end">
				<Button type="submit" icon={icons.check} disabled={saving || !name.trim()}>
					{saving ? t('common.saving') : t('spaceSettings.save')}
				</Button>
			</div>
		</form>

		{#if space.role === 'owner'}
			<SettingsSection title={t('spaceSettings.danger')} description={t('spaceSettings.dangerText')}>
				<SettingsRow label={t('spaceSettings.deleteLabel')} description={t('spaceSettings.deleteText')}>
					<Button
						variant="ghost-danger"
						icon={icons.remove}
						onclick={() => (confirmDelete = true)}
					>
						{t('spaceSettings.deleteButton')}
					</Button>
				</SettingsRow>
			</SettingsSection>
		{/if}
	</div>

	<ConfirmModal
		bind:open={confirmDelete}
		tone="danger"
		title={t('spaceSettings.confirm.title', { name: space.name })}
		description={t('spaceSettings.confirm.text')}
		confirmLabel={t('spaceSettings.deleteButton')}
		cancelLabel={t('spaceSettings.confirm.cancel')}
		onConfirm={runDelete}
	/>
{/if}
