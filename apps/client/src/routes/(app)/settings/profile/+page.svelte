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

	let profile = $state<UserProfile | null>(null);
	let profileName = $state('');
	let profileEmail = $state('');
	let profileError = $state('');
	let profileLoading = $state(false);

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordError = $state('');
	let passwordLoading = $state(false);

	const memberSince = $derived(
		profile?.created_at ? new Date(profile.created_at).toLocaleDateString() : '—'
	);

	const avatarNote = $derived(
		profile?.avatar_source === 'oidc'
			? 'Your photo comes from single sign-on. Change it there and it updates here within a few minutes.'
			: 'No photo in single sign-on yet — add one there and it appears here.'
	);

	onMount(loadProfile);

	async function loadProfile() {
		try {
			profile = await api.auth.me();
			profileName = profile.name ?? '';
			profileEmail = profile.email;
			userStore.value = profile;
		} catch (e) {
			profileError = e instanceof Error ? e.message : 'Failed to load profile';
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
			toast.success('Profile updated.');
		} catch (e) {
			profileError = e instanceof Error ? e.message : 'Failed to update profile';
			toast.danger(profileError);
		}
		profileLoading = false;
	}

	async function changePassword(event: SubmitEvent) {
		event.preventDefault();
		passwordError = '';
		if (newPassword !== confirmPassword) {
			passwordError = 'Passwords do not match.';
			return;
		}
		passwordLoading = true;
		try {
			await api.auth.changePassword(currentPassword, newPassword);
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			toast.success('Password changed.');
		} catch (e) {
			passwordError = e instanceof Error ? e.message : 'Failed to change password';
			toast.danger(passwordError);
		}
		passwordLoading = false;
	}

	async function logout() {
		await logoutSession();
		userStore.value = null;
		goto('/login');
	}
</script>

<svelte:head><title>Profile — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<ProfileCard
		name={profile?.name?.trim() || profile?.email || 'Account'}
		email={profile?.email}
		avatar={profile?.avatar_url || undefined}
		meta={[
			{ label: 'Member since', value: memberSince },
			{ label: 'Photo', value: profile?.avatar_source === 'oidc' ? 'Single sign-on' : 'None yet' }
		]}
	/>

	<SettingsSection title="Identity" description={avatarNote}>
		<form onsubmit={updateProfile} class="flex flex-col gap-4">
			<SettingsRow
				stacked
				label="Name"
				description="Shown on the documents you send and in the signing emails."
			>
				<Field>
					<Input bind:value={profileName} placeholder="Your name" aria-label="Name" />
				</Field>
			</SettingsRow>

			<SettingsRow
				stacked
				label="Email"
				description="Where signing notifications and test emails are delivered."
			>
				<Field error={profileError || undefined}>
					<Input
						bind:value={profileEmail}
						type="email"
						placeholder="you@example.com"
						aria-label="Email"
						required
					/>
				</Field>
			</SettingsRow>

			<div class="flex pt-1">
				<Button type="submit" icon={icons.check} disabled={profileLoading}>
					{profileLoading ? 'Saving…' : 'Save'}
				</Button>
			</div>
		</form>
	</SettingsSection>

	<SettingsSection
		title="Password"
		description="Plume never shows a password back to you, so changing one means proving you know the current one."
	>
		<form onsubmit={changePassword} class="flex flex-col gap-4">
			<SettingsRow stacked label="Current password">
				<Field>
					<Input
						bind:value={currentPassword}
						type="password"
						autocomplete="current-password"
						aria-label="Current password"
						required
					/>
				</Field>
			</SettingsRow>

			<SettingsRow stacked label="New password" description="At least twelve characters.">
				<Field>
					<Input
						bind:value={newPassword}
						type="password"
						autocomplete="new-password"
						aria-label="New password"
						minlength={8}
						required
					/>
				</Field>
			</SettingsRow>

			<SettingsRow stacked label="Confirm new password">
				<Field error={passwordError || undefined}>
					<Input
						bind:value={confirmPassword}
						type="password"
						autocomplete="new-password"
						aria-label="Confirm new password"
						minlength={8}
						required
					/>
				</Field>
			</SettingsRow>

			<div class="flex pt-1">
				<Button type="submit" icon={icons.key} disabled={passwordLoading}>
					{passwordLoading ? 'Changing…' : 'Change password'}
				</Button>
			</div>
		</form>
	</SettingsSection>

	<SettingsSection
		title="Session"
		description="Sessions are per browser. Signing out here leaves your other devices alone."
	>
		<SettingsRow
			label="Log out"
			description="Ends this session and returns you to the sign-in page. Documents already sent keep collecting signatures."
		>
			<Button variant="outline" icon={icons.logout} onclick={logout}>Log out</Button>
		</SettingsRow>
	</SettingsSection>
</div>
