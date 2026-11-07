<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		Button,
		ConfirmModal,
		SecretField,
		SettingsRow,
		SettingsSection,
		icons,
		toast
	} from '@facile/muse';
	import { api, clearToken } from '$lib';
	import type { UserProfile } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { formatDateTime } from '$lib/i18n/format';
	import { setSessionLocale, t } from '$lib/i18n/index.svelte';

	let profile = $state<UserProfile | null>(userStore.value);
	let confirmClear = $state(false);

	const memberSince = $derived(
		profile?.created_at ? formatDateTime(profile.created_at) : '—'
	);

	onMount(async () => {
		try {
			profile = await api.auth.me();
			userStore.value = profile;
		} catch {}
	});

	async function syncProfile() {
		try {
			await api.auth.syncProfile();
			profile = await api.auth.me();
			userStore.value = profile;
			toast.success(t('advanced.synced'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('advanced.syncFailed'));
		}
	}

	function clearLocalData() {
		for (const key of Object.keys(localStorage)) {
			if (key.startsWith('plume.')) localStorage.removeItem(key);
		}
		clearToken();
		userStore.value = null;
		setSessionLocale(null);
		toast.success(t('advanced.cleared'));
		goto('/login');
	}
</script>

<svelte:head><title>{t('advanced.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title={t('advanced.instance')} description={t('advanced.instanceText')}>
		<SettingsRow stacked>
			<SecretField label={t('advanced.instanceUrl')} value={page.url.origin} sensitive={false} class="w-full" />
		</SettingsRow>

		<SettingsRow label={t('api.basePath')} description={t('advanced.basePathText')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('advanced.account')} description={t('advanced.accountText')}>
		<SettingsRow stacked>
			<SecretField label={t('advanced.userId')} value={profile?.id ?? ''} sensitive={false} class="w-full" />
		</SettingsRow>

		<SettingsRow label={t('advanced.avatarSource')} description={t('advanced.avatarSourceText')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">{profile?.avatar_source || t('advanced.avatarLocal')}</span>
		</SettingsRow>

		<SettingsRow label={t('profile.memberSince')} description={t('advanced.memberSinceText')}>
			<span class="text-fc-sm text-fc-fg-muted">{memberSince}</span>
		</SettingsRow>

		<SettingsRow label={t('advanced.resync')} description={t('advanced.resyncText')}>
			<Button variant="outline" icon={icons.refresh} onclick={syncProfile}>
				{t('advanced.syncNow')}
			</Button>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title={t('advanced.danger')} description={t('advanced.dangerText')}>
		<SettingsRow label={t('advanced.clear')} description={t('advanced.clearText')}>
			<Button variant="danger" icon={icons.remove} onclick={() => (confirmClear = true)}>
				{t('advanced.clear')}
			</Button>
		</SettingsRow>
	</SettingsSection>
</div>

<ConfirmModal
	bind:open={confirmClear}
	tone="danger"
	title={t('advanced.confirm.title')}
	description={t('advanced.confirm.text')}
	confirmLabel={t('advanced.confirm.label')}
	onConfirm={clearLocalData}
/>
