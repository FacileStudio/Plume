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

	let profile = $state<UserProfile | null>(userStore.value);
	let confirmClear = $state(false);

	const memberSince = $derived(
		profile?.created_at ? new Date(profile.created_at).toLocaleString() : '—'
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
			toast.success('Profile re-read from single sign-on');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Failed to sync profile');
		}
	}

	function clearLocalData() {
		for (const key of Object.keys(localStorage)) {
			if (key.startsWith('plume.')) localStorage.removeItem(key);
		}
		clearToken();
		userStore.value = null;
		toast.success('Local data cleared.');
		goto('/login');
	}
</script>

<svelte:head><title>Advanced — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection title="Instance" description="What this browser is talking to.">
		<SettingsRow stacked>
			<SecretField label="Instance URL" value={page.url.origin} sensitive={false} class="w-full" />
		</SettingsRow>

		<SettingsRow label="API base path" description="Every request is scoped under it.">
			<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection title="Account" description="Facts an operator may ask you for.">
		<SettingsRow stacked>
			<SecretField label="User ID" value={profile?.id ?? ''} sensitive={false} class="w-full" />
		</SettingsRow>

		<SettingsRow label="Avatar source" description="Where your profile picture comes from.">
			<span class="font-fc-mono text-fc-sm text-fc-fg">{profile?.avatar_source || 'local'}</span>
		</SettingsRow>

		<SettingsRow label="Member since" description="When this account was created.">
			<span class="text-fc-sm text-fc-fg-muted">{memberSince}</span>
		</SettingsRow>

		<SettingsRow
			label="Re-read single sign-on"
			description="Pulls your name and photo from the identity provider again instead of waiting for the next automatic sync."
		>
			<Button variant="outline" icon={icons.refresh} onclick={syncProfile}>Sync now</Button>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="Danger zone"
		description="Irreversible from this browser. Your account, documents and signatures are untouched."
	>
		<SettingsRow
			label="Clear local data"
			description="Removes the stored session and the theme preference, then signs you out. Plume has no self-service account deletion — an administrator has to remove the account server-side."
		>
			<Button variant="danger" icon={icons.remove} onclick={() => (confirmClear = true)}>
				Clear local data
			</Button>
		</SettingsRow>
	</SettingsSection>
</div>

<ConfirmModal
	bind:open={confirmClear}
	tone="danger"
	title="Clear local data?"
	description="Every Plume preference stored in this browser is deleted and you are signed out — there is no undo. Your account, documents and signatures stay on the server."
	confirmLabel="Clear and sign out"
	onConfirm={clearLocalData}
/>
