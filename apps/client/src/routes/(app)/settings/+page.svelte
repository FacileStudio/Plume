<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib';
	import type { UserProfile, Webhook } from '$lib';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';
	import { userStore } from '$lib/stores/user.svelte';

	let profile = $state<UserProfile | null>(null);
	let profileName = $state('');
	let profileEmail = $state('');
	let profileError = $state('');
	let profileSuccess = $state('');
	let profileLoading = $state(false);

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordError = $state('');
	let passwordSuccess = $state('');
	let passwordLoading = $state(false);

	let webhooks = $state<Webhook[]>([]);
	let showWebhookForm = $state(false);
	let webhookUrl = $state('');
	let webhookSecret = $state('');
	let editingWebhookId = $state<number | null>(null);
	let webhookSaving = $state(false);
	let deletingWebhookId = $state<number | null>(null);

	let smtpHost = $state('');
	let smtpPort = $state(587);
	let smtpUsername = $state('');
	let smtpPassword = $state('');
	let smtpFromEmail = $state('');
	let smtpFromName = $state('');
	let smtpConfigured = $state(false);
	let smtpSaving = $state(false);
	let smtpTesting = $state(false);
	let smtpDeleting = $state(false);

	let reminderIntervalDays = $state(3);
	let reminderSaving = $state(false);

	let activeTab = $state<'profile' | 'email' | 'reminders' | 'webhooks'>('profile');
	const settingsTabs = [
		{ id: 'profile', label: 'Profile', icon: 'solar:user-linear' },
		{ id: 'email', label: 'Email', icon: 'solar:letter-linear' },
		{ id: 'reminders', label: 'Reminders', icon: 'solar:bell-linear' },
		{ id: 'webhooks', label: 'Webhooks', icon: 'solar:bolt-linear' }
	] as const;

	async function loadProfile() {
		try {
			profile = await api.auth.me();
			profileName = profile.name ?? '';
			profileEmail = profile.email;
		} catch (e: any) {
			profileError = e.message;
		}
	}

	async function updateProfile() {
		profileError = '';
		profileSuccess = '';
		profileLoading = true;
		try {
			profile = await api.auth.updateProfile({ name: profileName, email: profileEmail });
			profileName = profile.name ?? '';
			profileEmail = profile.email;
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

	async function loadSmtp() {
		try {
			const config = await api.smtp.get();
			if (!config) {
				smtpConfigured = false;
				return;
			}
			smtpHost = config.host;
			smtpPort = config.port;
			smtpUsername = config.username;
			smtpPassword = '';
			smtpFromEmail = config.from_email;
			smtpFromName = config.from_name;
			smtpConfigured = true;
		} catch {
			smtpConfigured = false;
		}
	}

	async function saveSmtp() {
		smtpSaving = true;
		try {
			await api.smtp.save({
				host: smtpHost,
				port: smtpPort,
				username: smtpUsername,
				password: smtpPassword,
				from_email: smtpFromEmail,
				from_name: smtpFromName
			});
			smtpPassword = '';
			smtpConfigured = true;
			toast.success('SMTP configuration saved');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to save SMTP configuration');
		}
		smtpSaving = false;
	}

	async function deleteSmtp() {
		smtpDeleting = true;
		try {
			await api.smtp.delete();
			smtpHost = '';
			smtpPort = 587;
			smtpUsername = '';
			smtpPassword = '';
			smtpFromEmail = '';
			smtpFromName = '';
			smtpConfigured = false;
			toast.success('SMTP configuration removed');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete SMTP configuration');
		}
		smtpDeleting = false;
	}

	async function testSmtp() {
		const email = userStore.value?.email;
		if (!email) {
			toast.error('No email address found for current user');
			return;
		}
		smtpTesting = true;
		try {
			await api.smtp.test(email);
			toast.success(`Test email sent to ${email}`);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to send test email');
		}
		smtpTesting = false;
	}

	async function loadWebhooks() {
		try {
			webhooks = await api.webhooks.list();
		} catch {}
	}

	function resetWebhookForm() {
		webhookUrl = '';
		webhookSecret = '';
		editingWebhookId = null;
		showWebhookForm = false;
	}

	function startEditWebhook(wh: Webhook) {
		webhookUrl = wh.url;
		webhookSecret = '';
		editingWebhookId = wh.id;
		showWebhookForm = true;
	}

	async function saveWebhook() {
		webhookSaving = true;
		try {
			if (editingWebhookId) {
				const existing = webhooks.find((w) => w.id === editingWebhookId);
				await api.webhooks.update(editingWebhookId, {
					url: webhookUrl,
					secret: webhookSecret,
					enabled: existing?.enabled ?? true
				});
			} else {
				await api.webhooks.create({
					url: webhookUrl,
					secret: webhookSecret
				});
			}
			resetWebhookForm();
			await loadWebhooks();
		} catch {}
		webhookSaving = false;
	}

	async function toggleWebhookEnabled(wh: Webhook) {
		try {
			await api.webhooks.update(wh.id, {
				url: wh.url,
				secret: '',
				enabled: !wh.enabled
			});
			await loadWebhooks();
		} catch {}
	}

	async function deleteWebhook(id: number) {
		try {
			await api.webhooks.delete(id);
			deletingWebhookId = null;
			await loadWebhooks();
		} catch {}
	}

	let testingWebhookId = $state<number | null>(null);
	async function testWebhook(id: number) {
		if (testingWebhookId !== null) return;
		testingWebhookId = id;
		try {
			await api.webhooks.test(id);
			toast.success('Test event delivered');
			await loadWebhooks();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Test delivery failed');
		}
		testingWebhookId = null;
	}

	const webhookEventTypes = [
		'document.created',
		'document.sent',
		'document.completed',
		'document.declined',
		'document.deleted',
		'signer.added',
		'signer.email_opened',
		'signer.viewed',
		'signer.signed',
		'signer.declined',
		'signer.reminded'
	];

	async function loadReminderSettings() {
		try {
			const me = await api.auth.me();
			reminderIntervalDays = me.reminder_interval_days ?? 3;
			userStore.value = me;
		} catch {}
	}

	async function saveReminderSettings() {
		const current = userStore.value;
		if (!current) {
			toast.error('Profile not loaded');
			return;
		}
		if (reminderIntervalDays < 0 || reminderIntervalDays > 30 || Number.isNaN(reminderIntervalDays)) {
			toast.error('Interval must be between 0 and 30 days');
			return;
		}
		reminderSaving = true;
		try {
			const updated = await api.auth.updateProfile({
				name: current.name,
				email: current.email,
				reminder_interval_days: reminderIntervalDays
			});
			userStore.value = updated;
			toast.success('Reminder settings saved');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to save reminder settings');
		}
		reminderSaving = false;
	}

	onMount(async () => {
		await Promise.all([loadProfile(), loadSmtp(), loadWebhooks(), loadReminderSettings()]);
	});
</script>

<svelte:head><title>Settings — Plume</title></svelte:head>

<div class="mb-6 border-b pb-5">
	<h1 class="text-2xl font-bold">Settings</h1>
	<p class="mt-1 text-sm text-muted-foreground">Manage your profile, email delivery, reminders, and webhooks.</p>
</div>

<div class="mb-6 flex gap-6 border-b">
	{#each settingsTabs as t}
		<button
			onclick={() => (activeTab = t.id)}
			class="relative -mb-px flex items-center gap-2 border-b-2 px-1 pb-3 text-sm font-medium transition-colors {activeTab === t.id
				? 'border-foreground text-foreground'
				: 'border-transparent text-muted-foreground hover:text-foreground'}"
		>
			<Icon icon={t.icon} class="h-4 w-4" />
			{t.label}
		</button>
	{/each}
</div>

<div class="max-w-2xl">
	{#if activeTab === 'profile'}
	<div class="space-y-6">
		<div class="rounded-lg border p-6">
			<div class="flex items-center gap-2 mb-4">
				<Icon icon="solar:user-linear" class="h-5 w-5" />
				<h2 class="text-lg font-semibold">Profile</h2>
			</div>

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
							bind:value={profileName}
							placeholder="Your name"
							class="w-full rounded-md border bg-background px-3 py-2 text-sm"
						/>
					</div>
					<div>
						<label for="email" class="block text-sm font-medium mb-1">Email</label>
						<input
							id="email"
							type="email"
							bind:value={profileEmail}
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
						class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
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
			{:else}
				<p class="text-destructive text-sm">{profileError}</p>
			{/if}
		</div>

		<div class="rounded-lg border p-6">
			<div class="flex items-center gap-2 mb-4">
				<Icon icon="solar:lock-keyhole-linear" class="h-5 w-5" />
				<h2 class="text-lg font-semibold">Change password</h2>
			</div>

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
					class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					<Icon icon="solar:lock-keyhole-linear" class="h-4 w-4" />
					{passwordLoading ? 'Changing...' : 'Change Password'}
				</button>
			</form>
		</div>
	</div>

	{/if}

	{#if activeTab === 'email'}
	<div class="rounded-lg border p-6">
		<div class="flex items-center gap-2 mb-1">
			<Icon icon="solar:letter-linear" class="h-5 w-5" />
			<h2 class="text-lg font-semibold">Email (SMTP)</h2>
		</div>
		<p class="text-sm text-muted-foreground mb-4">Configure SMTP to send signing invitations by email</p>

		<div class="space-y-4">
			<div>
				<label for="smtp-host" class="block text-sm font-medium mb-1">Host</label>
				<input
					id="smtp-host"
					type="text"
					bind:value={smtpHost}
					placeholder="smtp.example.com"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="smtp-port" class="block text-sm font-medium mb-1">Port</label>
				<input
					id="smtp-port"
					type="number"
					bind:value={smtpPort}
					placeholder="587"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="smtp-username" class="block text-sm font-medium mb-1">Username</label>
				<input
					id="smtp-username"
					type="text"
					bind:value={smtpUsername}
					placeholder="user@example.com"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="smtp-password" class="block text-sm font-medium mb-1">Password</label>
				<input
					id="smtp-password"
					type="password"
					bind:value={smtpPassword}
					placeholder={smtpConfigured ? '••••••••' : ''}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="smtp-from-email" class="block text-sm font-medium mb-1">From Email</label>
				<input
					id="smtp-from-email"
					type="email"
					bind:value={smtpFromEmail}
					placeholder="noreply@example.com"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div>
				<label for="smtp-from-name" class="block text-sm font-medium mb-1">From Name</label>
				<input
					id="smtp-from-name"
					type="text"
					bind:value={smtpFromName}
					placeholder="Plume"
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
			</div>
			<div class="flex gap-2 pt-1">
				<button
					onclick={saveSmtp}
					disabled={smtpSaving || !smtpHost}
					class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					<Icon icon="solar:diskette-linear" class="h-4 w-4" />
					{smtpSaving ? 'Saving…' : 'Save'}
				</button>
				{#if smtpConfigured}
					<button
						onclick={testSmtp}
						disabled={smtpTesting}
						class="flex items-center gap-1.5 rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<Icon icon="solar:test-tube-linear" class="h-4 w-4" />
						{smtpTesting ? 'Sending…' : 'Test'}
					</button>
					<button
						onclick={deleteSmtp}
						disabled={smtpDeleting}
						class="flex items-center gap-1.5 rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-red-500 transition-colors hover:bg-red-500 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<Icon icon="solar:trash-bin-trash-linear" class="h-4 w-4" />
						{smtpDeleting ? 'Deleting…' : 'Delete'}
					</button>
				{/if}
			</div>
		</div>
	</div>

	{/if}

	{#if activeTab === 'reminders'}
	<div class="rounded-lg border p-6">
		<div class="flex items-center gap-2 mb-1">
			<Icon icon="solar:bell-linear" class="h-5 w-5" />
			<h2 class="text-lg font-semibold">Reminders</h2>
		</div>
		<p class="text-sm text-muted-foreground mb-4">Automatically re-send signing invitations to pending signers</p>

		<div class="space-y-4">
			<div>
				<label for="reminder-interval" class="block text-sm font-medium mb-1">Reminder interval (days)</label>
				<input
					id="reminder-interval"
					type="number"
					min="0"
					max="30"
					bind:value={reminderIntervalDays}
					class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
				/>
				<p class="mt-1.5 text-xs text-muted-foreground">0 to disable automatic reminders</p>
			</div>
			<div class="flex gap-2 pt-1">
				<button
					onclick={saveReminderSettings}
					disabled={reminderSaving}
					class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
				>
					<Icon icon="solar:diskette-linear" class="h-4 w-4" />
					{reminderSaving ? 'Saving…' : 'Save'}
				</button>
			</div>
		</div>
	</div>

	{/if}

	{#if activeTab === 'webhooks'}
	<div class="rounded-lg border p-6">
		<div class="flex items-center justify-between mb-1">
			<h2 class="text-lg font-semibold">Webhooks</h2>
			{#if !showWebhookForm && webhooks.length > 0}
				<button
					onclick={() => { resetWebhookForm(); showWebhookForm = true; }}
					class="flex items-center gap-1.5 rounded-full bg-foreground px-3 py-1 text-sm font-medium text-background transition-colors hover:bg-foreground/90"
				>
					<Icon icon="mdi:plus" class="h-4 w-4" />
					Add
				</button>
			{/if}
		</div>
		<p class="text-sm text-muted-foreground mb-2">Receive notifications on every document and signer event</p>
		<details class="mb-4 text-sm">
			<summary class="cursor-pointer text-xs text-muted-foreground hover:text-foreground">Supported event types</summary>
			<div class="mt-2 flex flex-wrap gap-1.5">
				{#each webhookEventTypes as eventType}
					<code class="rounded bg-muted px-1.5 py-0.5 text-xs">{eventType}</code>
				{/each}
			</div>
			<p class="mt-2 text-xs text-muted-foreground">Each delivery is HMAC-SHA256 signed via the <code class="rounded bg-muted px-1 py-0.5 text-xs">x-plume-signature-256</code> header. Failed deliveries retry up to 3 times with backoff.</p>
		</details>

		{#if webhooks.length === 0 && !showWebhookForm}
			<button
				onclick={() => (showWebhookForm = true)}
				class="flex items-center gap-2 rounded-lg border border-dashed px-4 py-3 text-sm text-muted-foreground transition-colors hover:border-foreground/30 hover:text-foreground w-full justify-center"
			>
				<Icon icon="mdi:plus" class="h-4 w-4" />
				Add Webhook
			</button>
		{/if}

		{#if webhooks.length > 0}
			<div class="space-y-3 mb-4">
				{#each webhooks as wh}
					<div class="rounded-lg border p-4">
						{#if deletingWebhookId === wh.id}
							<div class="flex items-center justify-between">
								<p class="text-sm">Delete this webhook?</p>
								<div class="flex gap-2">
									<button
										onclick={() => deleteWebhook(wh.id)}
										class="rounded-full bg-red-500 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-red-600"
									>
										Delete
									</button>
									<button
										onclick={() => (deletingWebhookId = null)}
										class="rounded-full bg-muted px-3 py-1 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
									>
										Cancel
									</button>
								</div>
							</div>
						{:else}
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-medium">{wh.url}</p>
									<div class="mt-1.5 flex flex-wrap items-center gap-2">
										<span class="rounded-full px-2.5 py-0.5 text-xs font-medium {wh.enabled ? 'bg-foreground text-background' : 'bg-muted text-muted-foreground'}">
											{wh.enabled ? 'Active' : 'Disabled'}
										</span>
										{#if wh.last_sent_at}
											<span class="text-xs text-muted-foreground">Last delivered {new Date(wh.last_sent_at).toLocaleString()}</span>
										{/if}
									</div>
								</div>
								<div class="flex items-center gap-1.5 shrink-0">
									<button
										onclick={() => testWebhook(wh.id)}
										disabled={testingWebhookId === wh.id}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
										aria-label="Send test event"
										title="Send test event"
									>
										<Icon icon={testingWebhookId === wh.id ? 'solar:spinner-linear' : 'solar:test-tube-linear'} class="h-3.5 w-3.5 {testingWebhookId === wh.id ? 'animate-spin' : ''}" />
									</button>
									<button
										onclick={() => toggleWebhookEnabled(wh)}
										class="relative h-6 w-10 rounded-full transition-colors {wh.enabled ? 'bg-green-500' : 'bg-muted'}"
										aria-label="{wh.enabled ? 'Disable' : 'Enable'} webhook"
									>
										<span class="absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform {wh.enabled ? 'left-[18px]' : 'left-0.5'}"></span>
									</button>
									<button
										onclick={() => startEditWebhook(wh)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-foreground hover:bg-muted"
										aria-label="Edit webhook"
									>
										<Icon icon="solar:pen-linear" class="h-3.5 w-3.5" />
									</button>
									<button
										onclick={() => (deletingWebhookId = wh.id)}
										class="rounded-md p-1.5 text-muted-foreground transition-colors hover:text-red-500 hover:bg-muted"
										aria-label="Delete webhook"
									>
										<Icon icon="solar:trash-bin-trash-linear" class="h-3.5 w-3.5" />
									</button>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}

		{#if showWebhookForm}
			<div class="rounded-lg border p-4 space-y-4">
				<div>
					<label for="webhook-url" class="block text-sm font-medium mb-1">URL</label>
					<input
						id="webhook-url"
						type="text"
						bind:value={webhookUrl}
						placeholder="https://nook.example.com/webhook/plume"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div>
					<label for="webhook-secret" class="block text-sm font-medium mb-1">Secret</label>
					<input
						id="webhook-secret"
						type="password"
						bind:value={webhookSecret}
						placeholder="Shared HMAC secret"
						class="w-full rounded-md border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div class="flex gap-2 pt-1">
					<button
						onclick={saveWebhook}
						disabled={webhookSaving || !webhookUrl}
						class="flex items-center gap-1.5 rounded-full bg-foreground px-4 py-1.5 text-sm font-medium text-background transition-colors hover:bg-foreground/90 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<Icon icon="solar:diskette-linear" class="h-4 w-4" />
						{webhookSaving ? 'Saving…' : editingWebhookId ? 'Update' : 'Save'}
					</button>
					<button
						onclick={resetWebhookForm}
						class="flex items-center gap-1.5 rounded-full bg-muted px-4 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
					>
						<Icon icon="solar:close-circle-linear" class="h-4 w-4" />
						Cancel
					</button>
				</div>
			</div>
		{/if}
	</div>
	{/if}
</div>
