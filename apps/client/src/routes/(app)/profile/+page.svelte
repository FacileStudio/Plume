<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { UserProfile } from '$lib';
	import Icon from '@iconify/svelte';
	import { userStore } from '$lib/stores/user.svelte';

	let profile = $state<UserProfile | null>(null);
	let name = $state('');
	let email = $state('');
	let profileError = $state('');
	let profileSuccess = $state('');
	let profileLoading = $state(false);

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordError = $state('');
	let passwordSuccess = $state('');
	let passwordLoading = $state(false);

	onMount(async () => {
		try {
			profile = await api.auth.me();
			name = profile.name ?? '';
			email = profile.email;
		} catch (e: any) {
			profileError = e.message;
		}
	});

	async function updateProfile() {
		profileError = '';
		profileSuccess = '';
		profileLoading = true;
		try {
			profile = await api.auth.updateProfile({ name, email });
			name = profile.name ?? '';
			email = profile.email;
			userStore.value = profile;
			profileSuccess = 'Profile updated.';
		} catch (e: any) {
			profileError = e.message;
		} finally {
			profileLoading = false;
		}
	}

	async function changePassword() {
		passwordError = '';
		passwordSuccess = '';

		if (newPassword !== confirmPassword) {
			passwordError = 'Passwords do not match.';
			return;
		}

		passwordLoading = true;
		try {
			await api.auth.changePassword(currentPassword, newPassword);
			passwordSuccess = 'Password changed.';
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		} catch (e: any) {
			passwordError = e.message;
		} finally {
			passwordLoading = false;
		}
	}
</script>

<svelte:head><title>Profile — Plume</title></svelte:head>

<h1 class="text-2xl font-bold mb-6">Profile</h1>

<div class="space-y-8 max-w-lg">
	<div class="rounded-lg border p-4">
		<h2 class="text-lg font-semibold mb-4">Profile</h2>

		{#if profile}
			<div class="flex items-center gap-3 mb-4">
				{#if profile.avatar_url}
					<img src="/api{profile.avatar_url}" alt="Avatar" class="h-14 w-14 rounded-full border border-border object-cover" />
				{:else}
					<div class="flex h-14 w-14 items-center justify-center rounded-full border border-border bg-foreground text-sm font-semibold text-background">
						{(profile.name || profile.email).split(' ').map((w: string) => w[0]).join('').toUpperCase().slice(0, 2)}
					</div>
				{/if}
				<div>
					<p class="text-sm font-medium">{profile.name || profile.email}</p>
					{#if profile.avatar_source === 'oidc'}
						<p class="text-xs text-muted-foreground">Synced from SSO</p>
					{/if}
				</div>
			</div>

			<form onsubmit={updateProfile} class="space-y-3">
				<div>
					<label for="name" class="block text-sm font-medium mb-1">Name</label>
					<input
						id="name"
						bind:value={name}
						placeholder="Your name"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					/>
				</div>
				<div>
					<label for="email" class="block text-sm font-medium mb-1">Email</label>
					<input
						id="email"
						type="email"
						bind:value={email}
						placeholder="you@example.com"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm"
						required
					/>
				</div>

				{#if profileError}
					<p class="text-destructive text-sm">{profileError}</p>
				{/if}
				{#if profileSuccess}
					<p class="text-sm text-green-600">{profileSuccess}</p>
				{/if}

				<button
					type="submit"
					disabled={profileLoading}
					class="flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50"
				>
					<Icon icon="solar:diskette-linear" class="h-4 w-4" />
					{profileLoading ? 'Saving...' : 'Save'}
				</button>
			</form>

			<p class="text-xs text-muted-foreground mt-4">
				Member since {new Date(profile.created_at).toLocaleDateString()}
			</p>
		{:else if !profileError}
			<p class="text-sm text-muted-foreground">Loading...</p>
		{/if}
	</div>

	<div class="rounded-lg border p-4">
		<h2 class="text-lg font-semibold mb-4">Change Password</h2>

		<form onsubmit={changePassword} class="space-y-3">
			<div>
				<label for="current-password" class="block text-sm font-medium mb-1">Current Password</label>
				<input
					id="current-password"
					type="password"
					bind:value={currentPassword}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					required
				/>
			</div>
			<div>
				<label for="new-password" class="block text-sm font-medium mb-1">New Password</label>
				<input
					id="new-password"
					type="password"
					bind:value={newPassword}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					required
					minlength="8"
				/>
			</div>
			<div>
				<label for="confirm-password" class="block text-sm font-medium mb-1">Confirm New Password</label>
				<input
					id="confirm-password"
					type="password"
					bind:value={confirmPassword}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm"
					required
					minlength="8"
				/>
			</div>

			{#if passwordError}
				<p class="text-destructive text-sm">{passwordError}</p>
			{/if}
			{#if passwordSuccess}
				<p class="text-sm text-green-600">{passwordSuccess}</p>
			{/if}

			<button
				type="submit"
				disabled={passwordLoading}
				class="flex items-center gap-1.5 rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground disabled:opacity-50"
			>
				<Icon icon="solar:lock-keyhole-linear" class="h-4 w-4" />
				{passwordLoading ? 'Changing...' : 'Change Password'}
			</button>
		</form>
	</div>
</div>
