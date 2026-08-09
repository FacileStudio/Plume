<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, Field, Input, SettingsRow, SettingsSection, StatusDot, icons, toast } from '@facile/muse';
	import { api } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';

	let reminderIntervalDays = $state(3);
	let reminderSaving = $state(false);

	const remindersOn = $derived(reminderIntervalDays > 0);

	onMount(loadReminderSettings);

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
			toast.danger('Profile not loaded');
			return;
		}
		if (
			reminderIntervalDays < 0 ||
			reminderIntervalDays > 30 ||
			Number.isNaN(reminderIntervalDays)
		) {
			toast.danger('Interval must be between 0 and 30 days');
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
			toast.danger(e instanceof Error ? e.message : 'Failed to save reminder settings');
		}
		reminderSaving = false;
	}
</script>

<svelte:head><title>Notifications — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Reminders"
		description="Plume re-sends the signing invitation to anyone who has not signed yet. Sending needs a working SMTP relay."
	>
		<SettingsRow
			label="Status"
			description="Reminders run in the background, one pass per interval."
		>
			<StatusDot
				tone={remindersOn ? 'success' : 'neutral'}
				label={remindersOn
					? `Every ${reminderIntervalDays} day${reminderIntervalDays === 1 ? '' : 's'}`
					: 'Off'}
			/>
		</SettingsRow>

		<SettingsRow
			stacked
			label="Reminder interval"
			description="In days, between 0 and 30. Zero disables automatic reminders entirely."
		>
			<Field helper="0 to disable automatic reminders">
				<Input
					type="number"
					min={0}
					max={30}
					bind:value={reminderIntervalDays}
					aria-label="Reminder interval in days"
				/>
			</Field>
		</SettingsRow>

		<div class="flex pt-1">
			<Button icon={icons.check} disabled={reminderSaving} onclick={saveReminderSettings}>
				{reminderSaving ? 'Saving…' : 'Save'}
			</Button>
		</div>
	</SettingsSection>
</div>
