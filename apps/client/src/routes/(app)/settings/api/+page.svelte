<script lang="ts">
	import { Button, SettingsRow, SettingsSection, StatusDot, icons } from '@facile/muse';
</script>

<svelte:head><title>API — Plume</title></svelte:head>

<div class="flex flex-col gap-10">
	<SettingsSection
		title="Reference"
		description="Every route this instance serves, generated from the handlers themselves rather than kept by hand."
	>
		{#snippet actions()}
			<Button href="/docs" target="_blank" rel="noopener" icon={icons.code}>Open the reference</Button>
		{/snippet}

		<SettingsRow
			label="Base path"
			description="Every endpoint is scoped under it, on this same origin."
		>
			<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
		</SettingsRow>

		<SettingsRow
			label="OpenAPI document"
			description="The machine-readable schema behind the reference — feed it to a client generator."
		>
			<Button variant="outline" href="/docs/openapi.json" target="_blank" rel="noopener" icon={icons.download}>
				openapi.json
			</Button>
		</SettingsRow>
	</SettingsSection>

	<SettingsSection
		title="Authentication"
		description="How a request proves who it is. This is worth reading before you automate anything against Plume."
	>
		<SettingsRow
			label="Scheme"
			description="Send the session token as Authorization: Bearer <token>. There is no other accepted form."
		>
			<span class="font-fc-mono text-fc-sm text-fc-fg">Bearer</span>
		</SettingsRow>

		<SettingsRow
			label="API keys"
			description="Plume issues none. The only credential it accepts is the session token minted when you sign in, which expires 30 days later and dies with a sign-out."
		>
			<StatusDot tone="neutral" label="Not available" />
		</SettingsRow>

		<SettingsRow
			label="Automating against Plume"
			description="A session token is your browser login, not a machine credential — it carries everything you can do and cannot be scoped or revoked on its own. Have the script sign in through /api/auth/login with its own account instead of lifting a token out of this browser."
		>
			<StatusDot tone="warning" label="Use a dedicated account" />
		</SettingsRow>
	</SettingsSection>
</div>
