<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Button,
		Field,
		Input,
		ProfileCard,
		SettingsRow,
		SettingsSection,
		icons,
		toast
	} from '@facile/muse';
	import { api, logout as logoutSession } from '$lib';
	import type { UserProfile } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { setSessionLocale, t } from '$lib/i18n/index.svelte';

	let profile = $state<UserProfile | null>(null);
	let profileName = $state('');
	let profileEmail = $state('');
	let profileError = $state('');
	let profileLoading = $state(false);

	const memberSince = $derived(
		profile?.created_at ? formatDate(profile.created_at) : '—'
	);

	const avatarNote = $derived(
		profile?.avatar_source === 'oidc' ? t('profile.avatarSso') : t('profile.avatarNone')
	);

	onMount(loadProfile);

	async function loadProfile() {
		try {
			profile = await api.auth.me();
			profileName = profile.name ?? '';
			profileEmail = profile.email;
			userStore.value = profile;
		} catch (e) {
			profileError = e instanceof Error ? e.message : t('profile.loadFailed');
		}
	}

	async function updateProfile(event: SubmitEvent) {
		event.preventDefault();
		profileError = '';
		profileLoading = true;
		try {
			profile = await api.auth.updateProfile({ name: profileName, email: profileEmail });
			profileName = profile.name ?? '';
			profileEmail = profile.email;
			userStore.value = profile;
			toast.success(t('profile.updated'));
		} catch (e) {
			profileError = e instanceof Error ? e.message : t('profile.updateFailed');
			toast.danger(profileError);
		}
		profileLoading = false;
	}

	async function logout() {
		await logoutSession();
		userStore.value = null;
		setSessionLocale(null);
		goto('/login');
	}
</script>

<svelte:head><title>{t('profile.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<ProfileCard
		name={profile?.name?.trim() || profile?.email || t('profile.account')}
		email={profile?.email}
		avatar={profile?.avatar_url || undefined}
		meta={[
			{ label: t('profile.memberSince'), value: memberSince },
			{
				label: t('profile.photo'),
				value: profile?.avatar_source === 'oidc' ? t('profile.photoSso') : t('profile.photoNone')
			}
		]}
	/>

	<SettingsSection title={t('profile.identity')} description={avatarNote}>
		<form onsubmit={updateProfile} class="flex flex-col gap-4">
			<SettingsRow stacked label={t('common.name')} description={t('profile.nameText')}>
				<Field>
					<Input
						bind:value={profileName}
						placeholder={t('profile.namePlaceholder')}
						aria-label={t('common.name')}
					/>
				</Field>
			</SettingsRow>

			<SettingsRow stacked label={t('common.email')} description={t('profile.emailText')}>
				<Field error={profileError || undefined}>
					<Input
						bind:value={profileEmail}
						type="email"
						placeholder={t('profile.emailPlaceholder')}
						aria-label={t('common.email')}
						required
					/>
				</Field>
			</SettingsRow>

			<div class="flex pt-1">
				<Button type="submit" icon={icons.check} disabled={profileLoading}>
					{profileLoading ? t('common.saving') : t('common.save')}
				</Button>
			</div>
		</form>
	</SettingsSection>

	<SettingsSection title={t('profile.session')} description={t('profile.sessionText')}>
		<SettingsRow label={t('profile.logout')} description={t('profile.logoutText')}>
			<Button variant="outline" icon={icons.logout} onclick={logout}>{t('profile.logout')}</Button>
		</SettingsRow>
	</SettingsSection>
</div>
