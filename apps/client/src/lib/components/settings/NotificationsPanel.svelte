<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, Field, Input, NotificationsSection, SettingsRow, StatusDot, icons, toast } from '@facile/muse';
	import { api } from '$lib';
	import { userStore } from '$lib/stores/user.svelte';
	import { t } from '$lib/i18n/index.svelte';

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
			toast.danger(t('notifications.notLoaded'));
			return;
		}
		if (
			reminderIntervalDays < 0 ||
			reminderIntervalDays > 30 ||
			Number.isNaN(reminderIntervalDays)
		) {
			toast.danger(t('notifications.range'));
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
			toast.success(t('notifications.saved'));
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : t('notifications.saveFailed'));
		}
		reminderSaving = false;
	}
</script>

<svelte:head><title>{t('notifications.title')}</title></svelte:head>

<div class="flex flex-col gap-10">
	<NotificationsSection title={t('notifications.reminders')} description={t('notifications.remindersText')}>
		<SettingsRow label={t('common.status')} description={t('notifications.statusText')}>
			<StatusDot
				tone={remindersOn ? 'success' : 'neutral'}
				label={remindersOn
					? t('notifications.every', { count: reminderIntervalDays })
					: t('common.off')}
			/>
		</SettingsRow>

		<SettingsRow
			stacked
			label={t('notifications.interval')}
			description={t('notifications.intervalText')}
		>
			<Field helper={t('notifications.intervalHelp')}>
				<Input
					type="number"
					min={0}
					max={30}
					bind:value={reminderIntervalDays}
					aria-label={t('notifications.intervalAria')}
				/>
			</Field>
		</SettingsRow>

		<div class="flex pt-1">
			<Button icon={icons.check} disabled={reminderSaving} onclick={saveReminderSettings}>
				{reminderSaving ? t('common.saving') : t('common.save')}
			</Button>
		</div>
	</NotificationsSection>
</div>
