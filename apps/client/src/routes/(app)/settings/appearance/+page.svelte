<script lang="ts">
	import { OptionCards, SettingsRow, SettingsSection, icons } from '@facile/muse';
	import { theme, type ThemePreference } from '$lib/theme.svelte';

	const options = [
		{ value: 'system', label: 'System', icon: icons.monitor },
		{ value: 'light', label: 'Light', icon: icons.sun },
		{ value: 'dark', label: 'Dark', icon: icons.moon }
	];
</script>

<svelte:head><title>Appearance — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Theme"
		description="Applied to this browser. Every device you sign in from keeps its own choice."
	>
		<SettingsRow
			stacked
			label="Colour scheme"
			description="System follows your operating system setting and changes with it."
		>
			<OptionCards
				label="Colour scheme"
				name="theme-preference"
				{options}
				value={theme.preference}
				onSelect={(next) => theme.set(next as ThemePreference)}
			/>
		</SettingsRow>
	</SettingsSection>
</div>
